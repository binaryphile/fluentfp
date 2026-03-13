package slice

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

func TestZip(t *testing.T) {
	tests := []struct {
		name string
		as   []int
		bs   []string
		want []pair.Pair[int, string]
	}{
		{
			name: "equal length",
			as:   []int{1, 2, 3},
			bs:   []string{"a", "b", "c"},
			want: []pair.Pair[int, string]{
				{First: 1, Second: "a"},
				{First: 2, Second: "b"},
				{First: 3, Second: "c"},
			},
		},
		{
			name: "first shorter",
			as:   []int{1, 2},
			bs:   []string{"a", "b", "c"},
			want: []pair.Pair[int, string]{
				{First: 1, Second: "a"},
				{First: 2, Second: "b"},
			},
		},
		{
			name: "second shorter",
			as:   []int{1, 2, 3},
			bs:   []string{"a"},
			want: []pair.Pair[int, string]{
				{First: 1, Second: "a"},
			},
		},
		{
			name: "both empty",
			as:   []int{},
			bs:   []string{},
			want: []pair.Pair[int, string]{},
		},
		{
			name: "one empty",
			as:   []int{1, 2, 3},
			bs:   []string{},
			want: []pair.Pair[int, string]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Zip(tt.as, tt.bs)
			if !reflect.DeepEqual([]pair.Pair[int, string](got), tt.want) {
				t.Errorf("Zip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipWith(t *testing.T) {
	concat := func(a int, b string) string { return fmt.Sprintf("%d:%s", a, b) }
	tests := []struct {
		name string
		as   []int
		bs   []string
		want []string
	}{
		{
			name: "equal length",
			as:   []int{1, 2, 3},
			bs:   []string{"a", "b", "c"},
			want: []string{"1:a", "2:b", "3:c"},
		},
		{
			name: "first shorter",
			as:   []int{1},
			bs:   []string{"a", "b"},
			want: []string{"1:a"},
		},
		{
			name: "both empty",
			as:   []int{},
			bs:   []string{},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ZipWith(tt.as, tt.bs, concat)
			if !reflect.DeepEqual([]string(got), tt.want) {
				t.Errorf("ZipWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
