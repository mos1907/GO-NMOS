package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go-nmos/backend/internal/models"
)

type fakeAutomationRepo struct {
	jobs []models.AutomationJob

	detectCalls int
	saves       []struct {
		kind   string
		result []byte
	}
	updates []struct {
		jobID  string
		status string
		result []byte
	}
}

func (f *fakeAutomationRepo) ListAutomationJobs(ctx context.Context) ([]models.AutomationJob, error) {
	return f.jobs, nil
}

func (f *fakeAutomationRepo) DetectCollisions(ctx context.Context) ([]models.CollisionGroup, error) {
	f.detectCalls++
	return []models.CollisionGroup{
		{MulticastIP: "239.1.1.1", Port: 5004, Count: 2, FlowNames: []string{"A", "B"}},
	}, nil
}

func (f *fakeAutomationRepo) ListFlows(ctx context.Context, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error) {
	return nil, nil
}

func (f *fakeAutomationRepo) ListNMOSSenders(ctx context.Context, deviceID string) ([]models.NMOSSender, error) {
	return nil, nil
}

func (f *fakeAutomationRepo) SaveCheckerResult(ctx context.Context, kind string, result []byte) error {
	f.saves = append(f.saves, struct {
		kind   string
		result []byte
	}{kind: kind, result: append([]byte(nil), result...)})
	return nil
}

func (f *fakeAutomationRepo) UpdateAutomationJobRun(ctx context.Context, jobID, status string, result []byte) error {
	f.updates = append(f.updates, struct {
		jobID  string
		status string
		result []byte
	}{jobID: jobID, status: status, result: append([]byte(nil), result...)})
	return nil
}

func TestAutomationRunner_RunCycle_CollisionCheckRuns(t *testing.T) {
	repo := &fakeAutomationRepo{
		jobs: []models.AutomationJob{
			{JobID: "collision_check", Enabled: true, ScheduleType: "interval", ScheduleValue: "1", LastRunAt: nil},
		},
	}
	runner := NewAutomationRunner(repo)
	runner.runCycle(context.Background())

	if repo.detectCalls != 1 {
		t.Fatalf("expected DetectCollisions called 1 time, got %d", repo.detectCalls)
	}
	if len(repo.saves) != 1 || repo.saves[0].kind != "collisions" {
		t.Fatalf("expected SaveCheckerResult(kind=collisions) once, got %+v", repo.saves)
	}
	if len(repo.updates) != 1 || repo.updates[0].jobID != "collision_check" || repo.updates[0].status != "success" {
		t.Fatalf("expected UpdateAutomationJobRun success once, got %+v", repo.updates)
	}

	var payload map[string]any
	if err := json.Unmarshal(repo.updates[0].result, &payload); err != nil {
		t.Fatalf("expected valid json payload, got err=%v", err)
	}
	if payload["total_collisions"] == nil {
		t.Fatalf("expected total_collisions in payload, got %v", payload)
	}
}

func TestAutomationRunner_RunCycle_SkipsWhenIntervalNotReached(t *testing.T) {
	now := time.Now()
	repo := &fakeAutomationRepo{
		jobs: []models.AutomationJob{
			{JobID: "collision_check", Enabled: true, ScheduleType: "interval", ScheduleValue: "3600", LastRunAt: &now},
		},
	}
	runner := NewAutomationRunner(repo)
	runner.runCycle(context.Background())

	if repo.detectCalls != 0 {
		t.Fatalf("expected DetectCollisions not called, got %d", repo.detectCalls)
	}
	if len(repo.updates) != 0 {
		t.Fatalf("expected no UpdateAutomationJobRun, got %+v", repo.updates)
	}
}

func TestAutomationRunner_RunCycle_NMOSCheckPlaceholderRuns(t *testing.T) {
	repo := &fakeAutomationRepo{
		jobs: []models.AutomationJob{
			{JobID: "nmos_check", Enabled: true, ScheduleType: "interval", ScheduleValue: "1", LastRunAt: nil},
		},
	}
	runner := NewAutomationRunner(repo)
	runner.runCycle(context.Background())

	if repo.detectCalls != 0 {
		t.Fatalf("expected DetectCollisions not called, got %d", repo.detectCalls)
	}
	if len(repo.saves) != 1 || repo.saves[0].kind != "nmos" {
		t.Fatalf("expected SaveCheckerResult(kind=nmos) once, got %+v", repo.saves)
	}
	if len(repo.updates) != 1 || repo.updates[0].jobID != "nmos_check" || repo.updates[0].status != "success" {
		t.Fatalf("expected UpdateAutomationJobRun success once, got %+v", repo.updates)
	}
}
