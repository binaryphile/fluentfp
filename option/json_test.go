package option

import (
	"encoding/json"
	"testing"
)

func TestBasic_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		opt  Basic[int]
		want string
	}{
		{"ok value", Of(42), "42"},
		{"not ok", Basic[int]{}, "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.opt)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBasic_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantOk  bool
		wantVal int
	}{
		{"value", "42", true, 42},
		{"null", "null", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opt Basic[int]
			if err := json.Unmarshal([]byte(tt.json), &opt); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if v, ok := opt.Get(); ok != tt.wantOk || (ok && v != tt.wantVal) {
				t.Errorf("got (%v, %v), want (%v, %v)", v, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

type testStruct struct {
	Name  string     `json:"name"`
	Value Basic[int] `json:"value"`
}

func TestBasic_JSON_InStruct(t *testing.T) {
	// Marshal with value
	s := testStruct{Name: "test", Value: Of(42)}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(data) != `{"name":"test","value":42}` {
		t.Errorf("marshal: got %s", data)
	}

	// Unmarshal with value
	var s2 testStruct
	if err := json.Unmarshal([]byte(`{"name":"test","value":42}`), &s2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if v, ok := s2.Value.Get(); !ok || v != 42 {
		t.Errorf("unmarshal value: got %v, %v", v, ok)
	}

	// Unmarshal with null
	var s3 testStruct
	if err := json.Unmarshal([]byte(`{"name":"test","value":null}`), &s3); err != nil {
		t.Fatalf("unmarshal null error: %v", err)
	}
	if _, ok := s3.Value.Get(); ok {
		t.Error("unmarshal null: expected NotOk")
	}

	// Unmarshal with missing field
	var s4 testStruct
	if err := json.Unmarshal([]byte(`{"name":"test"}`), &s4); err != nil {
		t.Fatalf("unmarshal missing error: %v", err)
	}
	if _, ok := s4.Value.Get(); ok {
		t.Error("unmarshal missing: expected NotOk")
	}
}
