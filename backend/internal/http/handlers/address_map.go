package handlers

import (
	"net/http"
	"sort"
	"strings"
)

func (h *Handler) AddressMap(w http.ResponseWriter, r *http.Request) {
	flows, err := h.repo.ExportFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "address map failed"})
		return
	}

	type bucket struct {
		Subnet string           `json:"subnet"`
		Count  int              `json:"count"`
		Flows  map[string][]any `json:"flows"`
	}
	grouped := map[string]*bucket{}

	for _, f := range flows {
		if strings.TrimSpace(f.MulticastIP) == "" {
			continue
		}
		parts := strings.Split(f.MulticastIP, ".")
		if len(parts) < 3 {
			continue
		}
		subnet := parts[0] + "." + parts[1] + "." + parts[2] + ".0/24"
		b, ok := grouped[subnet]
		if !ok {
			b = &bucket{
				Subnet: subnet,
				Flows:  map[string][]any{},
			}
			grouped[subnet] = b
		}
		b.Count++
		b.Flows[f.MulticastIP] = append(b.Flows[f.MulticastIP], map[string]any{
			"id":           f.ID,
			"display_name": f.DisplayName,
			"flow_id":      f.FlowID,
			"port":         f.Port,
			"status":       f.FlowStatus,
		})
	}

	keys := make([]string, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	resp := make([]bucket, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, *grouped[k])
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total_subnets": len(resp),
		"items":         resp,
	})
}
