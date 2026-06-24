package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TestReport holds test results.
type TestReport struct {
	Results        []EndpointResult
	TotalEndpoints int
	Reachable      int
	Unreachable    int
	ErrorHandling  int
	DBTests        int
	DBPassed       int
	DBFailed       int
	mu             sync.Mutex
}

type EndpointResult struct {
	Method       string
	Path         string
	DBResult     string
	Notes        string
	StatusCode   int
	ResponseTime time.Duration
	Reachable    bool
	ErrorTested  bool
	DBTested     bool
}

const (
	BaseURL         = "http://localhost:3000"
	GraphQLEndpoint = "http://localhost:3000/graphql"
	ReportFile      = "endpoint_test_report.md"
	DBPath          = "./data/vyzorix.db"
)

// HTTP Client setup.
var client = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func prettyJSON(data []byte) string {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return string(data)
	}

	out, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return string(data)
	}

	return string(out)
}

func (r *TestReport) addResult(res ...EndpointResult) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, ep := range res {
		r.Results = append(r.Results, ep)
		r.TotalEndpoints++

		if ep.Reachable {
			r.Reachable++
		} else {
			r.Unreachable++
		}

		if ep.DBTested {
			r.DBTests++

			switch ep.DBResult {
			case "PASS":
				r.DBPassed++
			case "FAIL":
				r.DBFailed++
			}
		}

		if ep.ErrorTested {
			r.ErrorHandling++
		}
	}
}

func (r *TestReport) generateReport() string {
	var sb strings.Builder

	sb.WriteString("# Comprehensive API Test Report\n\n")
	fmt.Fprintf(&sb, "**Generated:** %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&sb, "**Base URL:** %s\n\n", BaseURL)

	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Value |\n|--------|-------|\n")
	fmt.Fprintf(&sb, "| Total Endpoints | %d |\n", r.TotalEndpoints)
	fmt.Fprintf(&sb, "| Reachable | %d |\n", r.Reachable)
	fmt.Fprintf(&sb, "| Unreachable | %d |\n", r.Unreachable)
	fmt.Fprintf(&sb, "| Error Cases Tested | %d |\n", r.ErrorHandling)
	fmt.Fprintf(&sb, "| DB Tests | %d |\n", r.DBTests)
	fmt.Fprintf(&sb, "| DB Passed | %d |\n", r.DBPassed)
	fmt.Fprintf(&sb, "| DB Failed | %d |\n", r.DBFailed)

	// Group by category
	sb.WriteString("\n## Detailed Results\n\n")

	// Group results by path prefix
	categories := map[string][]EndpointResult{}

	for _, res := range r.Results {
		cat := categorizePath(res.Path)
		categories[cat] = append(categories[cat], res)
	}

	for cat, resList := range categories {
		fmt.Fprintf(&sb, "### %s Endpoints\n\n", cat)
		sb.WriteString("| Method | Path | Status | Time | DB | Notes |\n")
		sb.WriteString("|--------|------|--------|------|-----|-------|\n")

		for _, res := range resList {
			dbStatus := "—"

			if res.DBTested {
				switch res.DBResult {
				case "PASS":
					dbStatus = "✅"
				case "FAIL":
					dbStatus = "❌"
				default:
					dbStatus = "❓"
				}
			}

			reachable := "✅"
			if !res.Reachable {
				reachable = "❌"
			}

			notes := res.Notes
			if len(notes) > 100 {
				notes = notes[:100] + "..."
			}

			notes = strings.ReplaceAll(notes, "\n", " ")
			fmt.Fprintf(&sb, "| %s | `%s` | %s %d | %s | %s | %s |\n",
				res.Method, res.Path, reachable, res.StatusCode,
				res.ResponseTime.Round(time.Millisecond), dbStatus, notes)
		}

		sb.WriteString("\n")
	}

	// DB Summary
	sb.WriteString("## Database Operations Summary\n\n")

	for _, res := range r.Results {
		if res.DBTested {
			fmt.Fprintf(&sb, "- %s %s: %s\n", res.Method, res.Path, res.DBResult)
		}
	}

	return sb.String()
}

func testRequest(method, url string, body interface{}) EndpointResult {
	var reqBody io.Reader

	if body != nil {
		if s, ok := body.(string); ok {
			reqBody = strings.NewReader(s)
		} else {
			b, err := json.Marshal(body)
			if err != nil {
				return EndpointResult{
					Method:    method,
					Path:      url,
					Reachable: false,
					Notes:     fmt.Sprintf("JSON marshal error: %v", err),
				}
			}

			reqBody = bytes.NewReader(b)
		}
	}

	start := time.Now()

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return EndpointResult{
			Method:    method,
			Path:      url,
			Reachable: false,
			Notes:     fmt.Sprintf("Request error: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	duration := time.Since(start)

	result := EndpointResult{
		Method:       method,
		Path:         url,
		ResponseTime: duration,
		Reachable:    err == nil,
	}

	if err != nil {
		result.Notes = fmt.Sprintf("Connection error: %v", err)
		return result
	}

	defer func() { _ = resp.Body.Close() }()

	result.StatusCode = resp.StatusCode
	bodyBytes, _ := io.ReadAll(resp.Body)

	if len(bodyBytes) > 0 {
		result.Notes = prettyJSON(bodyBytes)
	}

	return result
}

func checkDBFile() bool {
	_, err := os.Stat(DBPath)
	return err == nil
}

// ==================== TEST CATEGORIES ====================

func testHealthEndpoints() []EndpointResult {
	return []EndpointResult{
		testRequest("GET", BaseURL+"/health", nil),
		testRequest("GET", BaseURL+"/healthz", nil),
	}
}

func testVersionEndpoints() []EndpointResult {
	return []EndpointResult{
		testRequest("GET", BaseURL+"/api/v1/version", nil),
		testRequest("GET", BaseURL+"/api/v1/changelog", nil),
		testRequest("GET", BaseURL+"/api/v1/apk/test.apk", nil),
		testRequest("GET", BaseURL+"/bin/test.bin", nil),
	}
}

func testAuthEndpoints() []EndpointResult {
	results := make([]EndpointResult, 0, 30)
	base := "/v1/auth"

	// Login tests
	results = append(results, testRequest("POST", BaseURL+base+"/login",
		map[string]string{"email": "test@example.com", "password": "wrongpassword"}))
	results = append(results, testRequest("POST", BaseURL+base+"/login",
		map[string]string{"email": "invalid-email", "password": "pass"}))
	results = append(results, testRequest("POST", BaseURL+base+"/login",
		map[string]string{"email": "", "password": ""}))
	results = append(results, testRequest("POST", BaseURL+base+"/login", nil))

	// Register tests
	results = append(results, testRequest("POST", BaseURL+base+"/register",
		map[string]interface{}{"email": fmt.Sprintf("user-%d@test.com", time.Now().Unix()), "password": "ValidPass123!", "name": "Test User"}))
	results = append(results, testRequest("POST", BaseURL+base+"/register",
		map[string]string{"email": "not-an-email"}))
	results = append(results, testRequest("POST", BaseURL+base+"/register",
		map[string]string{"email": "short"}))

	// Password reset
	results = append(results, testRequest("POST", BaseURL+base+"/forgot-password",
		map[string]string{"email": "test@example.com"}))
	results = append(results, testRequest("POST", BaseURL+base+"/forgot-password",
		map[string]string{"email": "invalid"}))

	// Email verification
	results = append(results, testRequest("POST", BaseURL+base+"/verify-email",
		map[string]string{"token": "invalid-token"}))
	results = append(results, testRequest("POST", BaseURL+base+"/resend-verification",
		map[string]string{"email": "test@example.com"}))
	results = append(results, testRequest("POST", BaseURL+base+"/cancel-verification",
		map[string]string{"token": "invalid-token"}))
	results = append(results, testRequest("GET", BaseURL+base+"/poll-verification?email=test@example.com", nil))

	// MFA
	results = append(results, testRequest("GET", BaseURL+base+"/mfa/status", nil))
	results = append(results, testRequest("POST", BaseURL+base+"/mfa/enroll", nil))
	results = append(results, testRequest("POST", BaseURL+base+"/mfa/verify-setup",
		map[string]string{"code": "123456"}))
	results = append(results, testRequest("POST", BaseURL+base+"/mfa/enable",
		map[string]string{"code": "123456"}))
	results = append(results, testRequest("POST", BaseURL+base+"/mfa/disable",
		map[string]string{"code": "123456"}))
	results = append(results, testRequest("POST", BaseURL+base+"/mfa/verify-backup",
		map[string]string{"code": "12345678"}))
	results = append(results, testRequest("POST", BaseURL+base+"/mfa/regenerate-backup-codes", nil))

	// Logout
	results = append(results, testRequest("POST", BaseURL+base+"/logout", nil))

	// Me
	results = append(results, testRequest("GET", BaseURL+base+"/me", nil))
	results = append(results, testRequest("PATCH", BaseURL+base+"/me",
		map[string]string{"name": "Updated Name"}))
	results = append(results, testRequest("PATCH", BaseURL+base+"/me/settings",
		map[string]interface{}{"theme": "dark"}))

	// Lockout
	results = append(results, testRequest("GET", BaseURL+base+"/lockout/status", nil))

	// Admin operators
	results = append(results, testRequest("GET", BaseURL+base+"/admin/operators", nil))
	results = append(results, testRequest("POST", BaseURL+base+"/admin/operators",
		map[string]interface{}{"email": fmt.Sprintf("admin-%d@test.com", time.Now().Unix()), "password": "AdminPass123!", "name": "Test Admin", "role": "admin"}))
	results = append(results, testRequest("GET", BaseURL+base+"/admin/operators/1", nil))
	results = append(results, testRequest("PATCH", BaseURL+base+"/admin/operators/1",
		map[string]string{"name": "Updated Admin"}))
	results = append(results, testRequest("DELETE", BaseURL+base+"/admin/operators/1", nil))
	results = append(results, testRequest("POST", BaseURL+base+"/admin/lockout/unlock/1", nil))

	// Client credentials
	results = append(results, testRequest("POST", BaseURL+base+"/client-credentials",
		map[string]interface{}{"name": fmt.Sprintf("client-%d", time.Now().Unix())}))
	results = append(results, testRequest("GET", BaseURL+base+"/client-credentials", nil))

	// OAuth (just reachability)
	results = append(results, testRequest("GET", BaseURL+base+"/google", nil))
	results = append(results, testRequest("GET", BaseURL+base+"/github", nil))

	return results
}

func testDeviceEndpoints() []EndpointResult {
	results := make([]EndpointResult, 0, 10)
	deviceID := fmt.Sprintf("test-device-%d", time.Now().Unix())
	base := "/v1/device"

	// Register
	results = append(results, testRequest("POST", BaseURL+base+"/register",
		map[string]interface{}{"device_id": deviceID, "name": "Test Device", "platform": "android", "app_version": "1.0.0"}))
	results = append(results, testRequest("POST", BaseURL+base+"/register",
		map[string]string{"invalid": "data"}))
	results = append(results, testRequest("POST", BaseURL+base+"/register",
		map[string]string{"device_id": "", "name": ""}))

	// Status
	results = append(results, testRequest("GET", BaseURL+base+"/"+deviceID+"/status", nil))
	results = append(results, testRequest("GET", BaseURL+base+"/invalid-id/status", nil))

	// Get device
	results = append(results, testRequest("GET", BaseURL+base+"/"+deviceID, nil))

	// Update FCM token
	results = append(results, testRequest("PATCH", BaseURL+base+"/"+deviceID+"/fcm-token",
		map[string]string{"fcm_token": "new-fcm-token-123"}))

	// Delete
	results = append(results, testRequest("DELETE", BaseURL+base+"/"+deviceID, nil))
	results = append(results, testRequest("DELETE", BaseURL+base+"/nonexistent-id", nil))

	return results
}

func testCommandEndpoints() []EndpointResult {
	results := make([]EndpointResult, 0, 10)
	deviceID := fmt.Sprintf("cmd-test-device-%d", time.Now().Unix())
	base := "/v1"
	dispatchID := fmt.Sprintf("dispatch-%d", time.Now().Unix())

	// First register a device
	testRequest("POST", BaseURL+base+"/device/register",
		map[string]interface{}{"device_id": deviceID, "name": "Command Test", "platform": "android", "app_version": "1.0.0"})

	// Send command
	results = append(results, testRequest("POST", BaseURL+base+"/device/"+deviceID+"/command",
		map[string]interface{}{"command": "FORCE_SPEAKER", "args": map[string]int{"volume": 80}, "priority": 5}))
	results = append(results, testRequest("POST", BaseURL+base+"/device/"+deviceID+"/command",
		map[string]string{"invalid": "data"}))
	results = append(results, testRequest("POST", BaseURL+base+"/device/"+deviceID+"/command",
		map[string]string{"command": "", "args": "{}"}))

	// Get pending commands
	results = append(results, testRequest("GET", BaseURL+base+"/device/"+deviceID+"/commands/pending", nil))

	// Command status
	results = append(results, testRequest("GET", BaseURL+base+"/command/"+dispatchID+"/status", nil))

	// Retry command
	results = append(results, testRequest("POST", BaseURL+base+"/command/"+dispatchID+"/retry", nil))

	// Cancel command
	results = append(results, testRequest("DELETE", BaseURL+base+"/command/"+dispatchID, nil))

	return results
}

func testDashboardEndpoints() []EndpointResult {
	return []EndpointResult{
		testRequest("GET", BaseURL+"/v1/dashboard/devices", nil),
		testRequest("GET", BaseURL+"/v1/dashboard/devices?page=1&limit=10", nil),
		testRequest("GET", BaseURL+"/v1/dashboard/devices?page=999&limit=10", nil),
		testRequest("GET", BaseURL+"/v1/dashboard/devices/operator", nil),
		testRequest("GET", BaseURL+"/v1/device/count", nil),
	}
}

func testAdminEndpoints() []EndpointResult {
	results := make([]EndpointResult, 0, 5)
	clientID := fmt.Sprintf("admin-client-%d", time.Now().Unix())

	results = append(results, testRequest("GET", BaseURL+"/v1/admin/clients", nil))
	results = append(results, testRequest("GET", BaseURL+"/v1/admin/clients/"+clientID, nil))
	results = append(results, testRequest("PATCH", BaseURL+"/v1/admin/clients/"+clientID,
		map[string]interface{}{"name": "Updated Client", "enabled": true}))
	results = append(results, testRequest("POST", BaseURL+"/v1/admin/clients/"+clientID+"/rotate-key", nil))
	results = append(results, testRequest("DELETE", BaseURL+"/v1/admin/clients/"+clientID, nil))

	return results
}

func testWebSocket() []EndpointResult {
	return []EndpointResult{
		testRequest("GET", BaseURL+"/v1/device/test-device/stream", nil),
	}
}

func testErrorHandling() []EndpointResult {
	results := make([]EndpointResult, 0, 3)

	results = append(results, testRequest("GET", BaseURL+"/nonexistent", nil))
	results = append(results, testRequest("POST", BaseURL+"/v1/auth/login", `{"email": "not-json`))
	results = append(results, testRequest("GET", BaseURL+"/v1/device/invalid@id/status", nil))

	return results
}

func testMalformedData() []EndpointResult {
	results := make([]EndpointResult, 0, 6)

	tests := []struct {
		path     string
		body     string
		testType string
	}{
		{"/v1/auth/login", `{"email": "not-json`, "Invalid JSON"},
		{"/v1/auth/login", `{}`, "Empty body"},
		{"/v1/auth/login", `{"email":"a@b.com","password":"x"}`, "Short password"},
		{"/v1/auth/register", `{"email":"a@b.com","password":"Valid123!"}`, "Missing name"},
		{"/v1/device/register", `{"device_id":"","name":""}`, "Empty fields"},
		{"/v1/device/test/command", `{"command":"","args":{}}`, "Empty command"},
	}

	for _, test := range tests {
		res := testRequest("POST", BaseURL+test.path, test.body)
		res.ErrorTested = true
		res.Notes = fmt.Sprintf("Type: %s | Response: %s", test.testType, res.Notes)
		results = append(results, res)
	}

	return results
}

func testDatabaseOperations() []EndpointResult {
	results := make([]EndpointResult, 0, 8)

	if !checkDBFile() {
		results = append(results, EndpointResult{
			Path:      "database_check",
			Reachable: false,
			Notes:     "Database file not found at " + DBPath,
		})

		return results
	}

	// Create operator
	email := fmt.Sprintf("dbtest-%d@example.com", time.Now().Unix())
	res := testRequest("POST", BaseURL+"/v1/auth/admin/operators",
		map[string]interface{}{"email": email, "password": "TestPass123!", "name": "DB Test", "role": "user"})
	res.DBTested = true

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		res.DBResult = "PASS"
	} else {
		res.DBResult = "FAIL"
	}

	results = append(results, res)

	// List operators (verify read)
	res = testRequest("GET", BaseURL+"/v1/auth/admin/operators", nil)
	res.DBTested = true

	if res.StatusCode == 200 {
		res.DBResult = "PASS"
	} else {
		res.DBResult = "FAIL"
	}

	results = append(results, res)

	// Register device
	deviceID := fmt.Sprintf("dbtest-device-%d", time.Now().Unix())
	res = testRequest("POST", BaseURL+"/v1/device/register",
		map[string]interface{}{"device_id": deviceID, "name": "DB Test Device", "platform": "android", "app_version": "1.0.0"})
	res.DBTested = true

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		res.DBResult = "PASS"
	} else {
		res.DBResult = "FAIL"
	}

	results = append(results, res)

	// List devices
	res = testRequest("GET", BaseURL+"/v1/dashboard/devices", nil)
	res.DBTested = true

	if res.StatusCode == 200 {
		res.DBResult = "PASS"
	} else {
		res.DBResult = "FAIL"
	}

	results = append(results, res)

	// Get device count
	res = testRequest("GET", BaseURL+"/v1/device/count", nil)
	res.DBTested = true

	if res.StatusCode == 200 {
		res.DBResult = "PASS"
	} else {
		res.DBResult = "FAIL"
	}

	results = append(results, res)

	// Send command
	res = testRequest("POST", BaseURL+"/v1/device/"+deviceID+"/command",
		map[string]interface{}{"command": "TEST_CMD", "args": map[string]bool{"test": true}, "priority": 5})
	res.DBTested = true

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		res.DBResult = "PASS"
	} else {
		res.DBResult = "FAIL"
	}

	results = append(results, res)

	// Get pending commands
	res = testRequest("GET", BaseURL+"/v1/device/"+deviceID+"/commands/pending", nil)
	res.DBTested = true

	if res.StatusCode == 200 {
		res.DBResult = "PASS"
	} else {
		res.DBResult = "FAIL"
	}

	results = append(results, res)

	return results
}

// ==================== MAIN ====================

func RunAllTests() error {
	fmt.Println("🔍 Comprehensive API Testing Suite")
	fmt.Println("====================================")

	report := &TestReport{}

	fmt.Print("Testing health endpoints... ")

	results := testHealthEndpoints()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing version endpoints... ")

	results = testVersionEndpoints()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing auth endpoints... ")

	results = testAuthEndpoints()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing device endpoints... ")

	results = testDeviceEndpoints()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing command endpoints... ")

	results = testCommandEndpoints()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing dashboard endpoints... ")

	results = testDashboardEndpoints()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing admin endpoints... ")

	results = testAdminEndpoints()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing WebSocket... ")

	results = testWebSocket()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing error handling... ")

	results = testErrorHandling()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing malformed data... ")

	results = testMalformedData()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	fmt.Print("Testing database operations... ")

	results = testDatabaseOperations()
	report.addResult(results...)
	fmt.Printf("Done (%d)\n", len(results))

	// Generate report
	_ = os.MkdirAll("./data", 0755)
	reportContent := report.generateReport()
	reportPath := filepath.Join(".", ReportFile)
	_ = os.WriteFile(reportPath, []byte(reportContent), 0644)

	fmt.Println()
	fmt.Println("====================================")
	fmt.Println("📊 TEST SUMMARY")
	fmt.Println("====================================")
	fmt.Printf("Total Endpoints: %d\n", report.TotalEndpoints)
	fmt.Printf("Reachable: %d ✅\n", report.Reachable)
	fmt.Printf("Unreachable: %d ❌\n", report.Unreachable)
	fmt.Printf("Error Cases Tested: %d\n", report.ErrorHandling)
	fmt.Printf("DB Tests: %d (%d passed, %d failed)\n", report.DBTests, report.DBPassed, report.DBFailed)
	fmt.Printf("\n📄 Full report: %s\n", reportPath)

	return nil
}

func categorizePath(path string) string {
	switch {
	case strings.HasPrefix(path, "/v1/auth"):
		return "Auth"
	case strings.HasPrefix(path, "/v1/device"):
		return "Device"
	case strings.HasPrefix(path, "/v1/dashboard"):
		return "Dashboard"
	case strings.HasPrefix(path, "/v1/admin"):
		return "Admin"
	case strings.HasPrefix(path, "/v1/command"):
		return "Command"
	case strings.HasPrefix(path, "/api/v1"):
		return "API"
	case strings.HasPrefix(path, "/health"):
		return "Health"
	default:
		return "Other"
	}
}
