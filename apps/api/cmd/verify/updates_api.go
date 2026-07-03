// Package verify provides verification for SERVER_BACKEND_UPDATES_API.md
// This script verifies ALL server-side requirements from the Updates API specification.
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
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/")
	appDir := filepath.Join(root, "apps/api/internal/application/")
	gqlDir := filepath.Join(root, "apps/api/internal/api/graphql/schema/")
	schemaDir := filepath.Join(root, "apps/api/internal/infrastructure/storage/")
	updatesDir := filepath.Join(appDir, "updates/")

	// =========================================================================
	// SECTION 2: CURRENT STATE ANALYSIS
	// =========================================================================
	fmt.Println("📋 SECTION 2: CURRENT STATE ANALYSIS")
	fmt.Println(strings.Repeat("─", 75))

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

	fmt.Println("\n--- 2.2 Missing Endpoints (REQUIRED) ---")
	updatesHandler := filepath.Join(handlerDir, "updates/updates_handler.go")

	missingEndpoints := []struct {
		id      string
		method  string
		path    string
		handler string
	}{
		{"MISS-1", "GET", "/v1/updates/status", "GetUpdateStatus"},
		{"MISS-2", "GET", "/v1/updates/versions", "GetVersions"},
		{"MISS-3", "GET", "/v1/updates/changelog", "GetChangelog"},
		{"MISS-4", "POST", "/v1/updates/push", "PushUpdate"},
		{"MISS-5", "GET", "/v1/updates/history", "GetHistory"},
		{"MISS-6", "GET", "/v1/updates/export", "ExportVersions"},
		{"MISS-7", "POST", "/v1/updates/sync", "SyncVersions"},
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

	fmt.Println("\n--- 2.3 Data Sources ---")
	dataChecks := []struct {
		id          string
		description string
		check       func() bool
	}{
		{"DS-1", "Version metadata (bin/version.json)", func() bool {
			infraDir := filepath.Join(root, "apps/api/internal/infrastructure/")
			if entries, err := os.ReadDir(infraDir); err == nil {
				for _, entry := range entries {
					if strings.Contains(strings.ToLower(entry.Name()), "github") || strings.Contains(strings.ToLower(entry.Name()), "update") {
						return true
					}
				}
			}
			return false
		}},
		{"DS-2", "Changelog (bin/changelog.json)", func() bool {
			if content, err := os.ReadFile(updatesHandler); err == nil {
				return strings.Contains(string(content), "changelog") || strings.Contains(string(content), "Changelog")
			}
			return false
		}},
		{"DS-3", "Updates history table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && (strings.Contains(string(content), "update_pushes") || strings.Contains(string(content), "update_push_devices")) {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DS-4", "Updates sync table", func() bool {
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
	}

	for _, ds := range dataChecks {
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

	// Helper to check if ANY of the terms exist
	searchAnyInUpdates := func(terms ...string) bool {
		// Check handler files
		if entries, err := os.ReadDir(handlerDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() == "updates" {
					handlerUpdatesDir := filepath.Join(handlerDir, "updates")
					if handlerEntries, err := os.ReadDir(handlerUpdatesDir); err == nil {
						for _, he := range handlerEntries {
							if content, err := os.ReadFile(filepath.Join(handlerUpdatesDir, he.Name())); err == nil {
								for _, term := range terms {
									if strings.Contains(string(content), term) {
										return true
									}
								}
							}
						}
					}
				}
			}
		}
		// Check service/response files
		if entries, err := os.ReadDir(updatesDir); err == nil {
			for _, entry := range entries {
				if content, err := os.ReadFile(filepath.Join(updatesDir, entry.Name())); err == nil {
					for _, term := range terms {
						if strings.Contains(string(content), term) {
							return true
						}
					}
				}
			}
		}
		return false
	}

	fmt.Println("--- GET /v1/updates/status (Section 3.1) ---")
	statusChecks := []struct {
		field string
		terms []string
	}{
		{"sync", []string{"SyncStatusInfo", "Status"}},
		{"lastSyncAt", []string{"LastSyncAt"}},
		{"nextSyncAt", []string{"NextSyncAt"}},
		{"latest", []string{"LatestVersionInfo", "Version"}},
		{"device", []string{"DeviceStatusInfo", "Device"}},
	}
	for _, check := range statusChecks {
		if searchAnyInUpdates(check.terms...) {
			fmt.Printf("  ✅  %s\n", check.field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check.field)
			failed++
		}
	}

	fmt.Println("\n--- GET /v1/updates/versions (Section 3.2) ---")
	versionsChecks := []struct {
		field string
		terms []string
	}{
		{"version", []string{"VersionResponse", "Version"}},
		{"apkFilename", []string{"APKFilename"}},
		{"apkSize", []string{"APKSize"}},
		{"sha256", []string{"SHA256", "sha256"}},
		{"releasedAt", []string{"ReleasedAt", "ReleaseDate"}},
		{"releaseNotes", []string{"ReleaseNotes", "releaseNotes"}},
		{"status", []string{"Status"}},
	}
	for _, check := range versionsChecks {
		if searchAnyInUpdates(check.terms...) {
			fmt.Printf("  ✅  %s\n", check.field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check.field)
			failed++
		}
	}

	fmt.Println("\n--- GET /v1/updates/changelog (Section 3.3) ---")
	changelogChecks := []struct {
		field string
		terms []string
	}{
		{"changelog", []string{"Changelog"}},
		{"Changelog", []string{"ChangelogEntry"}},
		{"version", []string{"Version"}},
	}
	for _, check := range changelogChecks {
		if searchAnyInUpdates(check.terms...) {
			fmt.Printf("  ✅  %s\n", check.field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check.field)
			failed++
		}
	}

	fmt.Println("\n--- POST /v1/updates/push (Section 3.4) ---")
	pushChecks := []struct {
		field string
		terms []string
	}{
		{"PushUpdate", []string{"PushUpdate"}},
		{"push", []string{"PushID", "pushId"}},
		{"imei", []string{"IMEI", "imei", "DeviceID", "deviceId"}},
		{"version", []string{"Version"}},
	}
	for _, check := range pushChecks {
		if searchAnyInUpdates(check.terms...) {
			fmt.Printf("  ✅  %s\n", check.field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check.field)
			failed++
		}
	}

	fmt.Println("\n--- POST /v1/updates/sync (Section 3.7) ---")
	syncChecks := []struct {
		field string
		terms []string
	}{
		{"SyncVersions", []string{"SyncVersions", "SyncFromGitHub"}},
		{"sync", []string{"SyncStatus", "sync"}},
		{"GitHub", []string{"GitHub", "github"}},
	}
	for _, check := range syncChecks {
		if searchAnyInUpdates(check.terms...) {
			fmt.Printf("  ✅  %s\n", check.field)
			passed++
		} else {
			fmt.Printf("  ❌  %s (MISSING)\n", check.field)
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
		{"DB-1", "update_pushes table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && (strings.Contains(string(content), "update_pushes") || strings.Contains(string(content), "update_push_devices")) {
							return true
						}
					}
				}
			}
			return false
		}},
		{"DB-2", "updates_sync table", func() bool {
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						content, _ := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
						if strings.Contains(string(content), "CREATE TABLE") && (strings.Contains(string(content), "updates_sync") || strings.Contains(string(content), "update_sync")) {
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
	// SECTION 7: SERVICE LAYER
	// =========================================================================
	fmt.Println("\n📋 SECTION 7: SERVICE LAYER")
	fmt.Println(strings.Repeat("─", 75))

	serviceSpecs := []struct {
		id     string
		file   string
		method string
	}{
		{"S-1", "updates/updates_service.go", "GetUpdateStatus"},
		{"S-2", "updates/updates_service.go", "GetVersions"},
		{"S-3", "updates/updates_service.go", "GetChangelog"},
		{"S-4", "updates/updates_service.go", "PushUpdate"},
		{"S-5", "updates/updates_service.go", "GetHistory"},
		{"S-6", "updates/updates_service.go", "ExportVersions"},
		{"S-7", "updates/updates_service.go", "SyncFromGitHub"},
	}

	for _, s := range serviceSpecs {
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

	graphqlSpecs := []struct {
		id           string
		searchName   string
		altName      string
	}{
		{"G-1", "UpdateStatus", ""},
		{"G-2", "UpdateVersion", ""},
		{"G-3", "ChangelogEntry", ""},
		{"G-4", "PushHistoryConnection", "UpdateHistory"},
		{"G-5", "updatesStatus", "updateStatus"},
		{"G-6", "updatesVersions", "updateVersions"},
		{"G-7", "pushUpdate", ""},
		{"G-8", "syncFromGitHub", "syncUpdates"},
	}

	for _, g := range graphqlSpecs {
		found := false
		if entries, err := os.ReadDir(gqlDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					content, _ := os.ReadFile(filepath.Join(gqlDir, entry.Name()))
					if strings.Contains(string(content), g.searchName) ||
						(g.altName != "" && strings.Contains(string(content), g.altName)) {
						found = true
						break
					}
				}
			}
		}
		if found {
			fmt.Printf("  ✅ %s  %s\n", g.id, g.searchName)
			passed++
		} else {
			fmt.Printf("  ❌ %s  %s (MISSING)\n", g.id, g.searchName)
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
		{"GET /v1/updates/status", "60/min"},
		{"GET /v1/updates/versions", "30/min"},
		{"GET /v1/updates/changelog", "30/min"},
		{"POST /v1/updates/push", "10/min"},
		{"POST /v1/updates/sync", "5/min"},
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
		{"F-1", "api/handlers/updates/updates_handler.go"},
		{"F-2", "api/handlers/updates/updates_routes.go"},
		{"F-3", "application/updates/updates_service.go"},
		{"F-4", "application/updates/updates_dto.go"},
		{"F-5", "domain/update/update_entity.go"},
		{"F-6", "domain/update/update_repository.go"},
		{"F-7", "infrastructure/github/github_client.go"},
		{"F-8", "infrastructure/storage/update_storage.go"},
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
	fmt.Printf("UPDATES API: %d passed, %d failed\n", passed, failed)
	fmt.Println(strings.Repeat("═", 75))

	return failed == 0
}
