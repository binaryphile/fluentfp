// Package either provides a sum type representing a value of one of two types.
//
// Convention: Left represents failure/error, Right represents success.
// Mnemonic: "Right is right" (correct).
//
// Either is right-biased: Map operates on the Right value only.
package either
