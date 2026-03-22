package toc_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/toc"
)

func TestClassifyStep(t *testing.T) {
	tests := []struct {
		name           string
		prevConstraint string
		currConstraint string
		ropeActive     bool
		rebalancing    bool
		starving       bool
		want           toc.FocusingStep
	}{
		// Step 1: Identify — no constraint.
		{"no_constraint", "", "", false, false, false, toc.StepIdentify},
		{"constraint_lost", "embed", "", false, false, false, toc.StepIdentify},

		// Step 5: Prevent Inertia — constraint changed.
		{"constraint_changed", "embed", "walk", true, false, false, toc.StepPreventInertia},
		{"constraint_changed_with_rebalancing", "embed", "walk", true, true, false, toc.StepPreventInertia},
		{"constraint_changed_with_starvation", "embed", "walk", false, false, true, toc.StepPreventInertia},

		// Step 4: Elevate — rebalancer active.
		{"rebalancing", "embed", "embed", true, true, false, toc.StepElevate},
		{"rebalancing_and_starving", "embed", "embed", true, true, true, toc.StepElevate},

		// Step 2: Exploit — rope not active or drum starving.
		{"rope_not_active", "", "embed", false, false, false, toc.StepExploit},
		{"first_identification", "", "embed", false, false, false, toc.StepExploit},
		{"starving_with_rope", "embed", "embed", true, false, true, toc.StepExploit},
		{"starving_without_rope", "embed", "embed", false, false, true, toc.StepExploit},

		// Step 3: Subordinate — healthy steady state.
		{"steady_state", "embed", "embed", true, false, false, toc.StepSubordinate},
		{"first_steady_state", "", "embed", true, false, false, toc.StepSubordinate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toc.ClassifyStep(tt.prevConstraint, tt.currConstraint, tt.ropeActive, tt.rebalancing, tt.starving)
			if got != tt.want {
				t.Errorf("ClassifyStep(%q, %q, rope=%v, rebal=%v, starve=%v) = %v, want %v",
					tt.prevConstraint, tt.currConstraint, tt.ropeActive, tt.rebalancing, tt.starving, got, tt.want)
			}
		})
	}
}

func TestFocusingStepString(t *testing.T) {
	tests := []struct {
		step toc.FocusingStep
		want string
	}{
		{toc.StepIdentify, "identify"},
		{toc.StepExploit, "exploit"},
		{toc.StepSubordinate, "subordinate"},
		{toc.StepElevate, "elevate"},
		{toc.StepPreventInertia, "prevent-inertia"},
	}

	for _, tt := range tests {
		if got := tt.step.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.step, got, tt.want)
		}
	}
}
