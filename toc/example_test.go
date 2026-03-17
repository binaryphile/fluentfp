package toc_test

import (
	"context"
	"fmt"

	"github.com/binaryphile/fluentfp/toc"
)

func Example() {
	// double concatenates a string with itself.
	double := func(_ context.Context, s string) (string, error) {
		return s + s, nil
	}

	ctx := context.Background()

	// Start's ctx governs stage lifetime; Submit's ctx bounds only admission.
	stage := toc.Start(ctx, double, toc.Options[string]{Capacity: 3, Workers: 1})

	go func() {
		defer stage.CloseInput()

		for _, item := range []string{"a", "b", "c"} {
			if err := stage.Submit(ctx, item); err != nil {
				break
			}
		}
	}()

	for result := range stage.Out() {
		val, err := result.Unpack()
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(val)
	}

	if err := stage.Wait(); err != nil {
		fmt.Println("stage error:", err)
	}

	// Output:
	// aa
	// bb
	// cc
}
