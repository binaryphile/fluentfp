package slice_test

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/binaryphile/fluentfp/result"
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
		checkValue func(t *testing.T, pe *result.PanicError)
	}{
		{
			name:       "string panic",
			panicValue: "boom",
			checkValue: func(t *testing.T, pe *result.PanicError) {
				if pe.Value != "boom" {
					t.Errorf("PanicError.Value: got %v, want \"boom\"", pe.Value)
				}
			},
		},
		{
			name:       "error panic",
			panicValue: errors.New("wrapped error"),
			checkValue: func(t *testing.T, pe *result.PanicError) {
				if _, ok := pe.Value.(error); !ok {
					t.Error("PanicError.Value: expected error type")
				}
			},
		},
		{
			name:       "nil panic produces PanicNilError",
			panicValue: nil, // Go 1.21+: recover() returns *runtime.PanicNilError
			checkValue: func(t *testing.T, pe *result.PanicError) {
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

			var pe *result.PanicError
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

	done := make(chan []result.Result[int])

	go func() {
		got := slice.FanOut(ctx, workers, input, blockUntilGate)
		done <- []result.Result[int](got)
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

	var pe *result.PanicError
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

		var pe *result.PanicError
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

	var pe *result.PanicError
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
