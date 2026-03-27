package wrap_test

import (
	"context"
	"fmt"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

func ExampleFn_WithRetry() {
	// double doubles the input. Succeeds on first try.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// Retry up to 3 times with exponential backoff.
	resilient := wrap.Func(double).WithRetry(3, wrap.ExpBackoff(time.Millisecond), nil)

	got, _ := resilient(context.Background(), 5)
	fmt.Println(got)
	// Output: 10
}

func ExampleFn_WithBreaker() {
	// double doubles the input.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// Protect a dependency with a circuit breaker.
	breaker := wrap.NewBreaker(wrap.BreakerConfig{
		ResetTimeout: 10 * time.Second,
	})

	protected := wrap.Func(double).WithBreaker(breaker)
	got, _ := protected(context.Background(), 21)
	fmt.Println(got)
	// Output: 42
}

func ExampleFn_WithThrottle() {
	// double doubles the input.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// Allow at most 5 concurrent calls.
	limited := wrap.Func(double).WithThrottle(5)

	got, _ := limited(context.Background(), 3)
	fmt.Println(got)
	// Output: 6
}

func ExampleFn_WithMapError() {
	// fetchUser simulates a user lookup that fails.
	fetchUser := func(_ context.Context, id int) (string, error) {
		return "", fmt.Errorf("not found")
	}

	// annotate wraps errors with calling context.
	annotate := func(err error) error {
		return fmt.Errorf("fetchUser(%d): %w", 42, err)
	}

	wrapped := wrap.Func(fetchUser).WithMapError(annotate)
	_, err := wrapped(context.Background(), 42)
	fmt.Println(err)
	// Output: fetchUser(42): not found
}

func ExampleFn_WithOnError() {
	// fetchUser simulates a user lookup that fails.
	fetchUser := func(_ context.Context, id int) (string, error) {
		return "", fmt.Errorf("not found")
	}

	// logError prints the error without changing the return value.
	logError := func(err error) {
		fmt.Printf("logged: %v\n", err)
	}

	observed := wrap.Func(fetchUser).WithOnError(logError)
	_, err := observed(context.Background(), 1)
	fmt.Printf("returned: %v\n", err)
	// Output:
	// logged: not found
	// returned: not found
}

func ExampleFn_chain() {
	// fetchData simulates a remote call.
	fetchData := func(_ context.Context, key string) (string, error) {
		return fmt.Sprintf("data(%s)", key), nil
	}

	// Compose: retry transient errors, then limit concurrency.
	resilient := wrap.Func(fetchData).
		WithRetry(3, wrap.ExpBackoff(time.Millisecond), nil).
		WithThrottle(10)

	got, _ := resilient(context.Background(), "abc")
	fmt.Println(got)
	// Output: data(abc)
}
