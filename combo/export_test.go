package combo_test

import . "github.com/binaryphile/fluentfp/combo"
import "github.com/binaryphile/fluentfp/tuple/pair"

func _() {
	_ = CartesianProduct[int, string]
	_ = CartesianProductWith[int, string, pair.Pair[int, string]]
	_ = Combinations[int]
	_ = Permutations[int]
	_ = PowerSet[int]
	_ = SeqCartesianProduct[int, string]
	_ = SeqCartesianProductWith[int, string, pair.Pair[int, string]]
	_ = SeqCombinations[int]
	_ = SeqPermutations[int]
	_ = SeqPowerSet[int]
}
