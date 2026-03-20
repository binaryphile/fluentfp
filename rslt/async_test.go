package rslt_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
)

func TestRunAsyncSuccess(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		return 42, nil
	})
	val, err := a.Wait()
	if err != nil || val != 42 {
		t.Errorf("Wait() = (%d, %v), want (42, nil)", val, err)
	}
}

func TestRunAsyncError(t *testing.T) {
	sentinel := errors.New("fail")
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		return 0, sentinel
	})
	_, err := a.Wait()
	if err != sentinel {
		t.Errorf("Wait() error = %v, want %v", err, sentinel)
	}
}

func TestRunAsyncPanicRecovery(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		panic("boom")
	})
	_, err := a.Wait()
	if err == nil {
		t.Fatal("expected error from panic")
	}
	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *PanicError, got %T: %v", err, err)
	}
	if pe.Value != "boom" {
		t.Errorf("panic value = %v, want \"boom\"", pe.Value)
	}
	if len(pe.Stack) == 0 {
		t.Error("expected non-empty stack trace")
	}
}

func TestRunAsyncContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	a := rslt.RunAsync(ctx, func(ctx context.Context) (int, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	})

	cancel()
	_, err := a.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Wait() error = %v, want context.Canceled", err)
	}
}

func TestRunAsyncMultipleWait(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		return 7, nil
	})
	v1, e1 := a.Wait()
	v2, e2 := a.Wait()
	if v1 != v2 || e1 != e2 {
		t.Errorf("Wait returned different values: (%d,%v) vs (%d,%v)", v1, e1, v2, e2)
	}
}

func TestRunAsyncConcurrentWait(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 99, nil
	})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := a.Wait()
			if err != nil || val != 99 {
				t.Errorf("concurrent Wait() = (%d, %v), want (99, nil)", val, err)
			}
		}()
	}
	wg.Wait()
}

func TestRunAsyncDone(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		return 1, nil
	})

	select {
	case <-a.Done():
	case <-time.After(time.Second):
		t.Fatal("Done() did not close within timeout")
	}

	val, err := a.Wait()
	if err != nil || val != 1 {
		t.Errorf("after Done: Wait() = (%d, %v), want (1, nil)", val, err)
	}
}

func TestRunAsyncDoneWithSelect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	slow := rslt.RunAsync(ctx, func(ctx context.Context) (int, error) {
		time.Sleep(50 * time.Millisecond)
		return 1, nil
	})

	select {
	case <-slow.Done():
		val, err := slow.Wait()
		if err != nil || val != 1 {
			t.Errorf("Wait() = (%d, %v), want (1, nil)", val, err)
		}
	case <-ctx.Done():
		t.Fatal("timed out")
	}
}

func TestRunAsyncCopySafety(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})

	// Copy the handle before completion.
	b := a

	// Both should see the same result.
	v1, e1 := a.Wait()
	v2, e2 := b.Wait()
	if v1 != v2 || e1 != e2 {
		t.Errorf("copy diverged: (%d,%v) vs (%d,%v)", v1, e1, v2, e2)
	}
	if v1 != 42 {
		t.Errorf("value = %d, want 42", v1)
	}
}

func TestRunAsyncNilFnPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil fn")
		}
	}()
	rslt.RunAsync[int](context.Background(), nil)
}

func TestRunAsyncNilCtxPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil ctx")
		}
	}()
	rslt.RunAsync[int](nil, func(_ context.Context) (int, error) {
		return 0, nil
	})
}

func TestRunAsyncCopyAfterCompletion(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		return 99, nil
	})
	a.Wait()
	b := a
	v, err := b.Wait()
	if err != nil || v != 99 {
		t.Errorf("copy after completion: (%d, %v), want (99, nil)", v, err)
	}
}

func TestRunAsyncZeroValuePanics(t *testing.T) {
	t.Run("Wait", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
			msg, ok := r.(string)
			if !ok || msg != "rslt.AsyncResult: zero value is invalid; use rslt.RunAsync" {
				t.Errorf("wrong panic: %v", r)
			}
		}()
		var a rslt.AsyncResult[int]
		a.Wait()
	})

	t.Run("Done", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		var a rslt.AsyncResult[int]
		a.Done()
	})

	t.Run("String", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		var a rslt.AsyncResult[int]
		_ = a.String()
	})
}

func TestRunAsyncStringPending(t *testing.T) {
	ch := make(chan struct{})
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		<-ch
		return 42, nil
	})
	s := a.String()
	if s != "AsyncResult(pending)" {
		t.Errorf("pending: %q", s)
	}
	close(ch)
	a.Wait()
}

func TestRunAsyncStringOk(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		return 42, nil
	})
	a.Wait()
	s := a.String()
	if s != "AsyncResult(ok: 42)" {
		t.Errorf("ok: %q", s)
	}
}

func TestRunAsyncStringError(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		return 0, errors.New("fail")
	})
	a.Wait()
	s := a.String()
	if s != "AsyncResult(err: fail)" {
		t.Errorf("error: %q", s)
	}
}

func TestRunAsyncStringPanic(t *testing.T) {
	a := rslt.RunAsync(context.Background(), func(_ context.Context) (int, error) {
		panic("boom")
	})
	a.Wait()
	s := a.String()
	if s != "AsyncResult(err: panic: boom)" {
		t.Errorf("panic: %q", s)
	}
}
