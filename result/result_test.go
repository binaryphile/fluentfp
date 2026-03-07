package result_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/binaryphile/fluentfp/result"
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
		expectedMsg := "result.Err: error must not be nil"
		if got != expectedMsg {
			t.Errorf("got %q, want %q", got, expectedMsg)
		}
	}()

	result.Err[int](nil)
}

func TestConvert(t *testing.T) {
	// double doubles an int.
	double := func(n int) int { return n * 2 }

	tests := []struct {
		name    string
		result  result.Result[int]
		wantVal int
		wantOk  bool
	}{
		{
			name:    "ok transforms",
			result:  result.Ok(5),
			wantVal: 10,
			wantOk:  true,
		},
		{
			name:    "err passes through",
			result:  result.Err[int](errors.New("fail")),
			wantVal: 0,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.Convert(double)
			val, ok := got.Get()
			if val != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Convert: got (%d, %t), want (%d, %t)", val, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestMustGet(t *testing.T) {
	t.Run("ok returns value", func(t *testing.T) {
		got := result.Ok(42).MustGet()
		if got != 42 {
			t.Errorf("MustGet: got %d, want 42", got)
		}
	})

	t.Run("err panics", func(t *testing.T) {
		defer func() {
			v := recover()
			if v == nil {
				t.Fatal("expected panic, got none")
			}

			got, ok := v.(string)
			if !ok {
				t.Fatalf("expected string panic, got %T: %v", v, v)
			}

			// expectedMsg is the panic message for MustGet on Err.
			expectedMsg := "result: MustGet called on Err"
			if got != expectedMsg {
				t.Errorf("got %q, want %q", got, expectedMsg)
			}
		}()

		result.Err[int](errors.New("fail")).MustGet()
	})
}

func TestStandaloneMap(t *testing.T) {
	// itoa converts int to string.
	itoa := func(n int) string { return fmt.Sprintf("%d", n) }

	tests := []struct {
		name    string
		input   result.Result[int]
		wantVal string
		wantOk  bool
	}{
		{
			name:    "ok transforms cross-type",
			input:   result.Ok(42),
			wantVal: "42",
			wantOk:  true,
		},
		{
			name:    "err passes through",
			input:   result.Err[int](errors.New("fail")),
			wantVal: "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := result.Map(tt.input, itoa)
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
		input result.Result[int]
		want  int
	}{
		{
			name:  "ok dispatches to onOk",
			input: result.Ok(42),
			want:  42,
		},
		{
			name:  "err dispatches to onErr",
			input: result.Err[int](errors.New("fail")),
			want:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := result.Fold(tt.input, errLen, identity)
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
			pe := &result.PanicError{Value: tt.value}
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

	pe := &result.PanicError{Value: wrappedSentinel}

	// Wrap in Err to test the full chain: Err → PanicError → wrapped → sentinel
	r := result.Err[int](pe)
	err, ok := r.GetErr()
	if !ok {
		t.Fatal("expected Err result")
	}

	if !errors.Is(err, sentinel) {
		t.Error("errors.Is: PanicError chain does not reach sentinel")
	}

	var target *result.PanicError
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
		input   []result.Result[int]
		want    []int
		wantErr bool
	}{
		{
			name:    "all ok",
			input:   []result.Result[int]{result.Ok(1), result.Ok(2), result.Ok(3)},
			want:    []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "one err",
			input:   []result.Result[int]{result.Ok(1), result.Err[int](errors.New("fail")), result.Ok(3)},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "first err by index wins",
			input:   []result.Result[int]{result.Err[int](errors.New("first")), result.Err[int](errors.New("second"))},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty",
			input:   []result.Result[int]{},
			want:    []int{},
			wantErr: false,
		},
		{
			name:    "zero-value results treated as ok",
			input:   []result.Result[int]{{}, result.Ok(5)},
			want:    []int{0, 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := result.CollectAll(tt.input)
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
		input := []result.Result[int]{result.Err[int](firstErr), result.Err[int](errors.New("second"))}
		_, err := result.CollectAll(input)
		if !errors.Is(err, firstErr) {
			t.Errorf("CollectAll: got error %v, want %v", err, firstErr)
		}
	})
}

func TestCollectOk(t *testing.T) {
	tests := []struct {
		name  string
		input []result.Result[int]
		want  []int
	}{
		{
			name:  "all ok",
			input: []result.Result[int]{result.Ok(1), result.Ok(2), result.Ok(3)},
			want:  []int{1, 2, 3},
		},
		{
			name:  "mixed",
			input: []result.Result[int]{result.Ok(1), result.Err[int](errors.New("fail")), result.Ok(3)},
			want:  []int{1, 3},
		},
		{
			name:  "all err",
			input: []result.Result[int]{result.Err[int](errors.New("a")), result.Err[int](errors.New("b"))},
			want:  []int{},
		},
		{
			name:  "empty",
			input: []result.Result[int]{},
			want:  []int{},
		},
		{
			name:  "zero-value results included",
			input: []result.Result[int]{{}, result.Err[int](errors.New("fail")), result.Ok(5)},
			want:  []int{0, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := result.CollectOk(tt.input)
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
