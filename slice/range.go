package slice

import "math"

// minInt is math.MinInt — the most negative int value.
// -minInt overflows in two's complement, which breaks ceiling division.
const minInt = math.MinInt

// Range returns [0, 1, ..., end-1].
// Returns empty for end <= 0.
func Range(end int) Int {
	if end <= 0 {
		return Int{}
	}

	result := make([]int, end)
	for i := range end {
		result[i] = i
	}

	return result
}

// RangeFrom returns [start, start+1, ..., end-1].
// Returns empty for start >= end.
// Panics if the range is too large to allocate.
func RangeFrom(start, end int) Int {
	if start >= end {
		return Int{}
	}

	n := int(uint(end) - uint(start))
	if n <= 0 {
		panic("slice.RangeFrom: range too large")
	}

	result := make([]int, n)
	for i := range result {
		result[i] = start + i
	}

	return result
}

// RangeStep generates values starting at start, incrementing by step, stopping
// before reaching end (half-open interval). Panics if step is zero or math.MinInt
// (negation overflow). Returns empty when direction mismatches step sign.
func RangeStep(start, end, step int) Int {
	if step == 0 {
		panic("slice.RangeStep: step must not be zero")
	}

	// -step overflows for MinInt; guard before count formula uses it
	if step == minInt {
		panic("slice.RangeStep: step must not be math.MinInt (negation overflow)")
	}

	if step > 0 && start >= end {
		return Int{}
	}

	if step < 0 && start <= end {
		return Int{}
	}

	// Overflow-safe count using unsigned arithmetic.
	var count int
	if step > 0 {
		diff := uint(end) - uint(start)
		ustep := uint(step)
		ucount := (diff-1)/ustep + 1
		count = int(ucount)
		if count <= 0 {
			panic("slice.RangeStep: range too large")
		}
	} else {
		diff := uint(start) - uint(end)
		ustep := uint(-step) // safe: step != minInt checked above
		ucount := (diff-1)/ustep + 1
		count = int(ucount)
		if count <= 0 {
			panic("slice.RangeStep: range too large")
		}
	}

	result := make([]int, count)
	v := start
	for i := range result {
		result[i] = v
		v += step
	}

	return result
}
