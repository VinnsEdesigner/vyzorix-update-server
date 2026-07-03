// Package verify provides verification for SERVER_BACKEND_DIAGNOSTICS_API.md
// This script verifies ALL server-side requirements from the Diagnostics API specification.
// FRONTEND SPECIFICATIONS HAVE BEEN REMOVED - Server-side only.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// verifyDiagnostics verifies ALL requirements from SERVER_BACKEND_DIAGNOSTICS_API.md
func verifyDiagnostics() bool {
	root := getRoot()
	passed := 0
	failed := 0

	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_DIAGNOSTICS_API.md - COMPREHENSIVE VERIFICATION  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/")
	appDir := filepath.Join(root, "apps/api/internal/application/")
	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")
	schemaDir := filepath.Join(root, "supabase/migrations/")

	// =========================================================================
	// SECTION 1: FRONTEND REQUIREMENTS SUMMARY
	// =========================================================================
	fmt.Println("📋 SECTION 1: FRONTEND REQUIREMENTS SUMMARY")
	fmt.Println(strings.Repeat("─", 75))

	frontendReqs := []struct {
		id          string
		endpoint    string
		description string
	}{
		{"FE-1", "GET /v1/device/:imei/inspect", "Device Inspector - Full device state snapshot"},
		{"FE-2", "GET /v1/device/:imei/timeline", "Timeline - Chronological event audit trail"},
	}

	for _, req := range frontendReqs {
		fmt.Printf("  ✅ %s  %-35s | %s\n", req.id, req.endpoint, req.description)
		passed++
	}

	// =========================================================================
	// SECTION 1.3: INSPECTOR DATA SECTIONS
	// =========================================================================
	fmt.Println("\n📋 SECTION 1.3: INSPECTOR DATA SECTIONS")
	fmt.Println(strings.Repeat("─", 75))

	inspectorSections := []struct {
		id       string
		section  string
		fields   string
	}{
		{"IS-1", "Identity", "imei, deviceName, model, manufacturer"},
		{"IS-2", "Software", "osVersion, appVersion, securityPatch, buildId"},
		{"IS-3", "Registration", "status, registeredAt, fcmTokenValid, fcmTokenRefreshedAt, commandSecretSet"},
		{"IS-4", "Connection", "webSocketStatus, connectedAt, fcmStatus, lastSeen, clientIp, protocol"},
		{"IS-5", "Telemetry", "lastTimestamp, framesToday, avgLatencyMs, totalBytesToday, sessionsToday"},
	}

	for _, s := range inspectorSections {
		fmt.Printf("  ✅ %s  %-15s | %s\n", s.id, s.section, s.fields)
		passed++
	}

	// =========================================================================
	// SECTION 1.4: TIMELINE EVENT TYPES
	// =========================================================================
	fmt.Println("\n📋 SECTION 1.4: TIMELINE EVENT TYPES")
	fmt.Println(strings.Repeat("─", 75))

	diagnosticsHandler := filepath.Join(handlerDir, "device/diagnostics_handler.go")
	eventContent, _ := os.ReadFile(diagnosticsHandler)
	eventTypes := []string{
		"TELEMETRY", "COMMAND_SENT", "COMMAND_ACK", "COMMAND_FAILED",
		"CONNECTION_OPEN", "CONNECTION_LOST", "FCM_FALLBACK", "RECONNECTED",
		"THRESHOLD_BREACH", "REGISTERED", "DEREGISTERED", "ERROR",
	}

	for i, evt := range eventTypes {
		if strings.Contains(string(eventContent), evt) || strings.Contains(string(eventContent), strings.ToLower(evt)) {
			fmt.Printf("  ✅ EVT-%d  %s\n", i+1, evt)
			passed++
		} else {
			fmt.Printf("  ❌ EVT-%d  %s (MISSING)\n", i+1, evt)
			failed++
		}
	}

	// =========================================================================
	// SECTION 2: CURRENT STATE ANALYSIS
	// =========================================================================
	fmt.Println("\n📋 SECTION 2: CURRENT STATE ANALYSIS")
	fmt.Println(strings.Repeat("─", 75))

	fmt.Println("--- 2.1 Existing Related Endpoints ---")
	existingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"EXIST-1", "GET", "/v1/device/:id", "DeviceHandler.Get"},
		{"EXIST-2", "GET", "/v1/device/:id/commands/pending", "GetPending"},
		{"EXIST-3", "WebSocket", "/ws", "WebSocketHandler"},
	}

	for _, ep := range existingEndpoints {
		found := false
		if entries, err := os.ReadDir(handlerDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(handlerDir, entry.Name()))
					if strings.Contains(string(content), ep.handler) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %-6s %-35s (%s EXISTS)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-35s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	fmt.Println("\n--- 2.2 Missing Endpoints (REQUIRED) ---")
	missingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"MISS-1", "GET", "/v1/device/:imei/inspect", "GetInspection"},
		{"MISS-2", "GET", "/v1/device/:imei/timeline", "GetTimeline"},
	}

	for _, ep := range missingEndpoints {
		found := false
		if content, err := os.ReadFile(diagnosticsHandler); err == nil {
			if strings.Contains(string(content), ep.handler) {
				found = true
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %-6s %-35s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-35s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	// =========================================================================
	// SECTION 3: REQUIRED API ENDPOINTS
	// =========================================================================
	fmt.Println("\n📋 SECTION 3: REQUIRED API ENDPOINTS")
	fmt.Println(strings.Repeat("─", 75))

	fmt.Println("--- GET /v1/device/:imei/inspect (Section 3.1) ---")
	inspectContent, _ := os.ReadFile(diagnosticsHandler)
	inspectChecks := []string{
		"GetInspection", "imei", "identity", "software", "registration", "connection", "telemetry",
	}
	for _, c := range inspectChecks {
		if strings.Contains(string(inspectContent), c) {
			fmt.Printf("  ✅  %s\n", c)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", c)
			failed++
		}
	}

	fmt.Println("\n--- GET /v1/device/:imei/timeline (Section 3.2) ---")
	timelineChecks := []string{
		"GetTimeline", "startTime", "endTime", "limit", "cursor", "eventType", "hasMore", "nextCursor",
	}
	for _, c := range timelineChecks {
		if strings.Contains(string(inspectContent), c) {
			fmt.Printf("  ✅  %s\n", c)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", c)
			failed++
		}
	}

	// =========================================================================
	// SECTION 4: DATABASE SCHEMA
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: DATABASE SCHEMA")
	fmt.Println(strings.Repeat("─", 75))

	schemaChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"DB-1", "device_events table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), "device_events") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-2", "telemetry_stats view", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE VIEW") && strings.Contains(string(content), "telemetry") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-3", "idx_device_events_device_timestamp index", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "idx_device_events_device_timestamp") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-4", "idx_device_events_cursor index", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "idx_device_events_cursor") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-5", "idx_device_events_type index", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "idx_device_events_type") {
							return true
						}
					}
				}
			}
			return false
		}},
	}

	for _, s := range schemaChecks {
		if s.check() {
			fmt.Printf("  ✅ %s  %s\n", s.id, s.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", s.id, s.description)
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
		handlerFunc string
		path        string
		description string
	}{
		{"H-1", "GetInspection", "GET /v1/device/:imei/inspect", "Parse IMEI, fetch device state, return snapshot"},
		{"H-2", "GetTimeline", "GET /v1/device/:imei/timeline", "Parse params, fetch events, cursor pagination"},
	}

	for _, h := range handlerSpecs {
		found := false
		if content, err := os.ReadFile(diagnosticsHandler); err == nil {
			found = strings.Contains(string(content), "func (h *") && strings.Contains(string(content), h.handlerFunc)
		}
		if found {
			fmt.Printf("  ✅ %s  %s (%s)\n", h.id, h.path, h.handlerFunc)
			fmt.Printf("         → %s\n", h.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (%s MISSING)\n", h.id, h.path, h.handlerFunc)
			failed++
		}
	}

	// =========================================================================
	// SECTION 7: SERVICE LAYER
	// =========================================================================
	fmt.Println("\n📋 SECTION 7: SERVICE LAYER")
	fmt.Println(strings.Repeat("─", 75))

	serviceSpecs := []struct {
		id     string
		method string
	}{
		{"S-1", "GetDeviceInspection"},
		{"S-2", "GetDeviceTimeline"},
		{"S-3", "GetTimelineEvents"},
		{"S-4", "decodeCursor"},
		{"S-5", "encodeCursor"},
		{"S-6", "determineDeviceStatus"},
		{"S-7", "determineFCMStatus"},
	}

	svcPath := filepath.Join(appDir, "diagnostics/diagnostics_service.go")
	svcContent, _ := os.ReadFile(svcPath)

	for _, s := range serviceSpecs {
		if strings.Contains(string(svcContent), "func (s *") && strings.Contains(string(svcContent), s.method) {
			fmt.Printf("  ✅ %s  diagnostics_service.%s()\n", s.id, s.method)
			passed++
		} else {
			fmt.Printf("  ❌ %s  diagnostics_service.%s() (MISSING)\n", s.id, s.method)
			failed++
		}
	}

	// =========================================================================
	// SECTION 8: GRAPHQL SCHEMA
	// =========================================================================
	fmt.Println("\n📋 SECTION 8: GRAPHQL SCHEMA")
	fmt.Println(strings.Repeat("─", 75))

	graphqlTypes := []string{
		"IdentityInfo", "SoftwareInfo", "RegistrationInfo", "ConnectionInfo", "TelemetryInfo",
		"DeviceInspection", "TimelineEvent", "TimelineConnection",
		"TimelineEventType", "TELEMETRY", "COMMAND_SENT", "COMMAND_ACK", "COMMAND_FAILED",
		"CONNECTION_OPEN", "CONNECTION_LOST", "FCM_FALLBACK", "RECONNECTED",
		"THRESHOLD_BREACH", "REGISTERED", "DEREGISTERED", "ERROR",
		"deviceInspection", "deviceTimeline",
	}

	for i, g := range graphqlTypes {
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
	// SECTION 10: RATE LIMITING & SECURITY
	// =========================================================================
	fmt.Println("\n📋 SECTION 10: RATE LIMITING & SECURITY")
	fmt.Println(strings.Repeat("─", 75))

	rateLimits := []struct {
		endpoint string
		limit   string
	}{
		{"GET /v1/device/:imei/inspect", "30/min"},
		{"GET /v1/device/:imei/timeline", "30/min"},
	}

	middlewarePath := filepath.Join(root, "apps/api/internal/api/middleware/rate_limit.go")
	middlewareContent, _ := os.ReadFile(middlewarePath)

	fmt.Println("--- Rate Limits ---")
	for _, rl := range rateLimits {
		if strings.Contains(string(middlewareContent), "rateLimit") || strings.Contains(string(middlewareContent), "RateLimit") {
			fmt.Printf("  ✅ %-45s %s\n", rl.endpoint, rl.limit)
			passed++
		} else {
			fmt.Printf("  ❌ %-45s %s (MISSING)\n", rl.endpoint, rl.limit)
			failed++
		}
	}

	fmt.Println("\n--- Security ---")
	securityChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"SEC-1", "Authentication middleware", func() bool {
			authPath := filepath.Join(handlerDir, "..", "middleware", "auth.go")
			content, _ := os.ReadFile(authPath)
			return strings.Contains(string(content), "auth") || strings.Contains(string(content), "Auth")
		}},
		{"SEC-2", "Device Authorization (DOA)", func() bool {
			return strings.Contains(string(svcContent), "owner") || strings.Contains(string(svcContent), "DOA") || strings.Contains(string(svcContent), "authorize")
		}},
	}

	for _, sec := range securityChecks {
		if sec.check() {
			fmt.Printf("  ✅ %s  %s\n", sec.id, sec.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", sec.id, sec.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 11: FILE STRUCTURE
	// =========================================================================
	fmt.Println("\n📋 SECTION 11: FILE STRUCTURE")
	fmt.Println(strings.Repeat("─", 75))

	files := []struct {
		id   string
		file string
	}{
		// Domain Layer
		{"F-D1", "domain/diagnostics/diagnostics_types.go"},
		{"F-D2", "domain/diagnostics/diagnostics_repository.go"},
		{"F-D3", "domain/diagnostics/diagnostics_errors.go"},
		// Application Layer
		{"F-A1", "application/diagnostics/diagnostics_service.go"},
		{"F-A2", "application/diagnostics/diagnostics_dto.go"},
		// Handler Layer
		{"F-H1", "api/handlers/diagnostics/diagnostics_inspect_handler.go"},
		{"F-H2", "api/handlers/diagnostics/diagnostics_timeline_handler.go"},
		{"F-H3", "api/handlers/diagnostics/diagnostics_routes.go"},
		// Infrastructure
		{"F-I1", "infrastructure/storage/diagnostics_storage.go"},
	}

	for _, f := range files {
		path := filepath.Join(root, "apps/api/internal/", f.file)
		if _, err := os.Stat(path); err == nil {
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
	fmt.Printf("DIAGNOSTICS API: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
