package must

import (
	"errors"
	"os"
	"reflect"
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
			err:       errors.New(""),
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
				switch r := recover(); r {
				case nil:
					if tt.wantPanic {
						t.Errorf("BeNil() did not panic")
					}
				default:
					if !tt.wantPanic {
						t.Errorf("BeNil() panicked")
					}
				}

			}()

			BeNil(tt.err)
		})
	}
}

func TestGet(t *testing.T) {
	type args[T any] struct {
		t   T
		err error
	}
	type testCase[T any] struct {
		name      string
		args      args[T]
		want      T
		wantPanic bool
	}
	tests := []testCase[int]{
		{
			name:      "panic on error",
			args:      args[int]{err: errors.New("")},
			wantPanic: true,
		},
		{
			name: "return value on no error",
			args: args[int]{t: 1},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				switch r := recover(); r {
				case nil:
				default:
					if !tt.wantPanic {
						t.Errorf("Get() panicked")
					}
				}
			}()

			if got := Get(tt.args.t, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet2(t *testing.T) {
	type args[T any, T2 any] struct {
		t   T
		t2  T2
		err error
	}
	type testCase[T any, T2 any] struct {
		name      string
		args      args[T, T2]
		want      T
		want2     T2
		wantPanic bool
	}
	tests := []testCase[int, int]{
		{
			name:      "panic on error",
			args:      args[int, int]{err: errors.New("")},
			wantPanic: true,
		},
		{
			name:  "return values on no error",
			args:  args[int, int]{t: 1, t2: 2},
			want:  1,
			want2: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				switch r := recover(); r {
				case nil:
					if tt.wantPanic {
						t.Errorf("Get2() did not panic")
					}
				default:
					if !tt.wantPanic {
						t.Errorf("Get2() panicked")
					}
				}
			}()

			got, got2 := Get2(tt.args.t, tt.args.t2, tt.args.err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get2() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("Get2() got1 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetenv(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		want      string
		wantPanic bool
	}{
		{
			name:      "panic on non-existent key",
			key:       "NON_EXISTENT_KEY",
			wantPanic: true,
		},
		{
			name:      "panic on empty value",
			key:       "EMPTY_VALUE",
			value:     "",
			wantPanic: true,
		},
		{
			name:  "return value on extant key",
			key:   "EXTANT_KEY",
			value: "value",
			want:  "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				switch r := recover(); r {
				case nil:
					if tt.wantPanic {
						t.Errorf("Getenv() did not panic")
					}
				default:
					if !tt.wantPanic {
						t.Errorf("Getenv() panicked")
					}
				}
			}()

			switch tt.value {
			case "":
				if err := os.Unsetenv(tt.key); err != nil {
					t.Fatal("couldn't unset environment variable")
				}
			default:
				if err := os.Setenv(tt.key, tt.value); err != nil {
					t.Fatal("couldn't set environment variable")
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Fatal("couldn't unset environment variable")
					}
				}()
			}

			if got := Getenv(tt.key); got != tt.want {
				t.Errorf("Getenv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOf(t *testing.T) {
	type testCase[T any, R any] struct {
		name      string
		t         T
		fn        func(T) R
		want      R
		wantPanic bool
	}
	tests := []testCase[int, int]{
		{
			name:      "panic on error",
			fn:        Of(func(int) (int, error) { return 0, errors.New("") }),
			t:         0,
			wantPanic: true,
		},
		{
			name: "return value on no error",
			fn:   Of(func(int) (int, error) { return 1, nil }),
			t:    0,
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				switch r := recover(); r {
				case nil:
					if tt.wantPanic {
						t.Errorf("Of() did not panic")
					}
				default:
					if !tt.wantPanic {
						t.Errorf("Of() panicked")
					}
				}
			}()

			if got := tt.fn(tt.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Of() = %v, want %v", got, tt.want)
			}
		})
	}
}
