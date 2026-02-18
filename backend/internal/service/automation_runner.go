package service

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"go-nmos/backend/internal/models"
)

type AutomationRunner struct {
	repo automationRepo
}

// automationRepo is a narrow dependency surface for testability.
type automationRepo interface {
	ListAutomationJobs(ctx context.Context) ([]models.AutomationJob, error)
	DetectCollisions(ctx context.Context) ([]models.CollisionGroup, error)
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
		if job.ScheduleType != "interval" {
			continue
		}

		seconds, err := strconv.Atoi(job.ScheduleValue)
		if err != nil || seconds <= 0 {
			continue
		}
		if job.LastRunAt != nil && now.Sub(*job.LastRunAt) < time.Duration(seconds)*time.Second {
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
			// Placeholder for full NMOS diff implementation.
			payload := []byte(`{"message":"nmos check is scheduled but full diff is not implemented yet"}`)
			_ = r.repo.SaveCheckerResult(ctx, "nmos", payload)
			_ = r.repo.UpdateAutomationJobRun(ctx, job.JobID, "success", payload)
		}
	}
}
