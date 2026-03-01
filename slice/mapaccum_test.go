package slice

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMapAccum(t *testing.T) {
	t.Run("empty slice returns init and empty slice", func(t *testing.T) {
		add := func(acc int, x int) (int, int) { return acc + x, x }
		state, result := MapAccum([]int{}, 0, add)
		if state != 0 {
			t.Errorf("state = %d, want 0", state)
		}
		if len(result) != 0 {
			t.Errorf("result length = %d, want 0", len(result))
		}
	})

	t.Run("running deltas", func(t *testing.T) {
		// delta returns the current value as new state and the difference as output.
		delta := func(prev int, cur int) (int, int) { return cur, cur - prev }
		state, deltas := MapAccum([]int{10, 30, 35, 50}, 0, delta)
		if state != 50 {
			t.Errorf("state = %d, want 50", state)
		}
		want := []int{10, 20, 5, 15}
		if !reflect.DeepEqual([]int(deltas), want) {
			t.Errorf("deltas = %v, want %v", deltas, want)
		}
	})

	t.Run("running total with labels", func(t *testing.T) {
		// label returns the running sum as state and a formatted label as output.
		label := func(sum int, x int) (int, string) {
			sum += x
			return sum, fmt.Sprintf("cumulative: %d", sum)
		}
		state, labels := MapAccum([]int{5, 3, 7}, 0, label)
		if state != 15 {
			t.Errorf("state = %d, want 15", state)
		}
		want := []string{"cumulative: 5", "cumulative: 8", "cumulative: 15"}
		if !reflect.DeepEqual([]string(labels), want) {
			t.Errorf("labels = %v, want %v", labels, want)
		}
	})

	t.Run("single element", func(t *testing.T) {
		add := func(acc int, x int) (int, int) { return acc + x, x }
		state, result := MapAccum([]int{42}, 0, add)
		if state != 42 {
			t.Errorf("state = %d, want 42", state)
		}
		want := []int{42}
		if !reflect.DeepEqual([]int(result), want) {
			t.Errorf("result = %v, want %v", result, want)
		}
	})
}
