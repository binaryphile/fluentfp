package rslt_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/binaryphile/fluentfp/rslt"
)

func TestErrNilPanics(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, got none")
		}

		got, ok := v.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", v, v)
		}

		// expectedMsg is the panic message for Err(nil).
		expectedMsg := "rslt.Err: error must not be nil"
		if got != expectedMsg {
			t.Errorf("got %q, want %q", got, expectedMsg)
		}
	}()

	rslt.Err[int](nil)
}

func TestOf(t *testing.T) {
	// sentinelErr is a specific error for identity checking.
	sentinelErr := errors.New("fail")

	tests := []struct {
		name    string
		value   int
		err     error
		wantVal int
		wantOk  bool
	}{
		{
			name:    "nil error returns ok",
			value:   42,
			err:     nil,
			wantVal: 42,
			wantOk:  true,
		},
		{
			name:    "non-nil error returns err",
			value:   99,
			err:     sentinelErr,
			wantVal: 0,
			wantOk:  false,
		},
		{
			name:    "zero value with nil error returns ok",
			value:   0,
			err:     nil,
			wantVal: 0,
			wantOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rslt.Of(tt.value, tt.err)
			val, ok := got.Get()
			if val != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Of: got (%d, %t), want (%d, %t)", val, ok, tt.wantVal, tt.wantOk)
			}
		})
	}

	t.Run("preserves error identity", func(t *testing.T) {
		got := rslt.Of(0, sentinelErr)
		err, ok := got.GetErr()
		if !ok {
			t.Fatal("expected Err result")
		}
		if !errors.Is(err, sentinelErr) {
			t.Errorf("Of: error identity not preserved")
		}
	})
}

func TestTransform(t *testing.T) {
	// double doubles an int.
	double := func(n int) int { return n * 2 }

	tests := []struct {
		name    string
		result  rslt.Result[int]
		wantVal int
		wantOk  bool
	}{
		{
			name:    "ok transforms",
			result:  rslt.Ok(5),
			wantVal: 10,
			wantOk:  true,
		},
		{
			name:    "err passes through",
			result:  rslt.Err[int](errors.New("fail")),
			wantVal: 0,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.Transform(double)
			val, ok := got.Get()
			if val != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Transform: got (%d, %t), want (%d, %t)", val, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestMustGet(t *testing.T) {
	t.Run("ok returns value", func(t *testing.T) {
		got := rslt.Ok(42).MustGet()
		if got != 42 {
			t.Errorf("MustGet: got %d, want 42", got)
		}
	})

	t.Run("err panics with wrapped error", func(t *testing.T) {
		originalErr := errors.New("fail")

		defer func() {
			v := recover()
			if v == nil {
				t.Fatal("expected panic, got none")
			}

			err, ok := v.(error)
			if !ok {
				t.Fatalf("expected error panic, got %T: %v", v, v)
			}

			if !errors.Is(err, originalErr) {
				t.Errorf("panic error does not wrap original: got %v", err)
			}

			if !strings.Contains(err.Error(), "rslt.MustGet") {
				t.Errorf("panic error missing context: got %v", err)
			}
		}()

		rslt.Err[int](originalErr).MustGet()
	})
}

func TestMapErr(t *testing.T) {
	// wrapWithContext annotates an error with context.
	wrapWithContext := func(err error) error {
		return fmt.Errorf("context: %w", err)
	}

	t.Run("err transforms error", func(t *testing.T) {
		original := errors.New("fail")
		r := rslt.Err[int](original).MapErr(wrapWithContext)

		err, ok := r.GetErr()
		if !ok {
			t.Fatal("expected Err")
		}

		if !errors.Is(err, original) {
			t.Error("wrapped error does not chain to original")
		}

		if !strings.Contains(err.Error(), "context:") {
			t.Errorf("error missing context: %v", err)
		}
	})

	t.Run("ok passes through unchanged", func(t *testing.T) {
		r := rslt.Ok(42).MapErr(wrapWithContext)
		v, ok := r.Get()
		if !ok || v != 42 {
			t.Errorf("MapErr on Ok: got (%d, %t), want (42, true)", v, ok)
		}
	})
}

func TestStandaloneMap(t *testing.T) {
	// itoa converts int to string.
	itoa := func(n int) string { return fmt.Sprintf("%d", n) }

	tests := []struct {
		name    string
		input   rslt.Result[int]
		wantVal string
		wantOk  bool
	}{
		{
			name:    "ok transforms cross-type",
			input:   rslt.Ok(42),
			wantVal: "42",
			wantOk:  true,
		},
		{
			name:    "err passes through",
			input:   rslt.Err[int](errors.New("fail")),
			wantVal: "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rslt.Map(tt.input, itoa)
			val, ok := got.Get()
			if val != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Map: got (%q, %t), want (%q, %t)", val, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestFold(t *testing.T) {
	// errLen returns the length of an error message.
	errLen := func(e error) int { return len(e.Error()) }

	// identity returns its argument.
	identity := func(n int) int { return n }

	tests := []struct {
		name  string
		input rslt.Result[int]
		want  int
	}{
		{
			name:  "ok dispatches to onOk",
			input: rslt.Ok(42),
			want:  42,
		},
		{
			name:  "err dispatches to onErr",
			input: rslt.Err[int](errors.New("fail")),
			want:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rslt.Fold(tt.input, errLen, identity)
			if got != tt.want {
				t.Errorf("Fold: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPanicErrorUnwrap(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{
			name:    "error value unwraps",
			value:   errors.New("inner"),
			wantErr: true,
		},
		{
			name:    "string value returns nil",
			value:   "not an error",
			wantErr: false,
		},
		{
			name:    "int value returns nil",
			value:   42,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &rslt.PanicError{Value: tt.value}
			unwrapped := pe.Unwrap()
			if tt.wantErr && unwrapped == nil {
				t.Error("Unwrap: got nil, want error")
			}
			if !tt.wantErr && unwrapped != nil {
				t.Errorf("Unwrap: got %v, want nil", unwrapped)
			}
		})
	}
}

func TestPanicErrorChain(t *testing.T) {
	// sentinel is a specific error for errors.Is checking.
	sentinel := errors.New("sentinel")

	// wrappedSentinel wraps sentinel to test chain traversal.
	wrappedSentinel := fmt.Errorf("wrapped: %w", sentinel)

	pe := &rslt.PanicError{Value: wrappedSentinel}

	// Wrap in Err to test the full chain: Err → PanicError → wrapped → sentinel
	r := rslt.Err[int](pe)
	err, ok := r.GetErr()
	if !ok {
		t.Fatal("expected Err result")
	}

	if !errors.Is(err, sentinel) {
		t.Error("errors.Is: PanicError chain does not reach sentinel")
	}

	var target *rslt.PanicError
	if !errors.As(err, &target) {
		t.Error("errors.As: could not extract *PanicError from chain")
	}
	if target.Value != wrappedSentinel {
		t.Error("errors.As: extracted PanicError has wrong Value")
	}
}

func TestCollectAll(t *testing.T) {
	tests := []struct {
		name    string
		input   []rslt.Result[int]
		want    []int
		wantErr bool
	}{
		{
			name:    "all ok",
			input:   []rslt.Result[int]{rslt.Ok(1), rslt.Ok(2), rslt.Ok(3)},
			want:    []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "one err",
			input:   []rslt.Result[int]{rslt.Ok(1), rslt.Err[int](errors.New("fail")), rslt.Ok(3)},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "first err by index wins",
			input:   []rslt.Result[int]{rslt.Err[int](errors.New("first")), rslt.Err[int](errors.New("second"))},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty",
			input:   []rslt.Result[int]{},
			want:    []int{},
			wantErr: false,
		},
		{
			name:    "zero-value results treated as ok",
			input:   []rslt.Result[int]{{}, rslt.Ok(5)},
			want:    []int{0, 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rslt.CollectAll(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("CollectAll: got nil error, want error")
				}
				if got != nil {
					t.Errorf("CollectAll: got %v, want nil on error", got)
				}

				return
			}

			if err != nil {
				t.Errorf("CollectAll: unexpected error: %v", err)

				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("CollectAll: got len %d, want len %d", len(got), len(tt.want))

				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("CollectAll[%d]: got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}

	t.Run("first err by index returns that error", func(t *testing.T) {
		// firstErr is the error that should be returned (it appears first by index).
		firstErr := errors.New("first")
		input := []rslt.Result[int]{rslt.Err[int](firstErr), rslt.Err[int](errors.New("second"))}
		_, err := rslt.CollectAll(input)
		if !errors.Is(err, firstErr) {
			t.Errorf("CollectAll: got error %v, want %v", err, firstErr)
		}
	})
}

func TestCollectOk(t *testing.T) {
	tests := []struct {
		name  string
		input []rslt.Result[int]
		want  []int
	}{
		{
			name:  "all ok",
			input: []rslt.Result[int]{rslt.Ok(1), rslt.Ok(2), rslt.Ok(3)},
			want:  []int{1, 2, 3},
		},
		{
			name:  "mixed",
			input: []rslt.Result[int]{rslt.Ok(1), rslt.Err[int](errors.New("fail")), rslt.Ok(3)},
			want:  []int{1, 3},
		},
		{
			name:  "all err",
			input: []rslt.Result[int]{rslt.Err[int](errors.New("a")), rslt.Err[int](errors.New("b"))},
			want:  []int{},
		},
		{
			name:  "empty",
			input: []rslt.Result[int]{},
			want:  []int{},
		},
		{
			name:  "zero-value results included",
			input: []rslt.Result[int]{{}, rslt.Err[int](errors.New("fail")), rslt.Ok(5)},
			want:  []int{0, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rslt.CollectOk(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("CollectOk: got len %d, want len %d", len(got), len(tt.want))

				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("CollectOk[%d]: got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLift(t *testing.T) {
	// fallibleAtoi is a fallible string-to-int conversion.
	fallibleAtoi := func(s string) (int, error) {
		if s == "42" {
			return 42, nil
		}
		return 0, errors.New("not 42")
	}

	lifted := rslt.Lift(fallibleAtoi)

	t.Run("success wraps as ok", func(t *testing.T) {
		got := lifted("42")
		val, ok := got.Get()
		if !ok || val != 42 {
			t.Errorf("Lift: got (%d, %t), want (42, true)", val, ok)
		}
	})

	t.Run("error wraps as err", func(t *testing.T) {
		got := lifted("bad")
		if got.IsOk() {
			t.Error("Lift: expected Err result")
		}
	})
}

// --- Monadic Bind ---

func TestResultFlatMap(t *testing.T) {
	// validate returns Ok if positive, Err otherwise.
	validate := func(n int) rslt.Result[int] {
		if n > 0 {
			return rslt.Ok(n * 10)
		}
		return rslt.Err[int](errors.New("non-positive"))
	}

	t.Run("ok with fn returning ok", func(t *testing.T) {
		got := rslt.Ok(5).FlatMap(validate)
		val, ok := got.Get()
		if !ok || val != 50 {
			t.Errorf("FlatMap: got (%d, %t), want (50, true)", val, ok)
		}
	})

	t.Run("ok with fn returning err", func(t *testing.T) {
		got := rslt.Ok(-1).FlatMap(validate)
		if got.IsOk() {
			t.Error("FlatMap: expected Err when fn returns Err")
		}
	})

	t.Run("err short-circuits and preserves error", func(t *testing.T) {
		sentinelErr := errors.New("original")
		got := rslt.Err[int](sentinelErr).FlatMap(validate)
		err, ok := got.GetErr()
		if !ok {
			t.Fatal("FlatMap: expected Err result")
		}
		if !errors.Is(err, sentinelErr) {
			t.Errorf("FlatMap: error not preserved, got %v, want %v", err, sentinelErr)
		}
	})
}

func TestOrCall(t *testing.T) {
	t.Run("Ok returns value without calling", func(t *testing.T) {
		called := false
		r := rslt.Ok(42)
		got := r.OrCall(func() int { called = true; return 99 })
		if got != 42 {
			t.Errorf("OrCall() = %v, want 42", got)
		}
		if called {
			t.Error("OrCall() should not call function for Ok")
		}
	})

	t.Run("Err calls function", func(t *testing.T) {
		called := false
		r := rslt.Err[int](errors.New("fail"))
		got := r.OrCall(func() int { called = true; return 99 })
		if got != 99 {
			t.Errorf("OrCall() = %v, want 99", got)
		}
		if !called {
			t.Error("OrCall() should call function for Err")
		}
	})
}

func TestStandaloneFlatMap(t *testing.T) {
	// stringify returns Ok string if positive, Err otherwise.
	stringify := func(n int) rslt.Result[string] {
		if n > 0 {
			return rslt.Ok(fmt.Sprintf("%d", n*2))
		}
		return rslt.Err[string](errors.New("non-positive"))
	}

	t.Run("ok with fn returning ok", func(t *testing.T) {
		got := rslt.FlatMap(rslt.Ok(5), stringify)
		val, ok := got.Get()
		if !ok || val != "10" {
			t.Errorf("FlatMap: got (%q, %t), want (\"10\", true)", val, ok)
		}
	})

	t.Run("ok with fn returning err", func(t *testing.T) {
		got := rslt.FlatMap(rslt.Ok(-1), stringify)
		if got.IsOk() {
			t.Error("FlatMap: expected Err when fn returns Err")
		}
	})

	t.Run("err short-circuits and preserves error", func(t *testing.T) {
		sentinelErr := errors.New("original")
		got := rslt.FlatMap(rslt.Err[int](sentinelErr), stringify)
		err, ok := got.GetErr()
		if !ok {
			t.Fatal("FlatMap: expected Err result")
		}
		if !errors.Is(err, sentinelErr) {
			t.Errorf("FlatMap: error not preserved, got %v, want %v", err, sentinelErr)
		}
	})
}

func TestResultErr(t *testing.T) {
	t.Run("ok returns nil", func(t *testing.T) {
		r := rslt.Ok(42)
		if r.Err() != nil {
			t.Errorf("Ok.Err() = %v, want nil", r.Err())
		}
	})

	t.Run("err returns the error", func(t *testing.T) {
		sentinelErr := errors.New("fail")
		r := rslt.Err[int](sentinelErr)
		if r.Err() != sentinelErr {
			t.Errorf("Err.Err() = %v, want %v", r.Err(), sentinelErr)
		}
	})

	t.Run("zero value returns nil", func(t *testing.T) {
		var r rslt.Result[int]
		if r.Err() != nil {
			t.Errorf("zero Result.Err() = %v, want nil", r.Err())
		}
	})
}

func TestTap(t *testing.T) {
	t.Run("ok calls fn and returns same result", func(t *testing.T) {
		var captured int
		r := rslt.Ok(42).Tap(func(v int) { captured = v })
		v, ok := r.Get()
		if !ok || v != 42 {
			t.Errorf("Tap on Ok: got (%d, %t), want (42, true)", v, ok)
		}
		if captured != 42 {
			t.Errorf("Tap did not call fn: captured = %d, want 42", captured)
		}
	})

	t.Run("err skips fn and returns same result", func(t *testing.T) {
		called := false
		sentinelErr := errors.New("fail")
		r := rslt.Err[int](sentinelErr).Tap(func(int) { called = true })
		if called {
			t.Fatal("Tap should not call fn on Err")
		}
		if r.Err() != sentinelErr {
			t.Errorf("Tap on Err: error = %v, want %v", r.Err(), sentinelErr)
		}
	})
}

func TestTapErr(t *testing.T) {
	t.Run("err calls fn and returns same result", func(t *testing.T) {
		var captured error
		sentinelErr := errors.New("fail")
		r := rslt.Err[int](sentinelErr).TapErr(func(err error) { captured = err })
		if captured != sentinelErr {
			t.Errorf("TapErr did not call fn: captured = %v, want %v", captured, sentinelErr)
		}
		if r.Err() != sentinelErr {
			t.Errorf("TapErr changed error: got %v, want %v", r.Err(), sentinelErr)
		}
	})

	t.Run("ok skips fn and returns same result", func(t *testing.T) {
		called := false
		r := rslt.Ok(42).TapErr(func(error) { called = true })
		if called {
			t.Fatal("TapErr should not call fn on Ok")
		}
		v, ok := r.Get()
		if !ok || v != 42 {
			t.Errorf("TapErr on Ok: got (%d, %t), want (42, true)", v, ok)
		}
	})
}
