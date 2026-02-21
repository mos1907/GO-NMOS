package interop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// TestTarget represents a device or registry to test against
type TestTarget struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "node" | "registry"
	BaseURL     string `json:"base_url"`
	Vendor      string `json:"vendor,omitempty"`
	Model       string `json:"model,omitempty"`
	Description string `json:"description,omitempty"`
}

// TestResult represents the result of an interoperability test
type TestResult struct {
	TestName      string            `json:"test_name"`
	Target        TestTarget        `json:"target"`
	Passed        bool              `json:"passed"`
	Status        string            `json:"status"` // "pass" | "fail" | "warning" | "skip"
	Message       string            `json:"message"`
	Details       map[string]any    `json:"details,omitempty"`
	Duration      time.Duration     `json:"duration"`
	Timestamp     time.Time         `json:"timestamp"`
}

// TestSuite runs a series of interoperability tests
type TestSuite struct {
	client *http.Client
}

func NewTestSuite() *TestSuite {
	return &TestSuite{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RunNodeTests runs interoperability tests against an NMOS node
func (ts *TestSuite) RunNodeTests(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}

	// Test 1: IS-04 Node API discovery
	result := ts.testIS04NodeDiscovery(ctx, target)
	results = append(results, result)

	if !result.Passed {
		return results // Stop if basic discovery fails
	}

	// Test 2: IS-04 Resource enumeration
	results = append(results, ts.testIS04Resources(ctx, target)...)

	// Test 3: IS-05 Connection API discovery
	results = append(results, ts.testIS05Discovery(ctx, target)...)

	// Test 4: IS-05 Receiver operations
	results = append(results, ts.testIS05ReceiverOps(ctx, target)...)

	// Test 5: IS-08 Audio Channel Mapping (if available)
	results = append(results, ts.testIS08Discovery(ctx, target)...)

	// Test 6: IS-07 Events API (if available)
	results = append(results, ts.testIS07Discovery(ctx, target)...)

	return results
}

// RunRegistryTests runs interoperability tests against an NMOS registry
func (ts *TestSuite) RunRegistryTests(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}

	// Test 1: IS-04 Query API discovery
	result := ts.testIS04QueryDiscovery(ctx, target)
	results = append(results, result)

	if !result.Passed {
		return results
	}

	// Test 2: Query API resource queries
	results = append(results, ts.testIS04QueryResources(ctx, target)...)

	// Test 3: Registry health and consistency
	results = append(results, ts.testRegistryHealth(ctx, target)...)

	return results
}

// testIS04NodeDiscovery tests IS-04 Node API version discovery
func (ts *TestSuite) testIS04NodeDiscovery(ctx context.Context, target TestTarget) TestResult {
	start := time.Now()
	testName := "IS-04 Node API Discovery"

	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/node/", target.BaseURL))
	if err != nil {
		return TestResult{
			TestName:  testName,
			Target:    target,
			Passed:    false,
			Status:    "fail",
			Message:   fmt.Sprintf("Failed to discover IS-04 Node API versions: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	if len(versions) == 0 {
		return TestResult{
			TestName:  testName,
			Target:    target,
			Passed:    false,
			Status:    "fail",
			Message:   "No IS-04 Node API versions found",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	sort.Strings(versions)
	latestVersion := strings.Trim(versions[len(versions)-1], "/")

	return TestResult{
		TestName:  testName,
		Target:    target,
		Passed:    true,
		Status:    "pass",
		Message:   fmt.Sprintf("Discovered IS-04 Node API version: %s", latestVersion),
		Details: map[string]any{
			"versions":       versions,
			"latest_version": latestVersion,
		},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}

// testIS04Resources tests IS-04 resource enumeration
func (ts *TestSuite) testIS04Resources(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}

	// Get latest version
	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/node/", target.BaseURL))
	if err != nil || len(versions) == 0 {
		return results
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")

	// Test devices
	start := time.Now()
	devices, err := ts.fetchJSONArray(ctx, fmt.Sprintf("%s/x-nmos/node/%s/devices", target.BaseURL, version))
	results = append(results, TestResult{
		TestName:  "IS-04 Devices Enumeration",
		Target:    target,
		Passed:    err == nil,
		Status:    func() string { if err == nil { return "pass" }; return "fail" }(),
		Message:   fmt.Sprintf("Found %d devices", len(devices)),
		Details:   map[string]any{"count": len(devices)},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Test flows
	start = time.Now()
	flows, err := ts.fetchJSONArray(ctx, fmt.Sprintf("%s/x-nmos/node/%s/flows", target.BaseURL, version))
	results = append(results, TestResult{
		TestName:  "IS-04 Flows Enumeration",
		Target:    target,
		Passed:    err == nil,
		Status:    func() string { if err == nil { return "pass" }; return "fail" }(),
		Message:   fmt.Sprintf("Found %d flows", len(flows)),
		Details:   map[string]any{"count": len(flows)},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Test senders
	start = time.Now()
	senders, err := ts.fetchJSONArray(ctx, fmt.Sprintf("%s/x-nmos/node/%s/senders", target.BaseURL, version))
	results = append(results, TestResult{
		TestName:  "IS-04 Senders Enumeration",
		Target:    target,
		Passed:    err == nil,
		Status:    func() string { if err == nil { return "pass" }; return "fail" }(),
		Message:   fmt.Sprintf("Found %d senders", len(senders)),
		Details:   map[string]any{"count": len(senders)},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Test receivers
	start = time.Now()
	receivers, err := ts.fetchJSONArray(ctx, fmt.Sprintf("%s/x-nmos/node/%s/receivers", target.BaseURL, version))
	results = append(results, TestResult{
		TestName:  "IS-04 Receivers Enumeration",
		Target:    target,
		Passed:    err == nil,
		Status:    func() string { if err == nil { return "pass" }; return "fail" }(),
		Message:   fmt.Sprintf("Found %d receivers", len(receivers)),
		Details:   map[string]any{"count": len(receivers)},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	return results
}

// testIS05Discovery tests IS-05 Connection API discovery
func (ts *TestSuite) testIS05Discovery(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}
	start := time.Now()

	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/connection/", target.BaseURL))
	if err != nil {
		return []TestResult{{
			TestName:  "IS-05 Connection API Discovery",
			Target:    target,
			Passed:    false,
			Status:    "fail",
			Message:   fmt.Sprintf("IS-05 Connection API not found: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}}
	}

	if len(versions) == 0 {
		return []TestResult{{
			TestName:  "IS-05 Connection API Discovery",
			Target:    target,
			Passed:    false,
			Status:    "warning",
			Message:   "IS-05 Connection API not available",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}}
	}

	sort.Strings(versions)
	latestVersion := strings.Trim(versions[len(versions)-1], "/")

	results = append(results, TestResult{
		TestName:  "IS-05 Connection API Discovery",
		Target:    target,
		Passed:    true,
		Status:    "pass",
		Message:   fmt.Sprintf("Discovered IS-05 Connection API version: %s", latestVersion),
		Details: map[string]any{
			"versions":       versions,
			"latest_version": latestVersion,
		},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Test IS-05 receivers endpoint
	receiversStart := time.Now()
	receiversURL := fmt.Sprintf("%s/x-nmos/connection/%s/single/receivers", target.BaseURL, latestVersion)
	receivers, err := ts.fetchJSONArray(ctx, receiversURL)
	results = append(results, TestResult{
		TestName:  "IS-05 Receivers Endpoint",
		Target:    target,
		Passed:    err == nil,
		Status:    func() string { if err == nil { return "pass" }; return "fail" }(),
		Message:   fmt.Sprintf("IS-05 receivers endpoint accessible, found %d receivers", len(receivers)),
		Details:   map[string]any{"count": len(receivers), "endpoint": receiversURL},
		Duration:  time.Since(receiversStart),
		Timestamp: time.Now(),
	})

	return results
}

// testIS05ReceiverOps tests IS-05 receiver operations
func (ts *TestSuite) testIS05ReceiverOps(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}

	// Get IS-05 version
	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/connection/", target.BaseURL))
	if err != nil || len(versions) == 0 {
		return results
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")

	// Get receivers
	receiversURL := fmt.Sprintf("%s/x-nmos/connection/%s/single/receivers", target.BaseURL, version)
	receivers, err := ts.fetchJSONArray(ctx, receiversURL)
	if err != nil || len(receivers) == 0 {
		return []TestResult{{
			TestName:  "IS-05 Receiver Operations",
			Target:    target,
			Passed:    true,
			Status:    "skip",
			Message:   "No receivers available for testing",
			Duration:  0,
			Timestamp: time.Now(),
		}}
	}

	// Test receiver staging endpoint (read-only test)
	if len(receivers) > 0 {
		receiverID := ""
		if id, ok := receivers[0]["id"].(string); ok {
			receiverID = id
		}
		if receiverID != "" {
			stagedStart := time.Now()
			stagedURL := fmt.Sprintf("%s/x-nmos/connection/%s/single/receivers/%s/staged", target.BaseURL, version, receiverID)
			resp, err := ts.client.Get(stagedURL)
			if err == nil {
				resp.Body.Close()
				results = append(results, TestResult{
					TestName:  "IS-05 Receiver Staged Endpoint",
					Target:    target,
					Passed:    resp.StatusCode == 200 || resp.StatusCode == 404,
					Status:    func() string { if resp.StatusCode == 200 || resp.StatusCode == 404 { return "pass" }; return "fail" }(),
					Message:   fmt.Sprintf("Receiver staged endpoint accessible (status: %d)", resp.StatusCode),
					Details:   map[string]any{"receiver_id": receiverID, "status_code": resp.StatusCode},
					Duration:  time.Since(stagedStart),
					Timestamp: time.Now(),
				})
			}
		}
	}

	return results
}

// testIS08Discovery tests IS-08 Audio Channel Mapping API discovery
func (ts *TestSuite) testIS08Discovery(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}
	start := time.Now()

	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/channelmapping/", target.BaseURL))
	if err != nil {
		return []TestResult{{
			TestName:  "IS-08 Audio Channel Mapping Discovery",
			Target:    target,
			Passed:    true,
			Status:    "skip",
			Message:   "IS-08 Audio Channel Mapping API not available (optional)",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}}
	}

	if len(versions) == 0 {
		return results
	}

	sort.Strings(versions)
	latestVersion := strings.Trim(versions[len(versions)-1], "/")

	results = append(results, TestResult{
		TestName:  "IS-08 Audio Channel Mapping Discovery",
		Target:    target,
		Passed:    true,
		Status:    "pass",
		Message:   fmt.Sprintf("Discovered IS-08 API version: %s", latestVersion),
		Details: map[string]any{
			"versions":       versions,
			"latest_version": latestVersion,
		},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	return results
}

// testIS07Discovery tests IS-07 Events API discovery
func (ts *TestSuite) testIS07Discovery(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}
	start := time.Now()

	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/events/", target.BaseURL))
	if err != nil {
		return []TestResult{{
			TestName:  "IS-07 Events API Discovery",
			Target:    target,
			Passed:    true,
			Status:    "skip",
			Message:   "IS-07 Events API not available (optional)",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}}
	}

	if len(versions) == 0 {
		return results
	}

	sort.Strings(versions)
	latestVersion := strings.Trim(versions[len(versions)-1], "/")

	results = append(results, TestResult{
		TestName:  "IS-07 Events API Discovery",
		Target:    target,
		Passed:    true,
		Status:    "pass",
		Message:   fmt.Sprintf("Discovered IS-07 API version: %s", latestVersion),
		Details: map[string]any{
			"versions":       versions,
			"latest_version": latestVersion,
		},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	return results
}

// testIS04QueryDiscovery tests IS-04 Query API discovery
func (ts *TestSuite) testIS04QueryDiscovery(ctx context.Context, target TestTarget) TestResult {
	start := time.Now()

	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/query/", target.BaseURL))
	if err != nil {
		return TestResult{
			TestName:  "IS-04 Query API Discovery",
			Target:    target,
			Passed:    false,
			Status:    "fail",
			Message:   fmt.Sprintf("Failed to discover IS-04 Query API versions: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	if len(versions) == 0 {
		return TestResult{
			TestName:  "IS-04 Query API Discovery",
			Target:    target,
			Passed:    false,
			Status:    "fail",
			Message:   "No IS-04 Query API versions found",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	sort.Strings(versions)
	latestVersion := strings.Trim(versions[len(versions)-1], "/")

	return TestResult{
		TestName:  "IS-04 Query API Discovery",
		Target:    target,
		Passed:    true,
		Status:    "pass",
		Message:   fmt.Sprintf("Discovered IS-04 Query API version: %s", latestVersion),
		Details: map[string]any{
			"versions":       versions,
			"latest_version": latestVersion,
		},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}

// testIS04QueryResources tests IS-04 Query API resource queries
func (ts *TestSuite) testIS04QueryResources(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}

	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/query/", target.BaseURL))
	if err != nil || len(versions) == 0 {
		return results
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")

	// Test nodes query
	start := time.Now()
	nodes, err := ts.fetchJSONArray(ctx, fmt.Sprintf("%s/x-nmos/query/%s/nodes", target.BaseURL, version))
	results = append(results, TestResult{
		TestName:  "IS-04 Query: Nodes",
		Target:    target,
		Passed:    err == nil,
		Status:    func() string { if err == nil { return "pass" }; return "fail" }(),
		Message:   fmt.Sprintf("Query API nodes endpoint: %d nodes", len(nodes)),
		Details:   map[string]any{"count": len(nodes)},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Test devices query
	start = time.Now()
	devices, err := ts.fetchJSONArray(ctx, fmt.Sprintf("%s/x-nmos/query/%s/devices", target.BaseURL, version))
	results = append(results, TestResult{
		TestName:  "IS-04 Query: Devices",
		Target:    target,
		Passed:    err == nil,
		Status:    func() string { if err == nil { return "pass" }; return "fail" }(),
		Message:   fmt.Sprintf("Query API devices endpoint: %d devices", len(devices)),
		Details:   map[string]any{"count": len(devices)},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	return results
}

// testRegistryHealth tests registry health and consistency
func (ts *TestSuite) testRegistryHealth(ctx context.Context, target TestTarget) []TestResult {
	results := []TestResult{}

	versions, err := ts.fetchJSONList(ctx, fmt.Sprintf("%s/x-nmos/query/", target.BaseURL))
	if err != nil || len(versions) == 0 {
		return results
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")

	// Test health endpoint (if available)
	healthStart := time.Now()
	healthURL := fmt.Sprintf("%s/x-nmos/query/%s/health", target.BaseURL, version)
	resp, err := ts.client.Get(healthURL)
	if err == nil {
		resp.Body.Close()
		results = append(results, TestResult{
			TestName:  "Registry Health Endpoint",
			Target:    target,
			Passed:    resp.StatusCode == 200,
			Status:    func() string { if resp.StatusCode == 200 { return "pass" }; return "warning" }(),
			Message:   fmt.Sprintf("Health endpoint status: %d", resp.StatusCode),
			Details:   map[string]any{"status_code": resp.StatusCode},
			Duration:  time.Since(healthStart),
			Timestamp: time.Now(),
		})
	}

	return results
}

// Helper functions

func (ts *TestSuite) fetchJSONList(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var items []string
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func (ts *TestSuite) fetchJSONArray(ctx context.Context, url string) ([]map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var items []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}
