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
	writeJSON(w, http.StatusOK, map[string]int{
		"total_jobs":   len(jobs),
		"enabled_jobs": enabled,
	})
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
