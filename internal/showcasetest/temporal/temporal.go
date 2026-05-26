// Package temporal compile-checks the showcase entry for temporalio mutable_state_rebuilder.
// The showcase shows two event cases out of 40+; compile-check uses those two
// to verify the slice.Fold pattern and method-on-state shape work.
package temporal

import (
	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for temporalio types (abbreviated; only the 2 cases shown in the showcase) ---

type historyEvent struct {
	eventType                              int
	workflowExecutionStartedEventAttributes startedAttrs
	activityTaskScheduledEventAttributes   scheduledAttrs
}

func (e *historyEvent) GetEventType() int { return e.eventType }
func (e *historyEvent) GetWorkflowExecutionStartedEventAttributes() startedAttrs {
	return e.workflowExecutionStartedEventAttributes
}
func (e *historyEvent) GetActivityTaskScheduledEventAttributes() scheduledAttrs {
	return e.activityTaskScheduledEventAttributes
}

type startedAttrs struct{}
type scheduledAttrs struct{}

type WorkflowState struct{}

func (s WorkflowState) WithExecution(_ startedAttrs) WorkflowState         { return s }
func (s WorkflowState) WithScheduledActivity(_ scheduledAttrs) WorkflowState { return s }

const (
	eventTypeWorkflowExecutionStarted = 1
	eventTypeActivityTaskScheduled    = 2
)

// --- the fluentfp rewrite from docs/showcase.md (verbatim structural pattern) ---

func ApplyEvents(history []*historyEvent, initialState WorkflowState) WorkflowState {
	// applyEvent transitions workflow state based on a single event.
	// Each case is a pure state transition — no loop context needed.
	applyEvent := func(state WorkflowState, event *historyEvent) WorkflowState {
		switch event.GetEventType() {
		case eventTypeWorkflowExecutionStarted:
			return state.WithExecution(event.GetWorkflowExecutionStartedEventAttributes())
		case eventTypeActivityTaskScheduled:
			return state.WithScheduledActivity(event.GetActivityTaskScheduledEventAttributes())
		// ... (40+ more cases in the original; omitted in showcase + here)
		}
		return state
	}

	currentState := slice.Fold(history, initialState, applyEvent)
	return currentState
}
