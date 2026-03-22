package toc

import "fmt"

// FocusingStep represents which of Goldratt's Five Focusing Steps
// the pipeline is currently in.
//
// The steps are not a state machine — they are a classification of
// the current system condition based on observable signals. Consumers
// call [ClassifyStep] per interval from their own coordination logic.
type FocusingStep int

const (
	// StepIdentify is Step 1: identify the system's constraint.
	// No constraint has been identified yet — the analyzer is still
	// collecting data or no stage is saturated.
	StepIdentify FocusingStep = iota + 1

	// StepExploit is Step 2: exploit the constraint — don't waste it.
	// A constraint is identified but exploitation is incomplete: the
	// rope/buffer is not yet active, or the drum is starving.
	StepExploit

	// StepSubordinate is Step 3: subordinate everything else.
	// The constraint is identified, the rope is active, and the drum
	// is not starving. Non-constraints defer to the drum's pace.
	StepSubordinate

	// StepElevate is Step 4: elevate the constraint's capacity.
	// The rebalancer is actively moving resources to the constraint
	// (e.g., adding workers to the drum stage).
	StepElevate

	// StepPreventInertia is Step 5: if the constraint has moved,
	// go back to Step 1 — do not allow inertia to become the
	// system's constraint. The old rope must be rebuilt for the
	// new drum.
	//
	// Constraint migration protocol: cancel old [RopeController]
	// context, call [NewRopeController] with the new drum. EWMA
	// state starts fresh — old drum's signals are irrelevant.
	//
	// Migration is safe because the new rope controls a prefix of
	// the old chain. If the constraint moved upstream (e.g., embed
	// → walk), the new rope limits head → walk WIP, which naturally
	// limits what reaches embed. The old drum's protection is
	// implicitly preserved by reduced upstream supply.
	StepPreventInertia
)

func (s FocusingStep) String() string {
	switch s {
	case StepIdentify:
		return "identify"
	case StepExploit:
		return "exploit"
	case StepSubordinate:
		return "subordinate"
	case StepElevate:
		return "elevate"
	case StepPreventInertia:
		return "prevent-inertia"
	default:
		return fmt.Sprintf("FocusingStep(%d)", int(s))
	}
}

// ClassifyStep determines the current focusing step from system state.
// Pure function — no side effects.
//
// The caller provides prev and curr constraint names so the comparison
// logic lives with the classification. The caller tracks prevConstraint
// across intervals (typically one string variable updated per tick).
//
// Classification priority (highest first):
//  1. No constraint (currConstraint empty) → [StepIdentify]
//  2. Constraint changed (prev ≠ curr, both non-empty) → [StepReassess]
//  3. Rebalancer active → [StepElevate]
//  4. Drum starving or rope not active → [StepExploit]
//  5. Otherwise → [StepSubordinate]
func ClassifyStep(
	prevConstraint string,
	currConstraint string,
	ropeActive bool,
	rebalancing bool,
	starving bool,
) FocusingStep {
	if currConstraint == "" {
		return StepIdentify
	}

	if prevConstraint != "" && prevConstraint != currConstraint {
		return StepPreventInertia
	}

	if rebalancing {
		return StepElevate
	}

	if starving || !ropeActive {
		return StepExploit
	}

	return StepSubordinate
}
