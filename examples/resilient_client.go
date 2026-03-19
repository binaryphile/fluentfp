//go:build ignore

// Resilient API client in 20 lines using call.CircuitBreaker + call.Retry + call.MapErr.
//
// Run:
//
//	go run examples/resilient_client.go
//
// Three decorators compose by stacking — each wraps the previous function,
// preserving the same func(context.Context, T) (R, error) signature:
//
//   1. MapErr: classify errors (transient vs permanent)
//   2. Retry: retry transient errors with exponential backoff
//   3. WithBreaker: short-circuit when the dependency is unhealthy
//
// The order matters: breaker wraps retry so it sees one logical operation.
// A request that retries 3 times and still fails counts as one breaker failure.
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/binaryphile/fluentfp/call"
)

// Simulated errors.
var (
	errTimeout    = errors.New("request timeout")
	errBadGateway = errors.New("502 bad gateway")
	errNotFound   = errors.New("404 not found")
)

// isTransient returns true for errors worth retrying.
func isTransient(err error) bool {
	return errors.Is(err, errTimeout) || errors.Is(err, errBadGateway)
}

// fetchUser simulates an unreliable API call.
func fetchUser(_ context.Context, id string) (string, error) {
	switch rand.IntN(4) {
	case 0:
		return "", errTimeout
	case 1:
		return "", errBadGateway
	case 2:
		return "", errNotFound
	default:
		return fmt.Sprintf("User(%s)", id), nil
	}
}

// classifyError annotates errors with transient/permanent context.
func classifyError(err error) error {
	if isTransient(err) {
		return fmt.Errorf("transient: %w", err)
	}
	return fmt.Errorf("permanent: %w", err)
}

func main() {
	// Stack three decorators — each preserves func(ctx, T) (R, error):
	//   fetchUser → MapErr(classify) → Retry(3, backoff) → WithBreaker
	classified := call.MapErr(fetchUser, classifyError)
	retried := call.Retry(3, call.ExponentialBackoff(50*time.Millisecond), isTransient, classified)

	breaker := call.NewBreaker(call.BreakerConfig{
		ResetTimeout: 5 * time.Second,
		ReadyToTrip:  call.ConsecutiveFailures(2),
		OnStateChange: func(t call.Transition) {
			fmt.Printf("  breaker: %s → %s\n", t.From, t.To)
		},
	})
	safeFetch := call.WithBreaker(breaker, retried)

	// Try 10 requests.
	ctx := context.Background()
	for i := range 10 {
		user, err := safeFetch(ctx, fmt.Sprintf("user-%d", i))
		if err != nil {
			fmt.Printf("[%d] error: %v\n", i, err)
		} else {
			fmt.Printf("[%d] ok: %s\n", i, user)
		}
	}
}
