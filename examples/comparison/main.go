//go:build ignore

// Package main demonstrates 10 Go FP libraries doing the same task:
// filter active users, extract names, print them, return as []string.
// Then compares operations beyond filter+map: Find, Reduce, chaining, Unzip.
//
// Run: go run examples/comparison/main.go
// See comparison.md for summary table and benchmarks.
//
// Quick reference (T=Type-safe, C=Concise, M=Method exprs, F=Fluent):
//
//	Library        Lines  T C M F
//	fluentfp         5    ✓ ✓ ✓ ✓
//	lo              13    ✓ . . .   index params prevent method exprs
//	go-funk          4    . ✓ ✓ .   uses any, requires type assertions
//	go-linq         22    . . . ✓   uses any, painful slice conversion
//	underscore       4    ✓ ✓ ✓ .   best alternative
//	fp-go            6    ✓ ✓ ✓ .   curried API, lacks Each
//	go-functional    8    ✓ ✓ ✓ .   iterators are single-use
//	fpGo             9    ✓ . ~ .   Filter needs index, Map doesn't
//	fuego           12    . ✓ ~ ✓   streams are single-use, uses Any
//	gofp            20    . . . .   must convert to []any first
//
// Note: lof.Println (wraps fmt.Println) is available to all libraries,
// but only those accepting func(T) callbacks can use it directly.
// Libraries requiring func(T, int) must write their own wrapper.
//
// Structure: Each library is wrapped in a block { ... } with its own
// printActiveNames function. Blocks scope the function definitions,
// allowing each library to define identically-named functions without conflict.
package main

import (
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	fpgo "github.com/TeaEntityLab/fpGo/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/slice"
	"github.com/rbrahul/gofp"
	"github.com/repeale/fp-go"
	u "github.com/rjNemo/underscore"
	"github.com/samber/lo"
	"github.com/seborama/fuego/v12"
	"github.com/thoas/go-funk"
	"slices"
)

func main() {
	users := []User{
		{
			name:   "Ren",
			active: true,
		},
		{
			name:   "Stimpy",
			active: false,
		},
	}

	fmt.Print("github.com/binaryphile/fluentfp/slice\n")
	// === fluentfp (5 lines) ===
	// Method expressions, fluent chaining. lof.Println provided by library.
	{
		printActiveNames := func(users []User) []string { // signature automatically converts return type
			names := slice.From(users).
				KeepIf(User.IsActive).
				ToString(User.Name) // returns slice.String
			names.Each(lof.Println) // from lower-order function helper package lof
			return names            // signature converts slice.String to []string
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/samber/lo\n")
	// === lo (13 lines) ===
	// Requires index parameter in callbacks. Cannot use lof.Println.
	{
		printActiveNames := func(users []User) []string {
			userIsActive := func(u User, _ int) bool {
				return u.IsActive()
			}
			getName := func(u User, _ int) string {
				return u.Name()
			}
			printLn := func(s string, _ int) {
				fmt.Println(s)
			}
			actives := lo.Filter(users, userIsActive)
			names := lo.Map(actives, getName)
			lo.ForEach(names, printLn)
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/thoas/go-funk\n")
	// === go-funk (4 lines) ===
	// Requires type assertions (not type-safe). Runtime overhead from boxing/unboxing.
	// Can use lof.Println.
	{
		printActiveNames := func(users []User) []string {
			actives := funk.Filter(users, User.IsActive).([]User)
			names := funk.Map(actives, User.Name).([]string)
			funk.ForEach(names, lof.Println)
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/ahmetb/go-linq/v3\n")
	// === go-linq (22 lines) ===
	// Query objects require any wrappers. Runtime overhead from boxing/unboxing.
	// Painful to get back to []string.
	{
		printActiveNames := func(users []User) []string {
			userIsActive := func(user any) bool {
				return user.(User).IsActive()
			}
			name := func(user any) any {
				return user.(User).Name()
			}
			printLn := func(a any) {
				fmt.Println(a)
			}
			nameQuery := linq.
				From(users).
				Where(userIsActive).
				Select(name)
			nameQuery.ForEach(printLn)

			var anys []any
			nameQuery.ToSlice(&anys)
			names := make([]string, len(anys))
			for i, a := range anys {
				names[i] = a.(string)
			}
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/rjNemo/underscore\n")
	// === underscore (4 lines) ===
	// Best alternative. Not fluent but clean. Can use lof.Println.
	{
		printActiveNames := func(users []User) []string {
			actives := u.Filter(users, User.IsActive)
			names := u.Map(actives, User.Name)
			u.Each(names, lof.Println)
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/repeale/fp-go\n")
	// === fp-go (6 lines) ===
	// Curried API: Filter(pred)(slice). Lacks Each.
	{
		printActiveNames := func(users []User) []string {
			actives := fp.Filter(User.IsActive)(users) // Filter returns a function that is called with a slice
			names := fp.Map(User.Name)(actives)        // same for Map
			for _, name := range names {
				fmt.Println(name)
			}
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/BooleanCat/go-functional/v2/it\n")
	// === go-functional (8 lines) ===
	// Go 1.23+ iterators. Single-use (must collect before reuse).
	{
		printActiveNames := func(users []User) []string {
			userSeq := slices.Values(users)
			activeSeq := it.Filter(userSeq, User.IsActive)
			nameSeq := it.Map(activeSeq, User.Name)
			names := slices.Collect(nameSeq) // Collect first (consumes iterator)
			for _, name := range names {
				fmt.Println(name)
			}
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/TeaEntityLab/fpGo/v2\n")
	// === fpGo (9 lines) ===
	// Variadic args. Filter needs index wrapper, Map doesn't. Lacks Each.
	{
		printActiveNames := func(users []User) []string {
			userIsActive := func(u User, _ int) bool {
				return u.IsActive()
			}
			actives := fpgo.Filter(userIsActive, users...)
			names := fpgo.Map(User.Name, actives...)
			for _, name := range names {
				fmt.Println(name)
			}
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/seborama/fuego/v12\n")
	// === fuego (12 lines) ===
	// Stream-based. Single-use (like Java). Map must return fuego.Any.
	// Runtime overhead from boxing/unboxing.
	{
		printActiveNames := func(users []User) []string {
			getName := func(u User) fuego.Any { return u.Name() }
			nameStream := fuego.NewStreamFromSlice(users, 1).
				Filter(User.IsActive).
				Map(getName)
			toSlice := fuego.ToSlice[fuego.Any]()
			anys := fuego.Collect(nameStream, toSlice) // Collect first (consumes stream)
			names := make([]string, len(anys))
			for i, a := range anys {
				names[i] = a.(string)
				fmt.Println(names[i])
			}
			return names
		}
		printActiveNames(users)
	}

	fmt.Print("\ngithub.com/rbrahul/gofp\n")
	// === gofp (20 lines) ===
	// Must convert input to []any first. Runtime overhead from boxing/unboxing.
	{
		printActiveNames := func(users []User) []string {
			anyUsers := make([]any, len(users))
			for i, user := range users {
				anyUsers[i] = user
			}
			userIsActive := func(_ int, a any) bool {
				return a.(User).IsActive()
			}
			getName := func(_ int, a any) any {
				return a.(User).Name()
			}
			anyActives := gofp.Filter(anyUsers, userIsActive)
			anyNames := gofp.Map(anyActives, getName)

			names := make([]string, len(anyNames))
			for i, a := range anyNames {
				names[i] = a.(string)
				fmt.Println(names[i])
			}

			return names
		}
		printActiveNames(users)
	}

	// --- Beyond Filter+Map ---
	// Additional operations comparing fluentfp against the strongest competitor.

	chainUsers := []User{
		{name: "Ren", active: true, age: 25},
		{name: "Stimpy", active: false, age: 30},
		{name: "Ren", active: true, age: 28}, // duplicate name for Unique/Distinct demo
	}

	// === Find: fluentfp vs lo ===
	fmt.Print("\n--- Find ---\n")

	fmt.Print("\nfluentfp Find\n")
	{
		user := slice.From(users).Find(User.IsActive).Or(User{name: "nobody"})
		fmt.Println(user.Name())
	}

	fmt.Print("\nlo Find\n")
	{
		user, ok := lo.Find(users, func(u User) bool { return u.IsActive() })
		if !ok {
			user = User{name: "nobody"}
		}
		fmt.Println(user.Name())
	}

	// === Chaining: fluentfp vs go-linq ===
	fmt.Print("\n--- Multi-step chain ---\n")

	fmt.Print("\nfluentfp chain\n")
	{
		names := slice.From(chainUsers).
			KeepIf(User.IsActive).
			ToString(User.Name).
			Unique()
		names.Each(lof.Println)
	}

	fmt.Print("\ngo-linq chain\n")
	{
		userIsActive := func(user any) bool { return user.(User).IsActive() }
		getName := func(user any) any { return user.(User).Name() }
		var results []any
		linq.From(chainUsers).
			Where(userIsActive).
			Select(getName).
			Distinct().
			ToSlice(&results)
		names := make([]string, len(results))
		for i, r := range results {
			names[i] = r.(string)
		}
		for _, name := range names {
			fmt.Println(name)
		}
	}

	// === Unzip: fluentfp (unique) ===
	// Extracts multiple fields in a single O(n) traversal.
	// No other library has Unzip — lo requires one pass per field.
	fmt.Print("\n--- Multi-field extraction ---\n")

	fmt.Print("\nfluentfp Unzip2 (1 pass)\n")
	{
		names, ages := slice.Unzip2(chainUsers, User.Name, User.Age)
		fmt.Println("names:", []string(names))
		fmt.Println("ages:", []int(ages))
	}

	fmt.Print("\nlo Map (2 separate passes)\n")
	{
		names := lo.Map(chainUsers, func(u User, _ int) string { return u.Name() })
		ages := lo.Map(chainUsers, func(u User, _ int) int { return u.Age() })
		fmt.Println("names:", names)
		fmt.Println("ages:", ages)
	}
}

// User definition

type User struct {
	active bool
	name   string
	age    int
}

func (u User) Name() string {
	return u.name
}

func (u User) IsActive() bool {
	return u.active
}

func (u User) Age() int {
	return u.age
}
