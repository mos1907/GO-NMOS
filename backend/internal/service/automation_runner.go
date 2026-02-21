package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go-nmos/backend/internal/models"
	"go-nmos/backend/internal/schedule"
)

type AutomationRunner struct {
	repo automationRepo
}

// automationRepo is a narrow dependency surface for testability.
type automationRepo interface {
	ListAutomationJobs(ctx context.Context) ([]models.AutomationJob, error)
	DetectCollisions(ctx context.Context) ([]models.CollisionGroup, error)
	ListFlows(ctx context.Context, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error)
	ListNMOSSenders(ctx context.Context, deviceID string) ([]models.NMOSSender, error)
	SaveCheckerResult(ctx context.Context, kind string, result []byte) error
	UpdateAutomationJobRun(ctx context.Context, jobID, status string, result []byte) error
}

func NewAutomationRunner(repo automationRepo) *AutomationRunner {
	return &AutomationRunner{repo: repo}
}

func (r *AutomationRunner) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
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

func (r *AutomationRunner) runCycle(ctx context.Context) {
	jobs, err := r.repo.ListAutomationJobs(ctx)
	if err != nil {
		log.Printf("automation runner: list jobs failed: %v", err)
		return
	}

	now := time.Now()
	for _, job := range jobs {
		if !job.Enabled {
			continue
		}
		if !schedule.ShouldRun(job, now) {
			continue
		}

		switch job.JobID {
		case "collision_check":
			collisions, err := r.repo.DetectCollisions(ctx)
			if err != nil {
				_ = r.repo.UpdateAutomationJobRun(ctx, job.JobID, "error", []byte(`{"error":"collision check failed"}`))
				continue
			}
			payload, _ := json.Marshal(map[string]any{
				"total_collisions": len(collisions),
				"items":            collisions,
			})
			_ = r.repo.SaveCheckerResult(ctx, "collisions", payload)
			_ = r.repo.UpdateAutomationJobRun(ctx, job.JobID, "success", payload)

		case "nmos_check":
			flows, err := r.repo.ListFlows(ctx, 10000, 0, "updated_at", "desc")
			if err != nil {
				_ = r.repo.UpdateAutomationJobRun(ctx, job.JobID, "error", []byte(`{"error":"failed to list flows"}`))
				continue
			}
			senders, err := r.repo.ListNMOSSenders(ctx, "")
			if err != nil {
				_ = r.repo.UpdateAutomationJobRun(ctx, job.JobID, "error", []byte(`{"error":"failed to list NMOS senders"}`))
				continue
			}
			differences := make([]map[string]any, 0)
			for _, flow := range flows {
				if flow.FlowID == "" {
					continue
				}
				found := false
				for _, sender := range senders {
					if sender.FlowID == flow.FlowID {
						found = true
						break
					}
				}
				if !found {
					differences = append(differences, map[string]any{
						"flow_id":      flow.FlowID,
						"display_name": flow.DisplayName,
						"issue":        "flow_id not found in NMOS registry",
						"type":         "missing_sender",
					})
				}
			}
			payload := map[string]any{
				"total_differences": len(differences),
				"items":             differences,
				"checked_flows":     len(flows),
				"timeout_seconds":   5,
			}
			raw, _ := json.Marshal(payload)
			_ = r.repo.SaveCheckerResult(ctx, "nmos", raw)
			_ = r.repo.UpdateAutomationJobRun(ctx, job.JobID, "success", raw)
		}
	}
}
