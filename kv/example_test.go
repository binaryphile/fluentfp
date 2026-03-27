package kv_test

import (
	"fmt"
	"sort"
	"strings"

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

func ExampleKeys() {
	m := map[string]int{"go": 2009, "rust": 2010, "python": 1991}

	keys := kv.Keys(m)
	sort.Strings(keys)
	fmt.Println(keys)
	// Output: [go python rust]
}

func ExampleValues() {
	m := map[string]int{"alice": 90, "bob": 85, "carol": 95}

	vals := kv.Values(m)
	sort.Ints(vals)
	fmt.Println(vals)
	// Output: [85 90 95]
}

func ExampleMerge() {
	defaults := map[string]string{"host": "localhost", "port": "8080"}
	overrides := map[string]string{"port": "9090", "debug": "true"}

	merged := kv.Merge(defaults, overrides)

	// Print specific keys for deterministic output.
	fmt.Println(merged["host"], merged["port"], merged["debug"])
	// Output: localhost 9090 true
}

func ExampleInvert() {
	codes := map[string]string{"US": "United States", "GB": "United Kingdom"}

	inverted := kv.Invert(codes)

	// Print specific keys for deterministic output.
	fmt.Println(inverted["United States"], inverted["United Kingdom"])
	// Output: US GB
}

func ExampleMapValues() {
	m := map[string]string{"greeting": "hello", "name": "world"}

	upper := kv.MapValues(m, strings.ToUpper)

	// Print specific keys for deterministic output.
	fmt.Println(upper["greeting"], upper["name"])
	// Output: HELLO WORLD
}
