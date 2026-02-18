package handlers

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logWriteMu sync.Mutex

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

		apiLine := fmt.Sprintf("%s method=%s path=%s status=%d duration_ms=%d ip=%s user=%s",
			time.Now().Format(time.RFC3339), r.Method, r.URL.Path, rw.status, duration, r.RemoteAddr, username)
		_ = appendLogLine(filepath.Join(h.cfg.LogDir, "api.log"), apiLine)

		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
			auditLine := fmt.Sprintf("%s action=%s path=%s status=%d user=%s",
				time.Now().Format(time.RFC3339), r.Method, r.URL.Path, rw.status, username)
			_ = appendLogLine(filepath.Join(h.cfg.LogDir, "audit.log"), auditLine)
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
