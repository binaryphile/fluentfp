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

// enrichOrder computes the total by looking up prices per SKU.
// SKU "FAIL-PRICE" deterministically fails to simulate a service outage.
// Assumes SKUs are already validated — unknown SKUs are caught in validation.
func enrichOrder(_ context.Context, o Order) (Order, error) {
	total := 0
	for _, item := range o.Items {
		if item.SKU == "FAIL-PRICE" {
			return o, errPricingFailure
		}
		total += prices[item.SKU] * item.Quantity
	}
	o.TotalCents = total
	o.Status = "enriched"
	return o, nil
}

// ---------------------------------------------------------------------------
// Validation — top-level named functions
// ---------------------------------------------------------------------------

// Validators: each takes an Order and returns Result[Order].
// Ok means valid (pass through), Err means invalid (rest of chain skipped).
// web.BadRequest returns a 400 error that carries the HTTP status code.
// rslt.Ok / rslt.Err wrap the value or error in a Result.

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
// slice.From wraps a slice for chaining. Every returns true if the
// predicate holds for all elements. LineItem.HasPositiveQty is a
// method expression — Go turns the method into a plain function.
func itemsHavePositiveQty(o Order) rslt.Result[Order] {
	if !slice.From(o.Items).Every(LineItem.HasPositiveQty) {
		return rslt.Err[Order](web.BadRequest("all items must have positive quantity"))
	}
	return rslt.Ok(o)
}

// itemsHaveKnownSKUs validates all SKUs exist in the price table.
// This runs before the breaker-wrapped enrichment so bad input
// doesn't count as a service failure.
func itemsHaveKnownSKUs(o Order) rslt.Result[Order] {
	for _, item := range o.Items {
		if item.SKU == "FAIL-PRICE" {
			continue // synthetic failure SKU, not a validation error
		}
		if _, ok := prices[item.SKU]; !ok {
			return rslt.Err[Order](web.BadRequest(
				fmt.Sprintf("unknown SKU: %s", item.SKU)))
		}
	}
	return rslt.Ok(o)
}

// ---------------------------------------------------------------------------
// Error mapping
// ---------------------------------------------------------------------------

// mapDomainError maps domain errors to HTTP errors at the adapter boundary.
// web.Adapt calls this for any error that isn't already a *web.Error.
// Return (*web.Error, true) to handle, or (nil, false) to fall through to 500.
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
	return nil, false
}

// ---------------------------------------------------------------------------
// Background pipeline (toc) — best-effort post-processing
// ---------------------------------------------------------------------------

// logOrder logs the order for the audit trail and returns its ID.
func logOrder(_ context.Context, o Order) (string, error) {
	log.Printf("  postprocess: audit order %s (%d cents)", o.ID, o.TotalCents)
	return o.ID, nil
}

// countItems counts SKU line items for inventory tracking.
func countItems(_ context.Context, o Order) (int, error) {
	return len(o.Items), nil
}

// drainResults logs each result or error from a pipeline branch.
// r.Err() returns the error if this Result is Err, or nil if Ok.
// r.Get() returns (value, true) if Ok, or (zero, false) if Err.
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
	// toc.FromChan bridges plain chan Order → chan Result[Order] for toc.
	// toc.NewTee broadcasts each item to N branches (here 2).
	tee := toc.NewTee(ctx, toc.FromChan(postCh), 2)

	// toc.Pipe chains a processing function onto a branch's output.
	// Branch 0: audit log. Branch 1: inventory count.
	auditPipe := toc.Pipe(ctx, tee.Branch(0), logOrder, toc.Options[Order]{})
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

	// findOrder looks up an order by ID, returning 404 on miss.
	// option.New wraps (Order, bool) → Option; OkOr bridges to Result.
	findOrder := func(id string) rslt.Result[Order] {
		return option.New(s.get(id)).OkOr(web.NotFound("order not found"))
	}

	// withNewID assigns a sequential ID and sets initial status.
	// Not pure (idCounter.Add is a side effect), but returns a new
	// value — Transform is the right shape (func(T) T).
	withNewID := func(o Order) Order {
		o.ID = fmt.Sprintf("ord-%d", idCounter.Add(1))
		o.Status = "pending"
		return o
	}

	// --- Circuit breaker for pricing enrichment ---
	// call.NewBreaker creates a 3-state breaker (closed → open → half-open).
	// call.WithBreaker wraps enrichOrder — same signature, breaker invisible.
	// After 3 consecutive failures, calls are rejected with call.ErrCircuitOpen.

	breaker := call.NewBreaker(call.BreakerConfig{
		ResetTimeout: 10 * time.Second,
		ReadyToTrip:  call.ConsecutiveFailures(3),
		OnStateChange: func(t call.Transition) {
			log.Printf("breaker: %s → %s", t.From, t.To)
		},
	})
	enrichWithBreaker := call.WithBreaker(breaker, enrichOrder)

	// --- Best-effort post-processing pipeline ---
	// After the HTTP handler stores an order, it sends a copy here
	// (non-blocking) for background audit logging and inventory tracking.
	// startPipeline reads from this channel and fans out via toc.Tee.

	postCh := make(chan Order, 20)
	startPipeline(ctx, postCh)

	// --- Validation + error mapping ---

	// Steps chains validators — runs each in order, skips the rest on first error.
	validateOrder := web.Steps(
		hasCustomer, hasItems, itemsHavePositiveQty, itemsHaveKnownSKUs)
	// WithErrorMapper translates domain errors to HTTP errors once,
	// at the Adapt boundary — handlers just return the error.
	errorMapper := web.WithErrorMapper(mapDomainError)

	// --- POST /orders ---
	// Handlers return Result[Response] instead of writing to ResponseWriter.
	// Ok = success response, Err = error response. Adapt renders both.

	handleCreateOrder := func(req *http.Request) rslt.Result[web.Response] {
		// Get request ID from context (set by middleware).
		reqID := ctxval.From[RequestID](req.Context()).Or("unknown")

		// enrichWithBreaker is func(ctx, Order) (Order, error) — two args.
		// The pipeline needs func(Order) Result[Order] — one arg.
		// LiftCtx binds the context and wraps (Order, error) → Result[Order].
		enrich := rslt.LiftCtx(req.Context(), enrichWithBreaker)

		// Side effect: log on enrichment failure.
		logFailure := func(err error) {
			log.Printf("[%s] enrichment failed: %v", reqID, err)
		}

		// Side effect: persist and send to background pipeline.
		storeAndNotify := func(o Order) {
			s.put(o)
			log.Printf("[%s] created order %s (%d cents)", reqID, o.ID, o.TotalCents)
			select {
			case postCh <- o:
			default:
				log.Printf("[%s] post-processing channel full, skipping", reqID)
			}
		}

		// Pipeline: each step operates on the Result from the previous step.
		// If any step fails, the rest are skipped and the error propagates.
		orderResult := web.DecodeJSON[Order](req)    // parse JSON → Result[Order]
		storedResult := orderResult.
			FlatMap(validateOrder).            // validate (can fail → 400)
			Transform(withNewID).              // assign ID + status
			FlatMap(enrich).                   // call pricing (can fail → 502/503)
			TapErr(logFailure).                // on error: log it
			Tap(storeAndNotify)                // on success: persist + notify
		return rslt.Map(storedResult, web.Created[Order]) // wrap in 201 response
	}

	// --- GET /orders/{id} ---

	handleGetOrder := func(req *http.Request) rslt.Result[web.Response] {
		// Get path param as Option; missing -> Err(400).
		idResult := web.PathParam(req, "id").
			OkOr(web.BadRequest("missing order id"))
		// FlatMap: findOrder can fail (404), so it returns Result.
		foundResult := rslt.FlatMap(idResult, findOrder)
		// Map: web.OK always succeeds (just wraps in 200).
		return rslt.Map(foundResult, web.OK[Order])
	}

	// --- GET /orders?status=X&min_total=Y (cents) ---

	handleListOrders := func(req *http.Request) rslt.Result[web.Response] {
		q := req.URL.Query()

		// Parse optional query params.
		// NonEmpty: "" → not-ok, non-empty → ok. Get: unpack to (value, bool).
		status, hasStatus := option.NonEmpty(q.Get("status")).Get()

		// Parse min_total if present. FlatMapResult skips parsing when
		// the param is missing, parses it when present, and returns 400
		// if present but not a valid integer.
		parseMinTotal := func(raw string) rslt.Result[int] {
			return option.Atoi(raw).OkOr(web.BadRequest(
				fmt.Sprintf(
					"min_total must be an integer (cents), got %q",
					raw)))
		}
		rawMinTotalOption := option.NonEmpty(q.Get("min_total"))
		minTotalResult := option.FlatMapResult(rawMinTotalOption, parseMinTotal)
		// Unpack converts Result back to Go's (value, error) pair.
		mtOption, err := minTotalResult.Unpack()
		if err != nil {
			return rslt.Err[web.Response](err)
		}
		mt, hasMinTotal := mtOption.Get()

		// Named predicates for optional filters.
		hasMatchingStatus := func(o Order) bool { return o.Status == status }
		totalAtLeast := func(o Order) bool { return o.TotalCents >= mt }

		// Sort + conditional filter. KeepIf is like filter() — keeps
		// elements where the predicate returns true.
		orders := slice.SortBy(s.list(), orderNum)
		if hasStatus {
			orders = orders.KeepIf(hasMatchingStatus)
		}
		if hasMinTotal {
			orders = orders.KeepIf(totalAtLeast)
		}

		return rslt.Ok(web.OK(orders))
	}

	// --- Middleware: inject request ID via ctxval ---
	// ctxval.With stores a value keyed by its Go type — no sentinel keys.
	// Handlers retrieve it with ctxval.From[RequestID](ctx).

	var reqCounter atomic.Int64
	withRequestID := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := RequestID(fmt.Sprintf("req-%d", reqCounter.Add(1)))
			ctx := ctxval.With(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// --- Routes ---
	// web.Adapt converts our Result-returning handlers to standard
	// http.HandlerFunc. It renders Ok as JSON responses and Err as
	// JSON error responses. errorMapper translates domain errors.

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
