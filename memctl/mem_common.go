package memctl

import "runtime/metrics"

// readGoRuntimeTotal reads /memory/classes/total:bytes via runtime/metrics.
func readGoRuntimeTotal() (uint64, bool) {
	var s [1]metrics.Sample
	s[0].Name = "/memory/classes/total:bytes"
	metrics.Read(s[:])

	if s[0].Value.Kind() == metrics.KindUint64 {
		return s[0].Value.Uint64(), true
	}

	return 0, false
}
