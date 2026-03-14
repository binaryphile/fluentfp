package seq

import (
	"strconv"
	"testing"
)

func TestFilterMap(t *testing.T) {
	// parsePositive returns (parsed int, true) for positive integers, else (0, false).
	parsePositive := func(s string) (int, bool) {
		n, err := strconv.Atoi(s)
		if err != nil || n <= 0 {
			return 0, false
		}
		return n, true
	}

	tests := []struct {
		name string
		in   []string
		want []int
	}{
		{"all kept", []string{"1", "2", "3"}, []int{1, 2, 3}},
		{"all discarded", []string{"0", "-1", "abc"}, nil},
		{"mixed preserves order", []string{"abc", "2", "-1", "4", "0"}, []int{2, 4}},
		{"empty", []string{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterMap(From(tt.in), parsePositive).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("nil seq", func(t *testing.T) {
		got := FilterMap(Seq[string](nil), parsePositive).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("repeated iteration", func(t *testing.T) {
		s := FilterMap(Of("1", "abc", "3"), parsePositive)
		first := s.Collect()
		second := s.Collect()
		assertSliceEqual(t, first, []int{1, 3})
		assertSliceEqual(t, second, []int{1, 3})
	})

	t.Run("early termination", func(t *testing.T) {
		var calls int

		// countingParse tracks calls and parses positive ints.
		countingParse := func(s string) (int, bool) {
			calls++
			n, err := strconv.Atoi(s)
			if err != nil || n <= 0 {
				return 0, false
			}
			return n, true
		}

		got := FilterMap(Of("1", "2", "3", "4", "5"), countingParse).Take(2).Collect()
		assertSliceEqual(t, got, []int{1, 2})

		if calls > 2 {
			t.Errorf("fn called %d times, want at most 2", calls)
		}
	})

	t.Run("early termination with discards", func(t *testing.T) {
		var calls int

		// countingParse tracks calls and only keeps numeric strings.
		countingParse := func(s string) (int, bool) {
			calls++
			n, err := strconv.Atoi(s)
			if err != nil {
				return 0, false
			}
			return n, true
		}

		// Must discard "x" and "y" before finding enough kept values.
		got := FilterMap(Of("x", "1", "y", "2", "z"), countingParse).Take(2).Collect()
		assertSliceEqual(t, got, []int{1, 2})

		// Consumes x(discard), 1(keep), y(discard), 2(keep) = 4 calls.
		if calls != 4 {
			t.Errorf("fn called %d times, want exactly 4", calls)
		}
	})

	t.Run("callback not called on empty", func(t *testing.T) {
		var calls int

		// counting tracks calls.
		counting := func(s string) (int, bool) {
			calls++
			return 0, false
		}

		FilterMap(From([]string{}), counting).Collect()

		if calls != 0 {
			t.Errorf("fn called %d times on empty, want 0", calls)
		}
	})

	t.Run("infinite stream", func(t *testing.T) {
		// keepEven returns (v, true) when v is even.
		keepEven := func(v int) (int, bool) {
			if v%2 == 0 {
				return v, true
			}
			return 0, false
		}

		got := FilterMap(Generate(0, func(n int) int { return n + 1 }), keepEven).Take(3).Collect()
		assertSliceEqual(t, got, []int{0, 2, 4})
	})
}

func TestFilterMapNilFnPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	FilterMap[string, int](nil, nil)
}
