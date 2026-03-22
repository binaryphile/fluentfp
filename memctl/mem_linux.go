//go:build linux

package memctl

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// readSystemAvailable reads MemAvailable from /proc/meminfo.
func readSystemAvailable() (uint64, bool) {
	return readProcField("/proc/meminfo", "MemAvailable:")
}

// readProcessRSS reads VmRSS from /proc/self/status.
func readProcessRSS() (uint64, bool) {
	return readProcField("/proc/self/status", "VmRSS:")
}

// readProcField reads a kB-valued field from a /proc file.
func readProcField(path, prefix string) (uint64, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, prefix) {
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

		return kb * 1024, true // kB to bytes
	}

	return 0, false
}

// readCgroup reads cgroup v2 memory.current and memory.max.
// Returns (current, limit, ok). limit=0 means unlimited ("max").
func readCgroup() (uint64, uint64, bool) {
	// Try direct path first (works for most containers).
	current, limit, ok := readCgroupAt("/sys/fs/cgroup")
	if ok {
		return current, limit, true
	}

	// Fall back: read /proc/self/cgroup, extract v2 path.
	path, ok := cgroupV2Path()
	if !ok {
		return 0, 0, false
	}

	return readCgroupAt(filepath.Join("/sys/fs/cgroup", path))
}

func readCgroupAt(dir string) (uint64, uint64, bool) {
	currentBytes, err := os.ReadFile(filepath.Join(dir, "memory.current"))
	if err != nil {
		return 0, 0, false
	}

	maxBytes, err := os.ReadFile(filepath.Join(dir, "memory.max"))
	if err != nil {
		return 0, 0, false
	}

	current, err := strconv.ParseUint(strings.TrimSpace(string(currentBytes)), 10, 64)
	if err != nil {
		return 0, 0, false
	}

	maxStr := strings.TrimSpace(string(maxBytes))
	var limit uint64
	if maxStr != "max" {
		limit, err = strconv.ParseUint(maxStr, 10, 64)
		if err != nil {
			return 0, 0, false
		}
	}
	// "max" → limit=0 (unlimited)

	return current, limit, true
}

// cgroupV2Path extracts the cgroup v2 path from /proc/self/cgroup.
// Returns the path and true if cgroup v2 is detected.
func cgroupV2Path() (string, bool) {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", false
	}

	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		// cgroup v2: single line "0::/<path>"
		if strings.HasPrefix(line, "0::") {
			path := strings.TrimPrefix(line, "0::")
			return path, true
		}
	}

	return "", false
}
