package slice_test

import (
	"fmt"
	"strings"

	"github.com/binaryphile/fluentfp/slice"
)

func ExampleMapper_Find() {
	// isNegative reports whether n is less than zero.
	isNegative := func(n int) bool { return n < 0 }

	found, _ := slice.From([]int{3, 1, -4, 1, -5}).Find(isNegative).Get()
	fmt.Println(found)
	// Output: -4
}

func ExampleMapper_Any() {
	// above90 reports whether a score exceeds 90.
	above90 := func(n int) bool { return n > 90 }

	fmt.Println(slice.From([]int{72, 85, 93, 64}).Any(above90))
	// Output: true
}

func ExampleMapper_Every() {
	// nonEmpty reports whether a string is non-empty.
	nonEmpty := func(s string) bool { return s != "" }

	fmt.Println(slice.From([]string{"alice", "bob", "carol"}).Every(nonEmpty))
	// Output: true
}

func ExampleMapper_Transform() {
	uppers := slice.From([]string{"hello", "world"}).Transform(strings.ToUpper)
	fmt.Println(uppers)
	// Output: [HELLO WORLD]
}

func ExampleContains() {
	tags := []string{"bug", "urgent", "backend"}
	fmt.Println(slice.Contains(tags, "urgent"))
	// Output: true
}

func ExampleIndexOf() {
	idx, _ := slice.IndexOf([]string{"a", "b", "c", "d"}, "c").Get()
	fmt.Println(idx)
	// Output: 2
}

func ExampleUnique() {
	tags := slice.Unique([]string{"go", "rust", "go", "python", "rust"})
	fmt.Println(tags)
	// Output: [go rust python]
}

func ExampleUniqueBy() {
	type user struct {
		name  string
		email string
	}

	users := []user{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Alice2", "alice@example.com"},
	}

	// userEmail extracts the email field for deduplication.
	userEmail := func(u user) string { return u.email }

	unique := slice.UniqueBy(users, userEmail)
	fmt.Println(len(unique))
	// Output: 2
}

func ExampleNonZero() {
	sparse := slice.NonZero([]int{0, 1, 0, 2, 0, 3})
	fmt.Println(sparse)
	// Output: [1 2 3]
}

func ExampleMinBy() {
	// wordLen returns the length of a string.
	wordLen := func(s string) int { return len(s) }

	shortest, _ := slice.MinBy([]string{"go", "rust", "c", "python"}, wordLen).Get()
	fmt.Println(shortest)
	// Output: c
}
