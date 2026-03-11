package base

import "strings"

type Float64 []float64

// Sum returns the sum of all elements.
func (fs Float64) Sum() float64 {
	var sum float64
	for _, f := range fs {
		sum += f
	}
	return sum
}

// Max returns the largest element, or zero if the slice is empty.
func (fs Float64) Max() float64 {
	if len(fs) == 0 {
		return 0
	}
	m := fs[0]
	for _, v := range fs[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// Min returns the smallest element, or zero if the slice is empty.
func (fs Float64) Min() float64 {
	if len(fs) == 0 {
		return 0
	}
	m := fs[0]
	for _, v := range fs[1:] {
		if v < m {
			m = v
		}
	}
	return m
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

// Max returns the largest element, or zero if the slice is empty.
func (is Int) Max() int {
	if len(is) == 0 {
		return 0
	}
	m := is[0]
	for _, v := range is[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// Min returns the smallest element, or zero if the slice is empty.
func (is Int) Min() int {
	if len(is) == 0 {
		return 0
	}
	m := is[0]
	for _, v := range is[1:] {
		if v < m {
			m = v
		}
	}
	return m
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

// Matches returns true if ss contains any element in filter.
// Returns true if filter is empty (no constraint).
func (ss String) Matches(filter []string) bool {
	return len(filter) == 0 || ss.ContainsAny(filter)
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
