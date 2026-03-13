package slice

import "testing"

func TestIndexOf(t *testing.T) {
	t.Run("finds first occurrence", func(t *testing.T) {
		got, ok := IndexOf([]int{10, 20, 30}, 20).Get()
		if !ok || got != 1 {
			t.Errorf("IndexOf() = (%d, %v), want (1, true)", got, ok)
		}
	})

	t.Run("returns first when multiple", func(t *testing.T) {
		got, ok := IndexOf([]int{5, 3, 5, 3, 5}, 3).Get()
		if !ok || got != 1 {
			t.Errorf("IndexOf() = (%d, %v), want (1, true)", got, ok)
		}
	})

	t.Run("not found returns not-ok", func(t *testing.T) {
		_, ok := IndexOf([]int{1, 2, 3}, 99).Get()
		if ok {
			t.Error("IndexOf() = ok, want not-ok")
		}
	})

	t.Run("empty slice returns not-ok", func(t *testing.T) {
		_, ok := IndexOf([]int{}, 1).Get()
		if ok {
			t.Error("IndexOf() = ok, want not-ok")
		}
	})

	t.Run("nil slice returns not-ok", func(t *testing.T) {
		_, ok := IndexOf[int](nil, 1).Get()
		if ok {
			t.Error("IndexOf() = ok, want not-ok")
		}
	})
}

func TestLastIndexOf(t *testing.T) {
	t.Run("finds last occurrence", func(t *testing.T) {
		got, ok := LastIndexOf([]int{10, 20, 30}, 20).Get()
		if !ok || got != 1 {
			t.Errorf("LastIndexOf() = (%d, %v), want (1, true)", got, ok)
		}
	})

	t.Run("returns last when multiple", func(t *testing.T) {
		got, ok := LastIndexOf([]int{5, 3, 5, 3, 5}, 3).Get()
		if !ok || got != 3 {
			t.Errorf("LastIndexOf() = (%d, %v), want (3, true)", got, ok)
		}
	})

	t.Run("not found returns not-ok", func(t *testing.T) {
		_, ok := LastIndexOf([]int{1, 2, 3}, 99).Get()
		if ok {
			t.Error("LastIndexOf() = ok, want not-ok")
		}
	})

	t.Run("empty slice returns not-ok", func(t *testing.T) {
		_, ok := LastIndexOf([]int{}, 1).Get()
		if ok {
			t.Error("LastIndexOf() = ok, want not-ok")
		}
	})

	t.Run("nil slice returns not-ok", func(t *testing.T) {
		_, ok := LastIndexOf[int](nil, 1).Get()
		if ok {
			t.Error("LastIndexOf() = ok, want not-ok")
		}
	})
}
