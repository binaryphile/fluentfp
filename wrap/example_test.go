package wrap_test

import (
	"context"
	"fmt"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

func ExampleFn_With_retry() {
	// double doubles the input. Succeeds on first try.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	resilient := wrap.Func(double).With(wrap.Features{
		Retry: wrap.Retry(3, wrap.ExpBackoff(time.Millisecond), nil),
	})

	got, _ := resilient(context.Background(), 5)
	fmt.Println(got)
	// Output: 10
}

func ExampleFn_With_breaker() {
	// double doubles the input.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	breaker := wrap.NewBreaker(wrap.BreakerConfig{
		ResetTimeout: 10 * time.Second,
	})

	protected := wrap.Func(double).With(wrap.Features{Breaker: breaker})
	got, _ := protected(context.Background(), 21)
	fmt.Println(got)
	// Output: 42
}

func ExampleFn_With_throttle() {
	// double doubles the input.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	limited := wrap.Func(double).With(wrap.Features{Throttle: wrap.Throttle(5)})

	got, _ := limited(context.Background(), 3)
	fmt.Println(got)
	// Output: 6
}

func ExampleFn_With_mapError() {
	// fetchUser simulates a user lookup that fails.
	fetchUser := func(_ context.Context, id int) (string, error) {
		return "", fmt.Errorf("not found")
	}

	// annotate wraps errors with calling context.
	annotate := func(err error) error {
		return fmt.Errorf("fetchUser(%d): %w", 42, err)
	}

	wrapped := wrap.Func(fetchUser).With(wrap.Features{MapError: annotate})
	_, err := wrapped(context.Background(), 42)
	fmt.Println(err)
	// Output: fetchUser(42): not found
}

func ExampleFn_With_onError() {
	// fetchUser simulates a user lookup that fails.
	fetchUser := func(_ context.Context, id int) (string, error) {
		return "", fmt.Errorf("not found")
	}

	// logError prints the error without changing the return value.
	logError := func(err error) {
		fmt.Printf("logged: %v\n", err)
	}

	observed := wrap.Func(fetchUser).With(wrap.Features{OnError: logError})
	_, err := observed(context.Background(), 1)
	fmt.Printf("returned: %v\n", err)
	// Output:
	// logged: not found
	// returned: not found
}

func ExampleFn_With_combined() {
	// fetchData simulates a remote call.
	fetchData := func(_ context.Context, key string) (string, error) {
		return fmt.Sprintf("data(%s)", key), nil
	}

	// All features in one struct. Library controls decorator order.
	resilient := wrap.Func(fetchData).With(wrap.Features{
		Retry:    wrap.Retry(3, wrap.ExpBackoff(time.Millisecond), nil),
		Throttle: wrap.Throttle(10),
	})

	got, _ := resilient(context.Background(), "abc")
	fmt.Println(got)
	// Output: data(abc)
}
