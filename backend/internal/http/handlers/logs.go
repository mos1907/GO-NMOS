package handlers

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

var logWriteMu sync.Mutex

type structuredLogEntry struct {
	Timestamp     time.Time `json:"ts"`
	Level         string    `json:"level"`
	Kind          string    `json:"kind"`
	Component     string    `json:"component,omitempty"`
	Message       string    `json:"message,omitempty"`
	Method        string    `json:"method,omitempty"`
	Path          string    `json:"path,omitempty"`
	Status        int       `json:"status,omitempty"`
	DurationMs    int64     `json:"duration_ms,omitempty"`
	User          string    `json:"user,omitempty"`
	RequestID     string    `json:"request_id,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Site          string    `json:"site,omitempty"`
	RemoteIP      string    `json:"remote_ip,omitempty"`
}

func (h *Handler) Logs(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "api"
	}
	if kind != "api" && kind != "audit" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind must be api or audit"})
		return
	}
	lines := 200
	if v := r.URL.Query().Get("lines"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 5000 {
			lines = n
		}
	}
	content, err := tailLines(filepath.Join(h.cfg.LogDir, kind+".log"), lines)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read logs failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kind":  kind,
		"lines": content,
	})
}

func (h *Handler) DownloadLogs(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "api"
	}
	if kind != "api" && kind != "audit" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind must be api or audit"})
		return
	}
	filePath := filepath.Join(h.cfg.LogDir, kind+".log")
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Disposition", "attachment; filename="+kind+".log")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "open log file failed"})
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+kind+".log")
	_, _ = io.Copy(w, f)
}

func (h *Handler) RequestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Milliseconds()

		username := "-"
		claims, _ := r.Context().Value(userContextKey).(*AuthClaims)
		if claims != nil && claims.Username != "" {
			username = claims.Username
		}

		requestID := chimw.GetReqID(r.Context())
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = requestID
		}

		component := "-"
		if strings.HasPrefix(r.URL.Path, "/api/") {
			rest := strings.TrimPrefix(r.URL.Path, "/api/")
			if idx := strings.Index(rest, "/"); idx >= 0 {
				component = rest[:idx]
			} else if rest != "" {
				component = rest
			}
		}

		site := r.URL.Query().Get("site")

		apiEntry := structuredLogEntry{
			Timestamp:     time.Now().UTC(),
			Level:         "info",
			Kind:          "api",
			Component:     component,
			Method:        r.Method,
			Path:          r.URL.Path,
			Status:        rw.status,
			DurationMs:    duration,
			User:          username,
			RequestID:     requestID,
			CorrelationID: correlationID,
			Site:          site,
			RemoteIP:      r.RemoteAddr,
			Message:       "http_request",
		}
		if apiLine, err := json.Marshal(apiEntry); err == nil {
			_ = appendLogLine(filepath.Join(h.cfg.LogDir, "api.log"), string(apiLine))
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
			auditEntry := structuredLogEntry{
				Timestamp:     time.Now().UTC(),
				Level:         "info",
				Kind:          "audit",
				Component:     component,
				Method:        r.Method,
				Path:          r.URL.Path,
				Status:        rw.status,
				User:          username,
				RequestID:     requestID,
				CorrelationID: correlationID,
				Site:          site,
				RemoteIP:      r.RemoteAddr,
				Message:       "audit_event",
			}
			if auditLine, err := json.Marshal(auditEntry); err == nil {
				_ = appendLogLine(filepath.Join(h.cfg.LogDir, "audit.log"), string(auditLine))
			}
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func appendLogLine(filePath, line string) error {
	logWriteMu.Lock()
	defer logWriteMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}

func tailLines(filePath string, maxLines int) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines, nil
}
