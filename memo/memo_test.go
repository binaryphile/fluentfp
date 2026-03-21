package memo

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

// --- From (state machine) ---

func TestOf_ComputesOnce(t *testing.T) {
	var calls atomic.Int32
	fn := From(func() int {
		calls.Add(1)
		return 42
	})

	got1 := fn()
	got2 := fn()
	got3 := fn()

	if got1 != 42 || got2 != 42 || got3 != 42 {
		t.Errorf("expected 42 each time, got %d, %d, %d", got1, got2, got3)
	}
	if c := calls.Load(); c != 1 {
		t.Errorf("expected 1 call, got %d", c)
	}
}

func TestOf_Concurrent(t *testing.T) {
	var calls atomic.Int32
	fn := From(func() int {
		calls.Add(1)
		return 99
	})

	var wg sync.WaitGroup
	const goroutines = 50
	results := make([]int, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = fn()
		}()
	}
	wg.Wait()

	for i, r := range results {
		if r != 99 {
			t.Errorf("goroutine %d: expected 99, got %d", i, r)
		}
	}
	if c := calls.Load(); c != 1 {
		t.Errorf("expected 1 call, got %d", c)
	}
}

func TestOf_PanicRetry(t *testing.T) {
	var calls atomic.Int32
	fn := From(func() int {
		n := calls.Add(1)
		if n == 1 {
			panic("transient failure")
		}
		return 42
	})

	// First call panics.
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic on first call")
			}
			if r != "transient failure" {
				t.Fatalf("unexpected panic value: %v", r)
			}
		}()
		fn()
	}()

	// Second call retries and succeeds.
	got := fn()
	if got != 42 {
		t.Errorf("expected 42 after retry, got %d", got)
	}

	// Third call returns cached result.
	got = fn()
	if got != 42 {
		t.Errorf("expected cached 42, got %d", got)
	}
	if c := calls.Load(); c != 2 {
		t.Errorf("expected 2 calls (1 panic + 1 success), got %d", c)
	}
}

func TestOf_PermanentPanic(t *testing.T) {
	var calls atomic.Int32
	fn := From(func() int {
		calls.Add(1)
		panic("always fails")
	})

	for range 3 {
		func() {
			defer func() { recover() }()
			fn()
		}()
	}

	if c := calls.Load(); c != 3 {
		t.Errorf("expected 3 calls (each retried), got %d", c)
	}
}

func TestOf_NilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil fn")
		}
	}()
	From[int](nil)
}

// --- Fn (basic memoization) ---

func TestFn_CachesPerKey(t *testing.T) {
	var calls atomic.Int32
	fn := Fn(func(k string) int {
		calls.Add(1)
		return len(k)
	})

	if fn("abc") != 3 {
		t.Error("expected 3")
	}
	if fn("ab") != 2 {
		t.Error("expected 2")
	}
	// Repeat — should be cached.
	if fn("abc") != 3 {
		t.Error("expected cached 3")
	}
	if c := calls.Load(); c != 2 {
		t.Errorf("expected 2 calls (2 distinct keys), got %d", c)
	}
}

// --- FnErr (error-retry semantics) ---

func TestFnErr_CachesSuccess(t *testing.T) {
	var calls atomic.Int32
	fn := FnErr(func(k string) (int, error) {
		calls.Add(1)
		return len(k), nil
	})

	v, err := fn("abc")
	if err != nil || v != 3 {
		t.Errorf("expected (3, nil), got (%d, %v)", v, err)
	}
	v, err = fn("abc")
	if err != nil || v != 3 {
		t.Errorf("expected cached (3, nil), got (%d, %v)", v, err)
	}
	if c := calls.Load(); c != 1 {
		t.Errorf("expected 1 call, got %d", c)
	}
}

func TestFnErr_RetriesOnError(t *testing.T) {
	var calls atomic.Int32
	errTransient := errors.New("transient")

	fn := FnErr(func(k string) (int, error) {
		n := calls.Add(1)
		if n == 1 {
			return 0, errTransient
		}
		return len(k), nil
	})

	// First call returns error — not cached.
	_, err := fn("abc")
	if !errors.Is(err, errTransient) {
		t.Fatalf("expected transient error, got %v", err)
	}

	// Second call retries and succeeds — now cached.
	v, err := fn("abc")
	if err != nil || v != 3 {
		t.Errorf("expected (3, nil), got (%d, %v)", v, err)
	}

	// Third call returns cached success.
	v, err = fn("abc")
	if err != nil || v != 3 {
		t.Errorf("expected cached (3, nil), got (%d, %v)", v, err)
	}
	if c := calls.Load(); c != 2 {
		t.Errorf("expected 2 calls (1 error + 1 success), got %d", c)
	}
}

// --- FnWith / FnErrWith (custom cache) ---

func TestFnWith_UsesProvidedCache(t *testing.T) {
	cache := NewLRU[string, int](10)
	var calls atomic.Int32

	fn := FnWith(func(k string) int {
		calls.Add(1)
		return len(k)
	}, cache)

	fn("abc")
	fn("abc") // should use cache

	if c := calls.Load(); c != 1 {
		t.Errorf("expected 1 call with custom cache, got %d", c)
	}
}

func TestFnErrWith_UsesProvidedCache(t *testing.T) {
	cache := NewLRU[string, int](10)
	var calls atomic.Int32

	fn := FnErrWith(func(k string) (int, error) {
		calls.Add(1)
		return len(k), nil
	}, cache)

	fn("abc")
	fn("abc") // should use cache

	if c := calls.Load(); c != 1 {
		t.Errorf("expected 1 call with custom cache, got %d", c)
	}
}

// --- LRU eviction ---

func TestLRU_EvictsLeastRecentlyUsed(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Store("a", 1)
	cache.Store("b", 2)
	cache.Store("c", 3) // evicts "a"

	if _, ok := cache.Load("a"); ok {
		t.Error("expected 'a' to be evicted")
	}
	if v, ok := cache.Load("b"); !ok || v != 2 {
		t.Error("expected 'b' to still be present")
	}
	if v, ok := cache.Load("c"); !ok || v != 3 {
		t.Error("expected 'c' to still be present")
	}
}

func TestLRU_LoadRefreshesOrder(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Store("a", 1)
	cache.Store("b", 2)
	cache.Load("a")     // refresh "a" — now "b" is least recent
	cache.Store("c", 3) // evicts "b"

	if _, ok := cache.Load("b"); ok {
		t.Error("expected 'b' to be evicted after 'a' was refreshed")
	}
	if v, ok := cache.Load("a"); !ok || v != 1 {
		t.Errorf("expected 'a'=1, got %d, %v", v, ok)
	}
}

func TestLRU_StoreUpdatesExisting(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Store("a", 1)
	cache.Store("b", 2)
	cache.Store("a", 10) // update "a", moves to front
	cache.Store("c", 3)  // evicts "b" (least recent)

	if _, ok := cache.Load("b"); ok {
		t.Error("expected 'b' to be evicted")
	}
	if v, ok := cache.Load("a"); !ok || v != 10 {
		t.Errorf("expected updated 'a'=10, got %d", v)
	}
}

func TestLRU_CapacityOne(t *testing.T) {
	cache := NewLRU[string, int](1)

	cache.Store("a", 1)
	cache.Store("b", 2) // evicts "a"

	if _, ok := cache.Load("a"); ok {
		t.Error("expected 'a' evicted with capacity 1")
	}
	if v, ok := cache.Load("b"); !ok || v != 2 {
		t.Errorf("expected 'b'=2, got %d", v)
	}
}

func TestNewLRU_ZeroCapacityPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for capacity 0")
		}
	}()
	NewLRU[string, int](0)
}
