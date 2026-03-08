package option

import "testing"

func TestOption_Value(t *testing.T) {
	tests := []struct {
		name    string
		opt     Option[string]
		want    string
		wantNil bool
	}{
		{"ok value", Of("hello"), "hello", false},
		{"not ok", Option[string]{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.opt.Value()
			if err != nil {
				t.Fatalf("Value() error: %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("Value() = %v, want nil", got)
				}
				return
			}
			if got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOption_Scan(t *testing.T) {
	t.Run("int64 to int", func(t *testing.T) {
		var opt Option[int]
		if err := opt.Scan(int64(42)); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v, ok := opt.Get(); !ok || v != 42 {
			t.Errorf("Scan(int64(42)) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("nil to not-ok", func(t *testing.T) {
		var opt Option[int]
		if err := opt.Scan(nil); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if _, ok := opt.Get(); ok {
			t.Error("Scan(nil) should be not-ok")
		}
	})

	t.Run("bytes to string", func(t *testing.T) {
		var opt Option[string]
		if err := opt.Scan([]byte("hello")); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v, ok := opt.Get(); !ok || v != "hello" {
			t.Errorf("Scan([]byte) = (%v, %v), want (hello, true)", v, ok)
		}
	})

	t.Run("string to string", func(t *testing.T) {
		var opt Option[string]
		if err := opt.Scan("hello"); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if v, ok := opt.Get(); !ok || v != "hello" {
			t.Errorf("Scan(string) = (%v, %v), want (hello, true)", v, ok)
		}
	})

	t.Run("nil resets ok option", func(t *testing.T) {
		opt := Of(42)
		if err := opt.Scan(nil); err != nil {
			t.Fatalf("Scan error: %v", err)
		}
		if _, ok := opt.Get(); ok {
			t.Error("Scan(nil) on ok option should become not-ok")
		}
	})
}
