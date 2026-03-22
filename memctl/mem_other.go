//go:build !linux

package memctl

func readSystemAvailable() (uint64, bool) { return 0, false }
func readProcessRSS() (uint64, bool)      { return 0, false }
func readCgroup() (uint64, uint64, bool)   { return 0, 0, false }
