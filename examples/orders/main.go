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
	"github.com/binaryphile/fluentfp/toc"
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
	Status     string     `json:"status"`
	TotalCents int        `json:"total_cents"`
}

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

var prices = map[string]int{
	"WIDGET-1": 999,
	"GADGET-2": 2450,
	"GIZMO-3":  500,
}

// priceOrder computes the total by looking up prices per SKU.
// Assumes SKUs are already validated.
func priceOrder(_ context.Context, o Order) (Order, error) {
	total := 0
	for _, item := range o.Items {
		total += prices[item.SKU] * item.Quantity
	}
	o.TotalCents = total
	o.Status = "priced"
	return o, nil
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// Validate checks all business rules on the order. As a method,
// Order.Validate is a method expression usable in FlatMap chains.
func (o Order) Validate() rslt.Result[Order] {
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
		if _, ok := prices[item.SKU]; !ok {
			return rslt.Err[Order](web.BadRequest(
				fmt.Sprintf("unknown SKU: %s", item.SKU)))
		}
	}
	return rslt.Ok(o)
}

// ---------------------------------------------------------------------------
// Background pipeline (toc) — best-effort post-processing
// ---------------------------------------------------------------------------

func logOrder(_ context.Context, o Order) (string, error) {
	log.Printf("  postprocess: audit order %s (%d cents)", o.ID, o.TotalCents)
	return o.ID, nil
}

func countItems(_ context.Context, o Order) (int, error) {
	return len(o.Items), nil
}

func drainResults[T any](name string, ch <-chan rslt.Result[T]) {
	for r := range ch {
		if err := r.Err(); err != nil {
			log.Printf("  postprocess: %s error: %v", name, err)
			continue
		}
		v, _ := r.Get()
		log.Printf("  postprocess: %s result: %v", name, v)
	}
}

func startPipeline(ctx context.Context, postCh <-chan Order) {
	tee := toc.NewTee(ctx, toc.FromChan(postCh), 2)
	auditPipe := toc.Pipe(ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
	inventoryPipe := toc.Pipe(ctx, tee.Branch(1), countItems, toc.Options[Order]{})
	go drainResults("audit", auditPipe.Out())
	go drainResults("inventory", inventoryPipe.Out())
}

// ---------------------------------------------------------------------------
// Handler factories
// ---------------------------------------------------------------------------

func newCreateOrder(
	s *store,
	idCounter *atomic.Int64,
	postCh chan<- Order,
) web.Handler {
	withNewID := func(o Order) Order {
		o.ID = fmt.Sprintf("ord-%d", idCounter.Add(1))
		o.Status = "pending"
		return o
	}

	return func(req *http.Request) rslt.Result[web.Response] {
		reqID := ctxval.Get[RequestID](req.Context()).Or("unknown")

		lookupPrices := rslt.LiftCtx(req.Context(), priceOrder)

		logFailure := func(err error) {
			log.Printf("[%s] pricing failed: %v", reqID, err)
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

		orderResult := web.DecodeJSON[Order](req)
		storedResult := orderResult.
			FlatMap(Order.Validate).
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

func newListOrders(s *store) web.Handler {
	return func(req *http.Request) rslt.Result[web.Response] {
		q := req.URL.Query()

		status, hasStatus := option.NonEmpty(q.Get("status")).Get()

		parseMinTotal := func(raw string) rslt.Result[int] {
			return option.Atoi(raw).OkOr(web.BadRequest(
				fmt.Sprintf("min_total must be an integer (cents), got %q", raw)))
		}
		rawMinTotalOption := option.NonEmpty(q.Get("min_total"))
		minTotalResult := option.MapResult(rawMinTotalOption, parseMinTotal)
		mtOption, err := minTotalResult.Unpack()
		if err != nil {
			return rslt.Err[web.Response](err)
		}
		mt, hasMinTotal := mtOption.Get()

		hasMatchingStatus := func(o Order) bool { return o.Status == status }
		totalAtLeast := func(o Order) bool { return o.TotalCents >= mt }

		orders := slice.SortBy(s.list(), orderNum)
		if hasStatus {
			orders = orders.KeepIf(hasMatchingStatus)
		}
		if hasMinTotal {
			orders = orders.KeepIf(totalAtLeast)
		}

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

	postCh := make(chan Order, 20)
	startPipeline(ctx, postCh)

	handleCreateOrder := newCreateOrder(s, &idCounter, postCh)
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
