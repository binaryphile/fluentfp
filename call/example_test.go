package call_test

import (
	"context"
	"fmt"

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
