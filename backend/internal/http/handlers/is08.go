package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const is08ClientTimeout = 10 * time.Second

// IS08GetIO proxies GET to the device's IS-08 /io endpoint (inputs and outputs view).
// Query: base_url (required) – full IS-08 API base, e.g. http://host/x-nmos/channel_mapping/v1.0
func (h *Handler) IS08GetIO(w http.ResponseWriter, r *http.Request) {
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
	targetURL := baseURL + "/io"
	if _, err := url.Parse(targetURL); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid base_url"})
		return
	}
	client := &http.Client{Timeout: is08ClientTimeout}
	resp, err := client.Get(targetURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "request failed", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// IS08GetMapActive proxies GET to the device's IS-08 /map/active endpoint (current channel map).
func (h *Handler) IS08GetMapActive(w http.ResponseWriter, r *http.Request) {
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
	targetURL := baseURL + "/map/active"
	client := &http.Client{Timeout: is08ClientTimeout}
	resp, err := client.Get(targetURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "request failed", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// IS08PostMapActivations creates a new activation on the device (apply channel map).
// Body: { "base_url": "http://...", "activation": { "mode": "activate_immediate", "requested": { "output1": [ ... ] } } }
func (h *Handler) IS08PostMapActivations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var payload struct {
		BaseURL    string         `json:"base_url"`
		Activation map[string]any `json:"activation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	baseURL := strings.TrimSpace(payload.BaseURL)
	if baseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "base_url is required in body"})
		return
	}
	baseURL = strings.TrimRight(baseURL, "/")
	targetURL := baseURL + "/map/activations"
	bodyBytes, _ := json.Marshal(payload.Activation)
	client := &http.Client{Timeout: is08ClientTimeout}
	resp, err := client.Post(targetURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "request failed", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}
