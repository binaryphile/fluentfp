package pipeline_test

import (
	"context"
	"sort"
	"testing"

	"github.com/binaryphile/fluentfp/pipeline"
)

func TestMerge(t *testing.T) {
	ctx := context.Background()
	a := pipeline.FromSlice(ctx, []int{1, 3, 5})
	b := pipeline.FromSlice(ctx, []int{2, 4, 6})

	out := pipeline.Merge(ctx, a, b)

	var got []int

	for v := range out {
		got = append(got, v)
	}

	sort.Ints(got)

	want := []int{1, 2, 3, 4, 5, 6}
	if len(got) != len(want) {
		t.Fatalf("got %d items, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestMerge_singleInput(t *testing.T) {
	ctx := context.Background()
	a := pipeline.FromSlice(ctx, []int{1, 2, 3})

	out := pipeline.Merge(ctx, a)

	var got []int

	for v := range out {
		got = append(got, v)
	}

	if len(got) != 3 {
		t.Errorf("expected 3 items, got %d", len(got))
	}
}

func TestMerge_noInputs(t *testing.T) {
	ctx := context.Background()
	out := pipeline.Merge[int](ctx)

	count := 0

	for range out {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}
