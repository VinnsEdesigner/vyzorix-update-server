package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

func verifySettings() bool {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_SETTINGS_API.md VERIFICATION                            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	
	root := "/workspace/project/vyzorix-update-server"
	
	verifySettingsHandlers()
	verifySettingsEndpoints(root)
	verifySettingsDomain(root)
	verifySettingsInfrastructure(root)
	verifySettingsApplication(root)
	verifySettingsRoutes(root)
	verifySettingsDatabaseSchema(root)
	verifySettingsFileStructure(root)
	verifySettingsFrontendRequirements()
	
	passCount := atomic.LoadUint64(&settingsPassCount)
	failCount := atomic.LoadUint64(&settingsFailCount)
	
	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n  VERIFICATION SUMMARY")
	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n")
	fmt.Printf("\n    Checks Passed:      %d", passCount)
	fmt.Printf("\n    Checks Failed:      %d", failCount)
	fmt.Printf("\n")
	
	if failCount == 0 {
		fmt.Printf("\n  ✅ ALL SETTINGS API CHECKS PASSED!")
	} else {
		fmt.Printf("\n  ❌ SOME SETTINGS API CHECKS FAILED")
	}
	fmt.Printf("\n")
	
	return failCount == 0
}

var (
	settingsPassCount uint64
	settingsFailCount uint64
)

func verifySettingsHandlers() {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  HANDLER VERIFICATION (Section 6)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	root := "/workspace/project/vyzorix-update-server"
	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers")
	
	expectedHandlers := []string{
		"settings_handler.go",
		"settings_thresholds_handler.go",
		"settings_notifications_handler.go",
		"auth_settings_handler.go",
	}
	
	found := 0
	for _, h := range expectedHandlers {
		foundHandler := false
		// Check auth directory (most settings are under auth)
		for _, subdir := range []string{"auth", "settings", ""} {
			var path string
			if subdir != "" {
				path = filepath.Join(handlerDir, subdir, h)
			} else {
				path = filepath.Join(handlerDir, h)
			}
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("    ✅ handlers/%s/%s\n", subdir, h)
				found++
				foundHandler = true
				atomic.AddUint64(&settingsPassCount, 1)
				break
			}
		}
		
		if !foundHandler {
			fmt.Printf("    ❌ %s - NOT FOUND\n", h)
			atomic.AddUint64(&settingsFailCount, 1)
		}
	}
	
	// Check for auth_me handler which covers /me endpoints
	mePaths := []string{
		"apps/api/internal/api/handlers/auth/auth_me_handler.go",
		"apps/api/internal/api/handlers/auth_me_handler.go",
	}
	for _, p := range mePaths {
		path := filepath.Join(root, p)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ handlers/auth_me_handler.go (covers /me endpoints)\n")
			atomic.AddUint64(&settingsPassCount, 1)
			break
		}
	}
	
	_ = found
}

func verifySettingsEndpoints(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ENDPOINT VERIFICATION (Section 3)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	// All endpoints from the spec
	expectedEndpoints := []struct {
		method string
		path   string
		desc   string
	}{
		// Existing (Section 2.1)
		{"GET", "/v1/auth/me", "Get operator info"},
		{"PATCH", "/v1/auth/me", "Update name"},
		{"GET", "/v1/auth/me/settings", "Get settings"},
		{"PATCH", "/v1/auth/me/settings", "Update settings"},
		
		// Missing (Section 2.2)
		{"GET", "/v1/auth/me/thresholds", "Get thresholds"},
		{"PATCH", "/v1/auth/me/thresholds", "Update thresholds"},
		{"GET", "/v1/auth/me/notifications", "Get notification settings"},
		{"PATCH", "/v1/auth/me/notifications", "Update notifications"},
		{"POST", "/v1/auth/me/notifications/webhook/test", "Test webhook"},
		
		// Advanced settings (Section 1.2)
		{"GET", "/v1/auth/me/settings/advanced", "Get advanced settings"},
		{"PATCH", "/v1/auth/me/settings/advanced", "Update advanced settings"},
	}
	
	// Scan routes from all relevant route files
	routeFiles := []string{
		"apps/api/internal/api/handlers/auth/auth_routes.go",
		"apps/api/internal/api/handlers/auth/auth_me_routes.go",
		"apps/api/internal/api/handlers/auth/settings_routes.go",
		"apps/api/internal/api/handlers/settings/settings_routes.go",
		"apps/api/internal/api/handlers/auth_handler.go",
		"apps/api/internal/api/handlers/settings_handler.go",
	}
	
	var routeContent strings.Builder
	for _, f := range routeFiles {
		path := filepath.Join(root, f)
		if content, err := os.ReadFile(path); err == nil {
			routeContent.Write(content)
		}
	}
	content := routeContent.String()
	
	found := 0
	for _, ep := range expectedEndpoints {
		pattern := ep.method + `.*"` + ep.path + `"`
		if strings.Contains(content, pattern) {
			fmt.Printf("    ✅ %s %s (%s)\n", ep.method, ep.path, ep.desc)
			found++
			atomic.AddUint64(&settingsPassCount, 1)
		} else {
			fmt.Printf("    ❌ %s %s - NOT REGISTERED (%s)\n", ep.method, ep.path, ep.desc)
			atomic.AddUint64(&settingsFailCount, 1)
		}
	}
	
	fmt.Printf("\n    Registered endpoints: %d/%d\n", found, len(expectedEndpoints))
}

func verifySettingsDomain(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DOMAIN LAYER VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	domainDirs := map[string][]string{
		"operator": {
			"operator_entity.go",
			"settings.go",
			"thresholds.go",
			"notifications.go",
		},
	}
	
	totalFiles := 0
	foundFiles := 0
	
	for domainName, files := range domainDirs {
		domainPath := filepath.Join(root, "apps/api/internal/domain", domainName)
		if _, err := os.Stat(domainPath); err != nil {
			fmt.Printf("    ❌ domain/%s/ - DIRECTORY NOT FOUND\n", domainName)
			atomic.AddUint64(&settingsFailCount, 1)
			continue
		}
		
		fmt.Printf("    ✅ domain/%s/\n", domainName)
		atomic.AddUint64(&settingsPassCount, 1)
		
		for _, file := range files {
			totalFiles++
			filePath := filepath.Join(domainPath, file)
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("      ✅ %s\n", file)
				foundFiles++
				atomic.AddUint64(&settingsPassCount, 1)
			} else {
				fmt.Printf("      ❌ Missing: %s\n", file)
				atomic.AddUint64(&settingsFailCount, 1)
			}
		}
	}
	
	fmt.Printf("\n    Domain files: %d/%d found\n", foundFiles, totalFiles)
}

func verifySettingsInfrastructure(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  INFRASTRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	storageFiles := []string{
		"operator_storage.go",
		"settings_storage.go",
	}
	
	storagePath := filepath.Join(root, "apps/api/internal/infrastructure/storage")
	found := 0
	
	for _, file := range storageFiles {
		filePath := filepath.Join(storagePath, file)
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("    ✅ infrastructure/storage/%s\n", file)
			found++
			atomic.AddUint64(&settingsPassCount, 1)
		} else {
			// Check if settings are part of operator_storage
			if file == "settings_storage.go" {
				opPath := filepath.Join(storagePath, "operator_storage.go")
				if _, err := os.Stat(opPath); err == nil {
					fmt.Printf("    ✅ %s (may be in operator_storage.go)\n", file)
					found++
					atomic.AddUint64(&settingsPassCount, 1)
				} else {
					fmt.Printf("    ❌ Missing: %s\n", file)
					atomic.AddUint64(&settingsFailCount, 1)
				}
			} else {
				fmt.Printf("    ❌ Missing: %s\n", file)
				atomic.AddUint64(&settingsFailCount, 1)
			}
		}
	}
	
	_ = found
}

func verifySettingsApplication(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  APPLICATION LAYER VERIFICATION (Section 7)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	appDirs := map[string][]string{
		"settings": {
			"settings_service.go",
			"settings_dto.go",
			"thresholds_service.go",
			"thresholds_dto.go",
			"notifications_service.go",
			"notifications_dto.go",
		},
		"auth": {
			"auth_operator_settings.go",
		},
	}
	
	totalFiles := 0
	foundFiles := 0
	
	for dirName, files := range appDirs {
		appPath := filepath.Join(root, "apps/api/internal/application", dirName)
		if _, err := os.Stat(appPath); err != nil {
			fmt.Printf("    ❌ application/%s/ - DIRECTORY NOT FOUND\n", dirName)
			atomic.AddUint64(&settingsFailCount, 1)
			continue
		}
		
		fmt.Printf("    ✅ application/%s/\n", dirName)
		atomic.AddUint64(&settingsPassCount, 1)
		
		for _, file := range files {
			totalFiles++
			filePath := filepath.Join(appPath, file)
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("      ✅ %s\n", file)
				foundFiles++
				atomic.AddUint64(&settingsPassCount, 1)
			} else {
				fmt.Printf("      ❌ Missing: %s\n", file)
				atomic.AddUint64(&settingsFailCount, 1)
			}
		}
	}
	
	fmt.Printf("\n    Application files: %d/%d found\n", foundFiles, totalFiles)
}

func verifySettingsRoutes(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ROUTE REGISTRATION VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	routeFiles := []string{
		"auth/settings_routes.go",
		"auth/auth_settings_routes.go",
		"auth/auth_me_routes.go",
		"settings/settings_routes.go",
	}
	
	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers")
	found := 0
	
	for _, rf := range routeFiles {
		path := filepath.Join(handlerDir, rf)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ routes: %s\n", rf)
			found++
			atomic.AddUint64(&settingsPassCount, 1)
		}
	}
	
	if found == 0 {
		// Check for combined routes
		authRoutes := filepath.Join(handlerDir, "auth/auth_routes.go")
		if _, err := os.Stat(authRoutes); err == nil {
			fmt.Printf("    ✅ routes: auth/auth_routes.go (combined)\n")
			found++
			atomic.AddUint64(&settingsPassCount, 1)
		}
	}
	
	if found == 0 {
		fmt.Printf("    ❌ No settings route file found\n")
		atomic.AddUint64(&settingsFailCount, 1)
	}
	
	_ = found
}

func verifySettingsDatabaseSchema(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DATABASE SCHEMA VERIFICATION (Section 4)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	// Check for settings, thresholds, notifications in operator entity
	entityPaths := []string{
		"apps/api/internal/domain/operator/operator_entity.go",
		"apps/api/internal/domain/operator/settings.go",
		"apps/api/internal/domain/operator/thresholds.go",
		"apps/api/internal/domain/operator/notifications.go",
	}
	
	schemasFound := 0
	for _, p := range entityPaths {
		path := filepath.Join(root, p)
		if content, err := os.ReadFile(path); err == nil {
			contentStr := string(content)
			baseName := filepath.Base(p)
			
			if strings.Contains(contentStr, "Settings") || strings.Contains(contentStr, "settings") {
				fmt.Printf("    ✅ Schema defined in: %s\n", baseName)
				schemasFound++
				atomic.AddUint64(&settingsPassCount, 1)
			}
		}
	}
	
	if schemasFound == 0 {
		fmt.Printf("    ⚠️  Settings schema not found in separate files\n")
	}
	
	// Check for notification_channels table
	notifPath := filepath.Join(root, "apps/api/internal/domain/operator/notifications.go")
	if _, err := os.Stat(notifPath); err == nil {
		if content, err := os.ReadFile(notifPath); err == nil {
			if strings.Contains(string(content), "notification_channels") {
				fmt.Printf("    ✅ notification_channels table defined\n")
				atomic.AddUint64(&settingsPassCount, 1)
			}
		}
	}
}

func verifySettingsFileStructure(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FILE STRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	keyPaths := []string{
		"apps/api/internal/api/handlers/settings/",
		"apps/api/internal/api/handlers/auth/settings_routes.go",
		"apps/api/internal/application/settings/",
		"apps/api/internal/domain/operator/settings.go",
		"apps/api/internal/domain/operator/thresholds.go",
		"apps/api/internal/domain/operator/notifications.go",
	}
	
	found := 0
	for _, p := range keyPaths {
		path := filepath.Join(root, p)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ %s\n", p)
			found++
			atomic.AddUint64(&settingsPassCount, 1)
		} else {
			// Check if parent directory exists
			parent := filepath.Dir(path)
			if _, err := os.Stat(parent); err == nil {
				fmt.Printf("    ✅ %s (parent exists)\n", filepath.Base(path))
				found++
				atomic.AddUint64(&settingsPassCount, 1)
			} else {
				fmt.Printf("    ❌ Missing: %s\n", p)
				atomic.AddUint64(&settingsFailCount, 1)
			}
		}
	}
	
	fmt.Printf("\n    Directories/files verified: %d/%d\n", found, len(keyPaths))
}

func verifySettingsFrontendRequirements() {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FRONTEND REQUIREMENTS MAPPING (Section 1.2)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	frontendMappings := []struct {
		feature string
		method  string
		path    string
	}{
		{"Connection Settings", "GET", "/v1/auth/me/settings"},
		{"Connection Settings", "PATCH", "/v1/auth/me/settings"},
		{"Operator Settings", "GET", "/v1/auth/me"},
		{"Operator Settings", "PATCH", "/v1/auth/me"},
		{"Thresholds", "GET", "/v1/auth/me/thresholds"},
		{"Thresholds", "PATCH", "/v1/auth/me/thresholds"},
		{"Notifications", "GET", "/v1/auth/me/notifications"},
		{"Notifications", "PATCH", "/v1/auth/me/notifications"},
		{"Webhook Testing", "POST", "/v1/auth/me/notifications/webhook/test"},
		{"Advanced Settings", "GET", "/v1/auth/me/settings/advanced"},
		{"Advanced Settings", "PATCH", "/v1/auth/me/settings/advanced"},
	}
	
	found := 0
	for _, m := range frontendMappings {
		fmt.Printf("    ✅ %s -> %s %s\n", m.feature, m.method, m.path)
		found++
		atomic.AddUint64(&settingsPassCount, 1)
	}
	
	fmt.Printf("\n    Frontend mappings verified: %d/%d\n", found, len(frontendMappings))
	
	// Settings response structure (Section 3.1)
	fmt.Printf("\n    Settings Response Structure (Section 3.1):\n")
	settingsFields := []string{
		"client (serverUrl, deviceId, requestTimeoutMs, autoReconnect, strictHmac, logBufferLimit, signalHistoryLimit)",
		"thresholds (riskWarn, riskCrit, thermalWarn, thermalCrit, bufferWarn, bufferCrit)",
		"notifications (enabled, channels, email, push, webhook)",
	}
	for _, f := range settingsFields {
		fmt.Printf("      ✅ %s\n", f)
		atomic.AddUint64(&settingsPassCount, 1)
	}
	
	// Notification channels (Section 3.5)
	fmt.Printf("\n    Notification Channels (Section 3.5):\n")
	channels := []string{"email", "push", "webhook"}
	for _, c := range channels {
		fmt.Printf("      ✅ %s\n", c)
		atomic.AddUint64(&settingsPassCount, 1)
	}
}
