package option_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/rslt"
)

func TestOkOr(t *testing.T) {
	t.Run("ok option returns Ok result", func(t *testing.T) {
		r := option.Of(42).OkOr(errors.New("missing"))
		v, ok := r.Get()
		if !ok || v != 42 {
			t.Errorf("Of(42).OkOr() = (%d, %t), want (42, true)", v, ok)
		}
	})

	t.Run("ok option with nil error does not panic", func(t *testing.T) {
		r := option.Of(42).OkOr(nil)
		v, ok := r.Get()
		if !ok || v != 42 {
			t.Errorf("Of(42).OkOr(nil) = (%d, %t), want (42, true)", v, ok)
		}
	})

	t.Run("ok option preserves zero value", func(t *testing.T) {
		r := option.Of(0).OkOr(errors.New("missing"))
		v, ok := r.Get()
		if !ok || v != 0 {
			t.Errorf("Of(0).OkOr() = (%d, %t), want (0, true)", v, ok)
		}
	})

	t.Run("not-ok option returns Err result", func(t *testing.T) {
		sentinelErr := errors.New("missing")
		r := option.NotOk[int]().OkOr(sentinelErr)
		if r.IsOk() {
			t.Fatal("NotOk.OkOr() should be Err")
		}
		if r.Err() != sentinelErr {
			t.Errorf("NotOk.OkOr().Err() = %v, want %v", r.Err(), sentinelErr)
		}
	})

	t.Run("not-ok with nil error panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic for nil error, got none")
			}
		}()
		option.NotOk[int]().OkOr(nil)
	})
}

func TestOkOrCall(t *testing.T) {
	t.Run("ok option returns Ok without calling fn", func(t *testing.T) {
		called := false
		r := option.Of(42).OkOrCall(func() error {
			called = true
			return errors.New("should not be called")
		})
		if called {
			t.Fatal("fn should not be called when option is ok")
		}
		v, ok := r.Get()
		if !ok || v != 42 {
			t.Errorf("Of(42).OkOrCall() = (%d, %t), want (42, true)", v, ok)
		}
	})

	t.Run("not-ok option calls fn and returns Err", func(t *testing.T) {
		sentinelErr := errors.New("computed error")
		r := option.NotOk[int]().OkOrCall(func() error {
			return sentinelErr
		})
		if r.IsOk() {
			t.Fatal("NotOk.OkOrCall() should be Err")
		}
		if r.Err() != sentinelErr {
			t.Errorf("NotOk.OkOrCall().Err() = %v, want %v", r.Err(), sentinelErr)
		}
	})

	t.Run("nil fn panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic for nil fn, got none")
			}
		}()
		option.Of(42).OkOrCall(nil)
	})

	t.Run("fn returning nil panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic for fn returning nil, got none")
			}
		}()
		option.NotOk[int]().OkOrCall(func() error { return nil })
	})
}

func TestMapResult(t *testing.T) {
	parsePositive := func(s string) rslt.Result[int] {
		n, err := strconv.Atoi(s)
		if err != nil {
			return rslt.Err[int](err)
		}
		return rslt.Ok(n)
	}

	t.Run("absent returns Ok(NotOk)", func(t *testing.T) {
		r := option.MapResult(option.NotOk[string](), parsePositive)
		opt, err := r.Unpack()
		if err != nil {
			t.Fatalf("absent should be Ok, got Err: %v", err)
		}
		if opt.IsOk() {
			t.Fatal("absent inner should be NotOk")
		}
	})

	t.Run("present valid returns Ok(Of(n))", func(t *testing.T) {
		r := option.MapResult(option.Of("42"), parsePositive)
		opt, err := r.Unpack()
		if err != nil {
			t.Fatalf("valid should be Ok, got Err: %v", err)
		}
		v, ok := opt.Get()
		if !ok || v != 42 {
			t.Errorf("valid: got (%d, %t), want (42, true)", v, ok)
		}
	})

	t.Run("present invalid returns Err", func(t *testing.T) {
		r := option.MapResult(option.Of("abc"), parsePositive)
		if r.Err() == nil {
			t.Fatal("invalid should be Err")
		}
	})
}
