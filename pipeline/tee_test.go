package pipeline_test

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/binaryphile/fluentfp/pipeline"
)

func TestTee(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3})
	outs := pipeline.Tee(ctx, in, 3)

	if len(outs) != 3 {
		t.Fatalf("expected 3 outputs, got %d", len(outs))
	}

	var wg sync.WaitGroup
	results := make([][]int, 3)

	for i, out := range outs {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for v := range out {
				results[i] = append(results[i], v)
			}
		}()
	}

	wg.Wait()

	want := []int{1, 2, 3}

	for i, got := range results {
		if !reflect.DeepEqual(got, want) {
			t.Errorf("consumer %d: got %v, want %v", i, got, want)
		}
	}
}

func TestTee_panicsOnInvalidN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for n <= 0")
		}
	}()

	pipeline.Tee(context.Background(), make(<-chan int), 0)
}
