package slice

// ParallelMap returns the result of applying fn to each member of ts, using the specified
// number of worker goroutines. Order is preserved. The fn must be safe for concurrent use.
func (ts MapperTo[R, T]) ParallelMap(workers int, fn func(T) R) Mapper[R] {
	if len(ts) == 0 {
		return Mapper[R]{}
	}
	results := make([]R, len(ts))
	forBatches(len(ts), workers, func(_, start, end int) {
		for j := start; j < end; j++ {
			results[j] = fn(ts[j])
		}
	})
	return results
}

// ParallelKeepIf returns a new slice containing members for which fn returns true,
// using the specified number of worker goroutines. Order is preserved.
func (ts MapperTo[R, T]) ParallelKeepIf(workers int, fn func(T) bool) MapperTo[R, T] {
	if len(ts) == 0 {
		return MapperTo[R, T]{}
	}
	batchResults := make([][]T, min(workers, len(ts)))
	forBatches(len(ts), workers, func(idx, start, end int) {
		var result []T
		for j := start; j < end; j++ {
			if fn(ts[j]) {
				result = append(result, ts[j])
			}
		}
		batchResults[idx] = result
	})
	total := 0
	for _, b := range batchResults {
		total += len(b)
	}
	results := make([]T, 0, total)
	for _, b := range batchResults {
		results = append(results, b...)
	}
	return results
}

// ParallelEach applies fn to each member of ts, using the specified number of worker
// goroutines. The fn must be safe for concurrent use.
func (ts MapperTo[R, T]) ParallelEach(workers int, fn func(T)) {
	forBatches(len(ts), workers, func(_, start, end int) {
		for j := start; j < end; j++ {
			fn(ts[j])
		}
	})
}
