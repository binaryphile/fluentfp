//go:build linux

package toc

import (
	"os"
	"strconv"
	"strings"
)

// readRSS returns the process RSS from /proc/self/status.
// Returns (rss, true) on success, (0, false) on failure.
func readRSS() (uint64, bool) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, false
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}

		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, false
		}

		return kb * 1024, true // VmRSS is in kB
	}

	return 0, false
}
