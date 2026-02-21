package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-nmos/backend/internal/models"
)

// ListEvents returns events with optional filters (C.3). GET /events?source=&severity=&since=&limit=
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	severity := strings.TrimSpace(r.URL.Query().Get("severity"))
	sinceStr := strings.TrimSpace(r.URL.Query().Get("since"))
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	var since *time.Time
	if sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = &t
		}
	}
	list, err := h.repo.ListEvents(r.Context(), source, severity, since, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list events"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": list})
}

// CreateEvent ingests one event (C.3). POST /events. Body: source_url, source_id, severity, message, payload?, flow_id?, sender_id?, receiver_id?, job_id?
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var body struct {
		SourceURL  string           `json:"source_url"`
		SourceID   string           `json:"source_id"`
		Severity   string           `json:"severity"`
		Message    string           `json:"message"`
		Payload    json.RawMessage  `json:"payload"`
		FlowID     string           `json:"flow_id"`
		SenderID   string           `json:"sender_id"`
		ReceiverID string           `json:"receiver_id"`
		JobID      string           `json:"job_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	severity := strings.TrimSpace(body.Severity)
	if severity == "" {
		severity = "info"
	}
	if severity != "info" && severity != "warning" && severity != "error" && severity != "critical" {
		severity = "info"
	}
	e := models.Event{
		SourceURL:  strings.TrimSpace(body.SourceURL),
		SourceID:   strings.TrimSpace(body.SourceID),
		Severity:   severity,
		Message:    strings.TrimSpace(body.Message),
		Payload:    body.Payload,
		FlowID:     strings.TrimSpace(body.FlowID),
		SenderID:   strings.TrimSpace(body.SenderID),
		ReceiverID: strings.TrimSpace(body.ReceiverID),
		JobID:      strings.TrimSpace(body.JobID),
		CreatedAt:  time.Now(),
	}
	if len(e.Payload) == 0 {
		e.Payload = json.RawMessage("{}")
	}
	id, err := h.repo.InsertEvent(r.Context(), e)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to insert event"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "created_at": e.CreatedAt})
}

// GetIS07Sources proxies GET to an IS-07 device's /sources (list event sources). GET /events/is07/sources?base_url=
func (h *Handler) GetIS07Sources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	baseURL := strings.TrimSpace(r.URL.Query().Get("base_url"))
	if baseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "base_url is required"})
		return
	}
	baseURL = strings.TrimRight(baseURL, "/")
	// IS-07 Events API base e.g. http://host/x-nmos/events/v1.0 -> /sources
	targetURL := baseURL + "/sources"
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "request failed", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()
	// Return as JSON; if device returns 404/500, forward status
	var out any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		writeJSON(w, resp.StatusCode, map[string]any{"error": "invalid response", "status": resp.StatusCode})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(out)
}
