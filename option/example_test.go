package option_test

import (
	"fmt"
	"strconv"

	"github.com/binaryphile/fluentfp/option"
)

func ExampleOption_Or() {
	// NonZero wraps a value as ok if it's not the zero value.
	port := option.NonZero(8080).Or(3000)
	fmt.Println(port)

	missing := option.NonZero(0).Or(3000)
	fmt.Println(missing)
	// Output:
	// 8080
	// 3000
}

func ExampleOption_OrElse() {
	// Multi-level fallback: try primary, then secondary, then default.
	primary := option.NotOk[string]()
	secondary := option.Of("backup")

	// primaryLookup returns the primary option.
	primaryLookup := func() option.Option[string] { return primary }
	// secondaryLookup returns the secondary option.
	secondaryLookup := func() option.Option[string] { return secondary }

	result := primaryLookup().
		OrElse(secondaryLookup).
		Or("default")
	fmt.Println(result)
	// Output: backup
}

func ExampleNonZero() {
	fmt.Println(option.NonZero(42).IsOk())
	fmt.Println(option.NonZero(0).IsOk())
	fmt.Println(option.NonZero("hello").IsOk())
	fmt.Println(option.NonZero("").IsOk())
	// Output:
	// true
	// false
	// true
	// false
}

func ExampleOption_OrCall() {
	// expensiveDefault simulates a costly computation for the fallback value.
	expensiveDefault := func() string { return "computed-" + strconv.Itoa(42) }

	present := option.Of("cached").OrCall(expensiveDefault)
	fmt.Println(present)

	absent := option.NotOk[string]().OrCall(expensiveDefault)
	fmt.Println(absent)
	// Output:
	// cached
	// computed-42
}

func ExampleNew() {
	m := map[string]int{"alice": 30, "bob": 25}

	val, ok := m["alice"]
	ageOption := option.New(val, ok)
	fmt.Println(ageOption.Or(-1))

	val, ok = m["charlie"]
	ageOption = option.New(val, ok)
	fmt.Println(ageOption.Or(-1))
	// Output:
	// 30
	// -1
}

func ExampleNonEmpty() {
	// Simulate optional query parameter.
	query := "42"
	paramOption := option.NonEmpty(query)
	fmt.Println(paramOption.Or("default"))

	empty := ""
	paramOption = option.NonEmpty(empty)
	fmt.Println(paramOption.Or("default"))
	// Output:
	// 42
	// default
}

func ExampleNonNil() {
	name := "alice"
	present := option.NonNil(&name)
	fmt.Println(present.Or("unknown"))

	absent := option.NonNil[string](nil)
	fmt.Println(absent.Or("unknown"))
	// Output:
	// alice
	// unknown
}

func ExampleNonErr() {
	// strconv.Atoi returns (int, error) — NonErr bridges to Option.
	good := option.NonErr(strconv.Atoi("42"))
	fmt.Println(good.Or(-1))

	bad := option.NonErr(strconv.Atoi("abc"))
	fmt.Println(bad.Or(-1))
	// Output:
	// 42
	// -1
}

func ExampleOption_Get() {
	found, ok := option.NonZero("hello").Get()
	fmt.Printf("found=%q ok=%v\n", found, ok)

	found, ok = option.NonZero("").Get()
	fmt.Printf("found=%q ok=%v\n", found, ok)
	// Output:
	// found="hello" ok=true
	// found="" ok=false
}

func ExampleOption_KeepIf() {
	// isPositive reports whether n is greater than zero.
	isPositive := func(n int) bool { return n > 0 }

	fmt.Println(option.Of(42).KeepIf(isPositive).Or(-1))
	fmt.Println(option.Of(-5).KeepIf(isPositive).Or(-1))
	fmt.Println(option.NotOk[int]().KeepIf(isPositive).Or(-1))
	// Output:
	// 42
	// -1
	// -1
}

func ExampleOption_MustGet() {
	// MustGet is for values known to be present — panics if not-ok.
	cfg := option.Of("production")
	fmt.Println(cfg.MustGet())
	// Output: production
}

func ExampleLookup() {
	colors := map[string]string{"go": "blue", "rust": "orange"}

	fmt.Println(option.Lookup(colors, "go").Or("unknown"))
	fmt.Println(option.Lookup(colors, "python").Or("unknown"))
	// Output:
	// blue
	// unknown
}
