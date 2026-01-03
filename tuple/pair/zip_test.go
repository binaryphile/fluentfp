package pair

import (
	"reflect"
	"testing"
)

func TestZipWith(t *testing.T) {
	add := func(a, b int) int { return a + b }

	t.Run("transform pairs", func(t *testing.T) {
		as := []int{1, 2, 3}
		bs := []int{10, 20, 30}
		got := ZipWith(as, bs, add)
		want := []int{11, 22, 33}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ZipWith() = %v, want %v", got, want)
		}
	})

	t.Run("empty slices", func(t *testing.T) {
		as := []int{}
		bs := []int{}
		got := ZipWith(as, bs, add)
		if len(got) != 0 {
			t.Errorf("ZipWith() = %v, want empty slice", got)
		}
	})

	t.Run("unequal lengths panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("ZipWith() did not panic on unequal lengths")
			}
		}()
		as := []int{1, 2}
		bs := []int{10}
		ZipWith(as, bs, add)
	})
}

func TestZip(t *testing.T) {
	t.Run("equal length slices", func(t *testing.T) {
		as := []int{1, 2, 3}
		bs := []string{"a", "b", "c"}
		got := Zip(as, bs)
		want := []X[int, string]{
			{V1: 1, V2: "a"},
			{V1: 2, V2: "b"},
			{V1: 3, V2: "c"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Zip() = %v, want %v", got, want)
		}
	})

	t.Run("empty slices", func(t *testing.T) {
		as := []int{}
		bs := []string{}
		got := Zip(as, bs)
		if len(got) != 0 {
			t.Errorf("Zip() = %v, want empty slice", got)
		}
	})

	t.Run("nil slices", func(t *testing.T) {
		var as []int
		var bs []string
		got := Zip(as, bs)
		if len(got) != 0 {
			t.Errorf("Zip() = %v, want empty slice", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		as := []int{42}
		bs := []string{"hello"}
		got := Zip(as, bs)
		want := []X[int, string]{{V1: 42, V2: "hello"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Zip() = %v, want %v", got, want)
		}
	})

	t.Run("unequal lengths panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Zip() did not panic on unequal lengths")
			}
		}()
		as := []int{1, 2}
		bs := []string{"a"}
		Zip(as, bs)
	})
}
