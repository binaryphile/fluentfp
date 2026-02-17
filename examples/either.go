//go:build ignore

package main

import (
	"fmt"
	"strconv"

	"github.com/binaryphile/fluentfp/either"
)

// This example demonstrates the either package for sum types.
// Either holds exactly one of two types: Left or Right.
// Convention: Left = failure, Right = success. Mnemonic: "Right is right."

func main() {
	// === Creating Either Values ===

	// Right creates a success value
	ok42 := either.Right[string, int](42)

	// Left creates a failure value
	fail := either.Left[string, int]("fail")

	// === Extracting Values ===

	// Get returns the Right value with comma-ok pattern
	if fortyTwo, ok := ok42.Get(); ok {
		fmt.Println("fortyTwo:", fortyTwo) // 42
	}

	// GetLeft returns the Left value with comma-ok pattern
	if err, ok := fail.GetLeft(); ok {
		fmt.Println("Got error:", err) // Got error: fail
	}

	// GetOr returns the Right value or a default
	zero := fail.GetOr(0)
	fmt.Println("zero:", zero) // 0

	// LeftOr returns the Left value or a default
	noError := ok42.LeftOr("no error")
	fmt.Println("noError:", noError) // no error

	// === Folding (Exhaustive Handling) ===

	// Fold handles both cases and returns a single result.
	// The first function handles Left, the second handles Right.

	// formatError returns a user-friendly error message.
	formatError := func(err string) string {
		return fmt.Sprintf("Error: %s", err)
	}

	// formatSuccess returns a success message with the value.
	formatSuccess := func(n int) string {
		return fmt.Sprintf("Success: %d", n)
	}

	message := either.Fold(ok42, formatError, formatSuccess)
	fmt.Println(message) // Success: 42

	message = either.Fold(fail, formatError, formatSuccess)
	fmt.Println(message) // Error: fail

	// === Mapping ===

	// Map transforms the Right value (type-preserving)
	// doubleInt doubles an integer.
	doubleInt := func(n int) int { return n * 2 }
	doubled := ok42.Map(doubleInt)
	fmt.Println("Doubled:", doubled.GetOr(0)) // Doubled: 84

	// either.Map transforms to a different type
	// intToString converts an int to its string representation.
	intToString := func(n int) string { return strconv.Itoa(n) }
	asString := either.Map(ok42, intToString)
	fmt.Println("As string:", asString.GetOr("")) // As string: 42

	// === Side Effects with IfRight/IfLeft ===

	// IfRight executes a function only if Right
	// printValue prints a value to stdout.
	printValue := func(n int) { fmt.Println("Calling with:", n) }
	ok42.IfRight(printValue)  // Calling with: 42
	fail.IfRight(printValue)  // (nothing printed)

	// IfLeft executes a function only if Left
	// logError logs an error message.
	logError := func(err string) { fmt.Println("Error logged:", err) }
	fail.IfLeft(logError) // Error logged: fail
	ok42.IfLeft(logError) // (nothing printed)

	// === Practical Example: Parse with Error Context ===

	result := parsePositiveInt("-5")
	// handleParseError returns a fallback value for parse errors.
	handleParseError := func(err ParseError) int { return err.Default }
	// useValue returns the parsed value unchanged.
	useValue := func(n int) int { return n }
	finalValue := either.Fold(result, handleParseError, useValue)
	fmt.Println("Final value:", finalValue) // Final value: 0
}

// ParseError contains context about why parsing failed.
type ParseError struct {
	Input   string
	Reason  string
	Default int
}

// parsePositiveInt parses a string as a positive integer.
// Returns Left with error context if invalid, Right with value if valid.
func parsePositiveInt(s string) either.Either[ParseError, int] {
	n, err := strconv.Atoi(s)
	if err != nil {
		return either.Left[ParseError, int](ParseError{
			Input:   s,
			Reason:  "not a number",
			Default: 0,
		})
	}
	if n <= 0 {
		return either.Left[ParseError, int](ParseError{
			Input:   s,
			Reason:  "not positive",
			Default: 0,
		})
	}
	return either.Right[ParseError, int](n)
}
