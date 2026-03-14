package seq

import (
	"reflect"
	"testing"
)

func TestChunk(t *testing.T) {
	t.Run("exact multiple", func(t *testing.T) {
		got := Chunk(Of(1, 2, 3, 4), 2).Collect()
		want := [][]int{{1, 2}, {3, 4}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("remainder chunk", func(t *testing.T) {
		got := Chunk(Of(1, 2, 3, 4, 5), 2).Collect()
		want := [][]int{{1, 2}, {3, 4}, {5}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := Chunk(Of(1), 3).Collect()
		want := [][]int{{1}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("size greater than len", func(t *testing.T) {
		got := Chunk(Of(1, 2), 5).Collect()
		want := [][]int{{1, 2}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := Chunk(From([]int{}), 2).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("nil seq", func(t *testing.T) {
		got := Chunk(Seq[int](nil), 2).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("snapshot stability", func(t *testing.T) {
		chunks := Chunk(Of(1, 2, 3, 4), 2).Collect()

		// Mutating chunk 0 should not affect chunk 1.
		chunks[0][0] = 99

		if chunks[1][0] != 3 || chunks[1][1] != 4 {
			t.Errorf("mutation leaked: chunk[1] = %v, want [3, 4]", chunks[1])
		}
	})

	t.Run("repeated iteration", func(t *testing.T) {
		s := Chunk(Of(1, 2, 3), 2)
		first := s.Collect()
		second := s.Collect()

		want := [][]int{{1, 2}, {3}}
		if !reflect.DeepEqual(first, want) {
			t.Errorf("first = %v, want %v", first, want)
		}
		if !reflect.DeepEqual(second, want) {
			t.Errorf("second = %v, want %v", second, want)
		}
	})

	t.Run("size 1", func(t *testing.T) {
		got := Chunk(Of(1, 2, 3), 1).Collect()
		want := [][]int{{1}, {2}, {3}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("early termination exact consumption", func(t *testing.T) {
		var calls int

		source := Seq[int](func(yield func(int) bool) {
			for i := 1; i <= 10; i++ {
				calls++
				if !yield(i) {
					return
				}
			}
		})

		got := Chunk(source, 3).Take(2).Collect()
		want := [][]int{{1, 2, 3}, {4, 5, 6}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}

		// Take(2) chunks of size 3 should consume exactly 6 elements.
		if calls != 6 {
			t.Errorf("source consumed %d elements, want exactly 6", calls)
		}
	})

	t.Run("infinite stream", func(t *testing.T) {
		// inc increments by 1.
		inc := func(n int) int { return n + 1 }

		got := Chunk(Generate(1, inc), 3).Take(2).Collect()
		want := [][]int{{1, 2, 3}, {4, 5, 6}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestChunkPanicsOnZeroSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	Chunk(Seq[int](nil), 0)
}

func TestChunkPanicsOnNegativeSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	Chunk(Seq[int](nil), -1)
}
