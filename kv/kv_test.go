package kv

import (
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

func TestMapTo(t *testing.T) {
	t.Run("transforms entries using key and value", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := MapTo[string](m).Map(func(k string, v int) string {
			return k + "=" + string(rune('0'+v))
		})
		if len(got) != 3 {
			t.Fatalf("MapTo().Map() len = %d, want 3", len(got))
		}
		sort.Strings(got)
		want := []string{"a=1", "b=2", "c=3"}
		for i, w := range want {
			if got[i] != w {
				t.Errorf("MapTo().Map()[%d] = %v, want %v", i, got[i], w)
			}
		}
	})

	t.Run("result chains with Mapper methods", func(t *testing.T) {
		m := map[string]int{"x": 10, "y": 20, "z": 30}
		count := MapTo[int](m).Map(func(k string, v int) int {
			return len(k) + v
		}).KeepIf(func(n int) bool { return n > 15 }).Len()
		if count != 2 {
			t.Errorf("MapTo().Map().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := MapTo[string](map[string]int{}).Map(func(k string, v int) string {
			return k
		})
		if len(got) != 0 {
			t.Errorf("MapTo().Map() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := MapTo[string, string, int](nil).Map(func(k string, v int) string {
			return k
		})
		if len(got) != 0 {
			t.Errorf("MapTo().Map() = %v, want empty", got)
		}
	})
}

func TestValues(t *testing.T) {
	t.Run("extracts values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := Values(m)
		if len(got) != 3 {
			t.Fatalf("Values() len = %d, want 3", len(got))
		}
		sort.Ints(got)
		want := []int{1, 2, 3}
		for i, w := range want {
			if got[i] != w {
				t.Errorf("Values()[%d] = %v, want %v", i, got[i], w)
			}
		}
	})

	t.Run("result chains with Mapper methods", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		count := Values(m).KeepIf(func(n int) bool { return n > 1 }).Len()
		if count != 2 {
			t.Errorf("Values().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := Values(map[string]int{})
		if len(got) != 0 {
			t.Errorf("Values() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := Values[string, int](nil)
		if len(got) != 0 {
			t.Errorf("Values() = %v, want empty", got)
		}
	})
}

func TestKeys(t *testing.T) {
	t.Run("extracts keys", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := Keys(m)
		if len(got) != 3 {
			t.Fatalf("Keys() len = %d, want 3", len(got))
		}
		sort.Strings(got)
		want := []string{"a", "b", "c"}
		for i, w := range want {
			if got[i] != w {
				t.Errorf("Keys()[%d] = %v, want %v", i, got[i], w)
			}
		}
	})

	t.Run("result chains with Mapper methods", func(t *testing.T) {
		m := map[string]int{"alpha": 1, "b": 2, "charlie": 3}
		count := Keys(m).KeepIf(func(s string) bool { return len(s) > 1 }).Len()
		if count != 2 {
			t.Errorf("Keys().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := Keys(map[string]int{})
		if len(got) != 0 {
			t.Errorf("Keys() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := Keys[string, int](nil)
		if len(got) != 0 {
			t.Errorf("Keys() = %v, want empty", got)
		}
	})
}

func TestMapValues(t *testing.T) {
	t.Run("transforms values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := MapValues(m, strconv.Itoa)
		want := map[string]string{"a": "1", "b": "2", "c": "3"}
		if !reflect.DeepEqual(map[string]string(got), want) {
			t.Errorf("MapValues() = %v, want %v", got, want)
		}
	})

	t.Run("preserves keys exactly", func(t *testing.T) {
		m := map[string]int{"x": 10, "y": 20}
		got := MapValues(m, func(v int) int { return v * 2 })
		want := map[string]int{"x": 20, "y": 40}
		if !reflect.DeepEqual(map[string]int(got), want) {
			t.Errorf("MapValues() = %v, want %v", got, want)
		}
	})

	t.Run("empty map returns non-nil empty map", func(t *testing.T) {
		got := MapValues(map[string]int{}, strconv.Itoa)
		if got == nil {
			t.Error("MapValues() returned nil, want non-nil empty map")
		}
		if len(got) != 0 {
			t.Errorf("MapValues() len = %d, want 0", len(got))
		}
	})

	t.Run("nil map returns non-nil empty map", func(t *testing.T) {
		got := MapValues[string](nil, strconv.Itoa)
		if got == nil {
			t.Error("MapValues() returned nil, want non-nil empty map")
		}
		if len(got) != 0 {
			t.Errorf("MapValues() len = %d, want 0", len(got))
		}
	})

	t.Run("result chains with Entries methods", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		// isLong returns true if the string value has more than 1 character.
		isLong := func(_ string, v string) bool { return len(v) > 1 }
		got := MapValues(m, strconv.Itoa).KeepIf(isLong)
		if len(got) != 0 {
			t.Errorf("MapValues().KeepIf() = %v, want empty (all single-digit)", got)
		}
	})
}

func TestKeepIf(t *testing.T) {
	// valueOver1 returns true if the value exceeds 1.
	valueOver1 := func(_ string, v int) bool { return v > 1 }

	t.Run("keeps matching entries", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := From(m).KeepIf(valueOver1)
		want := map[string]int{"b": 2, "c": 3}
		if !reflect.DeepEqual(map[string]int(got), want) {
			t.Errorf("KeepIf() = %v, want %v", got, want)
		}
	})

	t.Run("chains with Values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := From(m).KeepIf(valueOver1).Values()
		sort.Ints(got)
		want := []int{2, 3}
		if !reflect.DeepEqual([]int(got), want) {
			t.Errorf("KeepIf().Values() = %v, want %v", got, want)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := From(map[string]int{}).KeepIf(valueOver1)
		if len(got) != 0 {
			t.Errorf("KeepIf() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := From[string, int](nil).KeepIf(valueOver1)
		if len(got) != 0 {
			t.Errorf("KeepIf() = %v, want empty", got)
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		m := map[string]int{"a": 0, "b": -1}
		got := From(m).KeepIf(valueOver1)
		if len(got) != 0 {
			t.Errorf("KeepIf() = %v, want empty", got)
		}
	})

	t.Run("all match returns all entries", func(t *testing.T) {
		m := map[string]int{"a": 5, "b": 10}
		got := From(m).KeepIf(valueOver1)
		want := map[string]int{"a": 5, "b": 10}
		if !reflect.DeepEqual(map[string]int(got), want) {
			t.Errorf("KeepIf() = %v, want %v", got, want)
		}
	})
}

func TestRemoveIf(t *testing.T) {
	// valueOver1 returns true if the value exceeds 1.
	valueOver1 := func(_ string, v int) bool { return v > 1 }

	t.Run("removes matching entries", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := From(m).RemoveIf(valueOver1)
		want := map[string]int{"a": 1}
		if !reflect.DeepEqual(map[string]int(got), want) {
			t.Errorf("RemoveIf() = %v, want %v", got, want)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := From(map[string]int{}).RemoveIf(valueOver1)
		if len(got) != 0 {
			t.Errorf("RemoveIf() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := From[string, int](nil).RemoveIf(valueOver1)
		if len(got) != 0 {
			t.Errorf("RemoveIf() = %v, want empty", got)
		}
	})
}

func TestInvert(t *testing.T) {
	t.Run("swaps keys and values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := Invert(m)
		want := map[int]string{1: "a", 2: "b", 3: "c"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Invert() = %v, want %v", got, want)
		}
	})

	t.Run("duplicate values produce map with fewer entries", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 1, "c": 2}
		got := Invert(m)
		if len(got) != 2 {
			t.Errorf("Invert() len = %d, want 2", len(got))
		}
		if got[2] != "c" {
			t.Errorf("Invert()[2] = %q, want \"c\"", got[2])
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := Invert(map[string]int{})
		if len(got) != 0 {
			t.Errorf("Invert() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := Invert[string, int](nil)
		if len(got) != 0 {
			t.Errorf("Invert() = %v, want empty", got)
		}
	})
}

func TestMerge(t *testing.T) {
	t.Run("combines two maps", func(t *testing.T) {
		a := map[string]int{"a": 1, "b": 2}
		b := map[string]int{"c": 3, "d": 4}
		got := Merge(a, b)
		want := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Merge() = %v, want %v", got, want)
		}
	})

	t.Run("later maps override earlier keys", func(t *testing.T) {
		a := map[string]int{"x": 1, "y": 2}
		b := map[string]int{"y": 99, "z": 3}
		got := Merge(a, b)
		want := map[string]int{"x": 1, "y": 99, "z": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Merge() = %v, want %v", got, want)
		}
	})

	t.Run("three maps", func(t *testing.T) {
		got := Merge(
			map[string]int{"a": 1},
			map[string]int{"b": 2},
			map[string]int{"c": 3},
		)
		want := map[string]int{"a": 1, "b": 2, "c": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Merge() = %v, want %v", got, want)
		}
	})

	t.Run("skips nil maps", func(t *testing.T) {
		got := Merge(map[string]int{"a": 1}, nil, map[string]int{"b": 2})
		want := map[string]int{"a": 1, "b": 2}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Merge() = %v, want %v", got, want)
		}
	})

	t.Run("no args returns empty", func(t *testing.T) {
		got := Merge[string, int]()
		if got == nil || len(got) != 0 {
			t.Errorf("Merge() = %v, want non-nil empty", got)
		}
	})

	t.Run("single map returns copy", func(t *testing.T) {
		orig := map[string]int{"a": 1}
		got := Merge(orig)
		want := map[string]int{"a": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Merge() = %v, want %v", got, want)
		}
		// Verify it's a copy
		got["a"] = 999
		if orig["a"] != 1 {
			t.Error("Merge() did not copy — mutating result affected input")
		}
	})
}

func TestPickByKeys(t *testing.T) {
	t.Run("picks subset", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := PickByKeys(m, []string{"a", "c"})
		want := map[string]int{"a": 1, "c": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PickByKeys() = %v, want %v", got, want)
		}
	})

	t.Run("ignores keys not in map", func(t *testing.T) {
		m := map[string]int{"a": 1}
		got := PickByKeys(m, []string{"a", "z"})
		want := map[string]int{"a": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PickByKeys() = %v, want %v", got, want)
		}
	})

	t.Run("no keys returns empty", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := PickByKeys(m, []string{})
		if len(got) != 0 {
			t.Errorf("PickByKeys() = %v, want empty", got)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := PickByKeys(map[string]int{}, []string{"a"})
		if len(got) != 0 {
			t.Errorf("PickByKeys() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := PickByKeys[string, int](nil, []string{"a"})
		if len(got) != 0 {
			t.Errorf("PickByKeys() = %v, want empty", got)
		}
	})
}

func TestOmitByKeys(t *testing.T) {
	t.Run("omits subset", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := OmitByKeys(m, []string{"b"})
		want := map[string]int{"a": 1, "c": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("OmitByKeys() = %v, want %v", got, want)
		}
	})

	t.Run("ignores keys not in map", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := OmitByKeys(m, []string{"z"})
		want := map[string]int{"a": 1, "b": 2}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("OmitByKeys() = %v, want %v", got, want)
		}
	})

	t.Run("omit all returns empty", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := OmitByKeys(m, []string{"a", "b"})
		if len(got) != 0 {
			t.Errorf("OmitByKeys() = %v, want empty", got)
		}
	})

	t.Run("no keys returns all", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := OmitByKeys(m, []string{})
		want := map[string]int{"a": 1, "b": 2}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("OmitByKeys() = %v, want %v", got, want)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := OmitByKeys(map[string]int{}, []string{"a"})
		if len(got) != 0 {
			t.Errorf("OmitByKeys() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := OmitByKeys[string, int](nil, []string{"a"})
		if len(got) != 0 {
			t.Errorf("OmitByKeys() = %v, want empty", got)
		}
	})
}

func TestToPairs(t *testing.T) {
	t.Run("converts map to pairs", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := ToPairs(m)
		if len(got) != 3 {
			t.Fatalf("ToPairs() len = %d, want 3", len(got))
		}
		sort.Slice(got, func(i, j int) bool { return got[i].First < got[j].First })
		if got[0].First != "a" || got[0].Second != 1 {
			t.Errorf("ToPairs()[0] = {%q, %d}, want {\"a\", 1}", got[0].First, got[0].Second)
		}
		if got[1].First != "b" || got[1].Second != 2 {
			t.Errorf("ToPairs()[1] = {%q, %d}, want {\"b\", 2}", got[1].First, got[1].Second)
		}
		if got[2].First != "c" || got[2].Second != 3 {
			t.Errorf("ToPairs()[2] = {%q, %d}, want {\"c\", 3}", got[2].First, got[2].Second)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := ToPairs(map[string]int{})
		if len(got) != 0 {
			t.Errorf("ToPairs() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := ToPairs[string, int](nil)
		if len(got) != 0 {
			t.Errorf("ToPairs() = %v, want empty", got)
		}
	})

	t.Run("single entry", func(t *testing.T) {
		got := ToPairs(map[string]int{"x": 42})
		if len(got) != 1 {
			t.Fatalf("ToPairs() len = %d, want 1", len(got))
		}
		if got[0].First != "x" || got[0].Second != 42 {
			t.Errorf("ToPairs()[0] = {%q, %d}, want {\"x\", 42}", got[0].First, got[0].Second)
		}
	})

	t.Run("result chains with Mapper methods", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		// hasValueOver1 returns true if the pair's value exceeds 1.
		hasValueOver1 := func(p pair.Pair[string, int]) bool { return p.Second > 1 }
		count := ToPairs(m).KeepIf(hasValueOver1).Len()
		if count != 2 {
			t.Errorf("ToPairs().KeepIf().Len() = %d, want 2", count)
		}
	})
}

func TestFromPairs(t *testing.T) {
	t.Run("converts pairs to map", func(t *testing.T) {
		ps := []pair.Pair[string, int]{
			pair.Of("a", 1),
			pair.Of("b", 2),
			pair.Of("c", 3),
		}
		got := FromPairs(ps)
		want := map[string]int{"a": 1, "b": 2, "c": 3}
		if !reflect.DeepEqual(map[string]int(got), want) {
			t.Errorf("FromPairs() = %v, want %v", got, want)
		}
	})

	t.Run("duplicate keys last wins", func(t *testing.T) {
		ps := []pair.Pair[string, int]{
			pair.Of("a", 1),
			pair.Of("a", 99),
		}
		got := FromPairs(ps)
		if got["a"] != 99 {
			t.Errorf("FromPairs() duplicate key = %d, want 99 (last wins)", got["a"])
		}
		if len(got) != 1 {
			t.Errorf("FromPairs() len = %d, want 1", len(got))
		}
	})

	t.Run("empty slice returns empty", func(t *testing.T) {
		got := FromPairs([]pair.Pair[string, int]{})
		if got == nil || len(got) != 0 {
			t.Errorf("FromPairs() = %v, want non-nil empty", got)
		}
	})

	t.Run("nil slice returns empty", func(t *testing.T) {
		got := FromPairs[string, int](nil)
		if got == nil || len(got) != 0 {
			t.Errorf("FromPairs() = %v, want non-nil empty", got)
		}
	})

	t.Run("single pair", func(t *testing.T) {
		ps := []pair.Pair[string, int]{pair.Of("x", 42)}
		got := FromPairs(ps)
		want := map[string]int{"x": 42}
		if !reflect.DeepEqual(map[string]int(got), want) {
			t.Errorf("FromPairs() = %v, want %v", got, want)
		}
	})

	t.Run("result chains with Entries methods", func(t *testing.T) {
		ps := []pair.Pair[string, int]{
			pair.Of("a", 1),
			pair.Of("b", 2),
			pair.Of("c", 3),
		}
		// valueOver1 returns true if the value exceeds 1.
		valueOver1 := func(_ string, v int) bool { return v > 1 }
		got := FromPairs(ps).KeepIf(valueOver1)
		if len(got) != 2 {
			t.Errorf("FromPairs().KeepIf() len = %d, want 2", len(got))
		}
	})

	t.Run("roundtrip preserves map", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := FromPairs(ToPairs(m))
		if !reflect.DeepEqual(map[string]int(got), m) {
			t.Errorf("FromPairs(ToPairs()) = %v, want %v", got, m)
		}
	})
}

func TestFrom(t *testing.T) {
	t.Run("Values matches standalone Values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := From(m).Values()
		if len(got) != 2 {
			t.Fatalf("From().Values() len = %d, want 2", len(got))
		}
	})

	t.Run("Keys matches standalone Keys", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := From(m).Keys()
		if len(got) != 2 {
			t.Fatalf("From().Keys() len = %d, want 2", len(got))
		}
	})
}
