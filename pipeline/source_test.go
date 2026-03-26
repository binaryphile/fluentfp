package pipeline_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/binaryphile/fluentfp/pipeline"
)

func TestFromSlice(t *testing.T) {
	ctx := context.Background()
	out := pipeline.FromSlice(ctx, []int{1, 2, 3})

	var got []int

	for v := range out {
		got = append(got, v)
	}

	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGenerate(t *testing.T) {
	ctx := context.Background()
	n := 0

	// counter generates sequential integers up to 5.
	counter := func() (int, bool) {
		n++

		if n > 5 {
			return 0, false
		}

		return n, true
	}

	out := pipeline.Generate(ctx, counter)

	var got []int

	for v := range out {
		got = append(got, v)
	}

	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
