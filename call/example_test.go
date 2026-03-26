package call_test

import (
	"context"
	"fmt"
	"time"

	"github.com/binaryphile/fluentfp/call"
)

func ExampleBracket() {
	// Simulate a semaphore with capacity 1.
	sem := make(chan struct{}, 1)

	// semAcquire acquires a slot, returns a release that frees it.
	semAcquire := func(_ context.Context, _ string) (func(), error) {
		sem <- struct{}{}
		return func() { <-sem }, nil
	}

	// process converts a string to uppercase with a greeting.
	process := func(_ context.Context, s string) (string, error) {
		return fmt.Sprintf("hello, %s", s), nil
	}

	guarded := call.From(process).With(call.Bracket[string, string](semAcquire))
	result, _ := guarded(context.Background(), "world")
	fmt.Println(result)
	// Output: hello, world
}

func ExampleRetrier() {
	// double doubles the input. Succeeds on first try.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// Retry up to 3 times with constant 1ms backoff.
	resilient := call.From(double).With(
		call.Retrier[int, int](3, call.ConstantBackoff(time.Millisecond), nil),
	)

	got, _ := resilient(context.Background(), 5)
	fmt.Println(got)
	// Output: 10
}

func ExampleCircuitBreaker() {
	// double doubles the input.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// Protect a dependency with a circuit breaker that trips after 5 consecutive failures.
	breaker := call.NewBreaker(call.BreakerConfig{
		ResetTimeout: 10 * time.Second,
	})

	protected := call.From(double).With(call.CircuitBreaker[int, int](breaker))
	got, _ := protected(context.Background(), 21)
	fmt.Println(got)
	// Output: 42
}

func ExampleThrottler() {
	// double doubles the input.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// Allow at most 5 concurrent calls.
	limited := call.From(double).With(call.Throttler[int, int](5))

	got, _ := limited(context.Background(), 3)
	fmt.Println(got)
	// Output: 6
}

func ExampleErrMapper() {
	// fetchUser simulates a user lookup that fails.
	fetchUser := func(_ context.Context, id int) (string, error) {
		return "", fmt.Errorf("not found")
	}

	// annotate wraps errors with calling context.
	annotate := func(err error) error {
		return fmt.Errorf("fetchUser(%d): %w", 42, err)
	}

	wrapped := call.From(fetchUser).With(call.ErrMapper[int, string](annotate))
	_, err := wrapped(context.Background(), 42)
	fmt.Println(err)
	// Output: fetchUser(42): not found
}

func ExampleOnError() {
	// fetchUser simulates a user lookup that fails.
	fetchUser := func(_ context.Context, id int) (string, error) {
		return "", fmt.Errorf("not found")
	}

	// logError prints the error without changing the return value.
	logError := func(err error) {
		fmt.Printf("logged: %v\n", err)
	}

	observed := call.From(fetchUser).With(call.OnError[int, string](logError))
	_, err := observed(context.Background(), 1)
	fmt.Printf("returned: %v\n", err)
	// Output:
	// logged: not found
	// returned: not found
}

func ExampleFunc_With() {
	// fetchData simulates a remote call.
	fetchData := func(_ context.Context, key string) (string, error) {
		return fmt.Sprintf("data(%s)", key), nil
	}

	// Compose: retry transient errors, then limit concurrency.
	// Innermost decorator (Retrier) runs first per call.
	resilient := call.From(fetchData).With(
		call.Retrier[string, string](3, call.ConstantBackoff(time.Millisecond), nil),
		call.Throttler[string, string](10),
	)

	got, _ := resilient(context.Background(), "abc")
	fmt.Println(got)
	// Output: data(abc)
}
