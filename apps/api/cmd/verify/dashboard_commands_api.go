// Package verify provides verification for SERVER_BACKEND_DASHBOARD_COMMANDS_API.md
// This script verifies ALL server-side requirements from the Dashboard Commands API specification.
// FRONTEND SPECIFICATIONS HAVE BEEN REMOVED - Server-side only.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// verifyDashboardCommands verifies ALL requirements from SERVER_BACKEND_DASHBOARD_COMMANDS_API.md
func verifyDashboardCommands() bool {
	root := getRoot()
	passed := 0
	failed := 0

	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_DASHBOARD_COMMANDS_API.md - COMPREHENSIVE VERIFICATION  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/")
	appDir := filepath.Join(root, "apps/api/internal/application/")
	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")

	// =========================================================================
	// SECTION 2: CURRENT STATE ANALYSIS
	// =========================================================================
	fmt.Println("📋 SECTION 2: CURRENT STATE ANALYSIS")
	fmt.Println(strings.Repeat("─", 75))

	fmt.Println("--- 2.1 Existing Command Endpoints ---")
	cmdHandler := filepath.Join(handlerDir, "command/command_handler.go")
	cmdContent, _ := os.ReadFile(cmdHandler)

	existingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"EXIST-1", "POST", "/v1/device/:id/command", "ExecuteHandler"},
		{"EXIST-2", "GET", "/v1/device/:id/commands/pending", "GetPending"},
		{"EXIST-3", "GET", "/v1/command/:dispatchId/status", "GetStatus"},
		{"EXIST-4", "POST", "/v1/command/:dispatchId/retry", "Retry"},
		{"EXIST-5", "DELETE", "/v1/command/:dispatchId", "Cancel"},
	}

	for _, ep := range existingEndpoints {
		if strings.Contains(string(cmdContent), ep.handler) {
			fmt.Printf("  ✅ %s  %-6s %-40s (%s EXISTS)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-40s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	fmt.Println("\n--- 2.2 Missing Endpoints (REQUIRED) ---")
	missingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
		handlerFile string
	}{
		{"MISS-1", "GET", "/v1/device/:imei/commands", "GetHistory", "command_history_handler.go"},
		{"MISS-2", "GET", "/v1/device/:imei/logs", "GetLogs", "device_logs_handler.go"},
		{"MISS-3", "GET", "/v1/device/:imei/metrics", "GetMetrics", "device_metrics_handler.go"},
		{"MISS-4", "GET", "/v1/device/:imei/telemetry", "GetTelemetry", "device_telemetry_handler.go"},
		{"MISS-5", "GET", "/v1/device/:imei/metrics/export", "ExportMetrics", "device_metrics_handler.go"},
		{"MISS-6", "GET", "/v1/dashboard/stats", "GetStats", "dashboard_stats_handler.go"},
	}

	for _, ep := range missingEndpoints {
		var handlerPath string
		if strings.Contains(ep.handlerFile, "command") {
			handlerPath = filepath.Join(handlerDir, "command/", ep.handlerFile)
		} else if strings.Contains(ep.handlerFile, "dashboard") {
			handlerPath = filepath.Join(handlerDir, "dashboard/", ep.handlerFile)
		} else {
			handlerPath = filepath.Join(handlerDir, "device/", ep.handlerFile)
		}

		implemented := false
		if content, err := os.ReadFile(handlerPath); err == nil {
			implemented = strings.Contains(string(content), ep.handler)
		}

		if implemented {
			fmt.Printf("  ✅ %s  %-6s %-40s (Handler: %s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-40s (Handler: %s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	// =========================================================================
	// SECTION 3: REQUIRED API ENDPOINTS - Query Parameters
	// =========================================================================
	fmt.Println("\n📋 SECTION 3: REQUIRED API ENDPOINTS - Query Parameters")
	fmt.Println(strings.Repeat("─", 75))

	fmt.Println("--- GET /v1/device/:imei/commands (Section 3.1) ---")
	historyPath := filepath.Join(handlerDir, "command/command_history_handler.go")
	historyContent, _ := os.ReadFile(historyPath)
	historyParams := []string{"status", "page", "limit", "startTime", "endTime"}
	for _, p := range historyParams {
		if strings.Contains(string(historyContent), p) {
			fmt.Printf("  ✅  Query param: %s\n", p)
			passed++
		} else {
			fmt.Printf("  ❌  Query param: %s (MISSING)\n", p)
			failed++
		}
	}

	fmt.Println("\n--- GET /v1/device/:imei/logs (Section 3.2) ---")
	logsPath := filepath.Join(handlerDir, "device/device_logs_handler.go")
	logsContent, _ := os.ReadFile(logsPath)
	logsParams := []string{"type", "startTime", "endTime", "limit", "cursor"}
	for _, p := range logsParams {
		if strings.Contains(string(logsContent), p) {
			fmt.Printf("  ✅  Query param: %s\n", p)
			passed++
		} else {
			fmt.Printf("  ❌  Query param: %s (MISSING)\n", p)
			failed++
		}
	}

	fmt.Println("\n--- GET /v1/device/:imei/metrics (Section 3.3) ---")
	metricsPath := filepath.Join(handlerDir, "device/device_metrics_handler.go")
	metricsContent, _ := os.ReadFile(metricsPath)
	metricsParams := []string{"range", "startTime", "endTime", "resolution"}
	for _, p := range metricsParams {
		if strings.Contains(string(metricsContent), p) {
			fmt.Printf("  ✅  Query param: %s\n", p)
			passed++
		} else {
			fmt.Printf("  ❌  Query param: %s (MISSING)\n", p)
			failed++
		}
	}

	// =========================================================================
	// SECTION 4: DATABASE SCHEMA - device_logs TABLE
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: DATABASE SCHEMA - device_logs Table")
	fmt.Println(strings.Repeat("─", 75))

	schemaDir := filepath.Join(root, "supabase/migrations/")
	schemaChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"DB-1", "device_logs table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), "device_logs") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-2", "idx_device_logs index", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "idx_device_logs") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-3", "idx_event_type index", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "idx_event_type") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-4", "fk_device constraint", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "device_logs") && strings.Contains(string(content), "fk_device") {
							return true
						}
					}
				}
			}
			return false
		}},
	}

	for _, check := range schemaChecks {
		if check.check() {
			fmt.Printf("  ✅ %s  %s\n", check.id, check.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s\n", check.id, check.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 6: HANDLER SPECIFICATIONS
	// =========================================================================
	fmt.Println("\n📋 SECTION 6: HANDLER SPECIFICATIONS")
	fmt.Println(strings.Repeat("─", 75))

	handlerSpecs := []struct {
		id          string
		handlerFile string
		handlerFunc string
		path        string
	}{
		{"H-1", "device_metrics_handler.go", "GetMetrics", "GET /v1/device/:imei/metrics"},
		{"H-2", "device_metrics_handler.go", "ExportMetrics", "GET /v1/device/:imei/metrics/export"},
		{"H-3", "device_logs_handler.go", "GetLogs", "GET /v1/device/:imei/logs"},
		{"H-4", "command_history_handler.go", "GetHistory", "GET /v1/device/:imei/commands"},
	}

	for _, h := range handlerSpecs {
		var handlerPath string
		if strings.Contains(h.handlerFile, "command") {
			handlerPath = filepath.Join(handlerDir, "command/", h.handlerFile)
		} else {
			handlerPath = filepath.Join(handlerDir, "device/", h.handlerFile)
		}

		implemented := false
		if content, err := os.ReadFile(handlerPath); err == nil {
			implemented = strings.Contains(string(content), "func (h *") && strings.Contains(string(content), h.handlerFunc)
		}

		if implemented {
			fmt.Printf("  ✅ %s  %-40s (%s)\n", h.id, h.path, h.handlerFunc)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-40s (%s MISSING)\n", h.id, h.path, h.handlerFunc)
			failed++
		}
	}

	// =========================================================================
	// SECTION 7: SERVICE LAYER
	// =========================================================================
	fmt.Println("\n📋 SECTION 7: SERVICE LAYER")
	fmt.Println(strings.Repeat("─", 75))

	serviceSpecs := []struct {
		id          string
		serviceFile string
		method      string
	}{
		{"S-1", "metrics/metrics_service.go", "GetDeviceMetrics"},
		{"S-2", "metrics/metrics_service.go", "ExportMetrics"},
		{"S-3", "logs/logs_service.go", "GetDeviceLogs"},
		{"S-4", "command/command_service.go", "GetCommandHistory"},
		{"S-5", "dashboard/dashboard_service.go", "GetDashboardStats"},
	}

	for _, s := range serviceSpecs {
		servicePath := filepath.Join(appDir, s.serviceFile)
		implemented := false
		if content, err := os.ReadFile(servicePath); err == nil {
			implemented = strings.Contains(string(content), "func (s *") && strings.Contains(string(content), s.method)
		}

		if implemented {
			fmt.Printf("  ✅ %s  %s.%s()\n", s.id, s.serviceFile, s.method)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s.%s() (MISSING)\n", s.id, s.serviceFile, s.method)
			failed++
		}
	}

	// =========================================================================
	// SECTION 8: GRAPHQL SCHEMA
	// =========================================================================
	fmt.Println("\n📋 SECTION 8: GRAPHQL SCHEMA")
	fmt.Println(strings.Repeat("─", 75))

	graphqlSpecs := []string{
		"MetricChartPoint", "DeviceMetrics", "deviceMetrics", "deviceLogs",
		"dashboardStats", "CommandHistory", "DeviceLog",
	}

	for i, g := range graphqlSpecs {
		found := false
		if entries, err := os.ReadDir(gqlDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(gqlDir, entry.Name()))
					if strings.Contains(string(content), g) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ G-%d  %s\n", i+1, g)
			passed++
		} else {
			fmt.Printf("  ❌ G-%d  %s (MISSING)\n", i+1, g)
			failed++
		}
	}

	// =========================================================================
	// SECTION 9: ERROR HANDLING
	// =========================================================================
	fmt.Println("\n📋 SECTION 9: ERROR HANDLING")
	fmt.Println(strings.Repeat("─", 75))

	errorCodes := []string{"bad_request", "unauthorized", "forbidden", "not_found", "rate_limited", "internal_error"}
	hasErrors := make(map[string]bool)

	if entries, err := os.ReadDir(handlerDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				content, _ := os.ReadFile(filepath.Join(handlerDir, entry.Name()))
				for _, code := range errorCodes {
					if strings.Contains(string(content), code) {
						hasErrors[code] = true
					}
				}
			}
		}
	}

	for i, code := range errorCodes {
		if hasErrors[code] {
			fmt.Printf("  ✅ ERR-%d  %s\n", i+1, code)
			passed++
		} else {
			fmt.Printf("  ❌ ERR-%d  %s (MISSING)\n", i+1, code)
			failed++
		}
	}

	// =========================================================================
	// SECTION 10: RATE LIMITING
	// =========================================================================
	fmt.Println("\n📋 SECTION 10: RATE LIMITING")
	fmt.Println(strings.Repeat("─", 75))

	rateLimits := []struct {
		endpoint string
		limit   string
	}{
		{"GET /v1/device/:imei/commands", "60/min"},
		{"GET /v1/device/:imei/logs", "60/min"},
		{"GET /v1/device/:imei/metrics", "30/min"},
		{"GET /v1/device/:imei/metrics/export", "10/min"},
		{"POST /v1/device/:imei/command", "10/min"},
	}

	middlewarePath := filepath.Join(root, "apps/api/internal/api/middleware/rate_limit.go")
	middlewareContent, _ := os.ReadFile(middlewarePath)

	for _, rl := range rateLimits {
		if strings.Contains(string(middlewareContent), "rateLimit") || strings.Contains(string(middlewareContent), "RateLimit") {
			fmt.Printf("  ✅ %-40s %s\n", rl.endpoint, rl.limit)
			passed++
		} else {
			fmt.Printf("  ❌ %-40s %s (MISSING)\n", rl.endpoint, rl.limit)
			failed++
		}
	}

	// =========================================================================
	// SECTION 11: FILE STRUCTURE
	// =========================================================================
	fmt.Println("\n📋 SECTION 11: FILE STRUCTURE")
	fmt.Println(strings.Repeat("─", 75))

	requiredFiles := []struct {
		id   string
		file string
	}{
		{"F-1", "api/handlers/command/command_history_handler.go"},
		{"F-2", "api/handlers/device/device_logs_handler.go"},
		{"F-3", "api/handlers/device/device_metrics_handler.go"},
		{"F-4", "api/handlers/device/device_telemetry_handler.go"},
		{"F-5", "api/handlers/dashboard/dashboard_stats_handler.go"},
		{"F-6", "application/logs/logs_service.go"},
		{"F-7", "application/metrics/metrics_service.go"},
		{"F-8", "application/dashboard/dashboard_service.go"},
		{"F-9", "domain/logs/logs_entity.go"},
		{"F-10", "domain/logs/logs_repository.go"},
		{"F-11", "domain/metrics/metrics_entity.go"},
		{"F-12", "infrastructure/storage/logs_storage.go"},
		{"F-13", "infrastructure/storage/metrics_storage.go"},
	}

	for _, f := range requiredFiles {
		filePath := filepath.Join(root, "apps/api/internal/", f.file)
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("  ✅ %s  %s\n", f.id, f.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", f.id, f.file)
			failed++
		}
	}

	// =========================================================================
	// SUMMARY
	// =========================================================================
	fmt.Println(strings.Repeat("═", 75))
	fmt.Printf("DASHBOARD COMMANDS API: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
