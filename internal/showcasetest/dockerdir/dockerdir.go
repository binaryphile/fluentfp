// Package dockerdir compile-checks the showcase entry for docker/cli.
package dockerdir

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/binaryphile/fluentfp/option"
)

// --- stubs for the docker/cli globals ---

var (
	initConfigDir       sync.Once
	configDir           string
	configFileDir       = ".docker"
	EnvOverrideConfigDir = "DOCKER_CONFIG"
)

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func Dir() string {
	initConfigDir.Do(func() {
		// defaultDir computes the config directory from the user's home directory.
		defaultDir := func() string { return filepath.Join(getHomeDir(), configFileDir) }
		configDir = option.Env(EnvOverrideConfigDir).OrCall(defaultDir)
	})
	return configDir
}

// stubbed getHomeDir — the original is shown separately in the showcase but
// not in the fluentfp rewrite. Compile-check only needs the symbol to exist.
func getHomeDir() string {
	home, _ := os.UserHomeDir()
	if home == "" && runtime.GOOS != "windows" {
		// Skipping the user.Current() branch — it isn't part of the showcase rewrite.
	}
	return home
}
