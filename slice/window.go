package slice

// Window returns sliding windows of size elements.
// Each window is a sub-slice sharing the backing array of ts — overlapping
// windows alias the same memory. Mutating an element in one window affects
// all windows that include that position. Capacity is clipped to window size,
// so append on a window will not corrupt adjacent windows or the source.
// Use Clone() on individual windows if you need fully independent copies.
// Panics if size <= 0. Returns empty for empty/nil input or when len(ts) < size.
func Window[T any](ts []T, size int) [][]T {
	if size <= 0 {
		panic("slice.Window: size must be positive")
	}

	if len(ts) < size {
		return [][]T{}
	}

	count := len(ts) - size + 1
	result := make([][]T, count)
	for i := range count {
		result[i] = ts[i : i+size : i+size]
	}

	return result
}
