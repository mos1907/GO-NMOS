package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go-nmos/backend/internal/models"
)

// ScheduledPlaybooksRunner executes scheduled playbook executions when their time arrives (E.2).
type ScheduledPlaybooksRunner struct {
	repo scheduledPlaybooksRepo
}

type scheduledPlaybooksRepo interface {
	ListPendingScheduledPlaybookExecutions(ctx context.Context, before time.Time) ([]models.ScheduledPlaybookExecution, error)
	GetPlaybook(ctx context.Context, id string) (*models.Playbook, error)
	UpdateScheduledPlaybookExecution(ctx context.Context, id int64, updates map[string]any) error
	CreatePlaybookExecution(ctx context.Context, exec models.PlaybookExecution) (int64, error)
	UpdatePlaybookExecution(ctx context.Context, execID int64, status string, result json.RawMessage) error
	UpsertReceiverConnection(ctx context.Context, conn models.ReceiverConnection) error
	DeleteReceiverConnection(ctx context.Context, receiverID, state, role string) error
}

func NewScheduledPlaybooksRunner(repo scheduledPlaybooksRepo) *ScheduledPlaybooksRunner {
	return &ScheduledPlaybooksRunner{repo: repo}
}

func (r *ScheduledPlaybooksRunner) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runCycle(ctx)
		}
	}
}

func (r *ScheduledPlaybooksRunner) runCycle(ctx context.Context) {
	now := time.Now()
	pending, err := r.repo.ListPendingScheduledPlaybookExecutions(ctx, now)
	if err != nil {
		log.Printf("scheduled playbooks runner: list pending failed: %v", err)
		return
	}

	for _, scheduled := range pending {
		r.executeScheduledPlaybook(ctx, scheduled)
	}
}

func (r *ScheduledPlaybooksRunner) executeScheduledPlaybook(ctx context.Context, scheduled models.ScheduledPlaybookExecution) {
	// Mark as executing (update status to prevent duplicate execution)
	r.repo.UpdateScheduledPlaybookExecution(ctx, scheduled.ID, map[string]any{
		"status": "executing",
	})

	// Get playbook
	playbook, err := r.repo.GetPlaybook(ctx, scheduled.PlaybookID)
	if err != nil {
		r.repo.UpdateScheduledPlaybookExecution(ctx, scheduled.ID, map[string]any{
			"status":      "failed",
			"executed_at": time.Now(),
			"result":      json.RawMessage(fmt.Sprintf(`{"error":"playbook not found: %v"}`, err)),
		})
		return
	}

	// Parse parameters
	var params map[string]any
	if err := json.Unmarshal(scheduled.Parameters, &params); err != nil {
		r.repo.UpdateScheduledPlaybookExecution(ctx, scheduled.ID, map[string]any{
			"status":      "failed",
			"executed_at": time.Now(),
			"result":      json.RawMessage(fmt.Sprintf(`{"error":"invalid parameters: %v"}`, err)),
		})
		return
	}

	// Create execution record
	exec := models.PlaybookExecution{
		PlaybookID: scheduled.PlaybookID,
		Parameters: scheduled.Parameters,
		Status:     "running",
		StartedAt:  time.Now(),
	}
	execID, err := r.repo.CreatePlaybookExecution(ctx, exec)
	if err != nil {
		r.repo.UpdateScheduledPlaybookExecution(ctx, scheduled.ID, map[string]any{
			"status":      "failed",
			"executed_at": time.Now(),
			"result":      json.RawMessage(fmt.Sprintf(`{"error":"failed to create execution: %v"}`, err)),
		})
		return
	}

	// Execute steps (similar to ExecutePlaybook handler logic)
	var steps []map[string]any
	if err := json.Unmarshal(playbook.Steps, &steps); err != nil {
		r.repo.UpdatePlaybookExecution(ctx, execID, "error", json.RawMessage(`{"error":"invalid steps format"}`))
		r.repo.UpdateScheduledPlaybookExecution(ctx, scheduled.ID, map[string]any{
			"status":       "failed",
			"executed_at":  time.Now(),
			"execution_id": execID,
			"result":       json.RawMessage(`{"error":"invalid playbook steps"}`),
		})
		return
	}

	var executionResults []map[string]any
	var executionError error

	for i, step := range steps {
		action, _ := step["action"].(string)
		if action == "" {
			continue
		}

		// Replace template variables
		stepJSON, _ := json.Marshal(step)
		stepStr := string(stepJSON)
		for key, value := range params {
			placeholder := fmt.Sprintf("{{%s}}", key)
			valueStr := fmt.Sprintf("%v", value)
			stepStr = fmt.Sprintf("%s", strings.ReplaceAll(stepStr, placeholder, valueStr))
		}
		var resolvedStep map[string]any
		if err := json.Unmarshal([]byte(stepStr), &resolvedStep); err != nil {
			executionError = fmt.Errorf("step %d: failed to resolve parameters", i+1)
			break
		}

		stepResult := map[string]any{
			"step":   i + 1,
			"action": action,
		}

		switch action {
		case "connect_receiver":
			receiverID, _ := resolvedStep["receiver_id"].(string)
			senderID, _ := resolvedStep["sender_id"].(string)
			if receiverID == "" || senderID == "" {
				executionError = fmt.Errorf("step %d: receiver_id and sender_id required", i+1)
				break
			}
			conn := models.ReceiverConnection{
				ReceiverID: receiverID,
				SenderID:   senderID,
				State:      "active",
				Role:       "master",
			}
			if err := r.repo.UpsertReceiverConnection(ctx, conn); err != nil {
				executionError = fmt.Errorf("step %d: failed to connect receiver: %v", i+1, err)
				stepResult["error"] = err.Error()
			} else {
				stepResult["success"] = true
			}

		case "disconnect_receiver":
			receiverID, _ := resolvedStep["receiver_id"].(string)
			if receiverID == "" {
				executionError = fmt.Errorf("step %d: receiver_id required", i+1)
				break
			}
			if err := r.repo.DeleteReceiverConnection(ctx, receiverID, "active", "master"); err != nil {
				executionError = fmt.Errorf("step %d: failed to disconnect receiver: %v", i+1, err)
				stepResult["error"] = err.Error()
			} else {
				stepResult["success"] = true
			}

		default:
			executionError = fmt.Errorf("step %d: unknown action: %s", i+1, action)
			stepResult["error"] = executionError.Error()
		}

		executionResults = append(executionResults, stepResult)
		if executionError != nil {
			break
		}
	}

	// Update execution records
	resultJSON, _ := json.Marshal(map[string]any{
		"steps": executionResults,
		"error": func() string {
			if executionError != nil {
				return executionError.Error()
			}
			return ""
		}(),
	})
	status := "success"
	if executionError != nil {
		status = "error"
	}
	completedAt := time.Now()
	r.repo.UpdatePlaybookExecution(ctx, execID, status, json.RawMessage(resultJSON))
	r.repo.UpdateScheduledPlaybookExecution(ctx, scheduled.ID, map[string]any{
		"status":       status,
		"executed_at": completedAt,
		"execution_id": execID,
		"result":      json.RawMessage(resultJSON),
	})
}
