// Package combo generates combinatorial structures over slices: Cartesian
// products, permutations, combinations (k-subsets), and power sets.
//
// Eager functions return [slice.Mapper] for fluent chaining. Seq variants
// (SeqPermutations, SeqPowerSet, etc.) return [seq.Seq] for lazy evaluation
// with early termination — use these for large inputs where materializing the
// full result is impractical.
package combo
