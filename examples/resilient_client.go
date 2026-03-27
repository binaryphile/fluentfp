//go:build ignore

// Resilient API client using wrap.Func().WithBreaker().WithRetry().WithMapError().
//
// Run:
//
//	go run examples/resilient_client.go
//
// Three decorators chain fluently — each returns the same wrap.Fn signature:
//
//   1. WithMapError: classify errors (transient vs permanent)
//   2. WithRetry: retry transient errors with exponential backoff
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

	"github.com/binaryphile/fluentfp/wrap"
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
	breaker := wrap.NewBreaker(wrap.BreakerConfig{
		ResetTimeout: 5 * time.Second,
		ReadyToTrip:  wrap.ConsecutiveFailures(2),
		OnStateChange: func(t wrap.Transition) {
			fmt.Printf("  breaker: %s → %s\n", t.From, t.To)
		},
	})

	// Chain three decorators — reads bottom to top at call time:
	//   fetchUser → MapError(classify) → Retry(3, backoff) → Breaker
	safeFetch := wrap.Func(fetchUser).
		WithMapError(classifyError).
		WithRetry(3, wrap.ExpBackoff(50*time.Millisecond), isTransient).
		WithBreaker(breaker)

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
