package slice_test

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/slice"
)

func TestFanOutTransformsAndPreservesOrder(t *testing.T) {
	// double doubles an int.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	got := slice.FanOut(context.Background(), 4, []int{1, 2, 3, 4, 5}, double)

	if len(got) != 5 {
		t.Fatalf("got len %d, want 5", len(got))
	}

	for i, r := range got {
		val, ok := r.Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", i)

			continue
		}

		want := (i + 1) * 2
		if val != want {
			t.Errorf("[%d]: got %d, want %d", i, val, want)
		}
	}
}

func TestFanOutErrorInFn(t *testing.T) {
	// sentinel is a specific error for identity checking.
	sentinel := errors.New("fail at 2")

	// errAtIndex returns sentinel for index 2, success for others.
	errAtIndex := func(_ context.Context, n int) (int, error) {
		if n == 2 {
			return 0, sentinel
		}

		return n * 10, nil
	}

	got := slice.FanOut(context.Background(), 4, []int{0, 1, 2, 3}, errAtIndex)

	for i, r := range got {
		if i == 2 {
			if r.IsOk() {
				t.Errorf("[%d]: expected Err, got Ok", i)

				continue
			}

			err, _ := r.GetErr()
			if !errors.Is(err, sentinel) {
				t.Errorf("[%d]: got %v, want sentinel error", i, err)
			}

			continue
		}

		val, ok := r.Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", i)

			continue
		}

		if val != i*10 {
			t.Errorf("[%d]: got %d, want %d", i, val, i*10)
		}
	}
}

func TestFanOutPanicInFn(t *testing.T) {
	tests := []struct {
		name       string
		panicValue any
		checkValue func(t *testing.T, pe *rslt.PanicError)
	}{
		{
			name:       "string panic",
			panicValue: "boom",
			checkValue: func(t *testing.T, pe *rslt.PanicError) {
				if pe.Value != "boom" {
					t.Errorf("PanicError.Value: got %v, want \"boom\"", pe.Value)
				}
			},
		},
		{
			name:       "error panic",
			panicValue: errors.New("wrapped error"),
			checkValue: func(t *testing.T, pe *rslt.PanicError) {
				if _, ok := pe.Value.(error); !ok {
					t.Error("PanicError.Value: expected error type")
				}
			},
		},
		{
			name:       "nil panic produces PanicNilError",
			panicValue: nil, // Go 1.21+: recover() returns *runtime.PanicNilError
			checkValue: func(t *testing.T, pe *rslt.PanicError) {
				if pe.Value == nil {
					t.Error("PanicError.Value: got nil, expected *runtime.PanicNilError")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// panicAt panics at index 1, returns normally otherwise.
			panicAt := func(_ context.Context, n int) (int, error) {
				if n == 1 {
					panic(tt.panicValue)
				}

				return n, nil
			}

			got := slice.FanOut(context.Background(), 1, []int{0, 1, 2}, panicAt)

			// Index 0 and 2 should be Ok.
			for _, idx := range []int{0, 2} {
				val, ok := got[idx].Get()
				if !ok {
					t.Errorf("[%d]: expected Ok, got Err", idx)

					continue
				}

				if val != idx {
					t.Errorf("[%d]: got %d, want %d", idx, val, idx)
				}
			}

			// Index 1 should be Err with *PanicError.
			if got[1].IsOk() {
				t.Fatal("[1]: expected Err, got Ok")
			}

			err, _ := got[1].GetErr()

			var pe *rslt.PanicError
			if !errors.As(err, &pe) {
				t.Fatalf("[1]: expected *PanicError, got %T: %v", err, err)
			}

			if len(pe.Stack) == 0 {
				t.Error("[1]: PanicError.Stack is empty")
			}

			tt.checkValue(t, pe)
		})
	}
}

func TestFanOutEmptyInput(t *testing.T) {
	neverCalled := func(_ context.Context, _ int) (int, error) {
		t.Fatal("fn should not be called on empty input")

		return 0, nil
	}

	got := slice.FanOut(context.Background(), 4, []int{}, neverCalled)
	if got == nil {
		t.Error("expected non-nil empty Mapper, got nil")
	}

	if len(got) != 0 {
		t.Errorf("expected empty Mapper, got len %d", len(got))
	}
}

func TestFanOutSequential(t *testing.T) {
	// double doubles an int.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	got := slice.FanOut(context.Background(), 1, []int{10, 20, 30}, double)

	for i, r := range got {
		val, ok := r.Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", i)

			continue
		}

		want := (i + 1) * 10 * 2
		if val != want {
			t.Errorf("[%d]: got %d, want %d", i, val, want)
		}
	}
}

func TestFanOutClampsN(t *testing.T) {
	maxSeen := assertMaxConcurrency(t, 100, 3)
	if maxSeen > 3 {
		t.Errorf("max concurrent %d > inputSize 3 — n was not clamped", maxSeen)
	}
}

func TestFanOutValidationPanics(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		ctx     context.Context
		fn      func(context.Context, int) (int, error)
		wantMsg string
	}{
		{
			name:    "n <= 0",
			n:       0,
			ctx:     context.Background(),
			fn:      func(_ context.Context, n int) (int, error) { return n, nil },
			wantMsg: "slice.FanOut: n must be > 0",
		},
		{
			name:    "n <= 0 even for empty input",
			n:       -1,
			ctx:     context.Background(),
			fn:      func(_ context.Context, n int) (int, error) { return n, nil },
			wantMsg: "slice.FanOut: n must be > 0",
		},
		{
			name:    "nil ctx",
			n:       1,
			ctx:     nil,
			fn:      func(_ context.Context, n int) (int, error) { return n, nil },
			wantMsg: "slice.FanOut: ctx must not be nil",
		},
		{
			name:    "nil fn",
			n:       1,
			ctx:     context.Background(),
			fn:      nil,
			wantMsg: "slice.FanOut: fn must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				v := recover()
				if v == nil {
					t.Fatal("expected panic, got none")
				}

				got, ok := v.(string)
				if !ok {
					t.Fatalf("expected string panic, got %T: %v", v, v)
				}

				if got != tt.wantMsg {
					t.Errorf("got %q, want %q", got, tt.wantMsg)
				}
			}()

			slice.FanOut(tt.ctx, tt.n, []int{}, tt.fn)
		})
	}
}

func TestFanOutAlreadyCancelledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var called atomic.Int32

	// trackCalled counts invocations.
	trackCalled := func(_ context.Context, n int) (int, error) {
		called.Add(1)

		return n, nil
	}

	got := slice.FanOut(ctx, 4, []int{1, 2, 3, 4, 5}, trackCalled)

	if called.Load() != 0 {
		t.Errorf("expected zero callbacks, got %d", called.Load())
	}

	for i, r := range got {
		if r.IsOk() {
			t.Errorf("[%d]: expected Err, got Ok", i)
		}

		err, _ := r.GetErr()
		if !errors.Is(err, context.Canceled) {
			t.Errorf("[%d]: got %v, want context.Canceled", i, err)
		}
	}
}

func TestFanOutCancellationMidFlight(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	const workers = 2
	input := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	// startedCount tracks how many callbacks have started (non-blocking).
	var startedCount atomic.Int32

	// gate blocks all callbacks until released.
	gate := make(chan struct{})

	// blockUntilGate signals start atomically then waits for gate.
	blockUntilGate := func(ctx context.Context, n int) (int, error) {
		startedCount.Add(1)
		<-gate

		return n, nil
	}

	done := make(chan []rslt.Result[int])

	go func() {
		got := slice.FanOut(ctx, workers, input, blockUntilGate)
		done <- []rslt.Result[int](got)
	}()

	// Spin until at least workers callbacks have started.
	for startedCount.Load() < int32(workers) {
		// busy-wait is acceptable in tests for short waits
	}

	// Cancel — in-flight callbacks will return when gate opens, unscheduled get Err.
	cancel()

	// Release all blocked callbacks.
	close(gate)

	got := <-done

	if len(got) != len(input) {
		t.Fatalf("got len %d, want %d", len(got), len(input))
	}

	var okCount, errCount int
	for _, r := range got {
		if r.IsOk() {
			okCount++
		} else {
			errCount++
		}
	}

	// At least workers started, at most workers+1 due to race window.
	if okCount < workers {
		t.Errorf("okCount %d < workers %d", okCount, workers)
	}

	if okCount > workers+1 {
		t.Errorf("okCount %d > workers+1 %d — cancellation not prompt", okCount, workers+1)
	}

	if errCount == 0 {
		t.Error("expected some Err results from cancellation")
	}

	// Verify Err results have context.Canceled.
	for i, r := range got {
		if r.IsErr() {
			err, _ := r.GetErr()
			if !errors.Is(err, context.Canceled) {
				t.Errorf("[%d]: got %v, want context.Canceled", i, err)
			}
		}
	}
}

func TestFanOutNilInputSlice(t *testing.T) {
	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	got := slice.FanOut(context.Background(), 4, ([]int)(nil), identity)
	if got == nil {
		t.Error("expected non-nil empty Mapper, got nil")
	}

	if len(got) != 0 {
		t.Errorf("expected empty, got len %d", len(got))
	}
}

func TestFanOutConcurrencyBound(t *testing.T) {
	maxSeen := assertMaxConcurrency(t, 3, 10)
	if maxSeen > 3 {
		t.Errorf("max concurrent %d > workers 3", maxSeen)
	}
}

// assertMaxConcurrency runs FanOut with the given workers and inputSize,
// blocking each callback on a barrier. Returns the maximum observed concurrency.
func assertMaxConcurrency(t *testing.T, workers, inputSize int) int32 {
	t.Helper()

	var current atomic.Int32
	var maxSeen atomic.Int32
	barrier := make(chan struct{})

	// trackConcurrency records max concurrent goroutines then waits on barrier.
	trackConcurrency := func(_ context.Context, _ int) (int, error) {
		c := current.Add(1)

		// CAS loop to update max.
		for {
			old := maxSeen.Load()
			if c <= old || maxSeen.CompareAndSwap(old, c) {
				break
			}
		}

		<-barrier
		current.Add(-1)

		return 0, nil
	}

	input := make([]int, inputSize)
	done := make(chan struct{})

	go func() {
		slice.FanOut(context.Background(), workers, input, trackConcurrency)
		close(done)
	}()

	// Release all workers repeatedly until done.
	go func() {
		for {
			select {
			case barrier <- struct{}{}:
			case <-done:
				return
			}
		}
	}()

	<-done

	return maxSeen.Load()
}

func TestFanOutOrderPreservation(t *testing.T) {
	// barriers ensures items complete in reverse order.
	const n = 5
	barriers := make([]chan struct{}, n)
	for i := range barriers {
		barriers[i] = make(chan struct{})
	}

	// reverseOrder forces item i to wait for item i+1 to complete.
	reverseOrder := func(_ context.Context, idx int) (int, error) {
		if idx < n-1 {
			<-barriers[idx]
		}

		val := idx * 100

		if idx > 0 {
			close(barriers[idx-1])
		}

		return val, nil
	}

	input := make([]int, n)
	for i := range input {
		input[i] = i
	}

	got := slice.FanOut(context.Background(), n, input, reverseOrder)

	for i, r := range got {
		val, ok := r.Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", i)

			continue
		}

		want := i * 100
		if val != want {
			t.Errorf("[%d]: got %d, want %d", i, val, want)
		}
	}
}

func TestFanOutPanicConcurrent(t *testing.T) {
	// panicAtThree panics at index 3, returns normally otherwise.
	panicAtThree := func(_ context.Context, n int) (int, error) {
		if n == 3 {
			panic("concurrent boom")
		}

		return n * 10, nil
	}

	got := slice.FanOut(context.Background(), 4, []int{0, 1, 2, 3, 4}, panicAtThree)

	if len(got) != 5 {
		t.Fatalf("got len %d, want 5", len(got))
	}

	// Non-panicking indices should be Ok with correct values.
	for _, idx := range []int{0, 1, 2, 4} {
		val, ok := got[idx].Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", idx)

			continue
		}

		if val != idx*10 {
			t.Errorf("[%d]: got %d, want %d", idx, val, idx*10)
		}
	}

	// Index 3 should be Err with *PanicError.
	if got[3].IsOk() {
		t.Fatal("[3]: expected Err, got Ok")
	}

	err, _ := got[3].GetErr()

	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("[3]: expected *PanicError, got %T: %v", err, err)
	}

	if pe.Value != "concurrent boom" {
		t.Errorf("[3]: PanicError.Value: got %v, want \"concurrent boom\"", pe.Value)
	}
}

func TestFanOutSequentialCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// cancelAtOne cancels ctx when processing item at index 1.
	// Sequential path checks ctx.Err() before each runItem call,
	// so items 2+ are guaranteed to see cancellation.
	cancelAtOne := func(_ context.Context, n int) (int, error) {
		if n == 1 {
			cancel()
		}

		return n * 10, nil
	}

	got := slice.FanOut(ctx, 1, []int{0, 1, 2, 3, 4}, cancelAtOne)

	if len(got) != 5 {
		t.Fatalf("got len %d, want 5", len(got))
	}

	// Items 0 and 1 should be Ok (1 runs before cancellation is checked).
	for _, idx := range []int{0, 1} {
		val, ok := got[idx].Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", idx)

			continue
		}

		if val != idx*10 {
			t.Errorf("[%d]: got %d, want %d", idx, val, idx*10)
		}
	}

	// Items 2-4 must be Err(context.Canceled) — ctx was cancelled before their iteration.
	for i := 2; i < 5; i++ {
		if got[i].IsOk() {
			t.Errorf("[%d]: expected Err, got Ok", i)

			continue
		}

		err, _ := got[i].GetErr()
		if !errors.Is(err, context.Canceled) {
			t.Errorf("[%d]: got %v, want context.Canceled", i, err)
		}
	}
}

func TestFanOutMixedStates(t *testing.T) {
	// sentinel is a specific error for identity checking.
	sentinel := errors.New("item error")

	// mixedFn produces different outcomes per index.
	mixedFn := func(_ context.Context, n int) (int, error) {
		switch n {
		case 0:
			return 100, nil // success
		case 1:
			return 0, sentinel // error
		case 2:
			panic("mixed panic") // panic
		case 3:
			return 300, nil // success
		default:
			return n, nil
		}
	}

	got := slice.FanOut(context.Background(), 4, []int{0, 1, 2, 3}, mixedFn)

	if len(got) != 4 {
		t.Fatalf("got len %d, want 4", len(got))
	}

	// Index 0: Ok(100)
	val, ok := got[0].Get()
	if !ok {
		t.Error("[0]: expected Ok, got Err")
	} else if val != 100 {
		t.Errorf("[0]: got %d, want 100", val)
	}

	// Index 1: Err(sentinel)
	if got[1].IsOk() {
		t.Error("[1]: expected Err, got Ok")
	} else {
		err, _ := got[1].GetErr()
		if !errors.Is(err, sentinel) {
			t.Errorf("[1]: got %v, want sentinel", err)
		}
	}

	// Index 2: Err(*PanicError)
	if got[2].IsOk() {
		t.Error("[2]: expected Err, got Ok")
	} else {
		err, _ := got[2].GetErr()

		var pe *rslt.PanicError
		if !errors.As(err, &pe) {
			t.Errorf("[2]: expected *PanicError, got %T: %v", err, err)
		}
	}

	// Index 3: Ok(300)
	val, ok = got[3].Get()
	if !ok {
		t.Error("[3]: expected Ok, got Err")
	} else if val != 300 {
		t.Errorf("[3]: got %d, want 300", val)
	}
}

// FanOutEach tests

func TestFanOutEachAppliesAndRecordsErrors(t *testing.T) {
	// failAtTwo returns an error for value 2, nil otherwise.
	failAtTwo := func(_ context.Context, n int) error {
		if n == 2 {
			return errors.New("fail")
		}

		return nil
	}

	errs := slice.FanOutEach(context.Background(), 4, []int{0, 1, 2, 3}, failAtTwo)

	if len(errs) != 4 {
		t.Fatalf("got len %d, want 4", len(errs))
	}

	for i, err := range errs {
		if i == 2 {
			if err == nil {
				t.Errorf("[%d]: expected error, got nil", i)
			}

			continue
		}

		if err != nil {
			t.Errorf("[%d]: expected nil, got %v", i, err)
		}
	}
}

func TestFanOutEachPanicWrapped(t *testing.T) {
	// panicAtOne panics at index 1.
	panicAtOne := func(_ context.Context, n int) error {
		if n == 1 {
			panic("boom")
		}

		return nil
	}

	errs := slice.FanOutEach(context.Background(), 1, []int{0, 1, 2}, panicAtOne)

	if errs[0] != nil {
		t.Errorf("[0]: expected nil, got %v", errs[0])
	}

	if errs[2] != nil {
		t.Errorf("[2]: expected nil, got %v", errs[2])
	}

	var pe *rslt.PanicError
	if !errors.As(errs[1], &pe) {
		t.Fatalf("[1]: expected *PanicError, got %T: %v", errs[1], errs[1])
	}

	if pe.Value != "boom" {
		t.Errorf("[1]: PanicError.Value: got %v, want \"boom\"", pe.Value)
	}
}

func TestFanOutEachNilFnPanics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, got none")
		}

		got, ok := v.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", v, v)
		}

		// wantMsg is the expected panic message for nil fn.
		wantMsg := "slice.FanOutEach: fn must not be nil"
		if got != wantMsg {
			t.Errorf("got %q, want %q", got, wantMsg)
		}
	}()

	slice.FanOutEach[int](context.Background(), 1, []int{1}, nil)
}

func TestFanOutEachValidationPanics(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		ctx     context.Context
		wantMsg string
	}{
		{
			name:    "n <= 0",
			n:       0,
			ctx:     context.Background(),
			wantMsg: "slice.FanOutEach: n must be > 0",
		},
		{
			name:    "nil ctx",
			n:       1,
			ctx:     nil,
			wantMsg: "slice.FanOutEach: ctx must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				v := recover()
				if v == nil {
					t.Fatal("expected panic, got none")
				}

				got, ok := v.(string)
				if !ok {
					t.Fatalf("expected string panic, got %T: %v", v, v)
				}

				if got != tt.wantMsg {
					t.Errorf("got %q, want %q", got, tt.wantMsg)
				}
			}()

			// noop is a no-op fn for validation-only tests.
			noop := func(_ context.Context, _ int) error { return nil }
			slice.FanOutEach(tt.ctx, tt.n, []int{}, noop)
		})
	}
}

func TestFanOutEachEmptyInput(t *testing.T) {
	noop := func(_ context.Context, _ int) error { return nil }

	errs := slice.FanOutEach(context.Background(), 4, []int{}, noop)
	if errs == nil {
		t.Error("expected non-nil empty []error, got nil")
	}

	if len(errs) != 0 {
		t.Errorf("expected empty, got len %d", len(errs))
	}
}

func TestFanOutEachNilInput(t *testing.T) {
	noop := func(_ context.Context, _ int) error { return nil }

	errs := slice.FanOutEach(context.Background(), 4, ([]int)(nil), noop)
	if errs == nil {
		t.Error("expected non-nil empty []error, got nil")
	}

	if len(errs) != 0 {
		t.Errorf("expected empty, got len %d", len(errs))
	}
}

func TestFanOutEachAlreadyCancelledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var called atomic.Int32

	// trackCalled counts invocations.
	trackCalled := func(_ context.Context, _ int) error {
		called.Add(1)

		return nil
	}

	errs := slice.FanOutEach(ctx, 4, []int{1, 2, 3}, trackCalled)

	if called.Load() != 0 {
		t.Errorf("expected zero callbacks, got %d", called.Load())
	}

	if len(errs) != 3 {
		t.Fatalf("got len %d, want 3", len(errs))
	}

	for i, err := range errs {
		if err == nil {
			t.Errorf("[%d]: expected error, got nil", i)

			continue
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("[%d]: got %v, want context.Canceled", i, err)
		}
	}
}

// FanOutWeighted tests

func TestFanOutWeightedBasic(t *testing.T) {
	// double doubles an int.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	got := slice.FanOutWeighted(context.Background(), 4, []int{1, 2, 3, 4, 5}, unitCost, double)

	if len(got) != 5 {
		t.Fatalf("got len %d, want 5", len(got))
	}

	for i, r := range got {
		val, ok := r.Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", i)

			continue
		}

		want := (i + 1) * 2
		if val != want {
			t.Errorf("[%d]: got %d, want %d", i, val, want)
		}
	}
}

func TestFanOutWeightedVariableCosts(t *testing.T) {
	// identity returns the value unchanged.
	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	// variableCost returns cost equal to the value itself.
	variableCost := func(n int) int { return n }

	input := []int{1, 2, 3}
	got := slice.FanOutWeighted(context.Background(), 10, input, variableCost, identity)

	if len(got) != 3 {
		t.Fatalf("got len %d, want 3", len(got))
	}

	for i, r := range got {
		val, ok := r.Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", i)

			continue
		}

		if val != input[i] {
			t.Errorf("[%d]: got %d, want %d", i, val, input[i])
		}
	}
}

func TestFanOutWeightedOrderPreservation(t *testing.T) {
	// barriers ensures items complete in reverse order.
	const n = 5
	barriers := make([]chan struct{}, n)
	for i := range barriers {
		barriers[i] = make(chan struct{})
	}

	// reverseOrder forces item i to wait for item i+1 to complete.
	reverseOrder := func(_ context.Context, idx int) (int, error) {
		if idx < n-1 {
			<-barriers[idx]
		}

		val := idx * 100

		if idx > 0 {
			close(barriers[idx-1])
		}

		return val, nil
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	input := make([]int, n)
	for i := range input {
		input[i] = i
	}

	got := slice.FanOutWeighted(context.Background(), n, input, unitCost, reverseOrder)

	for i, r := range got {
		val, ok := r.Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", i)

			continue
		}

		want := i * 100
		if val != want {
			t.Errorf("[%d]: got %d, want %d", i, val, want)
		}
	}
}

func TestFanOutWeightedConcurrencyBound(t *testing.T) {
	// Each item costs 3. Capacity is 6. So at most 2 items can run concurrently.
	var current atomic.Int32
	var maxSeen atomic.Int32
	barrier := make(chan struct{})

	// trackConcurrency records max concurrent goroutines then waits on barrier.
	trackConcurrency := func(_ context.Context, _ int) (int, error) {
		c := current.Add(1)

		for {
			old := maxSeen.Load()
			if c <= old || maxSeen.CompareAndSwap(old, c) {
				break
			}
		}

		<-barrier
		current.Add(-1)

		return 0, nil
	}

	// costThree returns 3 for every item.
	costThree := func(_ int) int { return 3 }

	input := make([]int, 6)
	done := make(chan struct{})

	go func() {
		slice.FanOutWeighted(context.Background(), 6, input, costThree, trackConcurrency)
		close(done)
	}()

	// Release all workers repeatedly until done.
	go func() {
		for {
			select {
			case barrier <- struct{}{}:
			case <-done:
				return
			}
		}
	}()

	<-done

	if maxSeen.Load() > 2 {
		t.Errorf("max concurrent %d > expected 2 (capacity 6, cost 3)", maxSeen.Load())
	}
}

func TestFanOutWeightedCostExceedsCapacityPanics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, got none")
		}

		got, ok := v.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", v, v)
		}

		wantMsg := "slice.FanOutWeighted: cost must be <= capacity"
		if got != wantMsg {
			t.Errorf("got %q, want %q", got, wantMsg)
		}
	}()

	// costFive returns 5 for every item — exceeds capacity of 3.
	costFive := func(_ int) int { return 5 }

	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	slice.FanOutWeighted(context.Background(), 3, []int{1}, costFive, identity)
}

func TestFanOutWeightedCostZeroPanics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, got none")
		}

		got, ok := v.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", v, v)
		}

		wantMsg := "slice.FanOutWeighted: cost must be > 0"
		if got != wantMsg {
			t.Errorf("got %q, want %q", got, wantMsg)
		}
	}()

	// costZero returns 0 — invalid.
	costZero := func(_ int) int { return 0 }

	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	slice.FanOutWeighted(context.Background(), 3, []int{1}, costZero, identity)
}

func TestFanOutWeightedNilCostPanics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, got none")
		}

		got, ok := v.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", v, v)
		}

		wantMsg := "slice.FanOutWeighted: cost must not be nil"
		if got != wantMsg {
			t.Errorf("got %q, want %q", got, wantMsg)
		}
	}()

	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	slice.FanOutWeighted(context.Background(), 3, []int{1}, nil, identity)
}

func TestFanOutWeightedValidationPanics(t *testing.T) {
	tests := []struct {
		name    string
		cap     int
		ctx     context.Context
		cost    func(int) int
		fn      func(context.Context, int) (int, error)
		wantMsg string
	}{
		{
			name:    "capacity <= 0",
			cap:     0,
			ctx:     context.Background(),
			cost:    func(_ int) int { return 1 },
			fn:      func(_ context.Context, n int) (int, error) { return n, nil },
			wantMsg: "slice.FanOutWeighted: capacity must be > 0",
		},
		{
			name:    "nil ctx",
			cap:     1,
			ctx:     nil,
			cost:    func(_ int) int { return 1 },
			fn:      func(_ context.Context, n int) (int, error) { return n, nil },
			wantMsg: "slice.FanOutWeighted: ctx must not be nil",
		},
		{
			name:    "nil fn",
			cap:     1,
			ctx:     context.Background(),
			cost:    func(_ int) int { return 1 },
			fn:      nil,
			wantMsg: "slice.FanOutWeighted: fn must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				v := recover()
				if v == nil {
					t.Fatal("expected panic, got none")
				}

				got, ok := v.(string)
				if !ok {
					t.Fatalf("expected string panic, got %T: %v", v, v)
				}

				if got != tt.wantMsg {
					t.Errorf("got %q, want %q", got, tt.wantMsg)
				}
			}()

			slice.FanOutWeighted(tt.ctx, tt.cap, []int{}, tt.cost, tt.fn)
		})
	}
}

func TestFanOutWeightedEmptyInput(t *testing.T) {
	neverCalled := func(_ context.Context, _ int) (int, error) {
		t.Fatal("fn should not be called on empty input")

		return 0, nil
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	got := slice.FanOutWeighted(context.Background(), 4, []int{}, unitCost, neverCalled)
	if got == nil {
		t.Error("expected non-nil empty Mapper, got nil")
	}

	if len(got) != 0 {
		t.Errorf("expected empty Mapper, got len %d", len(got))
	}
}

func TestFanOutWeightedNilInput(t *testing.T) {
	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	got := slice.FanOutWeighted(context.Background(), 4, ([]int)(nil), unitCost, identity)
	if got == nil {
		t.Error("expected non-nil empty Mapper, got nil")
	}

	if len(got) != 0 {
		t.Errorf("expected empty, got len %d", len(got))
	}
}

func TestFanOutWeightedAlreadyCancelledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var called atomic.Int32

	// trackCalled counts invocations.
	trackCalled := func(_ context.Context, n int) (int, error) {
		called.Add(1)

		return n, nil
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	got := slice.FanOutWeighted(ctx, 4, []int{1, 2, 3, 4, 5}, unitCost, trackCalled)

	if called.Load() != 0 {
		t.Errorf("expected zero callbacks, got %d", called.Load())
	}

	for i, r := range got {
		if r.IsOk() {
			t.Errorf("[%d]: expected Err, got Ok", i)
		}

		err, _ := r.GetErr()
		if !errors.Is(err, context.Canceled) {
			t.Errorf("[%d]: got %v, want context.Canceled", i, err)
		}
	}
}

func TestFanOutWeightedCancellationMidFlight(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	const capacity = 4
	input := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	var startedCount atomic.Int32
	gate := make(chan struct{})

	// blockUntilGate signals start atomically then waits for gate.
	blockUntilGate := func(ctx context.Context, n int) (int, error) {
		startedCount.Add(1)
		<-gate

		return n, nil
	}

	// costTwo returns 2 for every item. Capacity 4 means at most 2 items concurrent.
	costTwo := func(_ int) int { return 2 }

	done := make(chan []rslt.Result[int])

	go func() {
		got := slice.FanOutWeighted(ctx, capacity, input, costTwo, blockUntilGate)
		done <- []rslt.Result[int](got)
	}()

	// Spin until at least 2 callbacks have started (capacity/costPerItem = 4/2 = 2).
	for startedCount.Load() < 2 {
		runtime.Gosched()
	}

	cancel()
	close(gate)

	got := <-done

	if len(got) != len(input) {
		t.Fatalf("got len %d, want %d", len(got), len(input))
	}

	var okCount, errCount int
	for _, r := range got {
		if r.IsOk() {
			okCount++
		} else {
			errCount++
		}
	}

	if okCount < 2 {
		t.Errorf("okCount %d < 2", okCount)
	}

	if errCount == 0 {
		t.Error("expected some Err results from cancellation")
	}

	for i, r := range got {
		if r.IsErr() {
			err, _ := r.GetErr()
			if !errors.Is(err, context.Canceled) {
				t.Errorf("[%d]: got %v, want context.Canceled", i, err)
			}
		}
	}
}

func TestFanOutWeightedPanicRecovery(t *testing.T) {
	// panicAtTwo panics at index 2, returns normally otherwise.
	panicAtTwo := func(_ context.Context, n int) (int, error) {
		if n == 2 {
			panic("weighted boom")
		}

		return n * 10, nil
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	got := slice.FanOutWeighted(context.Background(), 4, []int{0, 1, 2, 3}, unitCost, panicAtTwo)

	if len(got) != 4 {
		t.Fatalf("got len %d, want 4", len(got))
	}

	for _, idx := range []int{0, 1, 3} {
		val, ok := got[idx].Get()
		if !ok {
			t.Errorf("[%d]: expected Ok, got Err", idx)

			continue
		}

		if val != idx*10 {
			t.Errorf("[%d]: got %d, want %d", idx, val, idx*10)
		}
	}

	if got[2].IsOk() {
		t.Fatal("[2]: expected Err, got Ok")
	}

	err, _ := got[2].GetErr()

	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("[2]: expected *PanicError, got %T: %v", err, err)
	}

	if pe.Value != "weighted boom" {
		t.Errorf("[2]: PanicError.Value: got %v, want \"weighted boom\"", pe.Value)
	}
}

func TestFanOutWeightedMixedOutcomes(t *testing.T) {
	// sentinel is a specific error for identity checking.
	sentinel := errors.New("item error")

	// mixedFn produces different outcomes per index, each with different cost.
	mixedFn := func(_ context.Context, n int) (int, error) {
		switch n {
		case 0:
			return 100, nil // success, cost 1
		case 1:
			return 0, sentinel // error, cost 2
		case 2:
			panic("mixed panic") // panic, cost 3
		case 3:
			return 300, nil // success, cost 1
		default:
			return n, nil
		}
	}

	// variableCost assigns different costs per item.
	variableCost := func(n int) int {
		costs := []int{1, 2, 3, 1}
		return costs[n]
	}

	got := slice.FanOutWeighted(context.Background(), 10, []int{0, 1, 2, 3}, variableCost, mixedFn)

	if len(got) != 4 {
		t.Fatalf("got len %d, want 4", len(got))
	}

	// Index 0: Ok(100)
	val, ok := got[0].Get()
	if !ok {
		t.Error("[0]: expected Ok, got Err")
	} else if val != 100 {
		t.Errorf("[0]: got %d, want 100", val)
	}

	// Index 1: Err(sentinel)
	if got[1].IsOk() {
		t.Error("[1]: expected Err, got Ok")
	} else {
		err, _ := got[1].GetErr()
		if !errors.Is(err, sentinel) {
			t.Errorf("[1]: got %v, want sentinel", err)
		}
	}

	// Index 2: Err(*PanicError)
	if got[2].IsOk() {
		t.Error("[2]: expected Err, got Ok")
	} else {
		err, _ := got[2].GetErr()

		var pe *rslt.PanicError
		if !errors.As(err, &pe) {
			t.Errorf("[2]: expected *PanicError, got %T: %v", err, err)
		}
	}

	// Index 3: Ok(300)
	val, ok = got[3].Get()
	if !ok {
		t.Error("[3]: expected Ok, got Err")
	} else if val != 300 {
		t.Errorf("[3]: got %d, want 300", val)
	}
}

// FanOutEachWeighted tests

func TestFanOutEachWeightedBasic(t *testing.T) {
	// failAtTwo returns an error for value 2, nil otherwise.
	failAtTwo := func(_ context.Context, n int) error {
		if n == 2 {
			return errors.New("fail")
		}

		return nil
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	errs := slice.FanOutEachWeighted(context.Background(), 4, []int{0, 1, 2, 3}, unitCost, failAtTwo)

	if len(errs) != 4 {
		t.Fatalf("got len %d, want 4", len(errs))
	}

	for i, err := range errs {
		if i == 2 {
			if err == nil {
				t.Errorf("[%d]: expected error, got nil", i)
			}

			continue
		}

		if err != nil {
			t.Errorf("[%d]: expected nil, got %v", i, err)
		}
	}
}

func TestFanOutEachWeightedPanicWrapped(t *testing.T) {
	// panicAtOne panics at index 1.
	panicAtOne := func(_ context.Context, n int) error {
		if n == 1 {
			panic("weighted each boom")
		}

		return nil
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	errs := slice.FanOutEachWeighted(context.Background(), 1, []int{0, 1, 2}, unitCost, panicAtOne)

	if errs[0] != nil {
		t.Errorf("[0]: expected nil, got %v", errs[0])
	}

	if errs[2] != nil {
		t.Errorf("[2]: expected nil, got %v", errs[2])
	}

	var pe *rslt.PanicError
	if !errors.As(errs[1], &pe) {
		t.Fatalf("[1]: expected *PanicError, got %T: %v", errs[1], errs[1])
	}

	if pe.Value != "weighted each boom" {
		t.Errorf("[1]: PanicError.Value: got %v, want \"weighted each boom\"", pe.Value)
	}
}

func TestFanOutEachWeightedValidationPanics(t *testing.T) {
	tests := []struct {
		name    string
		cap     int
		ctx     context.Context
		cost    func(int) int
		fn      func(context.Context, int) error
		wantMsg string
	}{
		{
			name:    "capacity <= 0",
			cap:     0,
			ctx:     context.Background(),
			cost:    func(_ int) int { return 1 },
			fn:      func(_ context.Context, _ int) error { return nil },
			wantMsg: "slice.FanOutEachWeighted: capacity must be > 0",
		},
		{
			name:    "nil ctx",
			cap:     1,
			ctx:     nil,
			cost:    func(_ int) int { return 1 },
			fn:      func(_ context.Context, _ int) error { return nil },
			wantMsg: "slice.FanOutEachWeighted: ctx must not be nil",
		},
		{
			name:    "nil cost",
			cap:     1,
			ctx:     context.Background(),
			cost:    nil,
			fn:      func(_ context.Context, _ int) error { return nil },
			wantMsg: "slice.FanOutEachWeighted: cost must not be nil",
		},
		{
			name:    "nil fn",
			cap:     1,
			ctx:     context.Background(),
			cost:    func(_ int) int { return 1 },
			fn:      nil,
			wantMsg: "slice.FanOutEachWeighted: fn must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				v := recover()
				if v == nil {
					t.Fatal("expected panic, got none")
				}

				got, ok := v.(string)
				if !ok {
					t.Fatalf("expected string panic, got %T: %v", v, v)
				}

				if got != tt.wantMsg {
					t.Errorf("got %q, want %q", got, tt.wantMsg)
				}
			}()

			slice.FanOutEachWeighted(tt.ctx, tt.cap, []int{}, tt.cost, tt.fn)
		})
	}
}

func TestFanOutEachCancellationMidFlight(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var startedCount atomic.Int32
	gate := make(chan struct{})

	const workers = 2

	// blockUntilGate signals start atomically then waits for gate.
	blockUntilGate := func(_ context.Context, _ int) error {
		startedCount.Add(1)
		<-gate

		return nil
	}

	input := []int{0, 1, 2, 3, 4, 5, 6, 7}
	done := make(chan []error)

	go func() {
		done <- slice.FanOutEach(ctx, workers, input, blockUntilGate)
	}()

	// Spin until at least workers callbacks have started.
	for startedCount.Load() < int32(workers) {
		runtime.Gosched()
	}

	cancel()
	close(gate)

	errs := <-done

	if len(errs) != len(input) {
		t.Fatalf("got len %d, want %d", len(errs), len(input))
	}

	var nilCount, errCount int
	for _, err := range errs {
		if err == nil {
			nilCount++
		} else {
			errCount++
		}
	}

	if nilCount < workers {
		t.Errorf("nilCount %d < workers %d", nilCount, workers)
	}

	if errCount == 0 {
		t.Error("expected some errors from cancellation")
	}

	// Verify error entries have context.Canceled.
	for i, err := range errs {
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("[%d]: got %v, want context.Canceled", i, err)
		}
	}
}

// --- FanOutAll tests ---

func TestFanOutAllSuccess(t *testing.T) {
	// double doubles an int.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	got, err := slice.FanOutAll(context.Background(), 4, []int{1, 2, 3, 4, 5}, double)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{2, 4, 6, 8, 10}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("[%d]: got %d, want %d", i, v, want[i])
		}
	}
}

func TestFanOutAllReturnsRootCauseNotSiblingCancellation(t *testing.T) {
	// sentinel is the real failure, occurring at the LAST index (3).
	// Items 0-2 are at earlier indices and return context.Canceled after cancellation.
	// Without root-cause tracking, CollectAll would return index 0's context.Canceled.
	sentinel := errors.New("boom")

	// barrier ensures all 4 goroutines are running before the failure fires.
	var barrier sync.WaitGroup
	barrier.Add(4)

	// failAtThree fails at index 3 after all items are running.
	failAtThree := func(ctx context.Context, n int) (int, error) {
		barrier.Done()
		barrier.Wait()

		if n == 3 {
			return 0, sentinel
		}

		<-ctx.Done()

		return 0, ctx.Err()
	}

	_, err := slice.FanOutAll(context.Background(), 4, []int{0, 1, 2, 3}, failAtThree)
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want root cause %v (not sibling context.Canceled)", err, sentinel)
	}
}

func TestFanOutAllPanicReturnsRootCauseNotSiblingCancellation(t *testing.T) {
	// Panic at index 3 (last); items 0-2 return context.Canceled.
	// Without root-cause tracking, CollectAll would return context.Canceled.
	var barrier sync.WaitGroup
	barrier.Add(4)

	// panicAtThree panics at index 3 after all items are running.
	panicAtThree := func(ctx context.Context, n int) (int, error) {
		barrier.Done()
		barrier.Wait()

		if n == 3 {
			panic("kaboom")
		}

		<-ctx.Done()

		return 0, ctx.Err()
	}

	_, err := slice.FanOutAll(context.Background(), 4, []int{0, 1, 2, 3}, panicAtThree)

	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("got %T (%v), want *rslt.PanicError (not sibling context.Canceled)", err, err)
	}
}

func TestFanOutAllPanicStackContainsOriginalSite(t *testing.T) {
	// The panic originates from within fn. The stack trace captured by FanOutAll
	// (via debug.Stack in the deferred recover) must include the call site where
	// fn panicked, not just the wrapper's recovery frame.
	//
	// Go renders local func literals as TestName.funcN, so we search for the
	// test function name and the panic message rather than a local variable name.

	// triggerPanic panics with a distinctive message.
	triggerPanic := func(_ context.Context, _ int) (int, error) {
		panic("distinctive_panic_marker")
	}

	_, err := slice.FanOutAll(context.Background(), 1, []int{1}, triggerPanic)

	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("expected PanicError, got %T: %v", err, err)
	}

	stack := string(pe.Stack)

	// The stack must include this test's name (proving the panic site is captured).
	if !strings.Contains(stack, "TestFanOutAllPanicStackContainsOriginalSite") {
		t.Errorf("stack trace missing original panic site:\n%s", stack)
	}
}

func TestFanOutAllCancelsOnError(t *testing.T) {
	var started atomic.Int32

	// sentinel is the error that triggers cancellation.
	sentinel := errors.New("boom")

	// failFirst fails on the first item, others block until cancelled.
	failFirst := func(ctx context.Context, n int) (int, error) {
		started.Add(1)

		if n == 0 {
			return 0, sentinel
		}

		<-ctx.Done()

		return 0, ctx.Err()
	}

	items := make([]int, 20)
	for i := range items {
		items[i] = i
	}

	_, err := slice.FanOutAll(context.Background(), 2, items, failFirst)
	if !errors.Is(err, sentinel) {
		t.Fatalf("got error %v, want %v", err, sentinel)
	}

	// With n=2 and early cancellation, far fewer than 20 should start.
	if got := started.Load(); got > 4 {
		t.Errorf("started %d items, expected at most 4 with early cancellation", got)
	}
}

func TestFanOutAllPanicCancelsRemaining(t *testing.T) {
	var started atomic.Int32

	// panicFirst panics on the first item, others block until cancelled.
	panicFirst := func(ctx context.Context, n int) (int, error) {
		started.Add(1)

		if n == 0 {
			panic("kaboom")
		}

		<-ctx.Done()

		return 0, ctx.Err()
	}

	items := make([]int, 20)
	for i := range items {
		items[i] = i
	}

	_, err := slice.FanOutAll(context.Background(), 2, items, panicFirst)
	if err == nil {
		t.Fatal("expected error from panic")
	}

	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("expected PanicError, got %T: %v", err, err)
	}

	// With n=2 and early cancellation on panic, far fewer than 20 should start.
	if got := started.Load(); got > 4 {
		t.Errorf("started %d items, expected at most 4 with early cancellation on panic", got)
	}
}

func TestFanOutAllEmptyInput(t *testing.T) {
	// identity returns the input unchanged.
	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	got, err := slice.FanOutAll(context.Background(), 4, []int{}, identity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("got len %d, want 0", len(got))
	}
}

func TestFanOutAllPreservesOrder(t *testing.T) {
	// delayByValue yields to force out-of-order completion.
	delayByValue := func(_ context.Context, n int) (int, error) {
		runtime.Gosched()

		return n * 10, nil
	}

	got, err := slice.FanOutAll(context.Background(), 4, []int{3, 1, 4, 1, 5}, delayByValue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{30, 10, 40, 10, 50}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("[%d]: got %d, want %d", i, v, want[i])
		}
	}
}

func TestFanOutAllCallerContextNotCancelled(t *testing.T) {
	// parentCtx must remain valid after FanOutAll returns with an error.
	parentCtx := context.Background()

	// alwaysFail returns an error.
	alwaysFail := func(_ context.Context, _ int) (int, error) {
		return 0, errors.New("fail")
	}

	_, err := slice.FanOutAll(parentCtx, 2, []int{1, 2, 3}, alwaysFail)
	if err == nil {
		t.Fatal("expected error")
	}

	if parentCtx.Err() != nil {
		t.Fatalf("parent context cancelled: %v", parentCtx.Err())
	}
}

func TestFanOutAllAlreadyCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// identity would succeed if called.
	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	_, err := slice.FanOutAll(ctx, 2, []int{1, 2, 3}, identity)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestFanOutAllMultipleConcurrentFailures(t *testing.T) {
	// sentinel is the error we want to see.
	sentinel := errors.New("first")

	// allFail has every item fail, but item 0 fails with sentinel.
	allFail := func(_ context.Context, n int) (int, error) {
		if n == 0 {
			return 0, sentinel
		}

		return 0, errors.New("other")
	}

	_, err := slice.FanOutAll(context.Background(), 1, []int{0, 1, 2}, allFail)

	// With sequential execution (n=1), item 0 fails first.
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want %v", err, sentinel)
	}
}

func TestFanOutAllValidationPanics(t *testing.T) {
	// identity is a valid function for non-panic paths.
	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	t.Run("n<=0", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()

		slice.FanOutAll(context.Background(), 0, []int{1}, identity)
	})

	t.Run("nil ctx", func(t *testing.T) {
		defer func() {
			got, ok := recover().(string)
			if !ok {
				t.Fatal("expected string panic")
			}

			want := "slice.FanOutAll: ctx must not be nil"
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}()

		slice.FanOutAll[int, int](nil, 2, []int{1}, identity) //nolint:staticcheck
	})

	t.Run("nil fn", func(t *testing.T) {
		defer func() {
			got, ok := recover().(string)
			if !ok {
				t.Fatal("expected string panic")
			}

			want := "slice.FanOutAll: fn must not be nil"
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}()

		slice.FanOutAll[int, int](context.Background(), 2, []int{1}, nil)
	})
}

// --- FanOutWeightedAll tests ---

func TestFanOutWeightedAllSuccess(t *testing.T) {
	// double doubles an int.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	got, err := slice.FanOutWeightedAll(context.Background(), 4, []int{1, 2, 3, 4, 5}, unitCost, double)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{2, 4, 6, 8, 10}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("[%d]: got %d, want %d", i, v, want[i])
		}
	}
}

func TestFanOutWeightedAllCancelsOnError(t *testing.T) {
	var started atomic.Int32

	// sentinel is the error that triggers cancellation.
	sentinel := errors.New("boom")

	// failFirst fails on the first item, others block until cancelled.
	failFirst := func(ctx context.Context, n int) (int, error) {
		started.Add(1)

		if n == 0 {
			return 0, sentinel
		}

		<-ctx.Done()

		return 0, ctx.Err()
	}

	items := make([]int, 20)
	for i := range items {
		items[i] = i
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	_, err := slice.FanOutWeightedAll(context.Background(), 2, items, unitCost, failFirst)
	if !errors.Is(err, sentinel) {
		t.Fatalf("got error %v, want %v", err, sentinel)
	}

	if got := started.Load(); got > 4 {
		t.Errorf("started %d items, expected at most 4 with early cancellation", got)
	}
}

func TestFanOutWeightedAllReturnsRootCause(t *testing.T) {
	// sentinel is the real failure at the last index.
	sentinel := errors.New("real cause")

	var barrier sync.WaitGroup
	barrier.Add(4)

	// failAtThree fails at index 3 after all items are running.
	failAtThree := func(ctx context.Context, n int) (int, error) {
		barrier.Done()
		barrier.Wait()

		if n == 3 {
			return 0, sentinel
		}

		<-ctx.Done()

		return 0, ctx.Err()
	}

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	_, err := slice.FanOutWeightedAll(context.Background(), 4, []int{0, 1, 2, 3}, unitCost, failAtThree)
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want root cause %v", err, sentinel)
	}
}

func TestFanOutWeightedAllValidationPanics(t *testing.T) {
	// identity is a valid function.
	identity := func(_ context.Context, n int) (int, error) { return n, nil }

	// unitCost returns 1 for every item.
	unitCost := func(_ int) int { return 1 }

	t.Run("nil ctx", func(t *testing.T) {
		defer func() {
			got, ok := recover().(string)
			if !ok {
				t.Fatal("expected string panic")
			}

			want := "slice.FanOutWeightedAll: ctx must not be nil"
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}()

		slice.FanOutWeightedAll[int, int](nil, 2, []int{1}, unitCost, identity) //nolint:staticcheck
	})

	t.Run("nil fn", func(t *testing.T) {
		defer func() {
			got, ok := recover().(string)
			if !ok {
				t.Fatal("expected string panic")
			}

			want := "slice.FanOutWeightedAll: fn must not be nil"
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}()

		slice.FanOutWeightedAll[int, int](context.Background(), 2, []int{1}, unitCost, nil)
	})
}
