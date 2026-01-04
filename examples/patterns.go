//go:build ignore

// Package main demonstrates FluentFP patterns vs conventional Go.
//
// Each example illustrates two key insights:
//
// 1. CONCERNS FACTORED, NOT ELIMINATED
//    The library does the same work (make, range, assign) - just once instead of
//    at every call site. You specify only what varies.
//
// 2. THE INVISIBLE FAMILIARITY DISCOUNT
//    A loop you've seen 10,000 times feels instant to parse - but that's learned
//    pattern recognition, not inherent simplicity. FluentFP expresses intent
//    without mechanics to parse.
package main

import (
	"fmt"
	"time"

	"github.com/binaryphile/fluentfp/slice"
)

func main() {
	devs := []Developer{
		{Name: "Alice", CurrentTicket: ""},
		{Name: "Bob", CurrentTicket: "TKT-123"},
		{Name: "Carol", CurrentTicket: ""},
	}

	fmt.Println("=== Pattern 1: Filter with Method Expression ===")
	fmt.Println("The most dramatic improvement - 1 line vs 7")
	fmt.Println()
	filterExample(devs)

	fmt.Println("\n=== Pattern 2: Field Extraction (Map) ===")
	fmt.Println("Extract a field from each element - 1 line vs 5")
	fmt.Println()
	extractExample()

	fmt.Println("\n=== Pattern 3: Filter + Count Chain ===")
	fmt.Println("Compose operations - reads top to bottom")
	fmt.Println()
	chainExample(devs)

	fmt.Println("\n=== Pattern 4: Fold (Reduce) ===")
	fmt.Println("Accumulate values - named reducer documents intent")
	fmt.Println()
	foldExample()
}

// =============================================================================
// Pattern 1: Filter with Method Expression
// =============================================================================

func filterExample(devs []Developer) {
	// FLUENTFP: What you write
	// --------------------------
	// One concept: "keep developers who are idle"
	idle := slice.From(devs).KeepIf(Developer.IsIdle)

	fmt.Printf("FluentFP: %d idle developers\n", len(idle))

	// CONVENTIONAL: What you'd write without FluentFP
	// ------------------------------------------------
	// Four concepts interleaved: declare, range, if, append
	//
	// var result []Developer
	// for _, d := range devs {
	//     if d.IsIdle() {
	//         result = append(result, d)
	//     }
	// }

	// THE INVISIBLE FAMILIARITY DISCOUNT:
	// You've seen that loop pattern thousands of times, so it feels instant.
	// But show it to a non-programmer alongside KeepIf(Developer.IsIdle) -
	// which one can they understand?

	// CONCERNS FACTORED, NOT ELIMINATED:
	// The library still does: make, range, if, append
	// The difference: written once in slice/mapper.go, not at every call site
	// You specify only what varies: the predicate (Developer.IsIdle)
}

// =============================================================================
// Pattern 2: Field Extraction (Map)
// =============================================================================

func extractExample() {
	history := []Snapshot{
		{Day: 1, PercentUsed: 10.5},
		{Day: 2, PercentUsed: 25.3},
		{Day: 3, PercentUsed: 42.1},
	}

	// FLUENTFP: What you write
	// --------------------------
	// One concept: "extract PercentUsed as float64s"
	values := slice.From(history).ToFloat64(Snapshot.GetPercentUsed)

	fmt.Printf("FluentFP: extracted %v\n", values)

	// CONVENTIONAL: What you'd write without FluentFP
	// ------------------------------------------------
	// Four concepts: make with len, range with index, assign by position, return
	//
	// result := make([]float64, len(history))
	// for i, s := range history {
	//     result[i] = s.PercentUsed
	// }

	// THE INVISIBLE FAMILIARITY DISCOUNT:
	// The index variable 'i' is bookkeeping you don't care about.
	// But you've written it so many times, you don't notice the noise.

	// CONCERNS FACTORED, NOT ELIMINATED:
	// Library does: make([]float64, len), for i, range, results[i] = fn(t)
	// You specify only: the extraction function (Snapshot.GetPercentUsed)
}

// =============================================================================
// Pattern 3: Filter + Count Chain
// =============================================================================

func chainExample(devs []Developer) {
	// FLUENTFP: What you write
	// --------------------------
	// Two operations, two lines, reads top to bottom
	idleCount := slice.From(devs).
		KeepIf(Developer.IsIdle).
		Len()

	fmt.Printf("FluentFP: %d idle developers\n", idleCount)

	// CONVENTIONAL: What you'd write without FluentFP
	// ------------------------------------------------
	// All concerns interleaved in one loop body
	//
	// count := 0
	// for _, d := range devs {
	//     if d.IsIdle() {
	//         count++
	//     }
	// }

	// THE INVISIBLE FAMILIARITY DISCOUNT:
	// After 6 months, you'll re-read the loop and ask: "what is this counting?"
	// The chain says it directly: KeepIf(IsIdle).Len()
}

// =============================================================================
// Pattern 4: Fold (Reduce)
// =============================================================================

func foldExample() {
	durations := []time.Duration{
		2 * time.Hour,
		3 * time.Hour,
		1 * time.Hour,
	}

	// FLUENTFP: What you write
	// --------------------------
	// Named reducer documents intent

	// sumDuration adds two durations.
	sumDuration := func(a, b time.Duration) time.Duration { return a + b }
	total := slice.Fold(durations, time.Duration(0), sumDuration)

	fmt.Printf("FluentFP: total = %v\n", total)

	// CONVENTIONAL: What you'd write without FluentFP
	// ------------------------------------------------
	// Mutation-based accumulation
	//
	// var total time.Duration
	// for _, d := range durations {
	//     total += d
	// }

	// This one's close in line count. The win is:
	// - Named reducer (sumDuration) is reusable and testable
	// - Zero value is explicit parameter, not implicit var declaration
	// - No mutation (functional style)
}

// =============================================================================
// Types (minimal, for demonstration)
// =============================================================================

// Developer represents a team member.
type Developer struct {
	Name          string
	CurrentTicket string
}

// IsIdle returns true if the developer has no assigned ticket.
func (d Developer) IsIdle() bool { return d.CurrentTicket == "" }

// Snapshot captures a metric at a point in time.
type Snapshot struct {
	Day         int
	PercentUsed float64
}

// GetPercentUsed returns the percentage value.
func (s Snapshot) GetPercentUsed() float64 { return s.PercentUsed }
