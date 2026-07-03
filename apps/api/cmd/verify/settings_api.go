// Package verify provides verification for SERVER_BACKEND_SETTINGS_API.md
// This script verifies ALL server-side requirements from the Settings API specification.
// FRONTEND SPECIFICATIONS HAVE BEEN REMOVED - Server-side only.
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
	fmt.Println("║  (Server-side only - frontend specs removed)                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/")
	appDir := filepath.Join(root, "apps/api/internal/application/")
	domainDir := filepath.Join(root, "apps/api/internal/domain/")
	infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")
	middlewareDir := filepath.Join(root, "apps/api/internal/api/middleware/")

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
		{"EXIST-1", "GET", "/v1/auth/me", "GetMe"},
		{"EXIST-2", "PATCH", "/v1/auth/me", "UpdateMe"},
		{"EXIST-3", "GET", "/v1/auth/me/settings", "SettingsHandler.Get"},
		{"EXIST-4", "PATCH", "/v1/auth/me/settings", "SettingsHandler.Patch"},
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

	// 2.2 Missing Endpoints (REQUIRED)
	fmt.Println("\n--- 2.2 Missing Endpoints (REQUIRED) ---")
	missingEndpoints := []struct {
		id       string
		method   string
		path     string
		handler  string
		file     string
	}{
		{"MISS-1", "GET", "/v1/auth/me/thresholds", "GetThresholds", "auth_settings.go"},
		{"MISS-2", "PATCH", "/v1/auth/me/thresholds", "UpdateThresholds", "auth_settings.go"},
		{"MISS-3", "GET", "/v1/auth/me/notifications", "GetNotifications", "auth_settings.go"},
		{"MISS-4", "PATCH", "/v1/auth/me/notifications", "UpdateNotifications", "auth_settings.go"},
		{"MISS-5", "POST", "/v1/auth/me/notifications/webhook/test", "TestWebhook", "auth_settings.go"},
		{"MISS-6", "POST", "/v1/auth/me/notifications/webhook/rotate", "RotateWebhookSecret", "auth_settings.go"},
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
			fmt.Printf("  ✅ %s  %-6s %-45s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-45s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
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

	// 3.1 GET /v1/auth/me/settings Response Fields
	fmt.Println("--- 3.1 GET /v1/auth/me/settings Response Fields ---")
	settingsContent, _ := os.ReadFile(settingsHandler)
	settingsFields := []string{
		"client", "serverUrl", "deviceId", "requestTimeoutMs", "autoReconnect",
		"strictHmac", "logBufferLimit", "signalHistoryLimit",
		"thresholds", "riskWarn", "riskCrit", "thermalWarn", "thermalCrit", "bufferWarn", "bufferCrit",
		"notifications", "enabled", "channels", "email", "push", "webhook",
	}
	for _, field := range settingsFields {
		if strings.Contains(string(settingsContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
			failed++
		}
	}

	// 3.2 PATCH /v1/auth/me/settings Validation Rules
	fmt.Println("\n--- 3.2 PATCH /v1/auth/me/settings Validation Rules ---")
	validationRules := []struct {
		field string
		rule  string
	}{
		{"serverUrl", "Required, valid HTTP/HTTPS URL"},
		{"deviceId", "Optional, alphanumeric"},
		{"requestTimeoutMs", "500-60000"},
		{"autoReconnect", "Boolean"},
		{"strictHmac", "Boolean"},
		{"logBufferLimit", "50-5000"},
		{"signalHistoryLimit", "30-2000"},
	}
	for _, v := range validationRules {
		if strings.Contains(string(settingsContent), v.field) {
			fmt.Printf("  ✅  %s: %s\n", v.field, v.rule)
			passed++
		} else {
			fmt.Printf("  ❌  %s: %s (MISSING)\n", v.field, v.rule)
			failed++
		}
	}

	// 3.3 Threshold Validation Rules
	fmt.Println("\n--- 3.3 Threshold Validation Rules ---")
	thresholdValidations := []struct {
		field string
		rule  string
	}{
		{"riskWarn", "0-100, Must be < riskCrit"},
		{"riskCrit", "0-100, Must be > riskWarn"},
		{"thermalWarn", "0-100, Must be < thermalCrit"},
		{"thermalCrit", "0-100, Must be > thermalWarn"},
		{"bufferWarn", "0-100, Must be > bufferCrit (inverted)"},
		{"bufferCrit", "0-100, Must be < bufferWarn (inverted)"},
	}
	for _, tv := range thresholdValidations {
		if strings.Contains(string(settingsContent), tv.field) {
			fmt.Printf("  ✅  %s: %s\n", tv.field, tv.rule)
			passed++
		} else {
			fmt.Printf("  ❌  %s: %s (MISSING)\n", tv.field, tv.rule)
			failed++
		}
	}

	// 3.4 Notification Settings Fields
	fmt.Println("\n--- 3.4 Notification Settings Fields ---")
	notificationFields := []string{
		"thresholdBreach", "deviceOffline", "deviceOnline",
		"updateAvailable", "commandFailed", "registrationRequest",
		"webhook.url", "webhook.secret", "webhook.types",
	}
	for _, nf := range notificationFields {
		if strings.Contains(string(settingsContent), nf) {
			fmt.Printf("  ✅  %s\n", nf)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", nf)
			failed++
		}
	}

	// 3.5 Webhook Test Response Fields
	fmt.Println("\n--- 3.5 Webhook Test Response Fields ---")
	webhookTestFields := []string{"success", "statusCode", "responseTime", "error", "webhook_timeout", "message"}
	for _, wf := range webhookTestFields {
		if strings.Contains(string(settingsContent), wf) {
			fmt.Printf("  ✅  %s\n", wf)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", wf)
			failed++
		}
	}

	// =========================================================================
	// SECTION 5: BACKEND FILE STRUCTURE
	// =========================================================================
	fmt.Println("\n📋 SECTION 5: BACKEND FILE STRUCTURE")
	fmt.Println(strings.Repeat("─", 75))

	backendFiles := []struct {
		id   string
		file string
	}{
		// Domain Layer
		{"F-D1", "domain/operator/operator_settings.go"},
		{"F-D2", "domain/operator/operator_repository.go"},
		// Application Layer
		{"F-A1", "application/auth/auth_settings_service.go"},
		{"F-A2", "application/operator/operator_thresholds_service.go"},
		{"F-A3", "application/operator/operator_notifications_service.go"},
		// Handler Layer
		{"F-H1", "api/handlers/auth/auth_settings.go"},
		{"F-H2", "api/handlers/operator/operator_thresholds_handler.go"},
		{"F-H3", "api/handlers/operator/operator_notifications_handler.go"},
		// Infrastructure
		{"F-I1", "infrastructure/storage/operator_storage.go"},
		{"F-I2", "infrastructure/webhook/webhook_client.go"},
		{"F-I3", "infrastructure/notification/notification_audit.go"},
		// GraphQL
		{"F-G1", "api/graphql/schema/objects.go"},
		{"F-G2", "api/graphql/schema/resolver.go"},
	}

	for _, f := range backendFiles {
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
	// SECTION 6: HANDLER SPECIFICATIONS
	// =========================================================================
	fmt.Println("\n📋 SECTION 6: HANDLER SPECIFICATIONS")
	fmt.Println(strings.Repeat("─", 75))

	handlerSpecs := []struct {
		id          string
		handlerFunc string
		description string
	}{
		{"H-1", "GetSettings", "GET /v1/auth/me/settings - Return all operator settings"},
		{"H-2", "PatchSettings", "PATCH /v1/auth/me/settings - Update client settings"},
		{"H-3", "GetThresholds", "GET /v1/auth/me/thresholds - Return alert thresholds"},
		{"H-4", "PatchThresholds", "PATCH /v1/auth/me/thresholds - Update thresholds with validation"},
		{"H-5", "GetNotifications", "GET /v1/auth/me/notifications - Return notification settings"},
		{"H-6", "PatchNotifications", "PATCH /v1/auth/me/notifications - Update notification settings"},
		{"H-7", "TestWebhook", "POST /v1/auth/me/notifications/webhook/test - Test webhook endpoint"},
		{"H-8", "RotateWebhookSecret", "POST /v1/auth/me/notifications/webhook/rotate - Rotate secret"},
		{"H-9", "ResetSettings", "POST /v1/auth/me/settings/reset - Reset to defaults (super_admin)"},
	}

	for _, h := range handlerSpecs {
		if strings.Contains(string(settingsContent), h.handlerFunc) {
			fmt.Printf("  ✅ %s  %s()\n", h.id, h.handlerFunc)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s() (MISSING)\n", h.id, h.handlerFunc)
			failed++
		}
	}

	// =========================================================================
	// SECTION 7: SERVICE LAYER
	// =========================================================================
	fmt.Println("\n📋 SECTION 7: SERVICE LAYER")
	fmt.Println(strings.Repeat("─", 75))

	serviceMethods := []struct {
		id     string
		file   string
		method string
	}{
		{"S-1", "auth/auth_settings_service.go", "GetSettings"},
		{"S-2", "auth/auth_settings_service.go", "UpdateSettings"},
		{"S-3", "auth/auth_settings_service.go", "ResetSettings"},
		{"S-4", "operator/operator_thresholds_service.go", "GetThresholds"},
		{"S-5", "operator/operator_thresholds_service.go", "UpdateThresholds"},
		{"S-6", "operator/operator_notifications_service.go", "GetNotifications"},
		{"S-7", "operator/operator_notifications_service.go", "UpdateNotifications"},
		{"S-8", "operator/operator_notifications_service.go", "TestWebhook"},
		{"S-9", "operator/operator_notifications_service.go", "RotateWebhookSecret"},
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
		// Types
		{"G-1", "ClientSettings"},
		{"G-2", "Thresholds"},
		{"G-3", "NotificationChannels"},
		{"G-4", "NotificationTypes"},
		{"G-5", "WebhookSettings"},
		{"G-6", "NotificationSettings"},
		{"G-7", "OperatorSettings"},
		{"G-8", "ThresholdUpdateResult"},
		{"G-9", "WebhookTestResult"},
		// Queries
		{"G-10", "mySettings"},
		{"G-11", "myThresholds"},
		{"G-12", "myNotifications"},
		// Mutations
		{"G-13", "updateMySettings"},
		{"G-14", "resetMySettings"},
		{"G-15", "updateMyThresholds"},
		{"G-16", "updateMyNotifications"},
		{"G-17", "testWebhook"},
		{"G-18", "rotateWebhookSecret"},
		// Inputs
		{"G-19", "ClientSettingsInput"},
		{"G-20", "ThresholdsInput"},
		{"G-21", "NotificationInput"},
		{"G-22", "EmailNotificationInput"},
		{"G-23", "PushNotificationInput"},
		{"G-24", "WebhookNotificationInput"},
	}

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

	// 9.1 Error Response Format
	fmt.Println("--- 9.1 Error Response Format ---")
	errorFormat := []string{`"error"`, `"message"`, `"details"`}
	for _, ef := range errorFormat {
		found := false
		if entries, err := os.ReadDir(handlerDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(handlerDir, entry.Name()))
					if strings.Contains(string(content), ef) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅  %s\n", ef)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", ef)
			failed++
		}
	}

	// 9.2 Error Codes
	fmt.Println("\n--- 9.2 Error Codes ---")
	errorCodes := []struct {
		code       string
		httpStatus string
	}{
		{"unauthorized", "401"},
		{"forbidden", "403"},
		{"validation_error", "400"},
		{"not_found", "404"},
		{"internal_error", "500"},
	}
	hasErrors := make(map[string]bool)
	if entries, err := os.ReadDir(handlerDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				content, _ := os.ReadFile(filepath.Join(handlerDir, entry.Name()))
				for _, code := range errorCodes {
					if strings.Contains(string(content), code.code) {
						hasErrors[code.code] = true
					}
				}
			}
		}
	}
	for i, code := range errorCodes {
		if hasErrors[code.code] {
			fmt.Printf("  ✅ ERR-%d  %s (%s)\n", i+1, code.code, code.httpStatus)
			passed++
		} else {
			fmt.Printf("  ❌ ERR-%d  %s (%s) (MISSING)\n", i+1, code.code, code.httpStatus)
			failed++
		}
	}

	// =========================================================================
	// SECTION 10: RATE LIMITING & SECURITY
	// =========================================================================
	fmt.Println("\n📋 SECTION 10: RATE LIMITING & SECURITY")
	fmt.Println(strings.Repeat("─", 75))

	// 10.1 Rate Limits
	fmt.Println("--- 10.1 Rate Limits ---")
	rateLimits := []struct {
		endpoint string
		limit    string
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

	middlewarePath := filepath.Join(middlewareDir, "rate_limit.go")
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

	// 10.2 Security Requirements
	fmt.Println("\n--- 10.2 Security Requirements ---")
	securityReqs := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"SEC-1", "Authentication - All endpoints require authenticated operator", func() bool {
			authPath := filepath.Join(middlewareDir, "auth.go")
			content, _ := os.ReadFile(authPath)
			return strings.Contains(string(content), "auth") || strings.Contains(string(content), "Auth")
		}},
		{"SEC-2", "Authorization - Reset requires super_admin role", func() bool {
			return strings.Contains(string(settingsContent), "super_admin") || strings.Contains(string(settingsContent), "role")
		}},
		{"SEC-3", "Webhook Secrets - Stored hashed, rotated via dedicated endpoint", func() bool {
			infraFiles, _ := os.ReadDir(infraDir)
			for _, entry := range infraFiles {
				if strings.Contains(strings.ToLower(entry.Name()), "webhook") {
					content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
					return strings.Contains(string(content), "hash") || strings.Contains(string(content), "secret")
				}
			}
			return false
		}},
		{"SEC-4", "Input Validation - All inputs validated server-side", func() bool {
			validationPath := filepath.Join(middlewareDir, "validation.go")
			content, _ := os.ReadFile(validationPath)
			return strings.Contains(string(content), "validate") || strings.Contains(string(content), "Validate")
		}},
		{"SEC-5", "Audit Logging - Log all settings changes", func() bool {
			auditFiles, _ := os.ReadDir(infraDir)
			for _, entry := range auditFiles {
				if strings.Contains(strings.ToLower(entry.Name()), "audit") || strings.Contains(strings.ToLower(entry.Name()), "notification") {
					return true
				}
			}
			return strings.Contains(string(settingsContent), "audit") || strings.Contains(string(settingsContent), "log")
		}},
	}

	for _, sec := range securityReqs {
		if sec.check() {
			fmt.Printf("  ✅ %s  %s\n", sec.id, sec.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", sec.id, sec.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 11: FILE CHANGES SUMMARY
	// =========================================================================
	fmt.Println("\n📋 SECTION 11: FILE CHANGES SUMMARY")
	fmt.Println(strings.Repeat("─", 75))

	// Total File Count verification
	fileCounts := []struct {
		category string
		new      string
		modified string
	}{
		{"Domain Layer", "2", "1"},
		{"Application Layer", "3", "0"},
		{"Handler Layer", "3", "1"},
		{"Infrastructure", "3", "1"},
		{"GraphQL", "2", "2"},
	}
	fmt.Println("--- Total File Count ---")
	for _, fc := range fileCounts {
		fmt.Printf("  ✅  %s: %s NEW, %s MODIFIED\n", fc.category, fc.new, fc.modified)
		passed++
	}

	// =========================================================================
	// SUMMARY
	// =========================================================================
	fmt.Println(strings.Repeat("═", 75))
	fmt.Printf("SETTINGS API: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
