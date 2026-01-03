package ternary

import "testing"

func TestElse(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		useThen   bool // false = use ThenCall
		thenVal   string
		elseVal   string
		want      string
	}{
		{
			name:      "condition false returns else value",
			condition: false,
			useThen:   true,
			thenVal:   "then",
			elseVal:   "else",
			want:      "else",
		},
		{
			name:      "condition true with Then returns then value",
			condition: true,
			useThen:   true,
			thenVal:   "then",
			elseVal:   "else",
			want:      "then",
		},
		{
			name:      "condition true with ThenCall returns then function result",
			condition: true,
			useThen:   false,
			thenVal:   "then",
			elseVal:   "else",
			want:      "then",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.useThen {
				got = If[string](tt.condition).Then(tt.thenVal).Else(tt.elseVal)
			} else {
				got = If[string](tt.condition).ThenCall(func() string { return tt.thenVal }).Else(tt.elseVal)
			}
			if got != tt.want {
				t.Errorf("Else() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElseCall(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		useThen   bool // false = use ThenCall
		thenVal   string
		elseVal   string
		want      string
	}{
		{
			name:      "condition false calls else function",
			condition: false,
			useThen:   true,
			thenVal:   "then",
			elseVal:   "else",
			want:      "else",
		},
		{
			name:      "condition true with Then returns then value",
			condition: true,
			useThen:   true,
			thenVal:   "then",
			elseVal:   "else",
			want:      "then",
		},
		{
			name:      "condition true with ThenCall returns then function result",
			condition: true,
			useThen:   false,
			thenVal:   "then",
			elseVal:   "else",
			want:      "then",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.useThen {
				got = If[string](tt.condition).Then(tt.thenVal).ElseCall(func() string { return tt.elseVal })
			} else {
				got = If[string](tt.condition).ThenCall(func() string { return tt.thenVal }).ElseCall(func() string { return tt.elseVal })
			}
			if got != tt.want {
				t.Errorf("ElseCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLazyEvaluation(t *testing.T) {
	t.Run("ThenCall not called when condition false", func(t *testing.T) {
		called := false
		If[string](false).ThenCall(func() string {
			called = true
			return "then"
		}).Else("else")
		if called {
			t.Error("ThenCall function was called when condition was false")
		}
	})

	t.Run("ElseCall not called when condition true", func(t *testing.T) {
		called := false
		If[string](true).Then("then").ElseCall(func() string {
			called = true
			return "else"
		})
		if called {
			t.Error("ElseCall function was called when condition was true")
		}
	})
}
