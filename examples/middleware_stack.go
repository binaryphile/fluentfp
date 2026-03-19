//go:build ignore

// Integration example: web + ctxval + hof.CircuitBreaker middleware stack.
//
// Run:
//
//	go run examples/middleware_stack.go
//
// Shows how fluentfp packages compose at the HTTP boundary:
//   - ctxval: inject request ID via middleware
//   - web: typed handlers returning Result[Response]
//   - hof: circuit breaker wrapping an external service call
//   - option: safe query parameter parsing
//   - rslt: FlatMap/Map for error-free composition
//
// The middleware, handler, breaker, and error mapper are all independent
// pieces that compose without coupling.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/ctxval"
	"github.com/binaryphile/fluentfp/hof"
	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/web"
)

// RequestID is a request-scoped correlation ID.
type RequestID string

// --- External service simulation ---

var errServiceDown = errors.New("service unavailable")

// lookupWeather simulates an external weather API.
func lookupWeather(_ context.Context, city string) (string, error) {
	if city == "atlantis" {
		return "", errServiceDown
	}
	return fmt.Sprintf("Sunny in %s, 72°F", city), nil
}

// --- Error mapping (defined once, shared across handlers) ---

// mapDomainError maps domain errors to HTTP errors at the adapter boundary.
func mapDomainError(err error) (*web.Error, bool) {
	if errors.Is(err, hof.ErrCircuitOpen) {
		return &web.Error{
			Status:  http.StatusServiceUnavailable,
			Message: "weather service temporarily unavailable",
			Code:    "SERVICE_UNAVAILABLE",
		}, true
	}
	return nil, false
}

func main() {
	// --- Circuit breaker ---

	breaker := hof.NewBreaker(hof.BreakerConfig{
		ResetTimeout: 10 * time.Second,
		ReadyToTrip:  hof.ConsecutiveFailures(2),
		OnStateChange: func(t hof.Transition) {
			log.Printf("breaker: %s → %s", t.From, t.To)
		},
	})
	safeWeather := hof.WithBreaker(breaker, lookupWeather)

	// --- Middleware: request ID ---

	var reqCounter atomic.Int64
	withRequestID := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := RequestID(fmt.Sprintf("req-%d", reqCounter.Add(1)))
			ctx := ctxval.With(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// --- Handler ---

	handleWeather := func(req *http.Request) rslt.Result[web.Response] {
		reqID := ctxval.From[RequestID](req.Context()).Or("unknown")
		city := option.NonEmpty(req.URL.Query().Get("city")).Or("london")

		log.Printf("[%s] looking up weather for %s", reqID, city)
		return rslt.Map(
			rslt.Of(safeWeather(req.Context(), city)),
			web.OK[string],
		)
	}

	// --- Routes ---

	errorMapper := web.WithErrorMapper(mapDomainError)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /weather", web.Adapt(handleWeather, errorMapper))

	server := &http.Server{
		Addr:              ":3001",
		Handler:           withRequestID(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on :3001")
	log.Println("try: curl 'http://localhost:3001/weather?city=paris'")
	log.Println("trip breaker: curl 'http://localhost:3001/weather?city=atlantis' (2x)")
	log.Fatal(server.ListenAndServe())
}
