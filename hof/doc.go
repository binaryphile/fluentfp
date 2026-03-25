// Package hof provides higher-order functions over plain function signatures:
// composition, partial application, independent application, and call
// coalescing.
//
// The organizing principle is the function shape. hof operates on plain
// signatures like func(A) B, func(A, B) C, and func(T). For decorators
// over the context-aware call shape func(context.Context, T) (R, error),
// see the [call] package.
//
// Based on Stone's "Algorithms: A Functional Programming Approach"
// (pipe, sect, cross).
package hof
