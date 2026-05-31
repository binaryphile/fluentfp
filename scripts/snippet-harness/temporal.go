//go:build ignore

// Package snippet is the verification harness for the temporal showcase
// entry in docs/showcase.md (temporalio/temporal
// mutable_state_rebuilder.applyEvents rewrite). The snippet is function-
// body code (applyEvent := + currentState := + return). The harness
// wraps it in ApplyEvents and stubs the temporalio event vocabulary
// (historyEvent + accessor methods, startedAttrs / scheduledAttrs,
// WorkflowState with state-transition methods, the two event-type
// constants used by the snippet's switch).
//
// Local lowercase names stand in for historypb.HistoryEvent and
// enumspb.EVENT_TYPE_* — same trade-off the annotation and sagas
// migrations make: lose the qualified-package texture in exchange for
// compile-verifiability without the upstream module.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"github.com/binaryphile/fluentfp/slice"
)

type startedAttrs struct{}
type scheduledAttrs struct{}

type historyEvent struct {
	eventType                               int
	workflowExecutionStartedEventAttributes startedAttrs
	activityTaskScheduledEventAttributes    scheduledAttrs
}

func (e *historyEvent) GetEventType() int { return e.eventType }
func (e *historyEvent) GetWorkflowExecutionStartedEventAttributes() startedAttrs {
	return e.workflowExecutionStartedEventAttributes
}
func (e *historyEvent) GetActivityTaskScheduledEventAttributes() scheduledAttrs {
	return e.activityTaskScheduledEventAttributes
}

const (
	eventTypeWorkflowExecutionStarted = 1
	eventTypeActivityTaskScheduled    = 2
)

// WorkflowState stubs temporalio's mutable workflow state aggregate.
// The two state-transition methods (one per event type shown in the
// snippet's switch) preserve the snippet's left-fold shape.
type WorkflowState struct{}

func (s WorkflowState) WithExecution(_ startedAttrs) WorkflowState           { return s }
func (s WorkflowState) WithScheduledActivity(_ scheduledAttrs) WorkflowState { return s }

func ApplyEvents(history []*historyEvent, initialState WorkflowState) WorkflowState {
	// __SNIPPET__
}
