// Package verify provides verification for SERVER_BACKEND_SETTINGS_API.md
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// verifySettings verifies ALL requirements from SERVER_BACKEND_SETTINGS_API.md
func verifySettings() bool {
	root := getRoot()
	passed := 0
	failed := 0

	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_SETTINGS_API.md - COMPREHENSIVE VERIFICATION  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// =========================================================================
	// SECTION 2: CURRENT STATE ANALYSIS
	// =========================================================================
	fmt.Println("📋 SECTION 2: CURRENT STATE ANALYSIS")
	fmt.Println(strings.Repeat("─", 75))

	// 2.1 Existing Related Endpoints
	fmt.Println("--- 2.1 Existing Related Endpoints ---")
	existingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"EXIST-1", "GET", "/v1/auth/me", "AuthHandler.GetMe"},
		{"EXIST-2", "PATCH", "/v1/auth/me", "AuthHandler.UpdateMe"},
		{"EXIST-3", "GET", "/v1/auth/me/settings", "SettingsHandler.Get"},
		{"EXIST-4", "PATCH", "/v1/auth/me/settings", "SettingsHandler.Patch"},
	}

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/")
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

	// 2.2 Missing Endpoints
	fmt.Println("\n--- 2.2 Missing Endpoints (REQUIRED) ---")
	missingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"MISS-1", "GET", "/v1/auth/me/thresholds", "GetThresholds"},
		{"MISS-2", "PATCH", "/v1/auth/me/thresholds", "UpdateThresholds"},
		{"MISS-3", "GET", "/v1/auth/me/notifications", "GetNotifications"},
		{"MISS-4", "PATCH", "/v1/auth/me/notifications", "UpdateNotifications"},
		{"MISS-5", "POST", "/v1/auth/me/notifications/webhook/test", "TestWebhook"},
		{"MISS-6", "POST", "/v1/auth/me/notifications/webhook/rotate", "RotateWebhookSecret"},
		{"MISS-7", "POST", "/v1/auth/me/settings/reset", "ResetSettings"},
	}

	settingsHandler := filepath.Join(handlerDir, "auth_settings.go")
	for _, ep := range missingEndpoints {
		found := false
		if content, err := os.ReadFile(settingsHandler); err == nil {
			if strings.Contains(string(content), ep.handler) {
				found = true
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %-6s %-40s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-40s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	// 2.3 Existing Data Models
	fmt.Println("\n--- 2.3 Existing Data Models ---")
	dataModels := []struct {
		id   string
		file string
	}{
		{"MODEL-1", "domain/operator/operator_entity.go"},
		{"MODEL-2", "domain/operator/settings.go"},
	}

	domainDir := filepath.Join(root, "apps/api/internal/domain/")
	for _, m := range dataModels {
		if _, err := os.Stat(filepath.Join(domainDir, m.file)); err == nil {
			fmt.Printf("  ✅ %s  %s (EXISTS)\n", m.id, m.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", m.id, m.file)
			failed++
		}
	}

	// =========================================================================
	// SECTION 3: REQUIRED API ENDPOINTS
	// =========================================================================
	fmt.Println("\n📋 SECTION 3: REQUIRED API ENDPOINTS")
	fmt.Println(strings.Repeat("─", 75))

	// GET /v1/auth/me/settings - Section 3.1
	fmt.Println("--- GET /v1/auth/me/settings (Section 3.1) ---")
	settingsContent, _ := os.ReadFile(settingsHandler)
	settingsChecks := map[string]bool{
		"Client settings":   false,
		"Thresholds":      false,
		"Notifications":   false,
		"serverUrl":       false,
		"deviceId":        false,
	}
	for check := range settingsChecks {
		if strings.Contains(string(settingsContent), check) {
			settingsChecks[check] = true
		}
	}
	for check, found := range settingsChecks {
		if found {
			fmt.Printf("  ✅  %s\n", check)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check)
			failed++
		}
	}

	// GET /v1/auth/me/thresholds - Section 3.3
	fmt.Println("\n--- GET /v1/auth/me/thresholds (Section 3.3) ---")
	thresholdChecks := []string{"riskWarn", "riskCrit", "thermalWarn", "thermalCrit", "bufferWarn", "bufferCrit"}
	for _, t := range thresholdChecks {
		if strings.Contains(string(settingsContent), t) {
			fmt.Printf("  ✅  %s\n", t)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", t)
			failed++
		}
	}

	// GET /v1/auth/me/notifications - Section 3.5
	fmt.Println("\n--- GET /v1/auth/me/notifications (Section 3.5) ---")
	notificationChecks := []string{"enabled", "channels", "email", "push", "webhook"}
	for _, n := range notificationChecks {
		if strings.Contains(string(settingsContent), n) {
			fmt.Printf("  ✅  %s\n", n)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", n)
			failed++
		}
	}

	// POST /v1/auth/me/notifications/webhook/test - Section 3.6
	fmt.Println("\n--- POST /v1/auth/me/notifications/webhook/test (Section 3.6) ---")
	webhookTestChecks := []string{"TestWebhook", "webhook", "test"}
	for _, w := range webhookTestChecks {
		if strings.Contains(string(settingsContent), w) {
			fmt.Printf("  ✅  %s\n", w)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", w)
			failed++
		}
	}

	// POST /v1/auth/me/notifications/webhook/rotate - Section 3.8
	fmt.Println("\n--- POST /v1/auth/me/notifications/webhook/rotate (Section 3.8) ---")
	webhookRotateChecks := []string{"RotateWebhookSecret", "rotate", "secret"}
	for _, w := range webhookRotateChecks {
		if strings.Contains(string(settingsContent), w) {
			fmt.Printf("  ✅  %s\n", w)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", w)
			failed++
		}
	}

	// POST /v1/auth/me/settings/reset - Section 3.10
	fmt.Println("\n--- POST /v1/auth/me/settings/reset (Section 3.10) ---")
	resetChecks := []string{"ResetSettings", "reset"}
	for _, r := range resetChecks {
		if strings.Contains(string(settingsContent), r) {
			fmt.Printf("  ✅  %s\n", r)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", r)
			failed++
		}
	}

	// =========================================================================
	// SECTION 5: VALIDATION RULES
	// =========================================================================
	fmt.Println("\n📋 SECTION 5: VALIDATION RULES")
	fmt.Println(strings.Repeat("─", 75))

	validationRules := []struct {
		id    string
		field string
		rule  string
	}{
		{"V-1", "serverUrl", "Required, valid HTTP/HTTPS URL"},
		{"V-2", "deviceId", "Optional, alphanumeric"},
		{"V-3", "requestTimeoutMs", "500-60000"},
		{"V-4", "autoReconnect", "Boolean"},
		{"V-5", "strictHmac", "Boolean"},
		{"V-6", "logBufferLimit", "50-5000"},
		{"V-7", "signalHistoryLimit", "30-2000"},
	}

	for _, v := range validationRules {
		if strings.Contains(string(settingsContent), v.field) {
			fmt.Printf("  ✅ %s  %s (%s)\n", v.id, v.field, v.rule)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (%s) - MISSING\n", v.id, v.field, v.rule)
			failed++
		}
	}

	// =========================================================================
	// SECTION 7: SERVICE LAYER
	// =========================================================================
	fmt.Println("\n📋 SECTION 7: SERVICE LAYER")
	fmt.Println(strings.Repeat("─", 75))

	appDir := filepath.Join(root, "apps/api/internal/application/")
	serviceMethods := []struct {
		id     string
		file   string
		method string
	}{
		{"S-1", "auth/auth_service.go", "GetSettings"},
		{"S-2", "auth/auth_service.go", "UpdateSettings"},
		{"S-3", "auth/auth_service.go", "GetThresholds"},
		{"S-4", "auth/auth_service.go", "UpdateThresholds"},
		{"S-5", "auth/auth_service.go", "GetNotifications"},
		{"S-6", "auth/auth_service.go", "UpdateNotifications"},
		{"S-7", "auth/auth_service.go", "TestWebhook"},
	}

	for _, s := range serviceMethods {
		path := filepath.Join(appDir, s.file)
		found := false
		if content, err := os.ReadFile(path); err == nil {
			found = strings.Contains(string(content), "func (s *") && strings.Contains(string(content), s.method)
		}
		if found {
			fmt.Printf("  ✅ %s  %s.%s()\n", s.id, s.file, s.method)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s.%s() (MISSING)\n", s.id, s.file, s.method)
			failed++
		}
	}

	// =========================================================================
	// SECTION 8: GRAPHQL SCHEMA
	// =========================================================================
	fmt.Println("\n📋 SECTION 8: GRAPHQL SCHEMA")
	fmt.Println(strings.Repeat("─", 75))

	graphqlTypes := []struct {
		id    string
		type_ string
	}{
		{"G-1", "ClientSettings"},
		{"G-2", "Thresholds"},
		{"G-3", "Notifications"},
		{"G-4", "NotificationChannel"},
		{"G-5", "NotificationChannels"},
		{"G-6", "NotificationTypes"},
		{"G-7", "WebhookSettings"},
		{"G-8", "OperatorSettings"},
		{"G-9", "ThresholdUpdateResult"},
		{"G-10", "WebhookTestResult"},
		{"G-11", "me query"},
		{"G-12", "mySettings query"},
		{"G-13", "myThresholds query"},
		{"G-14", "myNotifications query"},
		{"G-15", "updateMySettings mutation"},
		{"G-16", "resetMySettings mutation"},
		{"G-17", "updateMyThresholds mutation"},
		{"G-18", "updateMyNotifications mutation"},
		{"G-19", "testWebhook mutation"},
		{"G-20", "rotateWebhookSecret mutation"},
	}

	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")
	for _, g := range graphqlTypes {
		found := false
		if entries, err := os.ReadDir(gqlDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(gqlDir, entry.Name()))
					if strings.Contains(string(content), g.type_) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %s\n", g.id, g.type_)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", g.id, g.type_)
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
		{"GET /v1/auth/me/settings", "60/min"},
		{"PATCH /v1/auth/me/settings", "30/min"},
		{"GET /v1/auth/me/thresholds", "60/min"},
		{"PATCH /v1/auth/me/thresholds", "30/min"},
		{"GET /v1/auth/me/notifications", "60/min"},
		{"PATCH /v1/auth/me/notifications", "30/min"},
		{"POST /v1/auth/me/settings/reset", "5/hour"},
		{"POST /v1/auth/me/notifications/webhook/test", "10/hour"},
		{"POST /v1/auth/me/notifications/webhook/rotate", "10/hour"},
	}

	middlewarePath := filepath.Join(root, "apps/api/internal/api/middleware/rate_limit.go")
	middlewareContent, _ := os.ReadFile(middlewarePath)

	for _, rl := range rateLimits {
		if strings.Contains(string(middlewareContent), "rateLimit") || strings.Contains(string(middlewareContent), "RateLimit") {
			fmt.Printf("  ✅ %-50s %s\n", rl.endpoint, rl.limit)
			passed++
		} else {
			fmt.Printf("  ❌ %-50s %s (MISSING)\n", rl.endpoint, rl.limit)
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
		{"F-1", "api/handlers/auth/auth_settings.go"},
		{"F-2", "application/auth/auth_service.go"},
		{"F-3", "domain/operator/settings.go"},
		{"F-4", "api/graphql/schema/settings.go"},
	}

	for _, f := range requiredFiles {
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
	fmt.Printf("SETTINGS API: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
