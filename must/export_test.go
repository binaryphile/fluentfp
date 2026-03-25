package must_test

import . "github.com/binaryphile/fluentfp/must"

// Compile-time API verification
func _() {
	_ = BeNil
	_ = ErrEnvEmpty
	_ = ErrEnvUnset
	_ = ErrNilFunction
	_ = Get[int]
	_ = Get2[int, int]
	_ = NonEmptyEnv
	_ = From[int, int]
}
