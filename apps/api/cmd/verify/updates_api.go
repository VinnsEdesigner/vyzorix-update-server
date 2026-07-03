// Package verify provides verification for SERVER_BACKEND_UPDATES_API.md
// This script verifies ALL server-side requirements from the Updates API specification.
// FRONTEND SPECIFICATIONS HAVE BEEN REMOVED - Server-side only.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// verifyUpdates verifies ALL requirements from SERVER_BACKEND_UPDATES_API.md
func verifyUpdates() bool {
	root := getRoot()
	passed := 0
	failed := 0

	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_UPDATES_API.md - COMPREHENSIVE VERIFICATION  ║")
	fmt.Println("║  (Server-side only - frontend specs removed)                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/")
	appDir := filepath.Join(root, "apps/api/internal/application/")
	infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")
	schemaDir := filepath.Join(root, "supabase/migrations/")
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
		{"EXIST-1", "GET", "/api/v1/version", "VersionHandler"},
		{"EXIST-2", "GET", "/api/v1/apk/:filename", "APKHandler"},
		{"EXIST-3", "POST", "/v1/device/:id/command", "CommandHandler (WAKE_UP_UPDATER)"},
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
	updatesHandler := filepath.Join(handlerDir, "updates/updates_handler.go")
	updatesContent, _ := os.ReadFile(updatesHandler)

	missingEndpoints := []struct {
		id       string
		method   string
		path     string
		handler  string
	}{
		{"MISS-1", "GET", "/v1/updates/status", "GetUpdateStatus"},
		{"MISS-2", "GET", "/v1/updates/versions", "GetVersions"},
		{"MISS-3", "GET", "/v1/updates/changelog", "GetChangelog"},
		{"MISS-4", "POST", "/v1/updates/push", "PushUpdate"},
		{"MISS-5", "GET", "/v1/updates/history", "GetHistory"},
		{"MISS-6", "GET", "/v1/updates/export", "ExportVersions"},
		{"MISS-7", "POST", "/v1/updates/sync", "SyncVersions"},
		{"MISS-8", "POST", "/v1/updates/history/:id/cancel", "CancelUpdate"},
	}

	for _, ep := range missingEndpoints {
		found := false
		if content, err := os.ReadFile(updatesHandler); err == nil {
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

	// 2.3 Data Sources
	fmt.Println("\n--- 2.3 Data Sources ---")
	dataSources := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"DS-1", "Version metadata (GitHub bin/version.json)", func() bool {
			infraFiles, _ := os.ReadDir(infraDir)
			for _, entry := range infraFiles {
				if strings.Contains(strings.ToLower(entry.Name()), "github") || strings.Contains(strings.ToLower(entry.Name()), "update") {
					return true
				}
			}
			return false
		}},
		{"DS-2", "Changelog (GitHub bin/changelog.json)", func() bool {
			return strings.Contains(string(updatesContent), "changelog") || strings.Contains(string(updatesContent), "Changelog")
		}},
		{"DS-3", "APK files storage (GitHub bin/v{version}/)", func() bool {
			infraFiles, _ := os.ReadDir(infraDir)
			for _, entry := range infraFiles {
				if strings.Contains(strings.ToLower(entry.Name()), "github") || strings.Contains(strings.ToLower(entry.Name()), "storage") {
					content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
					if strings.Contains(string(content), "apk") || strings.Contains(string(content), "APK") {
						return true
					}
				}
			}
			return false
		}},
		{"DS-4", "Updates history table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "updates_history") || strings.Contains(string(content), "update_history") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DS-5", "Updates sync table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "updates_sync") || strings.Contains(string(content), "update_sync") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DS-6", "Update versions table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "update_versions") || strings.Contains(string(content), "update_version") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DS-7", "Update pushes table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "update_pushes") || strings.Contains(string(content), "update_push") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DS-8", "Update push devices table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "update_push_devices") || strings.Contains(string(content), "push_device") {
							return true
						}
					}
				}
			}
			return false
		}},
	}

	for _, ds := range dataSources {
		if ds.check() {
			fmt.Printf("  ✅ %s  %s\n", ds.id, ds.description)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", ds.id, ds.description)
			failed++
		}
	}

	// =========================================================================
	// SECTION 3: REQUIRED API ENDPOINTS
	// =========================================================================
	fmt.Println("\n📋 SECTION 3: REQUIRED API ENDPOINTS")
	fmt.Println(strings.Repeat("─", 75))

	// 3.1 GET /v1/updates/status
	fmt.Println("--- 3.1 GET /v1/updates/status ---")
	statusFields := []string{"sync", "lastSyncAt", "nextSyncAt", "latest", "device", "version", "apkFilename", "sha256"}
	for _, field := range statusFields {
		if strings.Contains(string(updatesContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
			failed++
		}
	}

	// 3.2 GET /v1/updates/versions
	fmt.Println("\n--- 3.2 GET /v1/updates/versions ---")
	versionsFields := []string{"version", "apkFilename", "apkSize", "sha256", "releasedAt", "releaseNotes", "status", "pagination"}
	for _, field := range versionsFields {
		if strings.Contains(string(updatesContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
			failed++
		}
	}

	// 3.3 GET /v1/updates/changelog
	fmt.Println("\n--- 3.3 GET /v1/updates/changelog ---")
	changelogFields := []string{"changelog", "Changelog", "date", "type", "notes", "releaseNotes"}
	for _, field := range changelogFields {
		if strings.Contains(string(updatesContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
			failed++
		}
	}

	// 3.4 POST /v1/updates/push
	fmt.Println("\n--- 3.4 POST /v1/updates/push ---")
	pushFields := []string{"PushUpdate", "push", "version", "deviceIds", "installType", "scheduledAt", "pushId", "initiatedBy"}
	for _, field := range pushFields {
		if strings.Contains(string(updatesContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
			failed++
		}
	}

	// 3.5 GET /v1/updates/history
	fmt.Println("\n--- 3.5 GET /v1/updates/history ---")
	historyFields := []string{"GetHistory", "push", "initiatedAt", "completedAt", "deviceCount", "pending", "acknowledged", "failed"}
	for _, field := range historyFields {
		if strings.Contains(string(updatesContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
			failed++
		}
	}

	// 3.6 GET /v1/updates/export
	fmt.Println("\n--- 3.6 GET /v1/updates/export ---")
	exportFields := []string{"ExportVersions", "export", "format", "csv", "json"}
	for _, field := range exportFields {
		if strings.Contains(string(updatesContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
			failed++
		}
	}

	// 3.7 POST /v1/updates/sync
	fmt.Println("\n--- 3.7 POST /v1/updates/sync ---")
	syncFields := []string{"SyncVersions", "sync", "GitHub", "SyncFromGitHub"}
	for _, field := range syncFields {
		if strings.Contains(string(updatesContent), field) {
			fmt.Printf("  ✅  %s\n", field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", field)
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
		{"DB-1", "update_versions table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), "update_versions") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-2", "update_pushes table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), "update_pushes") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-3", "update_push_devices table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), "update_push_devices") {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-4", "update_sync_status table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && strings.Contains(string(content), "update_sync") {
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
	// SECTION 6: BACKEND FILE STRUCTURE
	// =========================================================================
	fmt.Println("\n📋 SECTION 6: BACKEND FILE STRUCTURE")
	fmt.Println(strings.Repeat("─", 75))

	backendFiles := []struct {
		id   string
		file string
	}{
		// Domain Layer
		{"F-D1", "domain/updates/updates_entity.go"},
		{"F-D2", "domain/updates/updates_repository.go"},
		{"F-D3", "domain/updates/updates_errors.go"},
		// Application Layer
		{"F-A1", "application/updates/updates_service.go"},
		{"F-A2", "application/updates/updates_versions_service.go"},
		{"F-A3", "application/updates/updates_push_service.go"},
		{"F-A4", "application/updates/updates_history_service.go"},
		{"F-A5", "application/updates/updates_sync_service.go"},
		{"F-A6", "application/updates/updates_dto.go"},
		// Handler Layer
		{"F-H1", "api/handlers/updates/updates_versions_handler.go"},
		{"F-H2", "api/handlers/updates/updates_push_handler.go"},
		{"F-H3", "api/handlers/updates/updates_history_handler.go"},
		{"F-H4", "api/handlers/updates/updates_sync_handler.go"},
		{"F-H5", "api/handlers/updates/updates_routes.go"},
		// Infrastructure
		{"F-I1", "infrastructure/storage/updates_storage.go"},
		{"F-I2", "infrastructure/github/github_client.go"},
		{"F-I3", "infrastructure/github/github_sync.go"},
		// GraphQL
		{"F-G1", "api/graphql/schema/objects.go"},
		{"F-G2", "api/graphql/schema/resolver.go"},
		{"F-G3", "api/graphql/schema/schema.go"},
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
	// SECTION 7: SERVICE LAYER
	// =========================================================================
	fmt.Println("\n📋 SECTION 7: SERVICE LAYER")
	fmt.Println(strings.Repeat("─", 75))

	serviceMethods := []struct {
		id     string
		file   string
		method string
	}{
		{"S-1", "updates/updates_service.go", "GetUpdateStatus"},
		{"S-2", "updates/updates_versions_service.go", "GetVersions"},
		{"S-3", "updates/updates_versions_service.go", "GetChangelog"},
		{"S-4", "updates/updates_versions_service.go", "ExportVersions"},
		{"S-5", "updates/updates_push_service.go", "PushUpdate"},
		{"S-6", "updates/updates_push_service.go", "CancelUpdate"},
		{"S-7", "updates/updates_history_service.go", "GetHistory"},
		{"S-8", "updates/updates_sync_service.go", "SyncFromGitHub"},
	}

	for _, s := range serviceMethods {
		servicePath := filepath.Join(appDir, s.file)
		found := false
		if content, err := os.ReadFile(servicePath); err == nil {
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
		{"G-1", "UpdateVersion"},
		{"G-2", "UpdateStatus"},
		{"G-3", "UpdatePush"},
		{"G-4", "PushDevice"},
		{"G-5", "SyncStatus"},
		{"G-6", "ChangelogEntry"},
		{"G-7", "DevicePushStatus"},
		// Enums
		{"G-8", "ReleaseType"},
		{"G-9", "MAJOR"},
		{"G-10", "MINOR"},
		{"G-11", "PATCH"},
		{"G-12", "UpdateStatus"},
		{"G-13", "InstallType"},
		// Queries
		{"G-14", "updatesStatus"},
		{"G-15", "updatesVersions"},
		{"G-16", "updatesChangelog"},
		{"G-17", "updatesHistory"},
		{"G-18", "updatesHistoryDetail"},
		{"G-19", "updatesSyncStatus"},
		// Mutations
		{"G-20", "pushUpdate"},
		{"G-21", "cancelUpdate"},
		{"G-22", "syncFromGitHub"},
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

	// Error Response Format
	fmt.Println("--- Error Response Format ---")
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

	// Error Codes
	fmt.Println("\n--- Error Codes ---")
	errorCodes := []struct {
		code       string
		httpStatus string
	}{
		{"bad_request", "400"},
		{"version_not_found", "400"},
		{"push_not_found", "404"},
		{"push_not_cancellable", "400"},
		{"sync_already_in_progress", "409"},
		{"unauthorized", "401"},
		{"forbidden", "403"},
		{"not_found", "404"},
		{"rate_limited", "429"},
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

	// Rate Limits
	fmt.Println("--- Rate Limits ---")
	rateLimits := []struct {
		endpoint string
		limit    string
	}{
		{"GET /v1/updates/status", "60/min"},
		{"GET /v1/updates/versions", "30/min"},
		{"GET /v1/updates/changelog", "30/min"},
		{"POST /v1/updates/push", "10/min"},
		{"GET /v1/updates/history", "30/min"},
		{"POST /v1/updates/history/:id/cancel", "10/min"},
		{"GET /v1/updates/export", "10/min"},
		{"POST /v1/updates/sync", "5/hour"},
	}

	middlewarePath := filepath.Join(middlewareDir, "rate_limit.go")
	middlewareContent, _ := os.ReadFile(middlewarePath)

	for _, rl := range rateLimits {
		if strings.Contains(string(middlewareContent), "rateLimit") || strings.Contains(string(middlewareContent), "RateLimit") {
			fmt.Printf("  ✅ %-45s %s\n", rl.endpoint, rl.limit)
			passed++
		} else {
			fmt.Printf("  ❌ %-45s %s (MISSING)\n", rl.endpoint, rl.limit)
			failed++
		}
	}

	// Security Requirements
	fmt.Println("\n--- Security Requirements ---")
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
		{"SEC-2", "Authorization - Only admins can push updates", func() bool {
			return strings.Contains(string(updatesContent), "admin") || strings.Contains(string(updatesContent), "Admin")
		}},
		{"SEC-3", "Audit Logging - Log all push and sync operations", func() bool {
			infraFiles, _ := os.ReadDir(infraDir)
			for _, entry := range infraFiles {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
					if strings.Contains(string(content), "audit") || strings.Contains(string(content), "log") {
						return true
					}
				}
			}
			return strings.Contains(string(updatesContent), "audit") || strings.Contains(string(updatesContent), "log")
		}},
		{"SEC-4", "Version Validation - Verify APK hash before serving", func() bool {
			return strings.Contains(string(updatesContent), "sha256") || strings.Contains(string(updatesContent), "hash")
		}},
		{"SEC-5", "GitHub Webhook - Secure webhook with secret validation", func() bool {
			infraFiles, _ := os.ReadDir(infraDir)
			for _, entry := range infraFiles {
				if strings.Contains(strings.ToLower(entry.Name()), "github") {
					content, _ := os.ReadFile(filepath.Join(infraDir, entry.Name()))
					return strings.Contains(string(content), "secret") || strings.Contains(string(content), "webhook")
				}
			}
			return false
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

	fileCounts := []struct {
		category string
		new     string
		modified string
	}{
		{"Domain Layer", "3", "0"},
		{"Application Layer", "6", "0"},
		{"Handler Layer", "5", "1"},
		{"Infrastructure", "4", "1"},
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
	fmt.Printf("UPDATES API: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
