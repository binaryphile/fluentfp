// Package pmap compile-checks the showcase entry for Starship parallel module rendering.
package pmap

import (
	"os"
	"os/exec"
	"strings"

	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for the Starship-equivalent types ---

// Segment holds rendered output from one status module.
type Segment struct {
	Name  string
	Text  string
	Color string
}

// renderModule evaluates a single status module by name.
func renderModule(name string) Segment {
	switch name {
	case "git":
		out, _ := exec.Command("git", "branch", "--show-current").Output()
		return Segment{Name: name, Text: strings.TrimSpace(string(out)), Color: "green"}
	default:
		return Segment{Name: name, Text: "?"}
	}
}

// isEnabled returns true if the module has something to show in the current environment.
var isEnabled = func(name string) bool {
	switch name {
	case "git":
		return exec.Command("git", "rev-parse", "--git-dir").Run() == nil
	case "go":
		_, err := os.Stat("go.mod")
		return err == nil
	default:
		return true
	}
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func RenderSequential(enabledModules []string) []Segment {
	segments := slice.Map(enabledModules, renderModule)
	return segments
}

func RenderParallel(enabledModules []string) []Segment {
	segments := slice.PMap(enabledModules, 8, renderModule)
	return segments
}

func ActiveModules(allModules []string) []string {
	active := slice.From(allModules).PKeepIf(8, isEnabled)
	return active
}
