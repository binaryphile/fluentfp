// Package pair provides a generic two-element tuple and strict zip operations.
//
// [Pair] is a value type holding First and Second. Construct via [Of].
// [Zip] and [ZipWith] combine two equal-length slices element-wise,
// panicking on length mismatch rather than silently truncating.
package pair
