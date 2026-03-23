package slice

import (
	"errors"
	"reflect"
	"testing"
)

func TestFold(t *testing.T) {
	t.Run("sum integers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		sum := func(acc, x int) int { return acc + x }
		got := Fold(input, 0, sum)
		want := 15
		if got != want {
			t.Errorf("Fold() = %v, want %v", got, want)
		}
	})

	t.Run("sum floats", func(t *testing.T) {
		input := []float64{1.5, 2.5, 3.0}
		sum := func(acc, x float64) float64 { return acc + x }
		got := Fold(input, 0.0, sum)
		want := 7.0
		if got != want {
			t.Errorf("Fold() = %v, want %v", got, want)
		}
	})

	t.Run("empty slice returns initial", func(t *testing.T) {
		input := []int{}
		sum := func(acc, x int) int { return acc + x }
		got := Fold(input, 42, sum)
		want := 42
		if got != want {
			t.Errorf("Fold() = %v, want %v", got, want)
		}
	})

	t.Run("build map from slice", func(t *testing.T) {
		type device struct {
			MAC  string
			Name string
		}
		devices := []device{
			{MAC: "aa:bb", Name: "phone"},
			{MAC: "cc:dd", Name: "laptop"},
		}
		toMap := func(m map[string]device, d device) map[string]device {
			m[d.MAC] = d
			return m
		}
		got := Fold(devices, make(map[string]device), toMap)
		want := map[string]device{
			"aa:bb": {MAC: "aa:bb", Name: "phone"},
			"cc:dd": {MAC: "cc:dd", Name: "laptop"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Fold() = %v, want %v", got, want)
		}
	})

	t.Run("concatenate strings", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		concat := func(acc, s string) string { return acc + s }
		got := Fold(input, "", concat)
		want := "abc"
		if got != want {
			t.Errorf("Fold() = %v, want %v", got, want)
		}
	})

	t.Run("find max", func(t *testing.T) {
		input := []int{3, 1, 4, 1, 5, 9, 2, 6}
		max := func(acc, x int) int {
			if x > acc {
				return x
			}
			return acc
		}
		got := Fold(input, input[0], max)
		want := 9
		if got != want {
			t.Errorf("Fold() = %v, want %v", got, want)
		}
	})
}

func TestTryFold(t *testing.T) {
	errBoom := errors.New("boom")

	t.Run("empty slice", func(t *testing.T) {
		got, err := TryFold([]int{}, 42, func(acc, x int) (int, error) {
			return acc + x, nil
		})
		if err != nil || got != 42 {
			t.Errorf("TryFold(empty) = (%d, %v), want (42, nil)", got, err)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		got, err := TryFold(nil, 42, func(acc, x int) (int, error) {
			return acc + x, nil
		})
		if err != nil || got != 42 {
			t.Errorf("TryFold(nil) = (%d, %v), want (42, nil)", got, err)
		}
	})

	t.Run("successful fold", func(t *testing.T) {
		got, err := TryFold([]int{1, 2, 3, 4, 5}, 0, func(acc, x int) (int, error) {
			return acc + x, nil
		})
		if err != nil || got != 15 {
			t.Errorf("TryFold = (%d, %v), want (15, nil)", got, err)
		}
	})

	t.Run("stops on first error", func(t *testing.T) {
		var visited int
		got, err := TryFold([]int{1, 2, 3, 4, 5}, 0, func(acc, x int) (int, error) {
			visited++
			if x == 3 {
				return acc + x, errBoom
			}
			return acc + x, nil
		})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want errBoom", err)
		}
		// acc = 0+1+2+3 = 6 (fn returns acc+x even on error)
		if got != 6 {
			t.Errorf("acc = %d, want 6 (sum through failing element)", got)
		}
		if visited != 3 {
			t.Errorf("visited = %d, want 3 (stopped after error)", visited)
		}
	})

	t.Run("returns acc from failing call not previous", func(t *testing.T) {
		// fn returns a different acc on error than on success.
		got, err := TryFold([]int{1, 2, 3}, 0, func(acc, x int) (int, error) {
			if x == 2 {
				return 999, errBoom // returns 999, not acc+x
			}
			return acc + x, nil
		})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want errBoom", err)
		}
		if got != 999 {
			t.Errorf("acc = %d, want 999 (from failing fn call)", got)
		}
	})

	t.Run("single element success", func(t *testing.T) {
		got, err := TryFold([]int{7}, 0, func(acc, x int) (int, error) {
			return acc + x, nil
		})
		if err != nil || got != 7 {
			t.Errorf("TryFold = (%d, %v), want (7, nil)", got, err)
		}
	})

	t.Run("single element failure", func(t *testing.T) {
		got, err := TryFold([]int{7}, 0, func(acc, x int) (int, error) {
			return acc, errBoom
		})
		if !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want errBoom", err)
		}
		if got != 0 {
			t.Errorf("acc = %d, want 0 (initial, fn returned it unchanged)", got)
		}
	})

	t.Run("left to right order", func(t *testing.T) {
		got, err := TryFold([]string{"a", "b", "c"}, "", func(acc, s string) (string, error) {
			return acc + s, nil
		})
		if err != nil || got != "abc" {
			t.Errorf("TryFold = (%q, %v), want (abc, nil)", got, err)
		}
	})
}
