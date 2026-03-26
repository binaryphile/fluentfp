package slice_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/binaryphile/fluentfp/slice"
)

type ticket struct {
	id   string
	done bool
}

func (t ticket) IsDone() bool  { return t.done }
func (t ticket) GetID() string { return t.id }

func ExampleMapper_KeepIf() {
	// isEven reports whether n is divisible by 2.
	isEven := func(n int) bool { return n%2 == 0 }

	evens := slice.From([]int{1, 2, 3, 4, 5}).KeepIf(isEven)
	fmt.Println(evens)
	// Output: [2 4]
}

func ExampleMap() {
	// label formats an int as a labeled string.
	label := func(n int) string { return fmt.Sprintf("n=%d", n) }

	strs := slice.Map([]int{1, 2, 3}, label)
	fmt.Println(strs)
	// Output: [n=1 n=2 n=3]
}

func ExampleFold() {
	// sum adds two integers.
	sum := func(acc, n int) int { return acc + n }

	total := slice.Fold([]int{1, 2, 3, 4}, 0, sum)
	fmt.Println(total)
	// Output: 10
}

func ExampleFrom() {
	tickets := []ticket{
		{"T-1", true},
		{"T-2", false},
		{"T-3", true},
		{"T-4", false},
	}

	doneIDs := slice.From(tickets).
		KeepIf(ticket.IsDone).
		ToString(ticket.GetID)
	fmt.Println(doneIDs)
	// Output: [T-1 T-3]
}

func ExampleFlatMap() {
	// toWords splits a sentence into individual words.
	toWords := func(s string) []string { return strings.Fields(s) }

	words := slice.FlatMap([]string{"hello world", "go is fun"}, toWords)
	fmt.Println(words)
	// Output: [hello world go is fun]
}

func ExampleFilterMap() {
	// tryParseInt parses a string as an integer, reporting success via the bool.
	tryParseInt := func(s string) (int, bool) {
		n, err := strconv.Atoi(s)
		return n, err == nil
	}

	nums := slice.FilterMap([]string{"1", "abc", "3", "", "5"}, tryParseInt)
	fmt.Println(nums)
	// Output: [1 3 5]
}

func ExampleReduce() {
	// longer returns whichever string is longer.
	longer := func(a, b string) string {
		if len(b) > len(a) {
			return b
		}
		return a
	}

	longest, _ := slice.Reduce([]string{"go", "rust", "c"}, longer).Get()
	fmt.Println(longest)
	// Output: rust
}

func ExampleScan() {
	// addAmount adds a transaction amount to a running balance.
	addAmount := func(balance, amount int) int { return balance + amount }

	balances := slice.Scan([]int{10, -3, 5, -2}, 100, addAmount)
	fmt.Println(balances)
	// Output: [100 110 107 112 110]
}

func ExampleMapAccum() {
	// numberLine prepends a line number to the text.
	numberLine := func(n int, line string) (int, string) {
		return n + 1, fmt.Sprintf("%d: %s", n, line)
	}

	_, numbered := slice.MapAccum([]string{"foo", "bar", "baz"}, 1, numberLine)
	fmt.Println(numbered)
	// Output: [1: foo 2: bar 3: baz]
}
