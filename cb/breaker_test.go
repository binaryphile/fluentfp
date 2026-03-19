package cb_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/cb"
)

// fakeClock provides deterministic time control for tests.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{t: time.Now()}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.t = c.t.Add(d)
}

func TestWithBreaker(t *testing.T) {
	t.Run("passes through when closed", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
		})
		calls := 0
		// doubleIt doubles the input.
		doubleIt := func(_ context.Context, n int) (int, error) {
			calls++
			return n * 2, nil
		}

		wrapped := cb.WithBreaker(b, doubleIt)
		got, err := wrapped(context.Background(), 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("opens after consecutive failures", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(3),
		})
		// alwaysFail always returns an error.
		alwaysFail := func(_ context.Context, _ int) (int, error) {
			return 0, fmt.Errorf("fail")
		}

		wrapped := cb.WithBreaker(b, alwaysFail)

		for i := 0; i < 3; i++ {
			_, _ = wrapped(context.Background(), 1)
		}

		snap := b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open", snap.State)
		}
	})

	t.Run("returns ErrOpen when open", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
		})
		// alwaysFail always returns an error.
		alwaysFail := func(_ context.Context, _ int) (int, error) {
			return 0, fmt.Errorf("fail")
		}

		wrapped := cb.WithBreaker(b, alwaysFail)
		_, _ = wrapped(context.Background(), 1) // trips breaker

		_, err := wrapped(context.Background(), 1)
		if !errors.Is(err, cb.ErrOpen) {
			t.Fatalf("got error %v, want ErrOpen", err)
		}
	})

	t.Run("transitions to half-open after reset timeout", func(t *testing.T) {
		clock := newFakeClock()
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: 30 * time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			Clock:        clock.Now,
		})
		calls := 0
		// succeedAfterTrip succeeds on all calls after the first.
		succeedAfterTrip := func(_ context.Context, n int) (int, error) {
			calls++
			if calls == 1 {
				return 0, fmt.Errorf("fail")
			}
			return n * 2, nil
		}

		wrapped := cb.WithBreaker(b, succeedAfterTrip)
		_, _ = wrapped(context.Background(), 1) // trips breaker

		// Still open before timeout.
		clock.Advance(29 * time.Second)
		_, err := wrapped(context.Background(), 1)
		if !errors.Is(err, cb.ErrOpen) {
			t.Fatalf("expected ErrOpen before timeout, got %v", err)
		}

		// After timeout, should transition to half-open and admit probe.
		clock.Advance(2 * time.Second)
		got, err := wrapped(context.Background(), 5)
		if err != nil {
			t.Fatalf("unexpected error on probe: %v", err)
		}
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}
	})

	t.Run("closes on half-open success", func(t *testing.T) {
		clock := newFakeClock()
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			Clock:        clock.Now,
		})
		calls := 0
		// failOnceThenSucceed fails once, then succeeds.
		failOnceThenSucceed := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls == 1 {
				return 0, fmt.Errorf("fail")
			}
			return 42, nil
		}

		wrapped := cb.WithBreaker(b, failOnceThenSucceed)
		_, _ = wrapped(context.Background(), 1) // trip

		clock.Advance(2 * time.Second)
		_, _ = wrapped(context.Background(), 1) // half-open probe succeeds

		snap := b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed", snap.State)
		}
		if snap.Failures != 0 {
			t.Fatalf("failures = %d, want 0 (reset on close)", snap.Failures)
		}
		if !snap.OpenedAt.IsZero() {
			t.Fatalf("OpenedAt = %v, want zero (cleared on close)", snap.OpenedAt)
		}
	})

	t.Run("reopens on half-open failure", func(t *testing.T) {
		clock := newFakeClock()
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			Clock:        clock.Now,
		})
		// alwaysFail always returns an error.
		alwaysFail := func(_ context.Context, _ int) (int, error) {
			return 0, fmt.Errorf("fail")
		}

		wrapped := cb.WithBreaker(b, alwaysFail)
		_, _ = wrapped(context.Background(), 1) // trip

		clock.Advance(2 * time.Second)
		_, _ = wrapped(context.Background(), 1) // half-open probe fails

		snap := b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open", snap.State)
		}
	})

	t.Run("rejects concurrent requests during half-open", func(t *testing.T) {
		clock := newFakeClock()
		probeStarted := make(chan struct{})
		probeFinish := make(chan struct{})
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			Clock:        clock.Now,
		})
		// slowFn blocks until signaled.
		slowFn := func(_ context.Context, _ int) (int, error) {
			close(probeStarted)
			<-probeFinish
			return 42, nil
		}

		wrapped := cb.WithBreaker(b, slowFn)

		// Trip with a quick failure first.
		quickFail := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
			return 0, fmt.Errorf("fail")
		})
		_, _ = quickFail(context.Background(), 1)

		clock.Advance(2 * time.Second)

		// Launch probe in background.
		go func() {
			_, _ = wrapped(context.Background(), 1)
		}()
		<-probeStarted

		// Second request should be rejected.
		_, err := wrapped(context.Background(), 1)
		if !errors.Is(err, cb.ErrOpen) {
			t.Fatalf("got %v, want ErrOpen", err)
		}

		close(probeFinish)
	})

	t.Run("context cancellation does not count as failure", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
		})
		// cancellingFn returns context.Canceled.
		cancellingFn := func(_ context.Context, _ int) (int, error) {
			return 0, context.Canceled
		}

		wrapped := cb.WithBreaker(b, cancellingFn)
		_, _ = wrapped(context.Background(), 1)

		snap := b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed (context.Canceled should not trip)", snap.State)
		}
		if snap.Failures != 0 {
			t.Fatalf("failures = %d, want 0", snap.Failures)
		}
	})

	t.Run("context.Canceled during half-open probe releases probe slot", func(t *testing.T) {
		clock := newFakeClock()
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			Clock:        clock.Now,
		})
		calls := 0
		// cancelThenSucceed returns context.Canceled on first two calls, then succeeds.
		cancelThenSucceed := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls <= 2 {
				return 0, context.Canceled
			}
			return 42, nil
		}

		// Trip with a real failure.
		tripped := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
			return 0, fmt.Errorf("fail")
		})
		_, _ = tripped(context.Background(), 1)

		wrapped := cb.WithBreaker(b, cancelThenSucceed)

		// First probe returns context.Canceled — should release slot, not count as failure.
		clock.Advance(2 * time.Second)
		_, _ = wrapped(context.Background(), 1)

		snap := b.Snapshot()
		if snap.State != cb.StateHalfOpen {
			t.Fatalf("state = %v, want half-open (context.Canceled should not reopen)", snap.State)
		}

		// Second probe also cancelled — slot should be released again.
		_, _ = wrapped(context.Background(), 1)

		// Third probe succeeds — should close.
		got, err := wrapped(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42 {
			t.Fatalf("got %d, want 42", got)
		}

		snap = b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed", snap.State)
		}
	})

	t.Run("context.DeadlineExceeded counts as failure by default", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
		})
		// timeoutFn returns context.DeadlineExceeded.
		timeoutFn := func(_ context.Context, _ int) (int, error) {
			return 0, context.DeadlineExceeded
		}

		wrapped := cb.WithBreaker(b, timeoutFn)
		_, _ = wrapped(context.Background(), 1)

		snap := b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open (DeadlineExceeded should trip)", snap.State)
		}
	})

	t.Run("context cancelled before call", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		calls := 0
		// doubleIt doubles the input.
		doubleIt := func(_ context.Context, n int) (int, error) {
			calls++
			return n * 2, nil
		}

		wrapped := cb.WithBreaker(b, doubleIt)
		_, err := wrapped(ctx, 1)

		if err != context.Canceled {
			t.Fatalf("got %v, want context.Canceled", err)
		}
		if calls != 0 {
			t.Fatalf("fn called %d times, want 0", calls)
		}
	})

	t.Run("success resets consecutive failure count", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(3),
		})
		calls := 0
		// failTwiceThenSucceed fails twice, then succeeds.
		failTwiceThenSucceed := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls <= 2 {
				return 0, fmt.Errorf("fail")
			}
			return 42, nil
		}

		wrapped := cb.WithBreaker(b, failTwiceThenSucceed)
		_, _ = wrapped(context.Background(), 1) // fail 1
		_, _ = wrapped(context.Background(), 1) // fail 2
		_, _ = wrapped(context.Background(), 1) // success, resets consecutive

		snap := b.Snapshot()
		if snap.ConsecutiveFailures != 0 {
			t.Fatalf("consecutive failures = %d, want 0 after success", snap.ConsecutiveFailures)
		}
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed", snap.State)
		}
	})

	t.Run("ignored errors do not break consecutive streak", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(3),
			ShouldCount:  func(err error) bool { return err.Error() != "ignored" },
		})
		calls := 0
		// failIgnoreFail returns counted, ignored, counted errors.
		failIgnoreFail := func(_ context.Context, _ int) (int, error) {
			calls++
			switch calls {
			case 1:
				return 0, fmt.Errorf("counted")
			case 2:
				return 0, fmt.Errorf("ignored")
			case 3:
				return 0, fmt.Errorf("counted")
			default:
				return 0, fmt.Errorf("counted")
			}
		}

		wrapped := cb.WithBreaker(b, failIgnoreFail)
		_, _ = wrapped(context.Background(), 1) // counted fail, consecutive=1
		_, _ = wrapped(context.Background(), 1) // ignored, consecutive stays 1
		_, _ = wrapped(context.Background(), 1) // counted fail, consecutive=2

		snap := b.Snapshot()
		if snap.ConsecutiveFailures != 2 {
			t.Fatalf("consecutive = %d, want 2 (ignored errors are transparent)", snap.ConsecutiveFailures)
		}
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed (only 2 consecutive, need 3)", snap.State)
		}
	})
}

func TestWithBreakerShouldCount(t *testing.T) {
	errTransient := fmt.Errorf("transient")
	errPermanent := fmt.Errorf("permanent")

	t.Run("uncounted errors do not trip breaker", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			ShouldCount:  func(err error) bool { return err != errPermanent },
		})
		// returnPermanent always returns a permanent error.
		returnPermanent := func(_ context.Context, _ int) (int, error) {
			return 0, errPermanent
		}

		wrapped := cb.WithBreaker(b, returnPermanent)
		_, err := wrapped(context.Background(), 1)

		if err != errPermanent {
			t.Fatalf("got %v, want %v", err, errPermanent)
		}

		snap := b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed (permanent errors not counted)", snap.State)
		}
	})

	t.Run("counted errors trip breaker", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			ShouldCount:  func(err error) bool { return err == errTransient },
		})
		// returnTransient always returns a transient error.
		returnTransient := func(_ context.Context, _ int) (int, error) {
			return 0, errTransient
		}

		wrapped := cb.WithBreaker(b, returnTransient)
		_, _ = wrapped(context.Background(), 1)

		snap := b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open", snap.State)
		}
	})

	t.Run("uncounted error releases half-open probe", func(t *testing.T) {
		clock := newFakeClock()
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			ShouldCount:  func(err error) bool { return err != errPermanent },
			Clock:        clock.Now,
		})
		calls := 0
		// permanentThenSuccess returns permanent error on first call, then succeeds.
		permanentThenSuccess := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls == 1 {
				return 0, errPermanent
			}
			return 42, nil
		}

		wrapped := cb.WithBreaker(b, permanentThenSuccess)
		// Trip with a counted error first.
		countedFail := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
			return 0, errTransient
		})
		_, _ = countedFail(context.Background(), 1)

		clock.Advance(2 * time.Second)

		// Half-open probe returns uncounted error — should release probe slot.
		_, _ = wrapped(context.Background(), 1)

		// Next request should also be admitted as a new probe (slot was released).
		got, err := wrapped(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error on second probe: %v", err)
		}
		if got != 42 {
			t.Fatalf("got %d, want 42", got)
		}
	})
}

func TestWithBreakerOnStateChange(t *testing.T) {
	t.Run("fires on all transitions", func(t *testing.T) {
		clock := newFakeClock()
		var transitions []cb.Transition
		var mu sync.Mutex
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			Clock:        clock.Now,
			OnStateChange: func(t cb.Transition) {
				mu.Lock()
				transitions = append(transitions, t)
				mu.Unlock()
			},
		})
		calls := 0
		// failOnceThenSucceed fails once, then succeeds.
		failOnceThenSucceed := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls == 1 {
				return 0, fmt.Errorf("fail")
			}
			return 42, nil
		}

		wrapped := cb.WithBreaker(b, failOnceThenSucceed)

		// Closed → Open.
		_, _ = wrapped(context.Background(), 1)

		// Open → HalfOpen → Closed.
		clock.Advance(2 * time.Second)
		_, _ = wrapped(context.Background(), 1)

		mu.Lock()
		defer mu.Unlock()

		if len(transitions) != 3 {
			t.Fatalf("got %d transitions, want 3", len(transitions))
		}
		if transitions[0].From != cb.StateClosed || transitions[0].To != cb.StateOpen {
			t.Fatalf("transition 0: %v→%v, want closed→open", transitions[0].From, transitions[0].To)
		}
		if transitions[1].From != cb.StateOpen || transitions[1].To != cb.StateHalfOpen {
			t.Fatalf("transition 1: %v→%v, want open→half-open", transitions[1].From, transitions[1].To)
		}
		if transitions[2].From != cb.StateHalfOpen || transitions[2].To != cb.StateClosed {
			t.Fatalf("transition 2: %v→%v, want half-open→closed", transitions[2].From, transitions[2].To)
		}
	})
}

func TestWithBreakerComposition(t *testing.T) {
	t.Run("composes with Retry wrapping breaker", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(5),
		})
		calls := 0
		// failOnceThenDouble fails once, then doubles.
		failOnceThenDouble := func(_ context.Context, n int) (int, error) {
			calls++
			if calls < 2 {
				return 0, fmt.Errorf("fail")
			}
			return n * 2, nil
		}

		shouldRetry := func(err error) bool {
			return !errors.Is(err, cb.ErrOpen)
		}
		composed := cb.Retry(3, cb.ConstantBackoff(0), shouldRetry,
			cb.WithBreaker(b, failOnceThenDouble))
		got, err := composed(context.Background(), 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}
	})

	t.Run("composes with breaker wrapping Retry", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(2),
		})
		calls := 0
		// failOnceThenDouble fails once, then doubles.
		failOnceThenDouble := func(_ context.Context, n int) (int, error) {
			calls++
			if calls < 2 {
				return 0, fmt.Errorf("fail")
			}
			return n * 2, nil
		}

		composed := cb.WithBreaker(b,
			cb.Retry(3, cb.ConstantBackoff(0), nil, failOnceThenDouble))
		got, err := composed(context.Background(), 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}

		snap := b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed", snap.State)
		}
	})

	t.Run("composes with Throttle", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(5),
		})
		// doubleIt doubles the input.
		doubleIt := func(_ context.Context, n int) (int, error) {
			return n * 2, nil
		}

		composed := cb.WithBreaker(b, cb.Throttle(2, doubleIt))
		got, err := composed(context.Background(), 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 10 {
			t.Fatalf("got %d, want 10", got)
		}
	})
}

func TestWithBreakerRejectedCount(t *testing.T) {
	clock := newFakeClock()
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(1),
		Clock:        clock.Now,
	})
	// alwaysFail always returns an error.
	alwaysFail := func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fail")
	}

	wrapped := cb.WithBreaker(b, alwaysFail)
	_, _ = wrapped(context.Background(), 1) // trip

	// These should all be rejected.
	for i := 0; i < 5; i++ {
		_, _ = wrapped(context.Background(), 1)
	}

	snap := b.Snapshot()
	if snap.Rejected != 5 {
		t.Fatalf("rejected = %d, want 5", snap.Rejected)
	}
}

func TestWithBreakerConcurrency(t *testing.T) {
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(100), // high threshold to avoid tripping
	})
	var count atomic.Int32
	// incrementer increments and returns the count.
	incrementer := func(_ context.Context, _ int) (int, error) {
		count.Add(1)
		return int(count.Load()), nil
	}

	wrapped := cb.WithBreaker(b, incrementer)

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := wrapped(context.Background(), 1)
			if err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := count.Load(); got != 100 {
		t.Fatalf("fn called %d times, want 100", got)
	}
}

func TestNewBreakerPanics(t *testing.T) {
	t.Run("zero ResetTimeout", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		cb.NewBreaker(cb.BreakerConfig{ResetTimeout: 0})
	})

	t.Run("negative ResetTimeout", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		cb.NewBreaker(cb.BreakerConfig{ResetTimeout: -time.Second})
	})
}

func TestWithBreakerPanics(t *testing.T) {
	b := cb.NewBreaker(cb.BreakerConfig{ResetTimeout: time.Second})
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }

	t.Run("nil breaker", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		cb.WithBreaker[int, int](nil, doubleIt)
	})

	t.Run("nil fn", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		cb.WithBreaker[int, int](b, nil)
	})
}

func TestConsecutiveFailuresPanics(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		cb.ConsecutiveFailures(0)
	})

	t.Run("negative", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		cb.ConsecutiveFailures(-1)
	})
}

func TestWithBreakerSharedBreaker(t *testing.T) {
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(3),
	})

	failCount := 0
	// failingFn always returns an error.
	failingFn := func(_ context.Context, _ string) (string, error) {
		failCount++
		return "", fmt.Errorf("fail")
	}
	// successFn always succeeds.
	successFn := func(_ context.Context, n int) (int, error) {
		return n * 2, nil
	}

	wrappedA := cb.WithBreaker(b, failingFn)
	wrappedB := cb.WithBreaker(b, successFn)

	// Failures from wrappedA count toward the shared breaker.
	for i := 0; i < 3; i++ {
		_, _ = wrappedA(context.Background(), "x")
	}

	// wrappedB should also be blocked — same breaker is open.
	_, err := wrappedB(context.Background(), 5)
	if !errors.Is(err, cb.ErrOpen) {
		t.Fatalf("got %v, want ErrOpen (shared breaker should block both)", err)
	}
}

func TestBreakerStateString(t *testing.T) {
	tests := []struct {
		state cb.BreakerState
		want  string
	}{
		{cb.StateClosed, "closed"},
		{cb.StateOpen, "open"},
		{cb.StateHalfOpen, "half-open"},
		{cb.BreakerState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Fatalf("BreakerState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

// --- Epoch/generation tests ---

func TestWithBreakerStaleResultsIgnored(t *testing.T) {
	t.Run("stale failure from prior closed epoch does not affect new epoch", func(t *testing.T) {
		clock := newFakeClock()
		staleStarted := make(chan struct{})
		staleFinish := make(chan struct{})

		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(2),
			Clock:        clock.Now,
		})

		calls := 0
		// slowThenFail blocks on first call, fails fast on second, succeeds on third+.
		slowThenFail := func(_ context.Context, _ int) (int, error) {
			calls++
			switch calls {
			case 1:
				close(staleStarted)
				<-staleFinish
				return 0, fmt.Errorf("stale fail")
			case 2, 3:
				return 0, fmt.Errorf("fast fail")
			default:
				return 42, nil
			}
		}

		wrapped := cb.WithBreaker(b, slowThenFail)

		// Launch slow call A (gen 0, closed).
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = wrapped(context.Background(), 1)
		}()
		<-staleStarted

		// Fast calls B and C trip the breaker (gen 0 → gen 1: closed→open).
		_, _ = wrapped(context.Background(), 1) // calls=2, fail
		_, _ = wrapped(context.Background(), 1) // calls=3, fail → trips

		snap := b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open", snap.State)
		}

		// Wait for timeout, probe succeeds (gen 1→2: open→half-open, gen 2→3: half-open→closed).
		clock.Advance(2 * time.Second)
		_, err := wrapped(context.Background(), 1) // calls=4, succeeds
		if err != nil {
			t.Fatalf("probe error: %v", err)
		}

		snap = b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed after probe", snap.State)
		}

		// Release stale call A and wait for it to complete.
		close(staleFinish)
		wg.Wait()

		// Breaker should still be closed — stale result ignored.
		snap = b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed (stale failure should be ignored)", snap.State)
		}
		if snap.Failures != 0 {
			t.Fatalf("failures = %d, want 0 (stale failure should not be counted)", snap.Failures)
		}
		if snap.ConsecutiveFailures != 0 {
			t.Fatalf("consecutive = %d, want 0 (stale failure should not affect streak)", snap.ConsecutiveFailures)
		}
	})

	t.Run("stale success while open does not reset streak", func(t *testing.T) {
		clock := newFakeClock()
		staleStarted := make(chan struct{})
		staleFinish := make(chan struct{})

		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(2),
			Clock:        clock.Now,
		})

		calls := 0
		// slowSuccess blocks on first call (success), fails on subsequent.
		slowSuccess := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls == 1 {
				close(staleStarted)
				<-staleFinish
				return 42, nil
			}
			return 0, fmt.Errorf("fail")
		}

		wrapped := cb.WithBreaker(b, slowSuccess)

		// Launch slow success A (gen 0, closed).
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = wrapped(context.Background(), 1)
		}()
		<-staleStarted

		// Trip the breaker with fast failures.
		_, _ = wrapped(context.Background(), 1) // calls=2, fail
		_, _ = wrapped(context.Background(), 1) // calls=3, fail → trips

		snap := b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open", snap.State)
		}

		// Release stale success and wait for goroutine to complete.
		close(staleFinish)
		wg.Wait()

		snap = b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open (stale success should not close breaker)", snap.State)
		}
	})
}

// --- Panic safety tests ---

func TestWithBreakerPanicInFn(t *testing.T) {
	t.Run("half-open probe panic reopens breaker", func(t *testing.T) {
		clock := newFakeClock()
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
			Clock:        clock.Now,
		})
		calls := 0
		// panicAfterTrip panics on the second call (the half-open probe).
		panicAfterTrip := func(_ context.Context, _ int) (int, error) {
			calls++
			if calls == 1 {
				return 0, fmt.Errorf("fail")
			}
			panic("boom")
		}

		wrapped := cb.WithBreaker(b, panicAfterTrip)
		_, _ = wrapped(context.Background(), 1) // trip

		clock.Advance(2 * time.Second)

		func() {
			defer func() { recover() }()
			_, _ = wrapped(context.Background(), 1)
		}()

		snap := b.Snapshot()
		if snap.State != cb.StateOpen {
			t.Fatalf("state = %v, want open after panic in half-open probe", snap.State)
		}
	})

	t.Run("closed-state panic does not affect breaker", func(t *testing.T) {
		b := cb.NewBreaker(cb.BreakerConfig{
			ResetTimeout: time.Second,
			ReadyToTrip:  cb.ConsecutiveFailures(1),
		})
		// panicker always panics.
		panicker := func(_ context.Context, _ int) (int, error) {
			panic("boom")
		}

		wrapped := cb.WithBreaker(b, panicker)

		func() {
			defer func() { recover() }()
			_, _ = wrapped(context.Background(), 1)
		}()

		snap := b.Snapshot()
		if snap.State != cb.StateClosed {
			t.Fatalf("state = %v, want closed after panic in closed state", snap.State)
		}
		if snap.Failures != 0 {
			t.Fatalf("failures = %d, want 0 (closed-state panics should not count)", snap.Failures)
		}
	})
}

func TestWithBreakerReadyToTripReentrancy(t *testing.T) {
	var b *cb.Breaker
	b = cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip: func(s cb.Snapshot) bool {
			// Call b.Snapshot() from inside ReadyToTrip.
			// This would deadlock if ReadyToTrip were called under the lock.
			snap := b.Snapshot()
			return snap.ConsecutiveFailures >= 3
		},
	})
	// alwaysFail always returns an error.
	alwaysFail := func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fail")
	}

	wrapped := cb.WithBreaker(b, alwaysFail)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 3; i++ {
			_, _ = wrapped(context.Background(), 1)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deadlock: ReadyToTrip calling b.Snapshot() should not deadlock")
	}

	snap := b.Snapshot()
	if snap.State != cb.StateOpen {
		t.Fatalf("state = %v, want open", snap.State)
	}
}

func TestWithBreakerReadyToTripPanicDoesNotDeadlock(t *testing.T) {
	panicked := false
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip: func(cb.Snapshot) bool {
			if !panicked {
				panicked = true
				panic("bad predicate")
			}
			return false
		},
	})
	// alwaysFail always returns an error.
	alwaysFail := func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fail")
	}

	wrapped := cb.WithBreaker(b, alwaysFail)

	// First call panics in ReadyToTrip.
	func() {
		defer func() { recover() }()
		_, _ = wrapped(context.Background(), 1)
	}()

	// Second call must not deadlock.
	done := make(chan struct{})
	go func() {
		defer func() { recover() }()
		_, _ = wrapped(context.Background(), 1)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deadlock: mutex not released after ReadyToTrip panic")
	}
}

func TestWithBreakerShouldCountPanicReleasesProbe(t *testing.T) {
	clock := newFakeClock()

	// Use a flag to control ShouldCount behavior.
	shouldCountPanics := false
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(1),
		Clock:        clock.Now,
		ShouldCount: func(err error) bool {
			if shouldCountPanics {
				panic("bad ShouldCount")
			}
			return true
		},
	})

	wrapped := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fail")
	})

	// Trip the breaker normally.
	_, _ = wrapped(context.Background(), 1)

	snap := b.Snapshot()
	if snap.State != cb.StateOpen {
		t.Fatalf("state = %v, want open", snap.State)
	}

	// Enable ShouldCount panics for the half-open probe.
	shouldCountPanics = true
	clock.Advance(2 * time.Second)

	// Half-open probe: fn returns error, ShouldCount panics.
	func() {
		defer func() { recover() }()
		_, _ = wrapped(context.Background(), 1)
	}()

	// Probe slot should be released — next call should be admitted as a new probe.
	shouldCountPanics = false
	_, err := wrapped(context.Background(), 1)

	// If probe slot was leaked, this returns ErrOpen (breaker wedged).
	// Being admitted proves the slot was released.
	if errors.Is(err, cb.ErrOpen) {
		t.Fatal("probe slot leaked: got ErrOpen, want admission")
	}

	snap = b.Snapshot()
	if snap.State != cb.StateOpen {
		t.Fatalf("state = %v, want open (probe failure should reopen)", snap.State)
	}
}

func TestWithBreakerOnStateChangePanicOnAdmission(t *testing.T) {
	clock := newFakeClock()
	callbackPanics := false
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(1),
		Clock:        clock.Now,
		OnStateChange: func(cb.Transition) {
			if callbackPanics {
				panic("bad callback")
			}
		},
	})

	wrapped := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fail")
	})

	// Trip the breaker normally.
	_, _ = wrapped(context.Background(), 1)

	// Enable callback panics for the open→half-open transition.
	callbackPanics = true
	clock.Advance(2 * time.Second)

	// Admission triggers open→half-open, OnStateChange panics.
	func() {
		defer func() { recover() }()
		_, _ = wrapped(context.Background(), 1)
	}()

	// The probe finalization guard should have recorded a failure,
	// reopening the breaker. Probe slot should not be leaked.
	callbackPanics = false
	clock.Advance(2 * time.Second)

	// Next call should be admitted (not wedged).
	done := make(chan struct{})
	go func() {
		defer func() { recover() }()
		_, _ = wrapped(context.Background(), 1)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("breaker wedged after OnStateChange panic on admission")
	}
}

func TestWithBreakerOnStateChangePanicAfterCompletion(t *testing.T) {
	clock := newFakeClock()
	callCount := 0
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(1),
		Clock:        clock.Now,
		OnStateChange: func(t cb.Transition) {
			callCount++
			// Panic on the closed→open transition (completion path).
			if callCount == 1 {
				panic("bad callback on trip")
			}
		},
	})

	wrapped := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
		return 0, fmt.Errorf("fail")
	})

	// Trip — OnStateChange panics on the closed→open transition.
	func() {
		defer func() { recover() }()
		_, _ = wrapped(context.Background(), 1)
	}()

	// State should be open — transition was committed before the callback.
	snap := b.Snapshot()
	if snap.State != cb.StateOpen {
		t.Fatalf("state = %v, want open (transition committed before callback panic)", snap.State)
	}
}

func TestWithBreakerExactTimeoutBoundary(t *testing.T) {
	clock := newFakeClock()
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(1),
		Clock:        clock.Now,
	})
	calls := 0
	wrapped := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
		calls++
		if calls == 1 {
			return 0, fmt.Errorf("fail")
		}
		return 42, nil
	})

	// Trip.
	_, _ = wrapped(context.Background(), 1)

	// Advance exactly to the boundary (openedAt + resetTimeout).
	clock.Advance(time.Second)

	// Should transition to half-open and admit probe.
	got, err := wrapped(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected admission at exact boundary, got error: %v", err)
	}
	if got != 42 {
		t.Fatalf("got %d, want 42", got)
	}

	snap := b.Snapshot()
	if snap.State != cb.StateClosed {
		t.Fatalf("state = %v, want closed", snap.State)
	}
}

func TestWithBreakerOpenedAtSemantics(t *testing.T) {
	clock := newFakeClock()
	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip:  cb.ConsecutiveFailures(1),
		Clock:        clock.Now,
	})

	// Initially closed — OpenedAt should be zero.
	snap := b.Snapshot()
	if !snap.OpenedAt.IsZero() {
		t.Fatalf("OpenedAt = %v, want zero when closed", snap.OpenedAt)
	}

	calls := 0
	wrapped := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
		calls++
		if calls == 1 {
			return 0, fmt.Errorf("fail")
		}
		return 42, nil
	})

	// Trip — OpenedAt should be set.
	tripTime := clock.Now()
	_, _ = wrapped(context.Background(), 1)

	snap = b.Snapshot()
	if snap.OpenedAt != tripTime {
		t.Fatalf("OpenedAt = %v, want %v", snap.OpenedAt, tripTime)
	}

	// Close via probe — OpenedAt should be zero again.
	clock.Advance(2 * time.Second)
	_, _ = wrapped(context.Background(), 1)

	snap = b.Snapshot()
	if snap.State != cb.StateClosed {
		t.Fatalf("state = %v, want closed", snap.State)
	}
	if !snap.OpenedAt.IsZero() {
		t.Fatalf("OpenedAt = %v, want zero after closing", snap.OpenedAt)
	}
}

func TestWithBreakerTOCTOUInvalidation(t *testing.T) {
	// Reproduces the TOCTOU race: goroutine A records a failure that would trip
	// (consecutiveFails reaches threshold), but before tripIfCurrent commits,
	// goroutine B records a success that resets consecutiveFails. The breaker
	// must NOT open because the trip decision was based on stale metrics.
	//
	// We use a custom ReadyToTrip that blocks to create the interleaving window.

	tripCheck := make(chan struct{})
	tripResume := make(chan struct{})

	b := cb.NewBreaker(cb.BreakerConfig{
		ResetTimeout: time.Second,
		ReadyToTrip: func(s cb.Snapshot) bool {
			if s.ConsecutiveFailures >= 3 {
				// Signal that we're in the trip check, then block.
				// This creates the window for a concurrent success.
				tripCheck <- struct{}{}
				<-tripResume

				return s.ConsecutiveFailures >= 3
			}
			return false
		},
	})

	var calls atomic.Int32
	wrapped := cb.WithBreaker(b, func(_ context.Context, _ int) (int, error) {
		n := calls.Add(1)
		if n <= 3 {
			return 0, fmt.Errorf("fail %d", n)
		}
		return 42, nil
	})

	// Rack up 2 failures (sequential — no race on calls).
	_, _ = wrapped(context.Background(), 1) // fail 1
	_, _ = wrapped(context.Background(), 1) // fail 2

	// Goroutine A: records fail 3, triggers ReadyToTrip which blocks.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = wrapped(context.Background(), 1) // fail 3 → ReadyToTrip blocks
	}()

	// Wait for ReadyToTrip to be called.
	<-tripCheck

	// Goroutine B: success resets consecutiveFails while A is blocked in ReadyToTrip.
	_, err := wrapped(context.Background(), 1) // calls=4, success
	if err != nil {
		t.Fatalf("concurrent success failed: %v", err)
	}

	// Resume A — ReadyToTrip returns true, but tripIfCurrent detects
	// the metricsVersion mismatch and aborts the trip.
	close(tripResume)
	wg.Wait()

	snap := b.Snapshot()
	if snap.State != cb.StateClosed {
		t.Fatalf("state = %v, want closed (trip should have been invalidated by concurrent success)", snap.State)
	}
	if snap.ConsecutiveFailures != 0 {
		t.Fatalf("consecutive = %d, want 0 (success should have reset)", snap.ConsecutiveFailures)
	}
}
