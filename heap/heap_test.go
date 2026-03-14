package heap

import (
	"cmp"
	"slices"
	"testing"
)

// intAsc compares two ints in ascending order.
var intAsc = func(a, b int) int { return cmp.Compare(a, b) }

// intDesc compares two ints in descending order.
var intDesc = func(a, b int) int { return cmp.Compare(b, a) }

func TestCollect(t *testing.T) {
	tests := []struct {
		name   string
		insert []int
		cmp    func(int, int) int
		want   []int
	}{
		{"empty", nil, intAsc, nil},
		{"single", []int{5}, intAsc, []int{5}},
		{"ascending input", []int{1, 2, 3, 4, 5}, intAsc, []int{1, 2, 3, 4, 5}},
		{"descending input", []int{5, 4, 3, 2, 1}, intAsc, []int{1, 2, 3, 4, 5}},
		{"random order", []int{3, 1, 4, 1, 5, 9, 2, 6}, intAsc, []int{1, 1, 2, 3, 4, 5, 6, 9}},
		{"duplicates", []int{3, 3, 3, 1, 1}, intAsc, []int{1, 1, 3, 3, 3}},
		{"all equal", []int{7, 7, 7}, intAsc, []int{7, 7, 7}},
		{"desc comparator", []int{3, 1, 4, 1, 5}, intDesc, []int{5, 4, 3, 1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := From(tt.insert, tt.cmp)
			got := h.Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}
}

func TestDeleteMin(t *testing.T) {
	t.Run("repeated pop gives sorted output", func(t *testing.T) {
		h := From([]int{5, 3, 8, 1, 9, 2, 7}, intAsc)
		var got []int
		for !h.IsEmpty() {
			v, rest, ok := h.Pop()
			if !ok {
				t.Fatal("Pop returned false on non-empty heap")
			}

			got = append(got, v)
			h = rest
		}

		want := []int{1, 2, 3, 5, 7, 8, 9}
		assertSliceEqual(t, want, got)
	})

	t.Run("on empty returns empty", func(t *testing.T) {
		h := New[int](intAsc)
		h2 := h.DeleteMin()
		if !h2.IsEmpty() {
			t.Error("DeleteMin on empty should return empty")
		}

		if h2.cmp == nil {
			t.Error("DeleteMin should preserve comparator")
		}
	})

	t.Run("singleton preserves comparator", func(t *testing.T) {
		h := New[int](intAsc).Insert(42)
		h2 := h.DeleteMin()

		if !h2.IsEmpty() {
			t.Error("DeleteMin on singleton should return empty")
		}

		// comparator preserved — Insert should work
		h3 := h2.Insert(7)
		if h3.Len() != 1 {
			t.Fatalf("Insert after singleton DeleteMin: Len = %d; want 1", h3.Len())
		}

		v, ok := h3.Min().Get()
		if !ok || v != 7 {
			t.Errorf("Min after Insert = %d, %v; want 7, true", v, ok)
		}
	})
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"both non-empty", []int{1, 5, 9}, []int{2, 4, 6}, []int{1, 2, 4, 5, 6, 9}},
		{"left empty", nil, []int{3, 1, 2}, []int{1, 2, 3}},
		{"right empty", []int{3, 1, 2}, nil, []int{1, 2, 3}},
		{"both empty", nil, nil, nil},
		{"interleaved", []int{1, 3, 5}, []int{2, 4, 6}, []int{1, 2, 3, 4, 5, 6}},
		{"overlapping", []int{1, 2, 3}, []int{2, 3, 4}, []int{1, 2, 2, 3, 3, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := From(tt.a, intAsc)
			hb := From(tt.b, intAsc)
			merged := ha.Merge(hb)
			got := merged.Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}
}

func TestPersistence(t *testing.T) {
	t.Run("insert does not modify original", func(t *testing.T) {
		h1 := From([]int{3, 1, 5}, intAsc)
		h2 := h1.Insert(0)

		// h1 should still have min=1
		if v, ok := h1.Min().Get(); !ok || v != 1 {
			t.Errorf("h1.Min() = %d, %v; want 1, true", v, ok)
		}

		// h2 should have min=0
		if v, ok := h2.Min().Get(); !ok || v != 0 {
			t.Errorf("h2.Min() = %d, %v; want 0, true", v, ok)
		}

		// h1 size unchanged
		if h1.Len() != 3 {
			t.Errorf("h1.Len() = %d; want 3", h1.Len())
		}

		// h2 size incremented
		if h2.Len() != 4 {
			t.Errorf("h2.Len() = %d; want 4", h2.Len())
		}
	})

	t.Run("delete-min does not modify original", func(t *testing.T) {
		h1 := From([]int{1, 2, 3}, intAsc)
		h2 := h1.DeleteMin()

		// h1 still has min=1
		if v, ok := h1.Min().Get(); !ok || v != 1 {
			t.Errorf("h1.Min() = %d, %v; want 1, true", v, ok)
		}

		// h2 has min=2
		if v, ok := h2.Min().Get(); !ok || v != 2 {
			t.Errorf("h2.Min() = %d, %v; want 2, true", v, ok)
		}

		// Both can independently collect
		assertSliceEqual(t, []int{1, 2, 3}, h1.Collect())
		assertSliceEqual(t, []int{2, 3}, h2.Collect())
	})

	t.Run("merge does not modify originals", func(t *testing.T) {
		h1 := From([]int{1, 3}, intAsc)
		h2 := From([]int{2, 4}, intAsc)
		h3 := h1.Merge(h2)

		assertSliceEqual(t, []int{1, 3}, h1.Collect())
		assertSliceEqual(t, []int{2, 4}, h2.Collect())
		assertSliceEqual(t, []int{1, 2, 3, 4}, h3.Collect())
	})
}

func TestPop(t *testing.T) {
	t.Run("returns element and rest", func(t *testing.T) {
		h := From([]int{3, 1, 2}, intAsc)
		v, rest, ok := h.Pop()
		if !ok {
			t.Fatal("Pop returned false on non-empty heap")
		}

		if v != 1 {
			t.Errorf("Pop value = %d; want 1", v)
		}

		assertSliceEqual(t, []int{2, 3}, rest.Collect())
	})

	t.Run("on empty returns false", func(t *testing.T) {
		h := New[int](intAsc)
		v, rest, ok := h.Pop()
		if ok {
			t.Error("Pop should return false on empty heap")
		}

		if v != 0 {
			t.Errorf("Pop value = %d; want 0", v)
		}

		if !rest.IsEmpty() {
			t.Error("Pop rest should be empty")
		}
	})

	t.Run("empty pop preserves comparator", func(t *testing.T) {
		h := New[int](intAsc)
		_, rest, _ := h.Pop()

		// rest should still support Insert because comparator is preserved
		rest = rest.Insert(42)
		if rest.Len() != 1 {
			t.Fatalf("Insert after empty Pop: Len = %d; want 1", rest.Len())
		}

		v, ok := rest.Min().Get()
		if !ok || v != 42 {
			t.Errorf("Min after Insert = %d, %v; want 42, true", v, ok)
		}
	})

	t.Run("singleton pop preserves comparator", func(t *testing.T) {
		h := New[int](intAsc).Insert(42)
		v, rest, ok := h.Pop()
		if !ok || v != 42 {
			t.Fatalf("Pop() = %d, %v; want 42, true", v, ok)
		}

		// rest is now empty but should preserve comparator
		rest = rest.Insert(7)
		if rest.Len() != 1 {
			t.Fatalf("Insert after singleton Pop: Len = %d; want 1", rest.Len())
		}

		got, ok := rest.Min().Get()
		if !ok || got != 7 {
			t.Errorf("Min after Insert = %d, %v; want 7, true", got, ok)
		}
	})
}

func TestMin(t *testing.T) {
	t.Run("on non-empty", func(t *testing.T) {
		h := From([]int{5, 3, 8, 1}, intAsc)
		v, ok := h.Min().Get()
		if !ok || v != 1 {
			t.Errorf("Min() = %d, %v; want 1, true", v, ok)
		}
	})

	t.Run("on empty", func(t *testing.T) {
		h := New[int](intAsc)
		if h.Min().IsOk() {
			t.Error("Min on empty should be not-ok")
		}
	})

	t.Run("on zero value", func(t *testing.T) {
		var h Heap[int]
		if h.Min().IsOk() {
			t.Error("Min on zero value should be not-ok")
		}
	})
}

func TestFrom(t *testing.T) {
	t.Run("matches sequential inserts", func(t *testing.T) {
		items := []int{5, 3, 8, 1, 9, 2, 7, 4, 6}

		h1 := From(items, intAsc)

		h2 := New[int](intAsc)
		for _, v := range items {
			h2 = h2.Insert(v)
		}

		assertSliceEqual(t, h2.Collect(), h1.Collect())
	})
}

func TestLargeHeap(t *testing.T) {
	const n = 10_000
	items := make([]int, n)
	for i := range items {
		items[i] = n - i // descending: n, n-1, ..., 1
	}

	h := From(items, intAsc)

	if h.Len() != n {
		t.Fatalf("Len() = %d; want %d", h.Len(), n)
	}

	got := h.Collect()
	want := make([]int, n)
	for i := range want {
		want[i] = i + 1
	}

	assertSliceEqual(t, want, got)
}

func TestPanics(t *testing.T) {
	t.Run("New with nil comparator", func(t *testing.T) {
		assertPanics(t, func() { New[int](nil) })
	})

	t.Run("From with nil comparator", func(t *testing.T) {
		assertPanics(t, func() { From([]int{1}, nil) })
	})

	t.Run("zero value Insert", func(t *testing.T) {
		var h Heap[int]
		assertPanics(t, func() { h.Insert(1) })
	})

	t.Run("zero value DeleteMin", func(t *testing.T) {
		var h Heap[int]
		assertPanics(t, func() { h.DeleteMin() })
	})

	t.Run("zero value Merge", func(t *testing.T) {
		var h Heap[int]
		assertPanics(t, func() { h.Merge(h) })
	})
}

func TestZeroValue(t *testing.T) {
	var h Heap[int]

	if !h.IsEmpty() {
		t.Error("zero value should be empty")
	}

	if h.Len() != 0 {
		t.Errorf("zero value Len() = %d; want 0", h.Len())
	}

	if h.Min().IsOk() {
		t.Error("zero value Min() should be not-ok")
	}

	v, rest, ok := h.Pop()
	if ok {
		t.Error("zero value Pop should return false")
	}

	if v != 0 {
		t.Errorf("zero value Pop value = %d; want 0", v)
	}

	if !rest.IsEmpty() {
		t.Error("zero value Pop rest should be empty")
	}

	if h.Collect() != nil {
		t.Error("zero value Collect should return nil")
	}
}

// helpers

func assertSliceEqual[T cmp.Ordered](t *testing.T, want, got []T) {
	t.Helper()

	if !slices.Equal(want, got) {
		t.Errorf("got %v; want %v", got, want)
	}
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()

	fn()
}
