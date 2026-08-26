package types

import "errors"

// WorkflowState represents the finite states of a recovery workflow
type WorkflowState string

const (
	StateFailed              WorkflowState = "FAILED"
	StatePendingVerification WorkflowState = "PENDING_VERIFICATION"
	StateAnalyzing           WorkflowState = "ANALYZING"
	StatePlanned             WorkflowState = "PLANNED"
	StatePolicyCheck         WorkflowState = "POLICY_CHECK"
	StateScheduled           WorkflowState = "SCHEDULED"
	StateExecuting           WorkflowState = "EXECUTING"
	StateVerifying           WorkflowState = "VERIFYING"
	StateRecovered           WorkflowState = "RECOVERED"
	StateNextAttempt         WorkflowState = "NEXT_ATTEMPT"
	StateEscalated           WorkflowState = "ESCALATED"
	StateHalted              WorkflowState = "HALTED"
)

// AllowedTransitions maps a state to its valid subsequent states
var AllowedTransitions = map[WorkflowState][]WorkflowState{
	StateFailed: {
		StatePendingVerification,
		StateAnalyzing, // Direct transition if verification skipped
	},
	StatePendingVerification: {
		StateAnalyzing,
		StateRecovered, // If verification finds it was already successful
	},
	StateAnalyzing: {
		StatePlanned,
		StateEscalated, // If confidence is too low or unknown
		StateHalted,    // If unrecoverable
	},
	StatePlanned: {
		StatePolicyCheck,
	},
	StatePolicyCheck: {
		StateScheduled,
		StateEscalated, // Blocked by policy needing approval
		StateHalted,    // Blocked by policy permanently
	},
	StateScheduled: {
		StateExecuting,
		StateHalted, // If canceled before execution
	},
	StateExecuting: {
		StateVerifying,
	},
	StateVerifying: {
		StateRecovered,
		StateNextAttempt,
		StateEscalated,
		StateHalted,
	},
	StateNextAttempt: {
		StateAnalyzing,
		StatePendingVerification,
	},
	StateEscalated: {
		StateScheduled, // Human approved
		StateHalted,    // Human rejected
		StateRecovered, // Resolved externally
	},
	// Terminal states have no outbound transitions in automated flow
	StateRecovered: {},
	StateHalted:    {},
}

var ErrInvalidStateTransition = errors.New("invalid state transition")

// ValidateTransition checks if moving from 'from' to 'to' is allowed
func ValidateTransition(from, to WorkflowState) error {
	allowed, ok := AllowedTransitions[from]
	if !ok {
		return ErrInvalidStateTransition
	}
	
	for _, state := range allowed {
		if state == to {
			return nil
		}
	}
	
	return ErrInvalidStateTransition
}
