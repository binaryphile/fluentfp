// Package main demonstrates fluentfp composition across 6 packages
// in a curl-testable order processing service.
//
// Run:
//
//	go run ./examples/orders/
//
// Then:
//
//	curl -s -X POST http://localhost:3000/orders \
//	  -H 'Content-Type: application/json' \
//	  -d '{"customer":"Alice","items":[{"sku":"WIDGET-1","quantity":3}]}'
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/binaryphile/fluentfp/ctxval"
	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/slice"
	"github.com/binaryphile/fluentfp/web"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// RequestID is a request-scoped correlation ID stored via ctxval.
type RequestID string

// LineItem is a single SKU + quantity in an order.
type LineItem struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// HasPositiveQty returns true if the line item has quantity > 0.
func (li LineItem) HasPositiveQty() bool { return li.Quantity > 0 }

// Order is the core domain object. TotalCents uses integer cents
// to avoid floating-point currency errors.
type Order struct {
	ID         string     `json:"id"`
	Customer   string     `json:"customer"`
	Items      []LineItem `json:"items"`
	Status     OrderStatus `json:"status"`
	TotalCents int        `json:"total_cents"`
}

// OrderStatus tracks the lifecycle of an order.
type OrderStatus string

const (
	statusPending OrderStatus = "pending"
	statusPriced  OrderStatus = "priced"
)

// orderNum extracts the numeric suffix from an order ID for sorting.
func orderNum(o Order) int {
	s := strings.TrimPrefix(o.ID, "ord-")
	n, _ := strconv.Atoi(s)
	return n
}

// ---------------------------------------------------------------------------
// Store — single source of truth
// ---------------------------------------------------------------------------

type store struct {
	mu   sync.RWMutex
	byID map[string]Order
}

func newStore() *store {
	return &store{byID: make(map[string]Order)}
}

func (s *store) put(o Order) {
	s.mu.Lock()
	s.byID[o.ID] = o
	s.mu.Unlock()
}

func (s *store) get(id string) (Order, bool) {
	s.mu.RLock()
	o, ok := s.byID[id]
	s.mu.RUnlock()
	return o, ok
}

func (s *store) list() []Order {
	s.mu.RLock()
	result := make([]Order, 0, len(s.byID))
	for _, o := range s.byID {
		result = append(result, o)
	}
	s.mu.RUnlock()
	return result
}

// ---------------------------------------------------------------------------
// Pricing — simulated external service (prices in cents)
// ---------------------------------------------------------------------------

// priceOrder returns a pricing function that closes over the catalog.
func priceOrder(catalog map[string]int) func(context.Context, Order) (Order, error) {
	return func(_ context.Context, o Order) (Order, error) {
		total := 0
		for _, item := range o.Items {
			total += catalog[item.SKU] * item.Quantity
		}
		o.TotalCents = total
		o.Status = statusPriced
		return o, nil
	}
}

// ---------------------------------------------------------------------------
// Background post-processing — best-effort audit and metrics
// ---------------------------------------------------------------------------

func startPipeline(_ context.Context, postCh <-chan Order) {
	go func() {
		for o := range postCh {
			log.Printf("  postprocess: audit order %s (%d cents)", o.ID, o.TotalCents)
			log.Printf("  postprocess: metrics %d line items", len(o.Items))
		}
	}()
}

// ---------------------------------------------------------------------------
// Handler factories
// ---------------------------------------------------------------------------

func newCreateOrder(
	s *store,
	idCounter *atomic.Int64,
	postCh chan<- Order,
	catalog map[string]int,
) web.Handler {
	// validateOrder checks business rules, closing over the price catalog.
	validateOrder := func(o Order) rslt.Result[Order] {
		if o.Customer == "" {
			return rslt.Err[Order](web.BadRequest("customer is required"))
		}
		if len(o.Items) == 0 {
			return rslt.Err[Order](web.BadRequest("order must have at least one item"))
		}
		if !slice.From(o.Items).Every(LineItem.HasPositiveQty) {
			return rslt.Err[Order](web.BadRequest("all items must have positive quantity"))
		}
		for _, item := range o.Items {
			if _, ok := catalog[item.SKU]; !ok {
				return rslt.Err[Order](web.BadRequest(
					fmt.Sprintf("unknown SKU: %s", item.SKU)))
			}
		}
		return rslt.Ok(o)
	}

	priceFn := priceOrder(catalog)

	withNewID := func(o Order) Order {
		o.ID = fmt.Sprintf("ord-%d", idCounter.Add(1))
		o.Status = statusPending
		return o
	}

	return func(req *http.Request) rslt.Result[web.Response] {
		reqID := ctxval.Lookup[RequestID](req.Context()).Or("unknown")

		// lookupPrices binds the request context to the pricing call.
		lookupPrices := func(o Order) rslt.Result[Order] {
			return rslt.Of(priceFn(req.Context(), o))
		}

		logFailure := func(err error) {
			log.Printf("[%s] order failed: %v", reqID, err)
		}

		storeAndNotify := func(o Order) {
			s.put(o)
			log.Printf("[%s] created order %s (%d cents)", reqID, o.ID, o.TotalCents)
			select {
			case postCh <- o:
			default:
				log.Printf("[%s] post-processing channel full, skipping", reqID)
			}
		}

		order, err := web.DecodeJSON[Order](req)
		storedResult := rslt.Of(order, err).
			FlatMap(validateOrder).
			Transform(withNewID).
			FlatMap(lookupPrices).
			TapErr(logFailure).
			Tap(storeAndNotify)
		return rslt.Map(storedResult, web.Created[Order])
	}
}

func newGetOrder(s *store) web.Handler {
	findOrder := func(id string) rslt.Result[Order] {
		return option.New(s.get(id)).OkOr(web.NotFound("order not found"))
	}

	return func(req *http.Request) rslt.Result[web.Response] {
		idResult := web.PathParam(req, "id").OkOr(web.BadRequest("missing order id"))
		foundResult := rslt.FlatMap(idResult, findOrder)
		return rslt.Map(foundResult, web.OK[Order])
	}
}

// parseMinTotal parses a min_total query parameter as an integer (cents).
func parseMinTotal(raw string) rslt.Result[int] {
	msg := fmt.Sprintf("min_total must be an integer (cents), got %q", raw)
	return option.Atoi(raw).OkOr(web.BadRequest(msg))
}

func newListOrders(s *store) web.Handler {
	return func(req *http.Request) rslt.Result[web.Response] {
		q := req.URL.Query()

		status, hasStatus := option.NonEmpty(q.Get("status")).Get()

		rawMinTotalOption := option.NonEmpty(q.Get("min_total"))
		mtOption, err := option.MapResult(rawMinTotalOption, parseMinTotal).Unpack()
		if err != nil {
			return rslt.Err[web.Response](err)
		}
		mt, hasMinTotal := mtOption.Get()

		hasMatchingStatus := func(o Order) bool { return !hasStatus || o.Status == OrderStatus(status) }
		totalAtLeast := func(o Order) bool { return !hasMinTotal || o.TotalCents >= mt }

		orders := slice.SortBy(s.list(), orderNum).KeepIf(hasMatchingStatus).KeepIf(totalAtLeast)

		return rslt.Ok(web.OK(orders))
	}
}

// ---------------------------------------------------------------------------
// main — wiring
// ---------------------------------------------------------------------------

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s := newStore()
	var idCounter atomic.Int64

	catalog := map[string]int{
		"WIDGET-1": 999,
		"GADGET-2": 2450,
		"GIZMO-3":  500,
	}

	postCh := make(chan Order, 20)
	startPipeline(ctx, postCh)

	handleCreateOrder := newCreateOrder(s, &idCounter, postCh, catalog)
	handleGetOrder := newGetOrder(s)
	handleListOrders := newListOrders(s)

	var reqCounter atomic.Int64
	withRequestID := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := RequestID(fmt.Sprintf("req-%d", reqCounter.Add(1)))
			ctx := ctxval.With(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", web.Adapt(handleCreateOrder))
	mux.HandleFunc("GET /orders/{id}", web.Adapt(handleGetOrder))
	mux.HandleFunc("GET /orders", web.Adapt(handleListOrders))

	server := &http.Server{
		Addr:              ":3000",
		Handler:           withRequestID(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Println("listening on :3000")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
