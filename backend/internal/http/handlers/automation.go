package handlers

import (
	"encoding/json"
	"net/http"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListAutomationJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.repo.ListAutomationJobs(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list jobs failed"})
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (h *Handler) GetAutomationSummary(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.repo.ListAutomationJobs(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "summary failed"})
		return
	}
	enabled := 0
	for _, j := range jobs {
		if j.Enabled {
			enabled++
		}
	}
	// Include latest checker results (collisions & nmos difference count) if available
	collisions, _ := h.repo.GetLatestCheckerResult(r.Context(), "collisions")
	nmosDiff, _ := h.repo.GetLatestCheckerResult(r.Context(), "nmos")

	type summary struct {
		TotalJobs           int    `json:"total_jobs"`
		EnabledJobs         int    `json:"enabled_jobs"`
		CollisionCount      int    `json:"collision_count"`
		NMOSDifferenceCount int    `json:"nmos_difference_count"`
		LastUpdated         string `json:"last_updated,omitempty"`
	}

	resp := summary{
		TotalJobs:   len(jobs),
		EnabledJobs: enabled,
	}

	if collisions != nil && len(collisions.Result) > 0 {
		// collisions.Result is JSON; for dashboard we only expose "total groups"
		// { "groups": [ ... ] } → collision_count = len(groups)
		var payload map[string]any
		if err := json.Unmarshal(collisions.Result, &payload); err == nil {
			if groups, ok := payload["groups"].([]any); ok {
				resp.CollisionCount = len(groups)
				resp.LastUpdated = collisions.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
		}
	}
	if nmosDiff != nil && len(nmosDiff.Result) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(nmosDiff.Result, &payload); err == nil {
			if n, ok := payload["difference_count"].(float64); ok {
				resp.NMOSDifferenceCount = int(n)
				if resp.LastUpdated == "" {
					resp.LastUpdated = nmosDiff.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) PutAutomationJob(w http.ResponseWriter, r *http.Request) {
	var req models.AutomationJob
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.JobID == "" {
		req.JobID = "collision_check"
	}
	if req.JobType == "" {
		req.JobType = req.JobID
	}
	if req.ScheduleType == "" {
		req.ScheduleType = "interval"
	}
	if req.ScheduleValue == "" {
		req.ScheduleValue = "1800"
	}
	if req.LastRunResult == nil {
		req.LastRunResult = json.RawMessage(`{}`)
	}
	if err := h.repo.UpsertAutomationJob(r.Context(), req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update job failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) GetAutomationJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	job, err := h.repo.GetAutomationJob(r.Context(), jobID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (h *Handler) EnableAutomationJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	if err := h.repo.SetAutomationJobEnabled(r.Context(), jobID, true); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "enable failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) DisableAutomationJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	if err := h.repo.SetAutomationJobEnabled(r.Context(), jobID, false); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "disable failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
