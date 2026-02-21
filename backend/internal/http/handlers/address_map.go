package handlers

import (
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"go-nmos/backend/internal/models"
)

func (h *Handler) AddressMap(w http.ResponseWriter, r *http.Request) {
	flows, err := h.repo.ExportFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "address map failed"})
		return
	}

	type bucket struct {
		Subnet          string           `json:"subnet"`
		Count           int              `json:"count"`
		UsedIPs         int              `json:"used_ips"`         // Unique IP addresses used
		TotalIPs        int              `json:"total_ips"`         // Total IPs in /24 subnet (254)
		AvailableIPs    int              `json:"available_ips"`    // Available IPs
		UsagePercentage float64          `json:"usage_percentage"` // Usage percentage
		Flows           map[string][]any `json:"flows"`
		UsedIPList      []string         `json:"used_ip_list"`     // List of used IPs
		BucketID        *int64           `json:"bucket_id,omitempty"`        // Associated planner bucket ID
		BucketName      string           `json:"bucket_name,omitempty"`      // Associated planner bucket name
		BucketCIDR      string           `json:"bucket_cidr,omitempty"`       // Associated planner bucket CIDR
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
				Subnet:     subnet,
				Flows:      map[string][]any{},
				UsedIPList: []string{},
				TotalIPs:   254, // /24 subnet has 254 usable IPs (excluding .0 and .255)
			}
			grouped[subnet] = b
		}
		b.Count++
		
		// Track unique IPs
		ipExists := false
		for _, usedIP := range b.UsedIPList {
			if usedIP == f.MulticastIP {
				ipExists = true
				break
			}
		}
		if !ipExists {
			b.UsedIPList = append(b.UsedIPList, f.MulticastIP)
			b.UsedIPs++
		}
		
		b.Flows[f.MulticastIP] = append(b.Flows[f.MulticastIP], map[string]any{
			"id":           f.ID,
			"display_name": f.DisplayName,
			"flow_id":      f.FlowID,
			"port":         f.Port,
			"status":       f.FlowStatus,
		})
	}

	// Load all buckets to match subnets
	allBuckets, err := h.repo.ListAllBuckets(r.Context())
	if err != nil {
		// Log error but continue without bucket matching
		allBuckets = []models.AddressBucket{}
	}
	
	// Helper function to check if an IP is within a CIDR or IP range
	matchBucket := func(subnetStr string, bucket *models.AddressBucket) bool {
		// Extract base IP from subnet (e.g., "239.0.0.0/24" -> "239.0.0.0")
		subnetBaseIP := net.ParseIP(strings.TrimSuffix(subnetStr, "/24"))
		if subnetBaseIP == nil {
			return false
		}
		
		// Check CIDR match
		if bucket.CIDR != "" {
			_, ipNet, err := net.ParseCIDR(bucket.CIDR)
			if err == nil && ipNet.Contains(subnetBaseIP) {
				return true
			}
		}
		
		// Check IP range match
		if bucket.StartIP != "" && bucket.EndIP != "" {
			startIP := net.ParseIP(bucket.StartIP)
			endIP := net.ParseIP(bucket.EndIP)
			if startIP != nil && endIP != nil {
				subnetBase := subnetBaseIP.To4()
				start := startIP.To4()
				end := endIP.To4()
				if subnetBase != nil && start != nil && end != nil {
					if bytesCompare(subnetBase, start) >= 0 && bytesCompare(subnetBase, end) <= 0 {
						return true
					}
				}
			}
		}
		
		return false
	}
	
	// Calculate usage statistics for each bucket and match with planner buckets
	for _, b := range grouped {
		b.AvailableIPs = b.TotalIPs - b.UsedIPs
		if b.TotalIPs > 0 {
			b.UsagePercentage = float64(b.UsedIPs) / float64(b.TotalIPs) * 100
		}
		// Sort used IP list
		sort.Strings(b.UsedIPList)
		
		// Match with planner bucket
		for _, plannerBucket := range allBuckets {
			if matchBucket(b.Subnet, &plannerBucket) {
				b.BucketID = &plannerBucket.ID
				b.BucketName = plannerBucket.Name
				b.BucketCIDR = plannerBucket.CIDR
				break // Use first matching bucket
			}
		}
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

// GetSubnetDetailedAnalysis returns detailed analysis for a specific subnet
// GET /api/address-map/subnet/{subnet}/analysis
func (h *Handler) GetSubnetDetailedAnalysis(w http.ResponseWriter, r *http.Request) {
	subnetParam := r.URL.Query().Get("subnet")
	if subnetParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "subnet parameter is required"})
		return
	}

	// Parse subnet (e.g., "239.0.0.0/24")
	parts := strings.Split(subnetParam, "/")
	if len(parts) != 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subnet format (expected: x.x.x.x/24)"})
		return
	}

	baseIP := parts[0]
	ipParts := strings.Split(baseIP, ".")
	if len(ipParts) != 4 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid IP format"})
		return
	}

	// Get all flows
	flows, err := h.repo.ExportFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get flows"})
		return
	}

	// Find flows in this subnet
	usedIPs := make(map[string][]map[string]any)
	for _, f := range flows {
		if strings.TrimSpace(f.MulticastIP) == "" {
			continue
		}
		fParts := strings.Split(f.MulticastIP, ".")
		if len(fParts) == 4 {
			// Check if IP is in the same /24 subnet
			if fParts[0] == ipParts[0] && fParts[1] == ipParts[1] && fParts[2] == ipParts[2] {
				usedIPs[f.MulticastIP] = append(usedIPs[f.MulticastIP], map[string]any{
					"id":           f.ID,
					"display_name": f.DisplayName,
					"flow_id":      f.FlowID,
					"port":         f.Port,
					"status":       f.FlowStatus,
				})
			}
		}
	}

	// Generate list of used IPs
	usedIPList := make([]string, 0, len(usedIPs))
	for ip := range usedIPs {
		usedIPList = append(usedIPList, ip)
	}
	sort.Strings(usedIPList)

	// Generate available IPs
	usedIPSet := make(map[string]bool)
	for ip := range usedIPs {
		usedIPSet[ip] = true
	}

	availableIPs := []string{}
	for i := 1; i <= 254; i++ {
		ip := ipParts[0] + "." + ipParts[1] + "." + ipParts[2] + "." + strconv.Itoa(i)
		if !usedIPSet[ip] {
			availableIPs = append(availableIPs, ip)
		}
	}

	usedCount := len(usedIPs)
	totalIPs := 254
	availableCount := totalIPs - usedCount
	usagePercentage := float64(usedCount) / float64(totalIPs) * 100

	writeJSON(w, http.StatusOK, map[string]any{
		"subnet":           subnetParam,
		"used_ips":         usedCount,
		"available_ips":    availableCount,
		"total_ips":        totalIPs,
		"usage_percentage": usagePercentage,
		"used_ip_list":     usedIPList,
		"available_ip_list": availableIPs[:min(100, len(availableIPs))], // Limit to first 100 for performance
		"flows_by_ip":      usedIPs,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// bytesCompare compares two IP addresses byte by byte
func bytesCompare(a, b []byte) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}
