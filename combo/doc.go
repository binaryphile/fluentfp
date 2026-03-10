package combo

import "github.com/binaryphile/fluentfp/tuple/pair"

func _() {
	_ = CartesianProduct[int, string]
	_ = CartesianProductWith[int, string, pair.Pair[int, string]]
	_ = Combinations[int]
	_ = Permutations[int]
	_ = PowerSet[int]
}
