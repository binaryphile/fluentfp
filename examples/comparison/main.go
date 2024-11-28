//go:build ignore

package main

import (
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	fpgo "github.com/TeaEntityLab/fpGo/v2"
	"github.com/ahmetb/go-linq/v3"
	"github.com/binaryphile/fluentfp/fluent"
	"github.com/binaryphile/fluentfp/lof"
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
			Name:   "Ren",
			Active: true,
		},
		{
			Name:   "Stimpy",
			Active: false,
		},
	}

	// Each example prints the user's name if they are active.
	// Examples return a native slice of names to show interoperability with code that expects slices.

	fmt.Print("github.com/binaryphile/fluentfp/fluent\n")
	// fluentfp is the most concise library.
	// It is type-safe, which aligns with functional principles.
	// It is fluent, meaning you can chain operations.
	// It works with method expressions.
	{
		printActiveNames := func(users fluent.SliceOf[User]) []string { // signature automatically converts types
			names := users.
				KeepIf(User.IsActive).
				ToString(User.GetName) // returns fluent.SliceOf[string]
			names.Each(lof.Println) // helper function from fluentfp/hof
			return names            // signature converts fluent.SliceOf[string] to []string
		}

		_ = printActiveNames(users) // signature converts []User to fluent.SliceOf[User]
	}

	fmt.Print("\ngithub.com/samber/lo\n")
	// lo is the most popular library with over 17k GitHub stars.
	// lo is not concise because it requires function arguments to accept an index, which are painful to wrap.
	// It is type-safe.
	// It is not fluent.
	// It does not work with method expressions.
	{
		printActiveNames := func(users []User) []string {
			userIsActive := func(u User, _ int) bool {
				return u.IsActive()
			}
			getName := func(u User, _ int) string {
				return u.GetName()
			}
			printLn := func(s string, _ int) {
				fmt.Println(s)
			}
			actives := lo.Filter(users, userIsActive)
			names := lo.Map(actives, getName)
			lo.ForEach(names, printLn)
			return names
		}

		_ = printActiveNames(users)
	}

	fmt.Print("\ngithub.com/thoas/go-funk\n")
	// go-funk is the second most popular library with over 4k GitHub stars.
	// go-funk is nearly as concise as fluentfp, but requires type assertions.
	// It is not type-safe, which does not align with functional principles.
	// It is not fluent.
	// It works with method expressions.
	{
		printActiveNames := func(users []User) []string {
			actives := funk.Filter(users, User.IsActive).([]User)
			names := funk.Map(actives, User.GetName).([]string)
			funk.ForEach(names, lof.Println)
			return names
		}

		_ = printActiveNames(users)
	}

	fmt.Print("\ngithub.com/ahmetb/go-linq/v3\n")
	// go-linq is the third most popular library with over 3k GitHub stars.
	// go-linq is not concise and relies on Query objects, which are painful to get back to a slice.
	// It is not type-safe.
	// It is fluent.
	// It does not work with method expressions.
	{
		printActiveNames := func(users []User) []string {
			userIsActive := func(user any) bool {
				return user.(User).IsActive()
			}
			name := func(user any) any {
				return user.(User).GetName()
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

		_ = printActiveNames(users)
	}

	// None of the rest of the libraries have over 500 stars.

	fmt.Print("\ngithub.com/rjNemo/underscore\n")
	// underscore is the best alternative to fluentfp.
	// It is type-safe.
	// It is not quite as concise as fluentfp due to not being fluent.
	// It works with method expressions.
	{
		printActiveNames := func(users []User) []string {
			actives := u.Filter(users, User.IsActive)
			names := u.Map(actives, User.GetName)
			u.Each(names, lof.Println)
			return names
		}

		_ = printActiveNames(users)
	}

	fmt.Print("\ngithub.com/repeale/fp-go\n")
	// fp-go is not as concise, lacking Each.
	// It is type-safe.
	// It is not fluent.
	// It works with method expressions.
	{
		printActiveNames := func(users []User) []string {
			actives := fp.Filter(User.IsActive)(users) // Filter returns a function that is called with a slice
			names := fp.Map(User.GetName)(actives)     // same for Map
			for _, name := range names {
				fmt.Println(name)
			}
			return names
		}

		_ = printActiveNames(users)
	}

	fmt.Print("\ngithub.com/BooleanCat/go-functional/v2/it\n")
	// go-functional is not as concise and relies on Go 1.23+ iterators.
	// Converting back to slice is slightly awkward.
	// It is type-safe.
	// It is not fluent.
	// It works with method expressions.
	{
		printActiveNames := func(users []User) []string {
			userSeq := slices.Values(users)
			activeSeq := it.Filter(userSeq, User.IsActive)
			nameSeq := it.Map(activeSeq, User.GetName)
			for name := range nameSeq {
				fmt.Println(name)
			}
			names := slices.Collect(nameSeq)
			return names
		}

		_ = printActiveNames(users)
	}

	fmt.Print("\ngithub.com/TeaEntityLab/fpGo/v2\n")
	// fpGo is not as concise and requires variadic arguments, which is slightly awkward.
	// It lacks Each.
	// It is type-safe.
	// It is not fluent.
	// Map works with method expressions, but Filter does not.
	{
		printActiveNames := func(users []User) []string {
			userIsActive := func(u User, _ int) bool {
				return u.IsActive()
			}
			actives := fpgo.Filter(userIsActive, users...)
			names := fpgo.Map(User.GetName, actives...)
			for _, name := range names {
				fmt.Println(name)
			}
			return names
		}

		_ = printActiveNames(users)
	}

	fmt.Print("\ngithub.com/seborama/fuego/v12\n")
	// fuego is not concise and relies on a stream implementation.  Converting back to a slice is painful.
	// It is not type-safe.
	// It is fluent.
	// It does not work with method expressions, requiring function arguments to return fuego.Any.
	// Its collectors require manufacturing, which adds to the complexity of returning a slice.
	{
		printActiveNames := func(users []User) []string {
			getName := func(u User) fuego.Any {
				return u.GetName()
			}
			printLn := func(a fuego.Any) {
				fmt.Println(a)
			}
			nameStream := fuego.NewStreamFromSlice(users, 1).
				Filter(User.IsActive).
				Map(getName)
			nameStream.ForEach(printLn)

			toSlice := fuego.ToSlice[fuego.Any]()
			anys := fuego.Collect(nameStream, toSlice)
			names := make([]string, len(anys))
			for i, a := range anys {
				names[i] = a.(string)
			}
			return names
		}

		_ = printActiveNames(users)
	}

	fmt.Print("\ngithub.com/rbrahul/gofp\n")
	// gofp is not concise and relies heavily on `any` types, which is painful.
	// It is not type-safe.
	// It is not fluent.
	// It does not work with method expressions.
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
				return a.(User).GetName()
			}
			anyActives := gofp.Filter(anyUsers, userIsActive)
			anyNames := gofp.Map(anyActives, getName)
			for _, anyName := range anyNames {
				fmt.Println(anyName.(string))
			}

			names := make([]string, len(anyNames))
			for i, a := range anyNames {
				names[i] = a.(string)
			}
			return names
		}

		_ = printActiveNames(users)
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
