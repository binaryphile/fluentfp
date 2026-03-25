package kv_test

import (
	"fmt"
	"sort"

	"github.com/binaryphile/fluentfp/kv"
)

func ExampleMap() {
	m := map[string]int{"alice": 90, "bob": 85}

	// formatEntry formats a key-value pair as "name: score".
	formatEntry := func(name string, score int) string {
		return fmt.Sprintf("%s: %d", name, score)
	}

	entries := kv.Map(m, formatEntry)

	// Sort for deterministic output (map iteration order is random).
	sort.Strings(entries)
	fmt.Println(entries)
	// Output: [alice: 90 bob: 85]
}
