package option_test

import (
	"fmt"
	"strings"

	"github.com/binaryphile/fluentfp/option"
)

func ExampleMap() {
	// strLen returns the length of a string.
	strLen := func(s string) int { return len(s) }

	present := option.Map(option.Of("hello"), strLen)
	fmt.Println(present.Or(0))

	absent := option.Map(option.NotOk[string](), strLen)
	fmt.Println(absent.Or(0))
	// Output:
	// 5
	// 0
}

func ExampleFlatMap() {
	users := map[string]string{"alice": "eng", "bob": "sales"}
	depts := map[string]string{"eng": "Engineering", "sales": "Sales"}

	// lookupDept resolves a department code to its full name.
	lookupDept := func(code string) option.Option[string] {
		return option.Lookup(depts, code)
	}

	userDept := option.Lookup(users, "alice")
	deptName := option.FlatMap(userDept, lookupDept)
	fmt.Println(deptName.Or("unknown"))

	missingUser := option.Lookup(users, "charlie")
	deptName = option.FlatMap(missingUser, lookupDept)
	fmt.Println(deptName.Or("unknown"))
	// Output:
	// Engineering
	// unknown
}

func ExampleOption_Transform() {
	upper := option.Of("hello").Transform(strings.ToUpper)
	fmt.Println(upper.Or(""))

	absent := option.NotOk[string]().Transform(strings.ToUpper)
	fmt.Println(absent.Or("fallback"))
	// Output:
	// HELLO
	// fallback
}

func ExampleZipWith() {
	// fullName joins first and last name with a space.
	fullName := func(first, last string) string { return first + " " + last }

	first := option.Of("Alice")
	last := option.Of("Smith")
	fmt.Println(option.ZipWith(first, last, fullName).Or("anonymous"))

	first = option.NotOk[string]()
	fmt.Println(option.ZipWith(first, last, fullName).Or("anonymous"))
	// Output:
	// Alice Smith
	// anonymous
}

func ExampleOption_OkOr() {
	val, err := option.Of(42).OkOr(fmt.Errorf("missing")).Unpack()
	fmt.Println(val, err)

	val, err = option.NotOk[int]().OkOr(fmt.Errorf("missing")).Unpack()
	fmt.Println(val, err)
	// Output:
	// 42 <nil>
	// 0 missing
}
