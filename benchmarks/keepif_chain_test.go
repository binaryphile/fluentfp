// Package benchmarks_test compares multi-chain KeepIf performance across
// eager (slice.Mapper) and lazy (seq.Seq) evaluation strategies.
//
// # Why this benchmark exists
//
// Chaining multiple eager Mapper.KeepIf calls materializes an intermediate
// slice at each stage. For long filter pipelines (e.g. 7 chained predicates),
// this can produce significant allocation overhead compared to a single-pass
// loop. The seq package avoids intermediate materialization by composing
// closures over iter.Seq, deferring allocation until a terminal like Collect.
//
// # What this measures
//
// Five variants isolate different cost components:
//
//   - LoopDirect: single loop with inline conditions (best possible baseline)
//   - LoopFuncs: single loop calling predicate functions, preallocated cap
//     (isolates function-call overhead from allocation overhead)
//   - LoopFuncsNoHint: same as LoopFuncs but with no capacity hint (isolates
//     collection-growth overhead from evaluation strategy)
//   - MapperChain: eager Mapper.KeepIf chaining (intermediate slice per stage)
//   - SeqCollect: lazy seq.KeepIf chaining with terminal Collect
//
// Two predicate orderings test the effect of predicate order:
//
//   - SelectiveFirst: high-rejection predicates early (reduces later stage sizes)
//   - SelectiveLast: high-rejection predicates late (all stages process near-full input)
//
// Order affects both short-circuit evaluation (CPU) and per-predicate cost
// placement. Predicates are not identical in cost (field reads, modulo,
// division, closure capture), so ordering effects reflect both selectivity
// and cost differences.
//
// Control benchmarks use always-true predicates to attribute seq allocation
// overhead between closure creation and terminal materialization, with output
// size held constant.
//
// # Interpreting results
//
// Predicate order affects CPU time for all variants due to short-circuit
// evaluation and varying predicate cost. For eager chains, order also affects
// allocation volume because each stage allocates proportional to its input
// size. For seq, allocation volume is independent of order since only the
// terminal allocates.
//
// SeqCollect uses slices.Collect which has no size hint, so its allocation
// profile reflects both lazy composition and terminal collection growth
// strategy. Compare with LoopFuncsNoHint to separate these effects.
//
// These benchmarks use cheap integer predicates and regular synthetic data.
// Results are specific to this workload; confirm with repeated runs,
// additional data shapes, and stable execution environments before
// generalizing.
package benchmarks_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/binaryphile/fluentfp/seq"
	"github.com/binaryphile/fluentfp/slice"
)

// benchItem is a minimal struct for benchmarking filter chains.
// Fields support cheap integer predicates that isolate allocation
// overhead from predicate CPU cost.
type benchItem struct {
	ID     int
	Active bool
	Score  int
}

// sinkItems prevents dead-code elimination without interface-boxing overhead.
var sinkItems []benchItem

func makeItems(n int) []benchItem {
	items := make([]benchItem, n)
	for i := range n {
		items[i] = benchItem{
			ID:     i,
			Active: i%2 == 0,
			Score:  i % 100,
		}
	}
	return items
}

// predicates returns 7 predicate functions for benchItem filtering.
// n is the dataset size, used by size-relative predicates.
// All predicates use cheap integer comparisons.
func predicates(n int) [7]func(benchItem) bool {
	return [7]func(benchItem) bool{
		func(b benchItem) bool { return b.Active },           // ~50% pass
		func(b benchItem) bool { return b.ID%3 != 0 },        // ~67% pass
		func(b benchItem) bool { return b.Score > 20 },        // ~80% pass
		func(b benchItem) bool { return b.ID < n/2 },          // ~50% pass
		func(b benchItem) bool { return b.ID > n/10 },         // ~90% pass
		func(b benchItem) bool { return b.ID/100%2 == 0 },     // ~50% pass
		func(b benchItem) bool { return b.Score < 90 },        // ~90% pass
	}
}

// Predicate indices for readability.
const (
	pIsActive = iota
	pNotDivBy3
	pHighScore
	pSmallID
	pLargeID
	pEvenHundred
	pModerateScore
)

// selectiveFirst returns predicates ordered with high-rejection first.
func selectiveFirst(n int) [7]func(benchItem) bool {
	p := predicates(n)
	return [7]func(benchItem) bool{
		p[pIsActive],       // 50%
		p[pNotDivBy3],      // 67%
		p[pEvenHundred],    // 50%
		p[pSmallID],        // 50%
		p[pHighScore],      // 80%
		p[pModerateScore],  // 90%
		p[pLargeID],        // 90%
	}
}

// selectiveLast returns predicates ordered with high-rejection last.
func selectiveLast(n int) [7]func(benchItem) bool {
	p := predicates(n)
	return [7]func(benchItem) bool{
		p[pLargeID],        // 90%
		p[pModerateScore],  // 90%
		p[pHighScore],      // 80%
		p[pSmallID],        // 50%
		p[pEvenHundred],    // 50%
		p[pNotDivBy3],      // 67%
		p[pIsActive],       // 50%
	}
}

// alwaysTrue returns 7 always-true predicates for control benchmarks.
func alwaysTrue() [7]func(benchItem) bool {
	t := func(benchItem) bool { return true }
	return [7]func(benchItem) bool{t, t, t, t, t, t, t}
}

// --- Helpers ---

// applyAll runs all 7 predicates with short-circuit AND.
func applyAll(item benchItem, preds [7]func(benchItem) bool) bool {
	return preds[0](item) && preds[1](item) && preds[2](item) &&
		preds[3](item) && preds[4](item) && preds[5](item) && preds[6](item)
}

func filterLoop(items []benchItem, preds [7]func(benchItem) bool) []benchItem {
	var result []benchItem
	for _, item := range items {
		if applyAll(item, preds) {
			result = append(result, item)
		}
	}
	return result
}

func filterMapper(items []benchItem, preds [7]func(benchItem) bool) []benchItem {
	return []benchItem(slice.From(items).
		KeepIf(preds[0]).KeepIf(preds[1]).KeepIf(preds[2]).
		KeepIf(preds[3]).KeepIf(preds[4]).KeepIf(preds[5]).KeepIf(preds[6]))
}

func filterSeq(items []benchItem, preds [7]func(benchItem) bool) []benchItem {
	return seq.From(items).
		KeepIf(preds[0]).KeepIf(preds[1]).KeepIf(preds[2]).
		KeepIf(preds[3]).KeepIf(preds[4]).KeepIf(preds[5]).KeepIf(preds[6]).
		Collect()
}

// --- Correctness ---

// TestMultiFilter_Correctness verifies ALL benchmarked variants produce
// identical output for each size and predicate ordering.
func TestMultiFilter_Correctness(t *testing.T) {
	sizes := []int{100, 1000, 10000}
	orders := []struct {
		name       string
		preds      func(int) [7]func(benchItem) bool
		directKeep func(benchItem, int) bool
	}{
		{
			name:  "SelectiveFirst",
			preds: selectiveFirst,
			directKeep: func(item benchItem, n int) bool {
				return item.Active &&
					item.ID%3 != 0 &&
					item.ID/100%2 == 0 &&
					item.ID < n/2 &&
					item.Score > 20 &&
					item.Score < 90 &&
					item.ID > n/10
			},
		},
		{
			name:  "SelectiveLast",
			preds: selectiveLast,
			directKeep: func(item benchItem, n int) bool {
				return item.ID > n/10 &&
					item.Score < 90 &&
					item.Score > 20 &&
					item.ID < n/2 &&
					item.ID/100%2 == 0 &&
					item.ID%3 != 0 &&
					item.Active
			},
		},
	}

	for _, n := range sizes {
		for _, order := range orders {
			t.Run(fmt.Sprintf("%s/n=%d", order.name, n), func(t *testing.T) {
				items := makeItems(n)
				preds := order.preds(n)

				want := filterLoop(items, preds)
				mapperResult := filterMapper(items, preds)
				seqResult := filterSeq(items, preds)

				var loopDirectResult []benchItem
				for _, item := range items {
					if order.directKeep(item, n) {
						loopDirectResult = append(loopDirectResult, item)
					}
				}

				if !slices.Equal(want, loopDirectResult) {
					t.Errorf("LoopDirect: got %d items, want %d (contents differ)", len(loopDirectResult), len(want))
				}
				if !slices.Equal(want, mapperResult) {
					t.Errorf("MapperChain: got %d items, want %d (contents differ)", len(mapperResult), len(want))
				}
				if !slices.Equal(want, seqResult) {
					t.Errorf("SeqCollect: got %d items, want %d (contents differ)", len(seqResult), len(want))
				}
			})
		}
	}
}

// TestMultiFilter_SurvivorCounts documents per-stage survivor counts for
// each ordering and size to characterize actual selectivity.
func TestMultiFilter_SurvivorCounts(t *testing.T) {
	sizes := []int{100, 1000, 10000}

	for _, n := range sizes {
		items := makeItems(n)

		for _, order := range []struct {
			name  string
			preds [7]func(benchItem) bool
		}{
			{"SelectiveFirst", selectiveFirst(n)},
			{"SelectiveLast", selectiveLast(n)},
		} {
			t.Run(fmt.Sprintf("%s/n=%d", order.name, n), func(t *testing.T) {
				current := items
				for stage, pred := range order.preds {
					var next []benchItem
					for _, item := range current {
						if pred(item) {
							next = append(next, item)
						}
					}
					t.Logf("stage %d: %d → %d survivors", stage, len(current), len(next))
					current = next
				}
			})
		}
	}
}

// --- Benchmark helpers ---

// benchLoopDirect runs a single-pass loop with inline conditions.
// Duplication between SelectiveFirst and SelectiveLast is intentional:
// inlining the conditions IS the property being benchmarked.
// Timed region includes: iteration, predicate evaluation, append, sink assignment.
// Excluded: input construction, predicate creation.
func benchLoopDirectSelectiveFirst(b *testing.B, n int) {
	items := makeItems(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		result := make([]benchItem, 0, n)
		for _, item := range items {
			if item.Active &&
				item.ID%3 != 0 &&
				item.ID/100%2 == 0 &&
				item.ID < n/2 &&
				item.Score > 20 &&
				item.Score < 90 &&
				item.ID > n/10 {
				result = append(result, item)
			}
		}
		sinkItems = result
	}
}

func benchLoopDirectSelectiveLast(b *testing.B, n int) {
	items := makeItems(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		result := make([]benchItem, 0, n)
		for _, item := range items {
			if item.ID > n/10 &&
				item.Score < 90 &&
				item.Score > 20 &&
				item.ID < n/2 &&
				item.ID/100%2 == 0 &&
				item.ID%3 != 0 &&
				item.Active {
				result = append(result, item)
			}
		}
		sinkItems = result
	}
}

// Timed region includes: iteration, predicate dispatch via function values,
// append, sink assignment. Predicates are called inline (not via applyAll)
// to avoid an extra non-inlined function call layer.
func benchLoopFuncs(b *testing.B, n int, preds [7]func(benchItem) bool) {
	items := makeItems(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		result := make([]benchItem, 0, n)
		for _, item := range items {
			if preds[0](item) && preds[1](item) && preds[2](item) &&
				preds[3](item) && preds[4](item) && preds[5](item) && preds[6](item) {
				result = append(result, item)
			}
		}
		sinkItems = result
	}
}

// benchLoopFuncsNoHint is identical to benchLoopFuncs but starts with
// var result []benchItem instead of make([]T, 0, n). This isolates
// collection-growth overhead from evaluation strategy — compare with
// SeqCollect to see how much of seq's allocation profile comes from
// slices.Collect's lack of size hint vs lazy composition itself.
func benchLoopFuncsNoHint(b *testing.B, n int, preds [7]func(benchItem) bool) {
	items := makeItems(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		var result []benchItem
		for _, item := range items {
			if preds[0](item) && preds[1](item) && preds[2](item) &&
				preds[3](item) && preds[4](item) && preds[5](item) && preds[6](item) {
				result = append(result, item)
			}
		}
		sinkItems = result
	}
}

// Timed region includes: pipeline construction (From + 7x KeepIf), 7 intermediate
// slice allocations, predicate evaluation, sink assignment.
func benchMapperChain(b *testing.B, n int, preds [7]func(benchItem) bool) {
	items := makeItems(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkItems = filterMapper(items, preds)
	}
}

// Timed region includes: pipeline construction (From + 7x KeepIf closure wrapping),
// traversal with predicate evaluation, Collect terminal materialization.
func benchSeqCollect(b *testing.B, n int, preds [7]func(benchItem) bool) {
	items := makeItems(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkItems = filterSeq(items, preds)
	}
}

// --- Control benchmarks for seq alloc attribution ---
// These use alwaysTrue predicates to hold output size constant (= input size),
// isolating per-closure allocation overhead from output-size-dependent
// Collect growth.

func benchSeqNoFilters(b *testing.B, n int) {
	items := makeItems(n)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkItems = seq.From(items).Collect()
	}
}

func benchSeqOneFilterPassAll(b *testing.B, n int) {
	items := makeItems(n)
	always := alwaysTrue()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkItems = seq.From(items).KeepIf(always[0]).Collect()
	}
}

func benchSeqSevenFiltersPassAll(b *testing.B, n int) {
	items := makeItems(n)
	always := alwaysTrue()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkItems = seq.From(items).
			KeepIf(always[0]).KeepIf(always[1]).KeepIf(always[2]).
			KeepIf(always[3]).KeepIf(always[4]).KeepIf(always[5]).KeepIf(always[6]).
			Collect()
	}
}

// --- Table-driven benchmarks ---

func BenchmarkMultiFilter(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, n := range sizes {
		predsFirst := selectiveFirst(n)
		predsLast := selectiveLast(n)

		b.Run(fmt.Sprintf("LoopDirect/SelectiveFirst/n=%d", n), func(b *testing.B) {
			benchLoopDirectSelectiveFirst(b, n)
		})
		b.Run(fmt.Sprintf("LoopDirect/SelectiveLast/n=%d", n), func(b *testing.B) {
			benchLoopDirectSelectiveLast(b, n)
		})
		b.Run(fmt.Sprintf("LoopFuncs/SelectiveFirst/n=%d", n), func(b *testing.B) {
			benchLoopFuncs(b, n, predsFirst)
		})
		b.Run(fmt.Sprintf("LoopFuncs/SelectiveLast/n=%d", n), func(b *testing.B) {
			benchLoopFuncs(b, n, predsLast)
		})
		b.Run(fmt.Sprintf("LoopFuncsNoHint/SelectiveFirst/n=%d", n), func(b *testing.B) {
			benchLoopFuncsNoHint(b, n, predsFirst)
		})
		b.Run(fmt.Sprintf("LoopFuncsNoHint/SelectiveLast/n=%d", n), func(b *testing.B) {
			benchLoopFuncsNoHint(b, n, predsLast)
		})
		b.Run(fmt.Sprintf("MapperChain/SelectiveFirst/n=%d", n), func(b *testing.B) {
			benchMapperChain(b, n, predsFirst)
		})
		b.Run(fmt.Sprintf("MapperChain/SelectiveLast/n=%d", n), func(b *testing.B) {
			benchMapperChain(b, n, predsLast)
		})
		b.Run(fmt.Sprintf("SeqCollect/SelectiveFirst/n=%d", n), func(b *testing.B) {
			benchSeqCollect(b, n, predsFirst)
		})
		b.Run(fmt.Sprintf("SeqCollect/SelectiveLast/n=%d", n), func(b *testing.B) {
			benchSeqCollect(b, n, predsLast)
		})
	}
}

// BenchmarkSeqControls isolates seq allocation overhead using always-true
// predicates so output size equals input size across all variants.
// This separates per-closure overhead from output-dependent Collect growth.
func BenchmarkSeqControls(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("NoFilters/n=%d", n), func(b *testing.B) {
			benchSeqNoFilters(b, n)
		})
		b.Run(fmt.Sprintf("OneFilterPassAll/n=%d", n), func(b *testing.B) {
			benchSeqOneFilterPassAll(b, n)
		})
		b.Run(fmt.Sprintf("SevenFiltersPassAll/n=%d", n), func(b *testing.B) {
			benchSeqSevenFiltersPassAll(b, n)
		})
	}
}
