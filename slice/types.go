package slice

type Any = Mapper[any]
type Bool = Mapper[bool]
type Byte = Mapper[byte]
type Error = Mapper[error]
type Float32 = Mapper[float32]
type Float64 []float64

// Sum returns the sum of all elements.
func (fs Float64) Sum() float64 {
	var sum float64
	for _, f := range fs {
		sum += f
	}
	return sum
}

type Int = Mapper[int]
type Rune = Mapper[rune]
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
	set := make(map[string]bool, len(ss))
	for _, s := range ss {
		set[s] = true
	}
	return set
}
