package option_test

import (
	"errors"
	"testing"

	"github.com/binaryphile/fluentfp/option"
)

func TestOkOr(t *testing.T) {
	t.Run("ok option returns Ok result", func(t *testing.T) {
		opt := option.Of(42)
		r := option.OkOr(opt, errors.New("missing"))
		v, ok := r.Get()
		if !ok || v != 42 {
			t.Errorf("OkOr(Of(42)) = (%d, %t), want (42, true)", v, ok)
		}
	})

	t.Run("not-ok option returns Err result", func(t *testing.T) {
		opt := option.NotOk[int]()
		sentinelErr := errors.New("missing")
		r := option.OkOr(opt, sentinelErr)
		if r.IsOk() {
			t.Fatal("OkOr(NotOk) should be Err")
		}
		if r.Err() != sentinelErr {
			t.Errorf("OkOr(NotOk).Err() = %v, want %v", r.Err(), sentinelErr)
		}
	})

	t.Run("nil error panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic for nil error, got none")
			}
		}()
		option.OkOr(option.NotOk[int](), nil)
	})
}
