//go:build !linux

package toc

// readRSS is not available on non-Linux platforms.
func readRSS() (uint64, bool) {
	return 0, false
}
