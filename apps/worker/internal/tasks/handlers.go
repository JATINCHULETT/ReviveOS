package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/schemas"
	"github.com/reviveos/worker/internal/executor"
	"github.com/reviveos/worker/internal/pipeline"
)

type TaskHandler struct {
	pool     *pgxpool.Pool
	pipeline *pipeline.Pipeline
	executor *executor.RecoveryExecutor
}

func NewTaskHandler(pool *pgxpool.Pool) *TaskHandler {
	return &TaskHandler{
		pool:     pool,
		pipeline: pipeline.NewPipeline(pool),
		executor: executor.NewRecoveryExecutor(pool, nil),
	}
}

func (h *TaskHandler) RegisterRoutes(mux *asynq.ServeMux) {
	mux.HandleFunc(schemas.TaskRecoveryAnalyze, h.HandleAnalyze)
	mux.HandleFunc(schemas.TaskRecoveryPlan, h.HandlePlan)
	mux.HandleFunc(schemas.TaskRecoveryPolicy, h.HandlePolicy)
	mux.HandleFunc(schemas.TaskRecoverySchedule, h.HandleSchedule)
	mux.HandleFunc(schemas.TaskRecoveryExecute, h.HandleExecute)
	mux.HandleFunc(schemas.TaskRecoveryVerify, h.HandleVerify)
	mux.HandleFunc(schemas.TaskRecoveryComplete, h.HandleComplete)
	mux.HandleFunc("payment.failed", h.HandlePaymentFailed)
	mux.HandleFunc("test:task", h.HandleTestTask)
	mux.HandleFunc("test:ping", h.HandleTestTask)
}

func (h *TaskHandler) HandleAnalyze(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Handling task: %s | Payload: %s", t.Type(), string(t.Payload()))

	var payload struct {
		PaymentID string `json:"payment_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal analyze payload: %w", err)
	}

	if payload.PaymentID == "" {
		return fmt.Errorf("empty payment_id in payload")
	}

	res, err := h.pipeline.AnalyzePayment(ctx, payload.PaymentID)
	if err != nil {
		log.Printf("[Worker] ERROR in AnalyzePayment: %v", err)
		return err
	}

	log.Printf("[Worker] Analysis finished: Workflow=%s, Category=%s, Prob=%.4f", res.WorkflowID, res.FailureCategory, res.Probability)
	return nil
}

func (h *TaskHandler) HandlePaymentFailed(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Handling payment.failed: %s", string(t.Payload()))

	var event schemas.PaymentFailureEvent
	if err := json.Unmarshal(t.Payload(), &event); err == nil && event.PaymentID != "" {
		_, err := h.pipeline.AnalyzePayment(ctx, event.PaymentID)
		if err != nil {
			log.Printf("[Worker] ERROR analyzing payment %s: %v", event.PaymentID, err)
			return err
		}
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &raw); err == nil {
		if pid, ok := raw["payment_id"].(string); ok && pid != "" {
			_, err := h.pipeline.AnalyzePayment(ctx, pid)
			return err
		}
	}

	return nil
}

func (h *TaskHandler) HandlePlan(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Processing %s | Payload: %s", t.Type(), string(t.Payload()))
	return nil
}

func (h *TaskHandler) HandlePolicy(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Processing %s | Payload: %s", t.Type(), string(t.Payload()))
	return nil
}

func (h *TaskHandler) HandleSchedule(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Processing %s | Payload: %s", t.Type(), string(t.Payload()))
	return nil
}

func (h *TaskHandler) HandleExecute(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Handling task: %s | Payload: %s", t.Type(), string(t.Payload()))

	var payload struct {
		WorkflowID string `json:"workflow_id"`
		PaymentID  string `json:"payment_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal execute payload: %w", err)
	}

	workflowID := payload.WorkflowID
	if workflowID == "" && payload.PaymentID != "" {
		_ = h.pool.QueryRow(ctx, `
			SELECT id::text FROM recovery_workflows 
			WHERE payment_id::text = $1 
			ORDER BY created_at DESC LIMIT 1
		`, payload.PaymentID).Scan(&workflowID)
	}

	if workflowID == "" {
		return fmt.Errorf("missing workflow_id in execute payload")
	}

	res, err := h.executor.ExecuteWorkflow(ctx, workflowID)
	if err != nil {
		log.Printf("[Worker] ERROR executing workflow %s: %v", workflowID, err)
		return err
	}

	log.Printf("[Worker] Workflow execution finished: Workflow=%s, Rec=%s, Action=%s, Recovered=%v",
		res.WorkflowID, res.Reconciliation, res.ActionTaken, res.Recovered)
	return nil
}

func (h *TaskHandler) HandleVerify(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Handling task: %s | Payload: %s", t.Type(), string(t.Payload()))
	return h.HandleExecute(ctx, t)
}

func (h *TaskHandler) HandleComplete(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Processing %s | Payload: %s", t.Type(), string(t.Payload()))
	return nil
}

func (h *TaskHandler) HandleTestTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Worker] Test task received: Type=%s Payload=%s", t.Type(), string(t.Payload()))
	var payload map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &payload); err == nil {
		if msg, ok := payload["message"].(string); ok {
			log.Printf("[Worker] Test task message content: %s", msg)
		}
	}
	return nil
}
