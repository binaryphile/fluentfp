package base

import "sync"

// forBatches divides n items among workers and calls fn(batchIdx, start, end) for each batch.
// Panics if workers <= 0. Clamps workers to n. No-ops if n == 0.
// If workers == 1, calls fn once synchronously (no goroutine).
func forBatches(n, workers int, fn func(batchIdx, start, end int)) {
	if workers <= 0 {
		panic("fluentfp: workers must be > 0")
	}
	if n == 0 {
		return
	}
	if workers > n {
		workers = n
	}
	if workers == 1 {
		fn(0, 0, n)
		return
	}
	batchSize := (n + workers - 1) / workers
	var wg sync.WaitGroup
	for i := 0; i < n; i += batchSize {
		end := min(i+batchSize, n)
		wg.Add(1)
		go func(idx, start, end int) {
			defer wg.Done()
			fn(idx, start, end)
		}(i/batchSize, i, end)
	}
	wg.Wait()
}

// PMap returns the result of applying fn to each member of m, using the specified
// number of worker goroutines. Order is preserved. The fn must be safe for concurrent use.
func PMap[T, R any](m Mapper[T], workers int, fn func(T) R) Mapper[R] {
	if len(m) == 0 {
		return Mapper[R]{}
	}
	results := make([]R, len(m))
	forBatches(len(m), workers, func(_, start, end int) {
		for j := start; j < end; j++ {
			results[j] = fn(m[j])
		}
	})
	return results
}

// PKeepIf returns a new slice containing members for which fn returns true,
// using the specified number of worker goroutines. Order is preserved.
func (m Mapper[T]) PKeepIf(workers int, fn func(T) bool) Mapper[T] {
	if len(m) == 0 {
		return Mapper[T]{}
	}
	batchResults := make([][]T, min(workers, len(m)))
	forBatches(len(m), workers, func(idx, start, end int) {
		var result []T
		for j := start; j < end; j++ {
			if fn(m[j]) {
				result = append(result, m[j])
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

// PEach applies fn to each member of m, using the specified number of worker
// goroutines. The fn must be safe for concurrent use.
func (m Mapper[T]) PEach(workers int, fn func(T)) {
	forBatches(len(m), workers, func(_, start, end int) {
		for j := start; j < end; j++ {
			fn(m[j])
		}
	})
}
