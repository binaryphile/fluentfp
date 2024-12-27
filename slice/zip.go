package slice

import (
	"bitbucket.org/accelecon/charybdis/services/metrics-grpc-api/internal/fluentfp/tuple/pair"
)

func Zip[T, T2 any](ts []T, t2s []T2) []pair.Of[T, T2] {
	if len(ts) != len(t2s) {
		panic("zip: arguments must have same length")
	}

	result := make([]pair.Of[T, T2], len(ts))
	for i := range ts {
		result[i] = pair.Of[T, T2]{
			V1: ts[i],
			V2: t2s[i],
		}
	}

	return result
}
