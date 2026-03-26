package pipeline_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/binaryphile/fluentfp/pipeline"
)

func TestFilter(t *testing.T) {
	ctx := context.Background()

	// isEven reports whether n is divisible by 2.
	isEven := func(n int) bool { return n%2 == 0 }

	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5})
	out := pipeline.Filter(ctx, in, isEven)

	var got []int

	for v := range out {
		got = append(got, v)
	}

	want := []int{2, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFilter_empty(t *testing.T) {
	ctx := context.Background()
	isEven := func(n int) bool { return n%2 == 0 }

	in := pipeline.FromSlice(ctx, []int{})
	out := pipeline.Filter(ctx, in, isEven)

	count := 0

	for range out {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestFilter_noneMatch(t *testing.T) {
	ctx := context.Background()
	// alwaysFalse rejects everything.
	alwaysFalse := func(int) bool { return false }

	in := pipeline.FromSlice(ctx, []int{1, 2, 3})
	out := pipeline.Filter(ctx, in, alwaysFalse)

	count := 0

	for range out {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}
