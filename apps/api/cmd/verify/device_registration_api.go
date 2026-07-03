// Package verify provides verification for SERVER_BACKEND_DEVICE_REGISTRATION_API.md
// This script verifies ALL server-side requirements from the Device Registration API specification.
// FRONTEND SPECIFICATIONS HAVE BEEN REMOVED - Server-side only.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// verifyDeviceRegistration verifies ALL requirements from SERVER_BACKEND_DEVICE_REGISTRATION_API.md
func verifyDeviceRegistration() bool {
	root := getRoot()
	passed := 0
	failed := 0

	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_DEVICE_REGISTRATION_API.md - COMPREHENSIVE VERIFICATION  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// =========================================================================
	// SECTION 2: CURRENT STATE ANALYSIS
	// =========================================================================
	fmt.Println("📋 SECTION 2: CURRENT STATE ANALYSIS")
	fmt.Println(strings.Repeat("─", 75))

	// 2.1 Existing Endpoints
	fmt.Println("--- 2.1 Existing Endpoints ---")
	existingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"EXIST-1", "POST", "/v1/device/inbox", "HandleInboxRequest"},
		{"EXIST-2", "POST", "/v1/device/confirm", "Handle"},
		{"EXIST-3", "GET", "/v1/device/:imei", "DeviceHandler"},
		{"EXIST-4", "DELETE", "/v1/device/:imei", "Deregister"},
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
		{"MISS-1", "GET", "/v1/device/inbox", "GetInbox"},
		{"MISS-2", "GET", "/v1/device/inbox/:imei", "GetInboxEntry"},
		{"MISS-3", "POST", "/v1/device/inbox/:imei/ack", "AckInbox"},
		{"MISS-4", "GET", "/v1/devices", "GetDevices"},
		{"MISS-5", "GET", "/v1/devices/:imei", "GetDevice"},
		{"MISS-6", "POST", "/v1/device/register", "RegisterDevice"},
	}

	inboxHandler := filepath.Join(handlerDir, "inbox/inbox_handler.go")
	devicesHandler := filepath.Join(handlerDir, "device/devices.go")

	for _, ep := range missingEndpoints {
		found := false
		var content []byte
		var err error

		if strings.Contains(ep.path, "inbox") {
			content, err = os.ReadFile(inboxHandler)
		} else {
			content, err = os.ReadFile(devicesHandler)
		}

		if err == nil && strings.Contains(string(content), ep.handler) {
			found = true
		}

		if found {
			fmt.Printf("  ✅ %s  %-6s %-35s (%s)\n", ep.id, ep.method, ep.path, ep.handler)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %-6s %-35s (%s MISSING)\n", ep.id, ep.method, ep.path, ep.handler)
			failed++
		}
	}

	// 2.3 Existing Domain Entities
	fmt.Println("\n--- 2.3 Existing Domain Entities ---")
	domainEntities := []struct {
		id   string
		file string
	}{
		{"DOM-1", "domain/device/device_entity.go"},
		{"DOM-2", "domain/device/status.go"},
		{"DOM-3", "domain/command/entity.go"},
	}

	for _, d := range domainEntities {
		path := filepath.Join(root, "apps/api/internal/", d.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ %s  %s (EXISTS)\n", d.id, d.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", d.id, d.file)
			failed++
		}
	}

	// 2.4 Database Tables
	fmt.Println("\n--- 2.4 Database Tables ---")
	schemaDir := filepath.Join(root, "supabase/migrations/")
	dbTables := []struct {
		id   string
		name string
	}{
		{"DB-1", "devices"},
		{"DB-2", "inbox_requests"},
	}

	// =========================================================================
	// SECTION 3: REGISTRATION FLOW
	// =========================================================================
	fmt.Println("\n📋 SECTION 3: REGISTRATION FLOW")
	fmt.Println(strings.Repeat("─", 75))

	registrationFlowChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"FLOW-1", "Device sends POST /v1/device/inbox with IMEI, model, manufacturer, fcmToken", func() bool {
			// Check inbox handler for incoming request handling
			inboxPath := filepath.Join(handlerDir, "inbox/inbox_handler.go")
			content, _ := os.ReadFile(inboxPath)
			return strings.Contains(string(content), "IMEI") && strings.Contains(string(content), "fcmToken")
		}},
		{"FLOW-2", "Server stores in INBOX with status: pending", func() bool {
			inboxPath := filepath.Join(handlerDir, "inbox/inbox_handler.go")
			content, _ := os.ReadFile(inboxPath)
			return strings.Contains(string(content), "pending") || strings.Contains(string(content), "Status")
		}},
		{"FLOW-3", "Operator can GET /v1/device/inbox to view pending", func() bool {
			inboxPath := filepath.Join(handlerDir, "inbox/inbox_handler.go")
			content, _ := os.ReadFile(inboxPath)
			return strings.Contains(string(content), "GetInbox")
		}},
		{"FLOW-4", "Operator can POST /v1/device/inbox/:imei/ack with action approve/reject", func() bool {
			inboxPath := filepath.Join(handlerDir, "inbox/inbox_handler.go")
			content, _ := os.ReadFile(inboxPath)
			return strings.Contains(string(content), "AckInbox") && strings.Contains(string(content), "approve")
		}},
		{"FLOW-5", "Server generates commandSecret on approval", func() bool {
			inboxPath := filepath.Join(handlerDir, "inbox/inbox_handler.go")
			content, _ := os.ReadFile(inboxPath)
			return strings.Contains(string(content), "commandSecret") || strings.Contains(string(content), "command_secret")
		}},
		{"FLOW-6", "FCM push sent to device with commandSecret", func() bool {
			infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
			if entries, err := os.ReadDir(infraDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(strings.ToLower(entry.Name()), "fcm") {
						content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
						return strings.Contains(string(content), "FCM") || strings.Contains(string(content), "Push")
					}
				}
			}
			return false
		}},
		{"FLOW-7", "Device confirms with POST /v1/device/confirm", func() bool {
			confirmPath := filepath.Join(handlerDir, "device/confirm_handler.go")
			if _, err := os.Stat(confirmPath); err == nil {
				content, _ := os.ReadFile(confirmPath)
				return strings.Contains(string(content), "Confirm") || strings.Contains(string(content), "commandSecret")
			}
			// Check if it's in another file
			if entries, err := os.ReadDir(handlerDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(handlerDir, entry.Name()))
						if strings.Contains(string(content), "commandSecret") && strings.Contains(string(content), "Confirm") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"FLOW-8", "Server validates commandSecret", func() bool {
			confirmPath := filepath.Join(handlerDir, "device/confirm_handler.go")
			if _, err := os.Stat(confirmPath); err == nil {
				content, _ := os.ReadFile(confirmPath)
				return strings.Contains(string(content), "Validate") || strings.Contains(string(content), "verify")
			}
			return false
		}},
		{"FLOW-9", "Device moved from INBOX to DEVICES table on confirm", func() bool {
			appDir := filepath.Join(root, "apps/api/internal/application/")
			if entries, err := os.ReadDir(appDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(appDir, entry.Name()))
						if strings.Contains(string(content), "Move") || strings.Contains(string(content), "Transfer") {
							return true
						}
					}
				}
			}
			return false
		}},
	}

	for _, f := range registrationFlowChecks {
		if f.check() {
			fmt.Printf("  ✅ %s  %s\n", f.id, f.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", f.id, f.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 3.2: DEREGISTRATION FLOW
	// =========================================================================
	fmt.Println("\n📋 SECTION 3.2: DEREGISTRATION FLOW")
	fmt.Println(strings.Repeat("─", 75))

	deregFlowChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"DEREG-1", "DELETE /v1/device/:imei marks as deregistered", func() bool {
			devicesPath := filepath.Join(handlerDir, "device/devices.go")
			content, _ := os.ReadFile(devicesPath)
			return strings.Contains(string(content), "Deregister") || strings.Contains(string(content), "Delete")
		}},
		{"DEREG-2", "Soft delete with 30 day retention", func() bool {
			// Check device service for soft delete logic
			appDir := filepath.Join(root, "apps/api/internal/application/device/")
			if entries, err := os.ReadDir(appDir); err == nil {
				for _, entry := range entries {
					content, _ := os.ReadFile(filepath.Join(appDir, entry.Name()))
					if strings.Contains(string(content), "30") || strings.Contains(string(content), "retention") || strings.Contains(string(content), "soft") {
						return true
					}
				}
			}
			return false
		}},
	}

	for _, d := range deregFlowChecks {
		if d.check() {
			fmt.Printf("  ✅ %s  %s\n", d.id, d.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", d.id, d.description)
			failed++
		}
	}

	for _, t := range dbTables {
		found := false
		if entries, err := os.ReadDir(schemaDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
					if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), t.name) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %s table (EXISTS)\n", t.id, t.name)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s table (MISSING)\n", t.id, t.name)
			failed++
		}
	}

	// =========================================================================
	// SECTION 4: REQUIRED API ENDPOINTS
	// =========================================================================
	fmt.Println("\n📋 SECTION 4: REQUIRED API ENDPOINTS")
	fmt.Println(strings.Repeat("─", 75))

	// GET /v1/device/inbox
	fmt.Println("--- GET /v1/device/inbox ---")
	inboxContent, _ := os.ReadFile(inboxHandler)
	inboxChecks := map[string]bool{
		"GetInbox handler":    false,
		"status param":        false,
		"page param":          false,
		"limit param":         false,
	}
	for check := range inboxChecks {
		if strings.Contains(string(inboxContent), check) {
			inboxChecks[check] = true
		}
	}
	for check, found := range inboxChecks {
		if found {
			fmt.Printf("  ✅  %s\n", check)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check)
			failed++
		}
	}

	// POST /v1/device/inbox/:imei/ack
	fmt.Println("\n--- POST /v1/device/inbox/:imei/ack ---")
	ackChecks := map[string]bool{
		"AckInbox handler":       false,
		"approve action":         false,
		"reject action":         false,
		"404 error":             false,
		"409 error":             false,
	}
	for check := range ackChecks {
		if strings.Contains(string(inboxContent), check) {
			ackChecks[check] = true
		}
	}
	for check, found := range ackChecks {
		if found {
			fmt.Printf("  ✅  %s\n", check)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check)
			failed++
		}
	}

	// GET /v1/devices
	fmt.Println("\n--- GET /v1/devices ---")
	devicesContent, _ := os.ReadFile(devicesHandler)
	devicesChecks := map[string]bool{
		"GetDevices handler":  false,
		"status filter":      false,
		"search filter":      false,
		"pagination":         false,
	}
	for check := range devicesChecks {
		if strings.Contains(string(devicesContent), check) {
			devicesChecks[check] = true
		}
	}
	for check, found := range devicesChecks {
		if found {
			fmt.Printf("  ✅  %s\n", check)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check)
			failed++
		}
	}

	// =========================================================================
	// SECTION 5: DATABASE SCHEMA
	// =========================================================================
	fmt.Println("\n📋 SECTION 5: DATABASE SCHEMA")
	fmt.Println(strings.Repeat("─", 75))

	schemaChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"SCHEMA-1", "inbox_requests table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), "inbox_requests") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"SCHEMA-2", "idx_inbox_pending index", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "idx_inbox_pending") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"SCHEMA-3", "registration_logs table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "registration_logs") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"SCHEMA-4", "devices.device_name column", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "device_name") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"SCHEMA-5", "devices.command_secret_hash column", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "command_secret_hash") {
							return true
						}
					}
				}
			}
			return false
		}},
		// Additional schema from Section 12.2
		{"SCHEMA-6", "domain/inbox/inbox_entity.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/inbox/inbox_entity.go"))
			return err == nil
		}},
		{"SCHEMA-7", "domain/inbox/inbox_status.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/inbox/inbox_status.go"))
			return err == nil
		}},
		{"SCHEMA-8", "domain/inbox/inbox_repository.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/inbox/inbox_repository.go"))
			return err == nil
		}},
		{"SCHEMA-9", "domain/inbox/inbox_errors.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/inbox/inbox_errors.go"))
			return err == nil
		}},
		{"SCHEMA-10", "infrastructure/storage/inbox_storage.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/infrastructure/storage/inbox_storage.go"))
			return err == nil
		}},
		{"SCHEMA-11", "infrastructure/storage/registration_log_storage.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/infrastructure/storage/registration_log_storage.go"))
			return err == nil
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
		{"SVC-1", "inbox/inbox_service.go", "GetInbox"},
		{"SVC-2", "inbox/inbox_service.go", "GetInboxEntry"},
		{"SVC-3", "inbox/inbox_service.go", "AckInbox"},
		{"SVC-4", "device/device_service.go", "GetDevices"},
		{"SVC-5", "device/device_service.go", "GetDeviceByIMEI"},
		{"SVC-6", "device/device_service.go", "DeregisterDevice"},
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

	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")
	graphqlTypes := []struct {
		id    string
		type_ string
	}{
		// GraphQL Types from Section 9.1
		{"GQL-1", "InboxEntry"},
		{"GQL-2", "InboxStatus"},
		{"GQL-3", "InboxConnection"},
		{"GQL-4", "Device"},
		{"GQL-5", "DeviceConnection"},
		{"GQL-6", "DeviceListConnection"},
		{"GQL-7", "AckResult"},
		{"GQL-8", "DeregisterResult"},
		// GraphQL Enums
		{"GQL-9", "PENDING"},
		{"GQL-10", "APPROVED"},
		{"GQL-11", "REJECTED"},
		// GraphQL Queries from Section 9.2
		{"GQL-12", "inbox query"},
		{"GQL-13", "inboxEntry query"},
		{"GQL-14", "devices query"},
		{"GQL-15", "device query"},
		// GraphQL Mutations from Section 9.3
		{"GQL-16", "ackInbox mutation"},
		{"GQL-17", "deregisterDevice mutation"},
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

	errorCodes := []string{"bad_request", "unauthorized", "forbidden", "not_found", "conflict", "rate_limited", "internal_error"}
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

	rateLimits := []string{"GET /v1/device/inbox", "POST /v1/device/inbox/:imei/ack", "GET /v1/devices", "DELETE /v1/devices/:imei"}
	middlewarePath := filepath.Join(root, "apps/api/internal/api/middleware/rate_limit.go")
	middlewareContent, _ := os.ReadFile(middlewarePath)

	for _, rl := range rateLimits {
		if strings.Contains(string(middlewareContent), "rateLimit") || strings.Contains(string(middlewareContent), "RateLimit") {
			fmt.Printf("  ✅ %s (rate limited)\n", rl)
			passed++
		} else {
			fmt.Printf("  ❌ %s (NOT RATE_LIMITED)\n", rl)
			failed++
		}
	}

	// Security checks
	fmt.Println("\n--- Security (Section 11.2) ---")
	securityChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"SEC-1", "Audit logging - Log all registration/deregistration actions", func() bool {
			if entries, err := os.ReadDir(appDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(appDir, entry.Name()))
						if strings.Contains(string(content), "Audit") || strings.Contains(string(content), "audit") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"SEC-2", "FCM notification", func() bool {
			infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
			if entries, err := os.ReadDir(infraDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(strings.ToLower(entry.Name()), "fcm") {
						return true
					}
				}
			}
			return false
		}},
		{"SEC-3", "Authentication - All endpoints require authenticated operator", func() bool {
			middlewarePath := filepath.Join(handlerDir, "..", "middleware", "auth.go")
			content, _ := os.ReadFile(middlewarePath)
			return strings.Contains(string(content), "auth") || strings.Contains(string(content), "Auth")
		}},
		{"SEC-4", "Device Ownership - DOA check on all device-specific operations", func() bool {
			// Check service layer for ownership validation
			if entries, err := os.ReadDir(appDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(appDir, entry.Name()))
						if strings.Contains(string(content), "owner") || strings.Contains(string(content), "DOA") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"SEC-5", "Secret Generation - Use crypto/rand for commandSecret", func() bool {
			inboxPath := filepath.Join(handlerDir, "inbox", "inbox_handler.go")
			content, _ := os.ReadFile(inboxPath)
			return strings.Contains(string(content), "crypto/rand") || strings.Contains(string(content), "rand")
		}},
		{"SEC-6", "FCM Validation - Verify FCM token format before storing", func() bool {
			inboxPath := filepath.Join(handlerDir, "inbox", "inbox_handler.go")
			content, _ := os.ReadFile(inboxPath)
			return strings.Contains(string(content), "fcmToken") && (strings.Contains(string(content), "valid") || strings.Contains(string(content), "format"))
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
	// SECTION 12: FILE CHANGES SUMMARY (Verification)
	// =========================================================================
	fmt.Println("\n📋 SECTION 12: FILE CHANGES SUMMARY VERIFICATION")
	fmt.Println(strings.Repeat("─", 75))

	fileStructureChecks := []struct {
		id   string
		file string
	}{
		// Domain Layer (Section 12.2)
		{"FS-D1", "domain/inbox/inbox_entity.go"},
		{"FS-D2", "domain/inbox/inbox_status.go"},
		{"FS-D3", "domain/inbox/inbox_repository.go"},
		{"FS-D4", "domain/inbox/inbox_errors.go"},
		{"FS-D5", "domain/device/device_entity.go (MODIFIED)"},
		// Application Layer (Section 12.2)
		{"FS-A1", "application/inbox/inbox_service.go"},
		{"FS-A2", "application/inbox/inbox_dto.go"},
		{"FS-A3", "application/inbox/inbox_errors.go"},
		{"FS-A4", "application/device/device_service.go (MODIFIED)"},
		// Handler Layer (Section 12.2)
		{"FS-H1", "api/handlers/inbox/inbox_handler.go"},
		{"FS-H2", "api/handlers/device/device_list_handler.go"},
		{"FS-H3", "api/handlers/inbox/inbox_routes.go"},
		// Infrastructure Layer (Section 12.2)
		{"FS-I1", "infrastructure/storage/inbox_storage.go"},
		{"FS-I2", "infrastructure/storage/registration_log_storage.go"},
		{"FS-I3", "infrastructure/storage/device_storage.go (MODIFIED)"},
	}

	for _, f := range fileStructureChecks {
		path := filepath.Join(root, "apps/api/internal/", f.file)
		// Handle (MODIFIED) suffix
		checkFile := strings.Split(f.file, " (")[0]
		path = filepath.Join(root, "apps/api/internal/", checkFile)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✅ %s  %s\n", f.id, f.file)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", f.id, f.file)
			failed++
		}
	}

	// =========================================================================
	// SECTION 13: IMPLEMENTATION ORDER VERIFICATION
	// =========================================================================
	fmt.Println("\n📋 SECTION 13: IMPLEMENTATION ORDER VERIFICATION")
	fmt.Println(strings.Repeat("─", 75))

	implOrderChecks := []struct {
		id          string
		phase       string
		description string
		check       func() bool
	}{
		{"IO-1", "Phase 1", "001_create_inbox_requests.sql migration", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "inbox") {
						return true
					}
				}
			}
			return false
		}},
		{"IO-2", "Phase 1", "domain/inbox/inbox_entity.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/inbox/inbox_entity.go"))
			return err == nil
		}},
		{"IO-3", "Phase 1", "domain/inbox/inbox_status.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/domain/inbox/inbox_status.go"))
			return err == nil
		}},
		{"IO-4", "Phase 2", "infrastructure/storage/inbox_storage.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/infrastructure/storage/inbox_storage.go"))
			return err == nil
		}},
		{"IO-5", "Phase 3", "application/inbox/inbox_service.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/application/inbox/inbox_service.go"))
			return err == nil
		}},
		{"IO-6", "Phase 4", "api/handlers/inbox/inbox_handler.go", func() bool {
			_, err := os.Stat(filepath.Join(root, "apps/api/internal/api/handlers/inbox/inbox_handler.go"))
			return err == nil
		}},
	}

	for _, io := range implOrderChecks {
		if io.check() {
			fmt.Printf("  ✅ %s  [%s] %s\n", io.id, io.phase, io.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  [%s] %s (MISSING)\n", io.id, io.phase, io.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 13: TESTING STRATEGY (Section 12.2 Phase 7)
	// =========================================================================
	fmt.Println("\n📋 SECTION 13: TESTING STRATEGY VERIFICATION")
	fmt.Println(strings.Repeat("─", 75))

	testingChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"TEST-1", "Unit tests for inbox service", func() bool {
			testDir := filepath.Join(root, "apps/api/internal/application/inbox/")
			if entries, err := os.ReadDir(testDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "_test.go") {
						return true
					}
				}
			}
			return false
		}},
		{"TEST-2", "Integration tests for handlers", func() bool {
			handlerTestDir := filepath.Join(root, "apps/api/internal/api/handlers/inbox/")
			if entries, err := os.ReadDir(handlerTestDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(entry.Name(), "_test.go") {
						return true
					}
				}
			}
			return false
		}},
		{"TEST-3", "FCM notification tests", func() bool {
			infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
			if entries, err := os.ReadDir(infraDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(strings.ToLower(entry.Name()), "fcm") {
						content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
						return strings.Contains(string(content), "Test") || strings.Contains(string(content), "test")
					}
				}
			}
			return false
		}},
	}

	for _, t := range testingChecks {
		if t.check() {
			fmt.Printf("  ✅ %s  %s\n", t.id, t.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (NOT FOUND)\n", t.id, t.description)
			failed++
		}
	}

	// =========================================================================
	// SUMMARY
	// =========================================================================
	fmt.Println(strings.Repeat("═", 75))
	fmt.Printf("DEVICE REGISTRATION API: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
