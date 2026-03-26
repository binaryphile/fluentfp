package pipeline_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/binaryphile/fluentfp/pipeline"
)

func TestBatch_exactMultiple(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5, 6})
	out := pipeline.Batch(ctx, in, 3)

	var got [][]int

	for b := range out {
		got = append(got, b)
	}

	want := [][]int{{1, 2, 3}, {4, 5, 6}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBatch_partialFinal(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5})
	out := pipeline.Batch(ctx, in, 3)

	var got [][]int

	for b := range out {
		got = append(got, b)
	}

	want := [][]int{{1, 2, 3}, {4, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBatch_sizeOne(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3})
	out := pipeline.Batch(ctx, in, 1)

	var got [][]int

	for b := range out {
		got = append(got, b)
	}

	want := [][]int{{1}, {2}, {3}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBatch_empty(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{})
	out := pipeline.Batch(ctx, in, 3)

	count := 0

	for range out {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 batches, got %d", count)
	}
}

func TestBatch_panicsOnInvalidSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for size <= 0")
		}
	}()

	pipeline.Batch(context.Background(), make(<-chan int), 0)
}
