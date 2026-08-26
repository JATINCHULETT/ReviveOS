package schemas

const (
	TaskRecoveryAnalyze  = "recovery:analyze"
	TaskRecoveryPlan     = "recovery:plan"
	TaskRecoveryPolicy   = "recovery:policy"
	TaskRecoverySchedule = "recovery:schedule"
	TaskRecoveryExecute  = "recovery:execute"
	TaskRecoveryVerify   = "recovery:verify"
	TaskRecoveryComplete = "recovery:complete"
)

type RecoveryAnalyzePayload struct {
	PaymentID string `json:"payment_id"`
}

type RecoveryPlanPayload struct {
	WorkflowID string `json:"workflow_id"`
}

type RecoveryPolicyPayload struct {
	WorkflowID string `json:"workflow_id"`
	ActionID   string `json:"action_id"`
}

type RecoveryExecutePayload struct {
	ActionID string `json:"action_id"`
}

type RecoveryVerifyPayload struct {
	ActionID string `json:"action_id"`
}
