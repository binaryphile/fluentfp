package slice

// Chunk splits ts into sub-slices of at most size elements.
// The last chunk may have fewer than size elements.
//
// Each chunk is a sub-slice of ts, so element mutations through a chunk are
// visible through ts (and vice versa). Capacity is clipped to chunk length,
// so appending to a chunk allocates a new backing array rather than corrupting
// adjacent chunks or the source.
//
// Panics if size <= 0.
func Chunk[T any](ts []T, size int) [][]T {
	if size <= 0 {
		panic("slice.Chunk: size must be positive")
	}
	if len(ts) == 0 {
		return [][]T{}
	}
	chunks := make([][]T, 0, (len(ts)+size-1)/size)
	for i := 0; i < len(ts); i += size {
		end := i + size
		if end > len(ts) {
			end = len(ts)
		}
		chunks = append(chunks, ts[i:end:end])
	}
	return chunks
}
