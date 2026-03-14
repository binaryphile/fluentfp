package must

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestBeNil(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantPanic bool
	}{
		{
			name:      "panic on non-nil",
			err:       errors.New("test error"),
			wantPanic: true,
		},
		{
			name:      "don't panic on nil",
			err:       nil,
			wantPanic: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic {
					if r == nil {
						t.Fatal("BeNil() did not panic")
					}
					if r != tt.err {
						t.Errorf("BeNil() panicked with %v, want exact error %v", r, tt.err)
					}
				} else if r != nil {
					t.Errorf("BeNil() panicked unexpectedly: %v", r)
				}
			}()

			BeNil(tt.err)
		})
	}
}

func TestGet(t *testing.T) {
	t.Run("panic preserves exact error", func(t *testing.T) {
		sentinel := errors.New("sentinel")

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Get() did not panic")
			}
			if r != sentinel {
				t.Errorf("Get() panicked with %v, want exact error %v", r, sentinel)
			}
		}()

		Get(0, sentinel)
	})

	t.Run("return value on no error", func(t *testing.T) {
		if got := Get(42, nil); got != 42 {
			t.Errorf("Get() = %v, want 42", got)
		}
	})

	t.Run("errors.Is traverses chain after recovery", func(t *testing.T) {
		sentinel := errors.New("root cause")
		wrapped := fmt.Errorf("context: %w", sentinel)

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Get() did not panic")
			}
			err, ok := r.(error)
			if !ok {
				t.Fatalf("Get() panicked with %T, want error", r)
			}
			if !errors.Is(err, sentinel) {
				t.Error("errors.Is could not find sentinel through wrapper after recovery")
			}
		}()

		Get(0, wrapped)
	})
}

func TestGet2(t *testing.T) {
	t.Run("panic preserves exact error", func(t *testing.T) {
		sentinel := errors.New("sentinel")

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Get2() did not panic")
			}
			if r != sentinel {
				t.Errorf("Get2() panicked with %v, want exact error %v", r, sentinel)
			}
		}()

		Get2(0, 0, sentinel)
	})

	t.Run("return values on no error", func(t *testing.T) {
		got1, got2 := Get2(1, 2, nil)
		if got1 != 1 {
			t.Errorf("Get2() got1 = %v, want 1", got1)
		}
		if got2 != 2 {
			t.Errorf("Get2() got2 = %v, want 2", got2)
		}
	})
}

func TestNonEmptyEnv(t *testing.T) {
	t.Run("panic on unset wraps ErrEnvUnset", func(t *testing.T) {
		key := "MUST_TEST_UNSET"
		old, had := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if had {
				os.Setenv(key, old)
			} else {
				os.Unsetenv(key)
			}
		})

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("NonEmptyEnv() did not panic")
			}
			err, ok := r.(error)
			if !ok {
				t.Fatalf("NonEmptyEnv() panicked with %T, want error", r)
			}
			if !errors.Is(err, ErrEnvUnset) {
				t.Error("expected errors.Is(err, ErrEnvUnset)")
			}
		}()

		NonEmptyEnv(key)
	})

	t.Run("panic on empty wraps ErrEnvEmpty", func(t *testing.T) {
		t.Setenv("MUST_TEST_EMPTY", "")

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("NonEmptyEnv() did not panic")
			}
			err, ok := r.(error)
			if !ok {
				t.Fatalf("NonEmptyEnv() panicked with %T, want error", r)
			}
			if !errors.Is(err, ErrEnvEmpty) {
				t.Error("expected errors.Is(err, ErrEnvEmpty)")
			}
		}()

		NonEmptyEnv("MUST_TEST_EMPTY")
	})

	t.Run("return value on non-empty", func(t *testing.T) {
		t.Setenv("MUST_TEST_SET", "hello")

		if got := NonEmptyEnv("MUST_TEST_SET"); got != "hello" {
			t.Errorf("NonEmptyEnv() = %v, want hello", got)
		}
	})

	t.Run("panic message includes key", func(t *testing.T) {
		key := "MUST_TEST_KEY_IN_MSG"

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("NonEmptyEnv() did not panic")
			}
			err, ok := r.(error)
			if !ok {
				t.Fatalf("NonEmptyEnv() panicked with %T, want error", r)
			}
			msg := err.Error()
			if !strings.Contains(msg, key) {
				t.Errorf("panic message %q does not contain key %q", msg, key)
			}
		}()

		NonEmptyEnv(key)
	})
}

func TestOf(t *testing.T) {
	t.Run("panic on error preserves exact error", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		fn := Of(func(int) (int, error) { return 0, sentinel })

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Of-wrapped function did not panic")
			}
			if r != sentinel {
				t.Errorf("panicked with %v, want exact error %v", r, sentinel)
			}
		}()

		fn(0)
	})

	t.Run("return value on no error", func(t *testing.T) {
		fn := Of(func(n int) (int, error) { return n * 2, nil })

		if got := fn(21); got != 42 {
			t.Errorf("Of-wrapped function = %v, want 42", got)
		}
	})

	t.Run("argument forwarding", func(t *testing.T) {
		fn := Of(strconv.Atoi)

		if got := fn("123"); got != 123 {
			t.Errorf("Of(strconv.Atoi)(\"123\") = %v, want 123", got)
		}
	})

	t.Run("nil function panics immediately wrapping ErrNilFunction", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Of(nil) did not panic")
			}
			err, ok := r.(error)
			if !ok {
				t.Fatalf("Of(nil) panicked with %T, want error", r)
			}
			if !errors.Is(err, ErrNilFunction) {
				t.Error("expected errors.Is(err, ErrNilFunction)")
			}
		}()

		Of[int, int](nil)
	})
}

