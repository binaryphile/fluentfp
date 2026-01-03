package slice

import (
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
