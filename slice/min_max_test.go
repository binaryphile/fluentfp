package slice

import "testing"

func TestMinBy(t *testing.T) {
	type item struct {
		name  string
		score int
	}

	score := func(i item) int { return i.score }

	t.Run("nil returns not-ok", func(t *testing.T) {
		if _, ok := MinBy[item](nil, score).Get(); ok {
			t.Error("expected not-ok for nil slice")
		}
	})

	t.Run("empty returns not-ok", func(t *testing.T) {
		if _, ok := MinBy([]item{}, score).Get(); ok {
			t.Error("expected not-ok for empty slice")
		}
	})

	t.Run("single element", func(t *testing.T) {
		got, ok := MinBy([]item{{"a", 5}}, score).Get()
		if !ok || got.name != "a" {
			t.Errorf("got (%v, %v), want (a, true)", got.name, ok)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		items := []item{{"a", 5}, {"b", 1}, {"c", 3}}
		got, ok := MinBy(items, score).Get()
		if !ok || got.name != "b" {
			t.Errorf("got (%v, %v), want (b, true)", got.name, ok)
		}
	})

	t.Run("ties return first", func(t *testing.T) {
		items := []item{{"a", 1}, {"b", 1}, {"c", 2}}
		got, ok := MinBy(items, score).Get()
		if !ok || got.name != "a" {
			t.Errorf("got (%v, %v), want (a, true)", got.name, ok)
		}
	})

	t.Run("key called once per element", func(t *testing.T) {
		calls := 0
		countingKey := func(i item) int {
			calls++
			return i.score
		}

		MinBy([]item{{"a", 1}, {"b", 2}, {"c", 3}}, countingKey)

		if calls != 3 {
			t.Errorf("key called %d times, want 3", calls)
		}
	})
}

func TestMaxBy(t *testing.T) {
	type item struct {
		name  string
		score int
	}

	score := func(i item) int { return i.score }

	t.Run("nil returns not-ok", func(t *testing.T) {
		if _, ok := MaxBy[item](nil, score).Get(); ok {
			t.Error("expected not-ok for nil slice")
		}
	})

	t.Run("empty returns not-ok", func(t *testing.T) {
		if _, ok := MaxBy([]item{}, score).Get(); ok {
			t.Error("expected not-ok for empty slice")
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		items := []item{{"a", 5}, {"b", 1}, {"c", 3}}
		got, ok := MaxBy(items, score).Get()
		if !ok || got.name != "a" {
			t.Errorf("got (%v, %v), want (a, true)", got.name, ok)
		}
	})

	t.Run("ties return first", func(t *testing.T) {
		items := []item{{"a", 2}, {"b", 5}, {"c", 5}}
		got, ok := MaxBy(items, score).Get()
		if !ok || got.name != "b" {
			t.Errorf("got (%v, %v), want (b, true)", got.name, ok)
		}
	})
}
