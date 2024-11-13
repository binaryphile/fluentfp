package main

import (
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/TeaEntityLab/fpGo/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/binaryphile/fluentfp/fluent"
	"github.com/binaryphile/fluentfp/hof"
	"github.com/rbrahul/gofp"
	"github.com/repeale/fp-go"
	u "github.com/rjNemo/underscore"
	"github.com/samber/lo"
	"github.com/seborama/fuego/v12"
	"github.com/thoas/go-funk"
	"slices"
)

func main() {
	users := []User{{Name: "Ren", Active: true}}

	fmt.Print("github.com/binaryphile/fluentfp/fluent\n")
	// assign or type-convert the existing slice to a fluent slice and the methods become available
	var fluentUsers fluent.SliceOf[User] = users
	fluentUsers.
		KeepIf(User.IsActive).
		ToString(User.GetName).
		Each(hof.Println) // helper to convert string argument to `any` required by fmt.Println

	fmt.Print("\ngithub.com/rjNemo/underscore\n")
	actives := u.Filter(users, User.IsActive)
	names := u.Map(actives, User.GetName)
	u.Each(names, hof.Println)

	fmt.Print("\ngithub.com/repeale/fp-go\n")
	// fp-go offers Map and Filter factories that return their respective operations closed over the factory's argument.
	// The resulting functions are then called with the slice.
	actives = fp.Filter(User.IsActive)(users)
	names = fp.Map(User.GetName)(actives)
	for _, name := range names { // there's no each in fp-go, so loop instead
		fmt.Println(name)
	}

	fmt.Print("\ngithub.com/thoas/go-funk\n")
	actives = funk.Filter(users, User.IsActive).([]User)
	names = funk.Map(actives, User.GetName).([]string)
	funk.ForEach(names, hof.Println)

	fmt.Print("\ngithub.com/ahmetb/go-linq/v3\n")
	activeUserQuery := linq.From(users).Where(func(user any) bool {
		return user.(User).IsActive()
	})
	nameQuery := activeUserQuery.Select(func(user any) any {
		return user.(User).GetName()
	})
	nameQuery.ForEach(func(name any) {
		fmt.Println(name)
	})

	fmt.Print("\ngithub.com/seborama/fuego/v12\n")
	// fuego operates on streams, so a stream is created from the slice.
	// It is fluent but requires wrapping some functions to return `fuego.Any`.
	userGetName := func(u User) fuego.Any {
		return u.GetName()
	}
	printName := func(name fuego.Any) {
		fmt.Println(name)
	}
	fuego.NewStreamFromSlice(users, 1).
		Filter(User.IsActive). // Filter is an exception to wrapping since it expects bool
		Map(userGetName).
		ForEach(printName)

	fmt.Print("\ngithub.com/samber/lo\n")
	// lo requires function arguments to accept an index, so existing functions must be wrapped.
	userIsActiveIdx := func(u User, _ int) bool {
		return u.IsActive()
	}
	userGetNameIdx := func(u User, _ int) string {
		return u.GetName()
	}
	printlnIdx := func(s string, _ int) {
		fmt.Println(s)
	}
	actives = lo.Filter(users, userIsActiveIdx)
	names = lo.Map(actives, userGetNameIdx)
	lo.ForEach(names, printlnIdx)

	fmt.Print("\ngithub.com/BooleanCat/go-functional/v2/it\n")
	// go-functional generates iterators that you loop over.
	// It is for go 1.23 and on.
	userIter := slices.Values(users)
	activeUserIter := it.Filter(userIter, User.IsActive)
	namesIter := it.Map(activeUserIter, User.GetName)
	for name := range namesIter {
		fmt.Println(name)
	}

	fmt.Print("\ngithub.com/TeaEntityLab/fpGo/v2\n")
	// fpgo offers Map and Filter functions for slices.
	// Indexing is required, so existing functions must be wrapped.
	userIsActiveIdx = func(u User, _ int) bool {
		return u.IsActive()
	}
	actives = fpgo.Filter(userIsActiveIdx, users...) // variadic arguments are required
	names = fpgo.Map(User.GetName, users...)
	for _, name := range names { // there's no each
		fmt.Println(name)
	}

	fmt.Print("\ngithub.com/rbrahul/gofp\n")
	// gofp requires wrapping function arguments but also requires slices of `any`.
	// Converting a slice to `any` is done with a helper function just for this example.
	usersAny := AnySliceOf(users)
	userIsActiveAny := func(_ int, a any) bool {
		return a.(User).IsActive()
	}
	userGetNameAny := func(_ int, a any) any {
		return a.(User).GetName()
	}
	activesAny := gofp.Filter(usersAny, userIsActiveAny)
	namesAny := gofp.Map(activesAny, userGetNameAny)
	for _, name := range namesAny {
		fmt.Println(name)
	}
}

// User definition

type User struct {
	Active bool
	Name   string
}

func (u User) GetName() string {
	return u.Name
}

func (u User) IsActive() bool {
	return u.Active
}

// helpers

func AnySliceOf[T any](ts []T) []any {
	result := make([]any, len(ts))

	for i, t := range ts {
		result[i] = t
	}

	return result
}
