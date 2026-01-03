package slice

import (
	"reflect"
	"testing"
)

type person struct {
	Name string
	Age  int
	City string
	Zip  string
}

func (p person) GetName() string { return p.Name }
func (p person) GetAge() int     { return p.Age }
func (p person) GetCity() string { return p.City }
func (p person) GetZip() string  { return p.Zip }

func TestUnzip2(t *testing.T) {
	tests := []struct {
		name   string
		input  []person
		wantA  []string
		wantB  []int
	}{
		{
			name:   "empty slice",
			input:  []person{},
			wantA:  []string{},
			wantB:  []int{},
		},
		{
			name: "single element",
			input: []person{
				{Name: "Alice", Age: 30},
			},
			wantA: []string{"Alice"},
			wantB: []int{30},
		},
		{
			name: "multiple elements",
			input: []person{
				{Name: "Alice", Age: 30},
				{Name: "Bob", Age: 25},
				{Name: "Carol", Age: 35},
			},
			wantA: []string{"Alice", "Bob", "Carol"},
			wantB: []int{30, 25, 35},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := Unzip2(tt.input, person.GetName, person.GetAge)
			if !reflect.DeepEqual([]string(gotA), tt.wantA) {
				t.Errorf("Unzip2() gotA = %v, want %v", gotA, tt.wantA)
			}
			if !reflect.DeepEqual([]int(gotB), tt.wantB) {
				t.Errorf("Unzip2() gotB = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestUnzip3(t *testing.T) {
	tests := []struct {
		name   string
		input  []person
		wantA  []string
		wantB  []int
		wantC  []string
	}{
		{
			name:   "empty slice",
			input:  []person{},
			wantA:  []string{},
			wantB:  []int{},
			wantC:  []string{},
		},
		{
			name: "multiple elements",
			input: []person{
				{Name: "Alice", Age: 30, City: "NYC"},
				{Name: "Bob", Age: 25, City: "LA"},
			},
			wantA: []string{"Alice", "Bob"},
			wantB: []int{30, 25},
			wantC: []string{"NYC", "LA"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB, gotC := Unzip3(tt.input, person.GetName, person.GetAge, person.GetCity)
			if !reflect.DeepEqual([]string(gotA), tt.wantA) {
				t.Errorf("Unzip3() gotA = %v, want %v", gotA, tt.wantA)
			}
			if !reflect.DeepEqual([]int(gotB), tt.wantB) {
				t.Errorf("Unzip3() gotB = %v, want %v", gotB, tt.wantB)
			}
			if !reflect.DeepEqual([]string(gotC), tt.wantC) {
				t.Errorf("Unzip3() gotC = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}

func TestUnzip4(t *testing.T) {
	tests := []struct {
		name   string
		input  []person
		wantA  []string
		wantB  []int
		wantC  []string
		wantD  []string
	}{
		{
			name:   "empty slice",
			input:  []person{},
			wantA:  []string{},
			wantB:  []int{},
			wantC:  []string{},
			wantD:  []string{},
		},
		{
			name: "multiple elements",
			input: []person{
				{Name: "Alice", Age: 30, City: "NYC", Zip: "10001"},
				{Name: "Bob", Age: 25, City: "LA", Zip: "90001"},
			},
			wantA: []string{"Alice", "Bob"},
			wantB: []int{30, 25},
			wantC: []string{"NYC", "LA"},
			wantD: []string{"10001", "90001"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB, gotC, gotD := Unzip4(tt.input, person.GetName, person.GetAge, person.GetCity, person.GetZip)
			if !reflect.DeepEqual([]string(gotA), tt.wantA) {
				t.Errorf("Unzip4() gotA = %v, want %v", gotA, tt.wantA)
			}
			if !reflect.DeepEqual([]int(gotB), tt.wantB) {
				t.Errorf("Unzip4() gotB = %v, want %v", gotB, tt.wantB)
			}
			if !reflect.DeepEqual([]string(gotC), tt.wantC) {
				t.Errorf("Unzip4() gotC = %v, want %v", gotC, tt.wantC)
			}
			if !reflect.DeepEqual([]string(gotD), tt.wantD) {
				t.Errorf("Unzip4() gotD = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}
