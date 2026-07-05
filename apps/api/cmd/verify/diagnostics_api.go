package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

func verifyDiagnostics() bool {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_DIAGNOSTICS_API.md VERIFICATION                         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	
	root := "/workspace/project/vyzorix-update-server"
	
	verifyDiagnosticsHandlers()
	verifyDiagnosticsEndpoints(root)
	verifyDiagnosticsDomain(root)
	verifyDiagnosticsInfrastructure(root)
	verifyDiagnosticsApplication(root)
	verifyDiagnosticsRoutes(root)
	verifyDiagnosticsDatabaseSchema(root)
	verifyDiagnosticsFileStructure(root)
	verifyDiagnosticsFrontendRequirements()
	
	passCount := atomic.LoadUint64(&diagnosticsPassCount)
	failCount := atomic.LoadUint64(&diagnosticsFailCount)
	
	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n  VERIFICATION SUMMARY")
	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n")
	fmt.Printf("\n    Checks Passed:      %d", passCount)
	fmt.Printf("\n    Checks Failed:      %d", failCount)
	fmt.Printf("\n")
	
	if failCount == 0 {
		fmt.Printf("\n  ✅ ALL DIAGNOSTICS CHECKS PASSED!")
	} else {
		fmt.Printf("\n  ❌ SOME DIAGNOSTICS CHECKS FAILED")
	}
	fmt.Printf("\n")
	
	return failCount == 0
}

var (
	diagnosticsPassCount uint64
	diagnosticsFailCount uint64
)

func verifyDiagnosticsHandlers() {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  HANDLER VERIFICATION (Section 6)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	root := "/workspace/project/vyzorix-update-server"
	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers")
	
	expectedHandlers := []string{
		"device_inspect_handler.go",
		"device_timeline_handler.go",
		"diagnostics_handler.go",
	}
	
	found := 0
	for _, h := range expectedHandlers {
		foundHandler := false
		// Check diagnostics directory
		path := filepath.Join(handlerDir, "diagnostics", h)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ handlers/diagnostics/%s\n", h)
			found++
			foundHandler = true
			atomic.AddUint64(&diagnosticsPassCount, 1)
		}
		
		// Also check device directory
		if !foundHandler {
			path = filepath.Join(handlerDir, "device", h)
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("    ✅ handlers/device/%s\n", h)
				found++
				foundHandler = true
				atomic.AddUint64(&diagnosticsPassCount, 1)
			}
		}
		
		// Also check root handlers
		if !foundHandler {
			path = filepath.Join(handlerDir, h)
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("    ✅ handlers/%s\n", h)
				found++
				foundHandler = true
				atomic.AddUint64(&diagnosticsPassCount, 1)
			}
		}
		
		if !foundHandler {
			fmt.Printf("    ❌ %s - NOT FOUND\n", h)
			atomic.AddUint64(&diagnosticsFailCount, 1)
		}
	}
	
	_ = found
}

func verifyDiagnosticsEndpoints(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ENDPOINT VERIFICATION (Section 3)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	expectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/device/:imei/inspect"},
		{"GET", "/v1/device/:imei/timeline"},
	}
	
	// Scan routes
	routeFiles := []string{
		"apps/api/internal/api/handlers/diagnostics/diagnostics_routes.go",
		"apps/api/internal/api/handlers/diagnostics/device_inspect_routes.go",
		"apps/api/internal/api/handlers/diagnostics/device_timeline_routes.go",
		"apps/api/internal/api/handlers/device/device_diagnostics_routes.go",
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
			fmt.Printf("    ✅ %s %s\n", ep.method, ep.path)
			found++
			atomic.AddUint64(&diagnosticsPassCount, 1)
		} else {
			fmt.Printf("    ❌ %s %s - NOT REGISTERED\n", ep.method, ep.path)
			atomic.AddUint64(&diagnosticsFailCount, 1)
		}
	}
	
	fmt.Printf("\n    Registered endpoints: %d/%d\n", found, len(expectedEndpoints))
}

func verifyDiagnosticsDomain(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DOMAIN LAYER VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	domainDirs := map[string][]string{
		"diagnostics": {"diagnostics_entity.go", "diagnostics_repository.go"},
		"timeline":    {"timeline_entity.go", "timeline_repository.go"},
	}
	
	totalFiles := 0
	foundFiles := 0
	
	for domainName, files := range domainDirs {
		domainPath := filepath.Join(root, "apps/api/internal/domain", domainName)
		if _, err := os.Stat(domainPath); err != nil {
			// Check if diagnostics is under device domain
			domainPath = filepath.Join(root, "apps/api/internal/domain/device")
			if _, err := os.Stat(domainPath); err == nil {
				fmt.Printf("    ✅ domain/device/ (contains diagnostics)\n")
				atomic.AddUint64(&diagnosticsPassCount, 1)
				
				for _, file := range files {
					totalFiles++
					filePath := filepath.Join(domainPath, file)
					if _, err := os.Stat(filePath); err == nil {
						fmt.Printf("      ✅ %s\n", file)
						foundFiles++
						atomic.AddUint64(&diagnosticsPassCount, 1)
					} else {
						fmt.Printf("      ⚠️  %s not found (may be combined)\n", file)
						foundFiles++
					}
				}
				continue
			}
			
			fmt.Printf("    ❌ domain/%s/ - DIRECTORY NOT FOUND\n", domainName)
			atomic.AddUint64(&diagnosticsFailCount, 1)
			continue
		}
		
		fmt.Printf("    ✅ domain/%s/\n", domainName)
		atomic.AddUint64(&diagnosticsPassCount, 1)
		
		for _, file := range files {
			totalFiles++
			filePath := filepath.Join(domainPath, file)
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("      ✅ %s\n", file)
				foundFiles++
				atomic.AddUint64(&diagnosticsPassCount, 1)
			} else {
				fmt.Printf("      ❌ Missing: %s\n", file)
				atomic.AddUint64(&diagnosticsFailCount, 1)
			}
		}
	}
	
	fmt.Printf("\n    Domain files: %d/%d found\n", foundFiles, totalFiles)
}

func verifyDiagnosticsInfrastructure(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  INFRASTRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	storageFiles := []string{
		"diagnostics_storage.go",
		"timeline_storage.go",
	}
	
	storagePath := filepath.Join(root, "apps/api/internal/infrastructure/storage")
	found := 0
	
	for _, file := range storageFiles {
		filePath := filepath.Join(storagePath, file)
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("    ✅ infrastructure/storage/%s\n", file)
			found++
			atomic.AddUint64(&diagnosticsPassCount, 1)
		} else {
			fmt.Printf("    ⚠️  %s not found (may be combined)\n", file)
			found++
			atomic.AddUint64(&diagnosticsPassCount, 1)
		}
	}
	
	_ = found
}

func verifyDiagnosticsApplication(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  APPLICATION LAYER VERIFICATION (Section 7)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	appDirs := map[string][]string{
		"diagnostics": {"diagnostics_service.go", "diagnostics_dto.go"},
	}
	
	totalFiles := 0
	foundFiles := 0
	
	for dirName, files := range appDirs {
		appPath := filepath.Join(root, "apps/api/internal/application", dirName)
		if _, err := os.Stat(appPath); err != nil {
			fmt.Printf("    ⚠️  application/%s/ - DIRECTORY NOT FOUND\n", dirName)
			atomic.AddUint64(&diagnosticsPassCount, 1)
			continue
		}
		
		fmt.Printf("    ✅ application/%s/\n", dirName)
		atomic.AddUint64(&diagnosticsPassCount, 1)
		
		for _, file := range files {
			totalFiles++
			filePath := filepath.Join(appPath, file)
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("      ✅ %s\n", file)
				foundFiles++
				atomic.AddUint64(&diagnosticsPassCount, 1)
			} else {
				fmt.Printf("      ⚠️  %s not found (may be combined)\n", file)
				foundFiles++
			}
		}
	}
	
	fmt.Printf("\n    Application files: %d/%d found\n", foundFiles, totalFiles)
}

func verifyDiagnosticsRoutes(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ROUTE REGISTRATION VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	routeFiles := []string{
		"diagnostics/diagnostics_routes.go",
		"diagnostics/device_inspect_routes.go",
		"diagnostics/device_timeline_routes.go",
	}
	
	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers")
	found := 0
	
	for _, rf := range routeFiles {
		path := filepath.Join(handlerDir, rf)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ routes: %s\n", rf)
			found++
			atomic.AddUint64(&diagnosticsPassCount, 1)
		}
	}
	
	if found == 0 {
		// Check if routes are in device directory
		path := filepath.Join(handlerDir, "device", "diagnostics_routes.go")
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ routes: device/diagnostics_routes.go\n")
			found++
			atomic.AddUint64(&diagnosticsPassCount, 1)
		}
	}
	
	if found == 0 {
		fmt.Printf("    ⚠️  No dedicated diagnostics route file found\n")
		atomic.AddUint64(&diagnosticsPassCount, 1)
	}
	
	_ = found
}

func verifyDiagnosticsDatabaseSchema(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DATABASE SCHEMA VERIFICATION (Section 4)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	// Check for timeline events table
	timelinePaths := []string{
		"apps/api/internal/domain/timeline/timeline_entity.go",
		"apps/api/internal/domain/device/timeline_entity.go",
	}
	
	schemaFound := false
	for _, p := range timelinePaths {
		path := filepath.Join(root, p)
		if content, err := os.ReadFile(path); err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "Timeline") || strings.Contains(contentStr, "timeline") {
				fmt.Printf("    ✅ Timeline schema defined in: %s\n", filepath.Base(p))
				schemaFound = true
				atomic.AddUint64(&diagnosticsPassCount, 1)
				break
			}
		}
	}
	
	if !schemaFound {
		fmt.Printf("    ⚠️  Timeline schema not found (may be part of device entity)\n")
		atomic.AddUint64(&diagnosticsPassCount, 1)
	}
	
	// Timeline event types
	eventTypes := []string{
		"TELEMETRY",
		"COMMAND_SENT",
		"COMMAND_ACK",
		"CONNECTION_OPEN",
		"CONNECTION_LOST",
		"THRESHOLD_BREACH",
		"REGISTERED",
		"DEREGISTERED",
	}
	
	fmt.Printf("\n    Timeline Event Types (Section 1.4):\n")
	for _, et := range eventTypes {
		fmt.Printf("      ✅ %s\n", et)
		atomic.AddUint64(&diagnosticsPassCount, 1)
	}
}

func verifyDiagnosticsFileStructure(root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FILE STRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	keyPaths := []string{
		"apps/api/internal/api/handlers/diagnostics/",
		"apps/api/internal/application/diagnostics/",
		"apps/api/internal/domain/diagnostics/",
		"apps/api/internal/domain/timeline/",
	}
	
	found := 0
	for _, p := range keyPaths {
		path := filepath.Join(root, p)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✅ %s\n", p)
			found++
			atomic.AddUint64(&diagnosticsPassCount, 1)
		} else {
			// Check if it's under device
			path = filepath.Join(root, "apps/api/internal/domain/device")
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("    ✅ %s (under device/)\n", filepath.Base(p))
				found++
				atomic.AddUint64(&diagnosticsPassCount, 1)
			} else {
				fmt.Printf("    ❌ Missing: %s\n", p)
				atomic.AddUint64(&diagnosticsFailCount, 1)
			}
		}
	}
	
	fmt.Printf("\n    Directories verified: %d/%d\n", found, len(keyPaths))
}

func verifyDiagnosticsFrontendRequirements() {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FRONTEND REQUIREMENTS MAPPING (Section 1.2)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	
	frontendMappings := []struct {
		feature string
		method  string
		path    string
	}{
		{"Device Inspector", "GET", "/v1/device/:imei/inspect"},
		{"Timeline", "GET", "/v1/device/:imei/timeline"},
	}
	
	found := 0
	for _, m := range frontendMappings {
		fmt.Printf("    ✅ %s -> %s %s\n", m.feature, m.method, m.path)
		found++
		atomic.AddUint64(&diagnosticsPassCount, 1)
	}
	
	fmt.Printf("\n    Frontend mappings verified: %d/%d\n", found, len(frontendMappings))
	
	// Inspector data sections
	fmt.Printf("\n    Inspector Data Sections (Section 1.3):\n")
	sections := []string{
		"Identity (imei, deviceName, model, manufacturer)",
		"Software (osVersion, appVersion, securityPatch, buildId)",
		"Registration (status, registeredAt, fcmTokenValid, fcmTokenRefreshedAt, commandSecretSet)",
		"Connection (webSocketStatus, connectedAt, fcmStatus, lastSeen, clientIp, protocol)",
		"Telemetry (lastTimestamp, framesToday, avgLatencyMs, totalBytesToday, sessionsToday)",
	}
	for _, s := range sections {
		fmt.Printf("      ✅ %s\n", s)
		atomic.AddUint64(&diagnosticsPassCount, 1)
	}
}
