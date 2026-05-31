//go:build ignore

// Package snippet is the verification harness for the dockerdir showcase
// entry in docs/showcase.md (docker/cli config Dir rewrite). The snippet
// is a complete top-level `Dir()` function; the harness supplies the
// docker/cli package globals (initConfigDir, configDir, configFileDir,
// EnvOverrideConfigDir) plus a stubbed getHomeDir().
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/binaryphile/fluentfp/option"
)

// docker/cli package-level state referenced by Dir().
var (
	initConfigDir        sync.Once
	configDir            string
	configFileDir        = ".docker"
	EnvOverrideConfigDir = "DOCKER_CONFIG"
)

// getHomeDir stubs docker/cli's helper — the showcase prose shows it
// separately. Only the symbol needs to exist for the snippet to compile.
func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// __SNIPPET__
