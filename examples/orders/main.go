// Package main demonstrates fluentfp composition across 7 packages in a
// curl-testable order processing service.
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
//
// Trip the circuit breaker with SKU "FAIL-PRICE":
//
//	for i in 1 2 3; do curl -s -X POST http://localhost:3000/orders \
//	  -H 'Content-Type: application/json' \
//	  -d '{"customer":"Bob","items":[{"sku":"FAIL-PRICE","quantity":1}]}'; done
//	curl -s -X POST http://localhost:3000/orders \
//	  -H 'Content-Type: application/json' \
//	  -d '{"customer":"Carol","items":[{"sku":"WIDGET-1","quantity":1}]}'
//	# → 503: circuit breaker open
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
	"github.com/binaryphile/fluentfp/call"
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

// Order is the core domain object. TotalCents uses integer cents to avoid
// floating-point currency errors.
type Order struct {
	ID         string     `json:"id"`
	Customer   string     `json:"customer"`
	Items      []LineItem `json:"items"`
	Status     string     `json:"status"`
	TotalCents int        `json:"total_cents"`
}

// GetID returns the order ID (method expression: Order.GetID).
func (o Order) GetID() string { return o.ID }

// GetStatus returns the order status (method expression: Order.GetStatus).
func (o Order) GetStatus() string { return o.Status }

// GetTotalCents returns the order total in cents (method expression: Order.GetTotalCents).
func (o Order) GetTotalCents() int { return o.TotalCents }

// orderNum extracts the numeric suffix from an order ID for sorting.
// "ord-2" → 2, "ord-10" → 10. Returns 0 if parsing fails.
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

var errPricingFailure = errors.New("pricing service error")
var errUnknownSKU = errors.New("unknown SKU")

// enrichOrder computes the total by looking up prices per SKU.
// SKU "FAIL-PRICE" deterministically fails. Unknown SKUs return an error.
func enrichOrder(_ context.Context, o Order) (Order, error) {
	total := 0
	for _, item := range o.Items {
		if item.SKU == "FAIL-PRICE" {
			return o, errPricingFailure
		}
		price, ok := option.Lookup(prices, item.SKU).Get()
		if !ok {
			return o, fmt.Errorf("%w: %s", errUnknownSKU, item.SKU)
		}
		total += price * item.Quantity
	}
	o.TotalCents = total
	o.Status = "enriched"
	return o, nil
}

// ---------------------------------------------------------------------------
// Validation — top-level named functions
// ---------------------------------------------------------------------------

// hasCustomer validates that the customer field is non-empty.
func hasCustomer(o Order) rslt.Result[Order] {
	if o.Customer == "" {
		return rslt.Err[Order](web.BadRequest("customer is required"))
	}
	return rslt.Ok(o)
}

// hasItems validates that an order has at least one item.
func hasItems(o Order) rslt.Result[Order] {
	if len(o.Items) == 0 {
		return rslt.Err[Order](web.BadRequest("order must have at least one item"))
	}
	return rslt.Ok(o)
}

// itemsHavePositiveQty validates all items have quantity > 0.
func itemsHavePositiveQty(o Order) rslt.Result[Order] {
	if !slice.From(o.Items).Every(LineItem.HasPositiveQty) {
		return rslt.Err[Order](web.BadRequest("all items must have positive quantity"))
	}
	return rslt.Ok(o)
}

// ---------------------------------------------------------------------------
// Error mapping
// ---------------------------------------------------------------------------

// mapDomainError maps domain errors to HTTP errors at the adapter boundary.
func mapDomainError(err error) (*web.Error, bool) {
	if errors.Is(err, call.ErrCircuitOpen) {
		return &web.Error{
			Status:  http.StatusServiceUnavailable,
			Message: "pricing service unavailable",
			Code:    "SERVICE_UNAVAILABLE",
		}, true
	}
	if errors.Is(err, errPricingFailure) {
		return &web.Error{
			Status:  http.StatusBadGateway,
			Message: "pricing service error",
			Code:    "BAD_GATEWAY",
		}, true
	}
	if errors.Is(err, errUnknownSKU) {
		return &web.Error{
			Status:  http.StatusUnprocessableEntity,
			Message: err.Error(),
			Code:    "UNKNOWN_SKU",
		}, true
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// Background pipeline (toc) — best-effort post-processing
// ---------------------------------------------------------------------------

// logOrder logs the order for the audit trail and returns its ID.
func logOrder(_ context.Context, o Order) (string, error) {
	log.Printf("  postprocess: audit order %s (%d cents)", o.ID, o.TotalCents)
	return o.GetID(), nil
}

// countItems counts SKU line items for inventory tracking.
func countItems(_ context.Context, o Order) (int, error) {
	return len(o.Items), nil
}

// drainResults logs each result or error from a pipeline branch.
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

// startPipeline launches a long-lived toc pipeline that broadcasts each
// order to an audit branch and an inventory branch via Tee.
// Post-processing is best-effort: errors are logged, not propagated.
func startPipeline(ctx context.Context, postCh <-chan Order) {
	// Stage accepts orders and passes them through for downstream Tee.
	passthrough := func(_ context.Context, o Order) (Order, error) { return o, nil }
	stage := toc.Start(ctx, passthrough, toc.Options[Order]{
		Capacity: 10,
		Workers:  1,
	})

	// Feed channel → stage.
	go func() {
		for o := range postCh {
			if err := stage.Submit(ctx, o); err != nil {
				log.Printf("  postprocess: submit failed: %v", err)
			}
		}
		stage.CloseInput()
	}()

	// Tee: broadcast each order to two branches.
	tee := toc.NewTee(ctx, stage.Out(), 2)

	// Branch 0: audit log.
	auditPipe := toc.Pipe(ctx, tee.Branch(0), logOrder, toc.Options[Order]{})

	// Branch 1: inventory count.
	inventoryPipe := toc.Pipe(ctx, tee.Branch(1), countItems, toc.Options[Order]{})

	go drainResults("audit", auditPipe.Out())
	go drainResults("inventory", inventoryPipe.Out())
}

// ---------------------------------------------------------------------------
// HTTP handlers
// ---------------------------------------------------------------------------

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s := newStore()
	var idCounter atomic.Int64

	// withNewID sets the order ID and initial status.
	withNewID := func(o Order) Order {
		o.ID = fmt.Sprintf("ord-%d", idCounter.Add(1))
		o.Status = "pending"
		return o
	}

	// --- Circuit breaker for pricing enrichment ---

	breaker := call.NewBreaker(call.BreakerConfig{
		ResetTimeout: 10 * time.Second,
		ReadyToTrip:  call.ConsecutiveFailures(3),
		OnStateChange: func(t call.Transition) {
			log.Printf("breaker: %s → %s", t.From, t.To)
		},
	})
	enrichWithBreaker := call.WithBreaker(breaker, enrichOrder)

	// --- Best-effort post-processing pipeline ---

	postCh := make(chan Order, 20)
	startPipeline(ctx, postCh)

	// --- Validation + error mapping ---

	validateOrder := web.Steps(hasCustomer, hasItems, itemsHavePositiveQty)
	errorMapper := web.WithErrorMapper(mapDomainError)

	// --- POST /orders ---

	handleCreateOrder := func(req *http.Request) rslt.Result[web.Response] {
		reqID := ctxval.From[RequestID](req.Context()).Or("unknown")

		// enrich calls the breaker-wrapped pricing service.
		enrich := func(o Order) rslt.Result[Order] {
			return rslt.Of(enrichWithBreaker(req.Context(), o))
		}

		// logFailure logs enrichment errors.
		logFailure := func(err error) {
			log.Printf("[%s] enrichment failed: %v", reqID, err)
		}

		// storeAndNotify persists the order and sends to post-processing.
		storeAndNotify := func(o Order) {
			s.put(o)
			log.Printf("[%s] created order %s (%d cents)", reqID, o.ID, o.TotalCents)
			select {
			case postCh <- o:
			default:
				log.Printf("[%s] post-processing channel full, skipping", reqID)
			}
		}

		order := web.DecodeJSON[Order](req)
		stored := order.
			FlatMap(validateOrder).
			Transform(withNewID).
			FlatMap(enrich).
			TapErr(logFailure).
			Tap(storeAndNotify)
		return rslt.Map(stored, web.Created[Order])
	}

	// --- GET /orders/{id} ---

	handleGetOrder := func(req *http.Request) rslt.Result[web.Response] {
		id := req.PathValue("id")
		if id == "" {
			return rslt.Err[web.Response](web.BadRequest("missing order id"))
		}

		found := option.New(s.get(id)).OkOr(web.NotFound("order not found"))
		return rslt.Map(found, web.OK[Order])
	}

	// --- GET /orders?status=X&min_total=Y (cents) ---

	handleListOrders := func(req *http.Request) rslt.Result[web.Response] {
		q := req.URL.Query()

		status, hasStatus := option.NonEmpty(q.Get("status")).Get()
		minTotalOption := option.FlatMap(option.NonEmpty(q.Get("min_total")), option.Atoi)
		if raw, ok := option.NonEmpty(q.Get("min_total")).Get(); ok {
			if _, ok := minTotalOption.Get(); !ok {
				return rslt.Err[web.Response](web.BadRequest(
					fmt.Sprintf("min_total must be an integer (cents), got %q", raw)))
			}
		}
		mt, hasMinTotal := minTotalOption.Get()

		// hasMatchingStatus checks if order status matches the filter.
		hasMatchingStatus := func(o Order) bool { return o.Status == status }
		// totalAtLeast checks if order total meets the minimum.
		totalAtLeast := func(o Order) bool { return o.TotalCents >= mt }

		sorted := slice.SortBy(s.list(), orderNum)
		orders := sorted.
			KeepIfWhen(hasStatus, hasMatchingStatus).
			KeepIfWhen(hasMinTotal, totalAtLeast)

		return rslt.Ok(web.OK(orders))
	}

	// --- Middleware: inject request ID via ctxval ---

	var reqCounter atomic.Int64
	withRequestID := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := RequestID(fmt.Sprintf("req-%d", reqCounter.Add(1)))
			ctx := ctxval.With(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// --- Routes ---

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", web.Adapt(handleCreateOrder, errorMapper))
	mux.HandleFunc("GET /orders/{id}", web.Adapt(handleGetOrder, errorMapper))
	mux.HandleFunc("GET /orders", web.Adapt(handleListOrders, errorMapper))

	server := &http.Server{
		Addr:              ":3000",
		Handler:           withRequestID(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Shutdown: stop accepting requests, drain in-flight handlers.
	// Context cancellation propagates to toc pipeline via ctx.
	// postCh is not closed here — it is drained by context cancellation
	// in the pipeline, avoiding send-on-closed-channel panics.
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
