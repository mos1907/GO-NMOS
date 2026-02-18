package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type nmosDiscoverRequest struct {
	BaseURL string `json:"base_url"`
}

type nmosResource struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	FlowID       string `json:"flow_id"`
	ManifestHREF string `json:"manifest_href"`
	Format       string `json:"format,omitempty"`
}

func (h *Handler) DiscoverNMOS(w http.ResponseWriter, r *http.Request) {
	baseURL := strings.TrimSpace(r.URL.Query().Get("base_url"))
	if baseURL == "" && r.Method == http.MethodPost {
		var req nmosDiscoverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			baseURL = strings.TrimSpace(req.BaseURL)
		}
	}
	if baseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "base_url is required"})
		return
	}
	baseURL = strings.TrimRight(baseURL, "/")

	versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/node/", baseURL))
	if err != nil || len(versions) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not discover IS-04 versions"})
		return
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")

	sendersData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/senders", baseURL, version))
	receiversData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/receivers", baseURL, version))
	flowsData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/flows", baseURL, version))

	flowFormatByID := map[string]string{}
	for _, item := range flowsData {
		id, _ := item["id"].(string)
		format, _ := item["format"].(string)
		if id != "" {
			flowFormatByID[id] = format
		}
	}

	senders := make([]nmosResource, 0, len(sendersData))
	for _, s := range sendersData {
		flowID, _ := s["flow_id"].(string)
		senders = append(senders, nmosResource{
			ID:           asString(s["id"]),
			Label:        fallback(asString(s["label"]), asString(s["id"])),
			Description:  asString(s["description"]),
			FlowID:       flowID,
			ManifestHREF: asString(s["manifest_href"]),
			Format:       flowFormatByID[flowID],
		})
	}

	receivers := make([]nmosResource, 0, len(receiversData))
	for _, rec := range receiversData {
		receivers = append(receivers, nmosResource{
			ID:          asString(rec["id"]),
			Label:       fallback(asString(rec["label"]), asString(rec["id"])),
			Description: asString(rec["description"]),
			Format:      asString(rec["format"]),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"base_url":     baseURL,
		"is04_version": version,
		"senders":      senders,
		"receivers":    receivers,
		"counts": map[string]int{
			"senders":   len(senders),
			"receivers": len(receivers),
			"flows":     len(flowsData),
		},
	})
}

func (h *Handler) CheckFlowNMOS(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid flow id"})
		return
	}
	flow, err := h.repo.GetFlowByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}
	baseURL := strings.TrimSpace(r.URL.Query().Get("base_url"))
	if baseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "base_url is required"})
		return
	}

	baseURL = strings.TrimRight(baseURL, "/")
	versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/node/", baseURL))
	if err != nil || len(versions) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "nmos node unavailable"})
		return
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")
	sendersData, err := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/senders", baseURL, version))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "senders read failed"})
		return
	}

	matches := []map[string]any{}
	for _, s := range sendersData {
		if asString(s["flow_id"]) == flow.FlowID {
			matches = append(matches, map[string]any{
				"id":            asString(s["id"]),
				"label":         fallback(asString(s["label"]), asString(s["id"])),
				"manifest_href": asString(s["manifest_href"]),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"flow_id":       flow.FlowID,
		"display_name":  flow.DisplayName,
		"nmos_matches":  matches,
		"matched_count": len(matches),
		"is_match":      len(matches) > 0,
	})
}

type nmosApplyRequest struct {
	ConnectionURL string `json:"connection_url"`
	SenderID      string `json:"sender_id,omitempty"`
}

func (h *Handler) ApplyFlowNMOS(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid flow id"})
		return
	}
	flow, err := h.repo.GetFlowByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}
	var req nmosApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req.ConnectionURL = strings.TrimSpace(req.ConnectionURL)
	if req.ConnectionURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "connection_url is required"})
		return
	}
	parsedURL, err := url.Parse(req.ConnectionURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "connection_url must be a valid absolute URL"})
		return
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "connection_url must start with http:// or https://"})
		return
	}

	patchBody := map[string]any{
		"activation":    map[string]any{"mode": "activate_immediate"},
		"master_enable": true,
		"transport_params": []map[string]any{
			{
				"multicast_ip":     flow.MulticastIP,
				"source_ip":        flow.SourceIP,
				"destination_port": flow.Port,
				"rtp_enabled":      true,
			},
		},
	}
	if req.SenderID != "" {
		patchBody["sender_id"] = req.SenderID
	}

	client := &http.Client{Timeout: 8 * time.Second}
	payload, _ := json.Marshal(patchBody)
	httpReq, _ := http.NewRequest(http.MethodPatch, parsedURL.String(), strings.NewReader(string(payload)))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "patch request failed"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error":       "nmos apply returned non-2xx status",
			"status_code": resp.StatusCode,
			"body":        string(body),
			"request":     patchBody,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(body),
		"request":     patchBody,
	})
}

func (h *Handler) fetchJSONList(url string) ([]string, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var versions []string
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func (h *Handler) fetchJSONArray(url string) ([]map[string]any, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func parsePathInt64(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}
