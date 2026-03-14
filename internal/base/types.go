package base

import (
	"math"
	"strings"

	"github.com/binaryphile/fluentfp/option"
)

type Float64 []float64

// Sum returns the sum of all elements.
func (fs Float64) Sum() float64 {
	var sum float64
	for _, f := range fs {
		sum += f
	}
	return sum
}

// Max returns the largest non-NaN element, or not-ok if the slice is empty
// or contains only NaN values. NaN elements are skipped.
// Signed zeros: -0.0 and +0.0 are equal per IEEE 754; the result depends on
// input order when both are present. Use [math.Copysign] if sign matters.
func (fs Float64) Max() option.Option[float64] {
	var m float64
	found := false
	for _, v := range fs {
		if math.IsNaN(v) {
			continue
		}
		if !found || v > m {
			m = v
			found = true
		}
	}
	if !found {
		return option.NotOk[float64]()
	}
	return option.Of(m)
}

// Min returns the smallest non-NaN element, or not-ok if the slice is empty
// or contains only NaN values. NaN elements are skipped.
// Signed zeros: -0.0 and +0.0 are equal per IEEE 754; the result depends on
// input order when both are present. Use [math.Copysign] if sign matters.
func (fs Float64) Min() option.Option[float64] {
	var m float64
	found := false
	for _, v := range fs {
		if math.IsNaN(v) {
			continue
		}
		if !found || v < m {
			m = v
			found = true
		}
	}
	if !found {
		return option.NotOk[float64]()
	}
	return option.Of(m)
}

type Int []int

// Sum returns the sum of all elements.
func (is Int) Sum() int {
	var sum int
	for _, v := range is {
		sum += v
	}
	return sum
}

// Max returns the largest element, or not-ok if the slice is empty.
func (is Int) Max() option.Option[int] {
	if len(is) == 0 {
		return option.NotOk[int]()
	}
	m := is[0]
	for _, v := range is[1:] {
		if v > m {
			m = v
		}
	}
	return option.Of(m)
}

// Min returns the smallest element, or not-ok if the slice is empty.
func (is Int) Min() option.Option[int] {
	if len(is) == 0 {
		return option.NotOk[int]()
	}
	m := is[0]
	for _, v := range is[1:] {
		if v < m {
			m = v
		}
	}
	return option.Of(m)
}

type String []string

// Unique returns a new slice with duplicate strings removed, preserving order.
func (ss String) Unique() String {
	seen := make(map[string]bool)
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// Contains returns true if ss contains target.
func (ss String) Contains(target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}

// ContainsAny returns true if ss contains any element in targets.
// Returns false if either slice is empty.
func (ss String) ContainsAny(targets []string) bool {
	if len(targets) == 0 {
		return false
	}
	set := String(targets).ToSet()
	for _, s := range ss {
		if set[s] {
			return true
		}
	}
	return false
}

// Each calls fn for every element.
func (ss String) Each(fn func(string)) {
	for _, s := range ss {
		fn(s)
	}
}

// Len returns the length of the slice.
func (ss String) Len() int {
	return len(ss)
}

// ToSet returns a map with each string as a key set to true.
func (ss String) ToSet() map[string]bool {
	return ToSet([]string(ss))
}

// Join concatenates the elements with sep between each.
func (ss String) Join(sep string) string {
	return strings.Join(ss, sep)
}

// NonEmpty removes empty strings.
func (ss String) NonEmpty() String {
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
