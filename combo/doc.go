// Package combo generates combinatorial structures over slices: Cartesian
// products, permutations, combinations (k-subsets), and power sets.
//
// All functions return [slice.Mapper] for fluent chaining. Results grow
// factorially or exponentially — intended for small inputs or test generation,
// not production data volumes.
package combo
