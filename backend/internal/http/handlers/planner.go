package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-nmos/backend/internal/models"
	"go-nmos/backend/internal/repository"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListPrivilegedBuckets(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListRootBuckets(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list buckets failed"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) ListBucketChildren(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bucket id"})
		return
	}
	items, err := h.repo.ListChildBuckets(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list children failed"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) CreateParentBucket(w http.ResponseWriter, r *http.Request) {
	var req models.AddressBucket
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req.BucketType = "parent"
	id, err := h.repo.CreateBucket(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create parent failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) CreateChildBucket(w http.ResponseWriter, r *http.Request) {
	var req models.AddressBucket
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req.BucketType = "child"
	id, err := h.repo.CreateBucket(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create child failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (h *Handler) PatchBucket(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bucket id"})
		return
	}
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := h.repo.UpdateBucket(r.Context(), id, payload); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update bucket failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bucket id"})
		return
	}
	if err := h.repo.DeleteBucket(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete bucket failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) ExportBuckets(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ExportBuckets(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "export buckets failed"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) ImportBuckets(w http.ResponseWriter, r *http.Request) {
	var payload []models.AddressBucket
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	n, err := h.repo.ImportBuckets(r.Context(), payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "import buckets failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"imported": n})
}

// GetBucketUsageStats returns usage statistics for a bucket
// GET /api/address/buckets/{id}/usage
func (h *Handler) GetBucketUsageStats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bucket id"})
		return
	}

	stats, err := h.repo.GetBucketUsageStats(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get usage stats"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// ListAllBuckets returns all buckets (root and child) for selection in forms
// GET /api/address/buckets/all
func (h *Handler) ListAllBuckets(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListAllBuckets(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list all buckets failed"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// GetBucketFlows returns all flows assigned to a specific bucket
// GET /api/address/buckets/{id}/flows
func (h *Handler) GetBucketFlows(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bucket id"})
		return
	}

	flows, err := h.repo.ListFlowsFiltered(r.Context(), repository.FlowListFilters{BucketID: &id}, 1000, 0, "updated_at", "desc")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get bucket flows"})
		return
	}

	writeJSON(w, http.StatusOK, flows)
}
