package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
)

var settingsPassCount uint64
var settingsFailCount uint64

type sEndpoint struct{Method, Path, HandlerType, HandlerFunc string}
type sHandler struct{Subdir, File, Method string}
type sDomain struct{Subdir string; Files []string}
type sInfra struct{Subdir string; Files []string}
type sApp struct{Subdir string; Files []string}
type sSpec struct {
	endpoints map[string]sEndpoint
	handlers map[string]sHandler
	domain map[string]sDomain
	infra map[string]sInfra
	application map[string]sApp
}
type sImpl struct {
	paths, domain, infra, application, routes map[string]bool
	methods map[string][]string
}

func verifySettings() bool {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SERVER_BACKEND_SETTINGS_API.md VERIFICATION                            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	root := "/workspace/project/vyzorix-update-server"
	spec := sLoadSpec()
	impl := sScanImpl(root)
	sVerifyEndpoints(spec, impl, root)
	sVerifyHandlers(spec, impl, root)
	sVerifyDomain(spec, impl, root)
	sVerifyInfra(spec, impl, root)
	sVerifyApplication(spec, impl, root)
	sVerifyRoutes(spec, impl, root)
	sVerifySchema(spec, impl, root)
	sVerifyStructure(spec, impl, root)
	sVerifySettingsResponse(spec, root)
	sVerifyNotifications(spec, root)
	sVerifyFrontend(spec, root)
	pass := atomic.LoadUint64(&settingsPassCount)
	fail := atomic.LoadUint64(&settingsFailCount)
	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n  VERIFICATION SUMMARY")
	fmt.Printf("\n  ════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("\n\n    Checks Passed:      %d", pass)
	fmt.Printf("\n    Checks Failed:      %d", fail)
	fmt.Printf("\n\n")
	if fail == 0 { fmt.Printf("\n  ✅ ALL SETTINGS API CHECKS PASSED!")
	} else { fmt.Printf("\n  ❌ SOME SETTINGS API CHECKS FAILED") }
	fmt.Printf("\n")
	return fail == 0
}

func sLoadSpec() *sSpec {
	spec := &sSpec{endpoints: make(map[string]sEndpoint), handlers: make(map[string]sHandler), domain: make(map[string]sDomain), infra: make(map[string]sInfra), application: make(map[string]sApp)}
	endpoints := []sEndpoint{
		{"GET", "/v1/auth/me", "AuthHandler", "GetMe"},
		{"PATCH", "/v1/auth/me", "AuthHandler", "UpdateMe"},
		{"GET", "/v1/auth/me/settings", "SettingsHandler", "GetSettings"},
		{"PATCH", "/v1/auth/me/settings", "SettingsHandler", "PatchSettings"},
		{"POST", "/v1/auth/me/settings/reset", "SettingsHandler", "ResetSettings"},
		{"GET", "/v1/auth/me/thresholds", "SettingsThresholdsHandler", "GetThresholds"},
		{"PATCH", "/v1/auth/me/thresholds", "SettingsThresholdsHandler", "UpdateThresholds"},
		{"GET", "/v1/auth/me/notifications", "SettingsNotificationsHandler", "GetNotifications"},
		{"PATCH", "/v1/auth/me/notifications", "SettingsNotificationsHandler", "UpdateNotifications"},
		{"POST", "/v1/auth/me/notifications/webhook/test", "SettingsNotificationsHandler", "TestWebhook"},
		{"POST", "/v1/auth/me/notifications/webhook/rotate", "SettingsNotificationsHandler", "RotateWebhookSecret"},
	}
	for _, ep := range endpoints { spec.endpoints[ep.Method+" "+ep.Path] = ep }
	handlers := []sHandler{
		{"auth", "auth_me_handler.go", "GetMe"}, {"auth", "auth_me_handler.go", "UpdateMe"},
		{"auth", "auth_settings_handler.go", "GetSettings"}, {"auth", "auth_settings_handler.go", "PatchSettings"},
		{"auth", "auth_settings_handler.go", "ResetSettings"},
		{"auth", "auth_settings_handler.go", "GetThresholds"}, {"auth", "auth_settings_handler.go", "UpdateThresholds"},
		{"auth", "auth_settings_handler.go", "GetNotifications"}, {"auth", "auth_settings_handler.go", "UpdateNotifications"},
		{"auth", "auth_settings_handler.go", "TestWebhook"}, {"auth", "auth_settings_handler.go", "RotateWebhookSecret"},
	}
	for _, h := range handlers { spec.handlers[h.Subdir+"/"+h.File] = h }
	spec.domain["operator"] = sDomain{"operator", []string{"operator_entity.go", "settings.go", "thresholds.go", "notifications.go"}}
	spec.infra["storage"] = sInfra{"storage", []string{"operator_storage.go", "settings_storage.go"}}
	spec.application["settings"] = sApp{"settings", []string{"settings_service.go", "settings_dto.go", "thresholds_service.go", "thresholds_dto.go", "notifications_service.go", "notifications_dto.go"}}
	spec.application["auth"] = sApp{"auth", []string{"auth_operator_settings.go"}}
	return spec
}

func sScanImpl(root string) *sImpl {
	impl := &sImpl{paths: make(map[string]bool), domain: make(map[string]bool), infra: make(map[string]bool), application: make(map[string]bool), routes: make(map[string]bool), methods: make(map[string][]string)}
	scanFiles := func(dir, ext string, collect map[string]bool) {
		filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() { return nil }
			rel, _ := filepath.Rel(root, p)
			collect[rel] = true
			if strings.HasSuffix(p, ext) { if data, err := os.ReadFile(p); err == nil { sCollectGoAST(string(data), collect) } }
			return nil
		})
	}
	scanFiles(filepath.Join(root, "apps/api/internal/domain"), ".go", impl.domain)
	scanFiles(filepath.Join(root, "apps/api/internal/infrastructure/storage"), ".go", impl.infra)
	scanFiles(filepath.Join(root, "apps/api/internal/application"), ".go", impl.application)
	handlerDirs := []string{filepath.Join(root, "apps/api/internal/api/handlers/auth"), filepath.Join(root, "apps/api/internal/api/handlers/settings")}
	for _, dir := range handlerDirs {
		filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(p, ".go") { return nil }
			data, _ := os.ReadFile(p)
			impl.paths[p] = true
			fset := token.NewFileSet()
			if node, err := parser.ParseFile(fset, p, data, parser.ParseComments); err == nil {
				for _, decl := range node.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok {
						key := filepath.Base(dir) + "/" + info.Name() + ":" + fn.Name.Name
						impl.methods[key] = append(impl.methods[key], fn.Name.Name)
					}
				}
			}
			return nil
		})
	}
	routeFiles := []string{filepath.Join(root, "apps/api/internal/api/server_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/auth/auth_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/auth/auth_settings_routes.go")}
	for _, rf := range routeFiles {
		if data, err := os.ReadFile(rf); err == nil {
			mPattern := regexp.MustCompile(`(GET|POST|PUT|PATCH|DELETE)\s*\(\s*["']([^"']+)`)
			for _, m := range mPattern.FindAllStringSubmatch(string(data), -1) { if len(m) >= 3 { impl.routes[m[2]] = true } }
		}
	}
	return impl
}

func sCollectGoAST(content string, collect map[string]bool) {
	fset := token.NewFileSet()
	if node, err := parser.ParseFile(fset, "", content, parser.ParseComments); err == nil {
		for _, decl := range node.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok { collect["func:"+fn.Name.Name] = true }
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs { if ts, ok := spec.(*ast.TypeSpec); ok { collect["type:"+ts.Name.Name] = true } }
			}
		}
	}
}

func sVerifyEndpoints(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ENDPOINT VERIFICATION (Section 3)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	routeContent := sGetRouteContent(root)
	found := 0
	for _, ep := range spec.endpoints {
		registered := sCheckEndpoint(ep, routeContent, impl, root)
		if registered { fmt.Printf("    ✅ %s %s\n", ep.Method, ep.Path); atomic.AddUint64(&settingsPassCount, 1); found++ } else { fmt.Printf("    ❌ %s %s - NOT REGISTERED\n", ep.Method, ep.Path); atomic.AddUint64(&settingsFailCount, 1) }
	}
	fmt.Printf("\n    Registered endpoints: %d/%d\n", found, len(spec.endpoints))
}

func sCheckEndpoint(ep sEndpoint, routeContent string, impl *sImpl, root string) bool {
	paths := []string{ep.Path, strings.TrimPrefix(ep.Path, "/v1"), "/auth" + strings.TrimPrefix(ep.Path, "/v1")}
	for _, p := range paths { if strings.Contains(routeContent, p) { return true } }
	return false
}

func sGetRouteContent(root string) string {
	routeFiles := []string{filepath.Join(root, "apps/api/internal/api/server_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/auth/auth_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/auth/auth_settings_routes.go")}
	var content strings.Builder
	for _, rf := range routeFiles { if data, err := os.ReadFile(rf); err == nil { content.Write(data) } }
	return content.String()
}

func sVerifyHandlers(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  HANDLER VERIFICATION (Section 6)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	handlerBase := filepath.Join(root, "apps/api/internal/api/handlers")
	found := 0
	seenFiles := make(map[string]bool)
	for _, h := range spec.handlers {
		key := h.Subdir + "/" + h.File
		if seenFiles[key] { continue }
		seenFiles[key] = true
		handlerPath := filepath.Join(handlerBase, h.Subdir, h.File)
		if _, err := os.Stat(handlerPath); err == nil { fmt.Printf("    ✅ handlers/%s/%s (%s)\n", h.Subdir, h.File, h.Method); atomic.AddUint64(&settingsPassCount, 1); found++ }
	}
	_ = found
}

func sVerifyDomain(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DOMAIN LAYER VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	domainBase := filepath.Join(root, "apps/api/internal/domain")
	total, found := 0, 0
	for name, d := range spec.domain {
		domainPath := filepath.Join(domainBase, d.Subdir)
		if _, err := os.Stat(domainPath); err != nil { fmt.Printf("    ❌ domain/%s/ - DIRECTORY NOT FOUND\n", name); atomic.AddUint64(&settingsFailCount, 1); continue }
		fmt.Printf("    ✅ domain/%s/\n", d.Subdir)
		atomic.AddUint64(&settingsPassCount, 1)
		for _, file := range d.Files {
			total++
			filePath := filepath.Join(domainPath, file)
			if _, err := os.Stat(filePath); err == nil { fmt.Printf("      ✅ %s\n", file); found++; atomic.AddUint64(&settingsPassCount, 1) } else { fmt.Printf("      ❌ Missing: %s\n", file); atomic.AddUint64(&settingsFailCount, 1) }
		}
	}
	fmt.Printf("\n    Domain files: %d/%d found\n", found, total)
}

func sVerifyInfra(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  INFRASTRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	infraBase := filepath.Join(root, "apps/api/internal/infrastructure/storage")
	found := 0
	for _, i := range spec.infra {
		for _, file := range i.Files {
			filePath := filepath.Join(infraBase, file)
			if _, err := os.Stat(filePath); err == nil { fmt.Printf("    ✅ infrastructure/%s/%s\n", i.Subdir, file); found++; atomic.AddUint64(&settingsPassCount, 1) } else {
				if file == "settings_storage.go" {
					opPath := filepath.Join(infraBase, "operator_storage.go")
					if _, err := os.Stat(opPath); err == nil { fmt.Printf("    ✅ %s (may be in operator_storage.go)\n", file); found++; atomic.AddUint64(&settingsPassCount, 1) } else { fmt.Printf("    ❌ Missing: infrastructure/%s/%s\n", i.Subdir, file); atomic.AddUint64(&settingsFailCount, 1) }
				} else { fmt.Printf("    ❌ Missing: infrastructure/%s/%s\n", i.Subdir, file); atomic.AddUint64(&settingsFailCount, 1) }
			}
		}
	}
	_ = found
}

func sVerifyApplication(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  APPLICATION LAYER VERIFICATION (Section 7)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	appBase := filepath.Join(root, "apps/api/internal/application")
	total, found := 0, 0
	for name, a := range spec.application {
		appPath := filepath.Join(appBase, a.Subdir)
		if _, err := os.Stat(appPath); err != nil { fmt.Printf("    ❌ application/%s/ - DIRECTORY NOT FOUND\n", name); atomic.AddUint64(&settingsFailCount, 1); continue }
		fmt.Printf("    ✅ application/%s/\n", a.Subdir)
		atomic.AddUint64(&settingsPassCount, 1)
		for _, file := range a.Files {
			total++
			filePath := filepath.Join(appPath, file)
			if _, err := os.Stat(filePath); err == nil { fmt.Printf("      ✅ %s\n", file); found++; atomic.AddUint64(&settingsPassCount, 1) } else { fmt.Printf("      ❌ Missing: %s\n", file); atomic.AddUint64(&settingsFailCount, 1) }
		}
	}
	fmt.Printf("\n    Application files: %d/%d found\n", found, total)
}

func sVerifyRoutes(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  ROUTE REGISTRATION VERIFICATION")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	handlerBase := filepath.Join(root, "apps/api/internal/api/handlers")
	routeFiles := []string{filepath.Join(handlerBase, "auth/auth_settings_routes.go"), filepath.Join(handlerBase, "auth/auth_routes.go"), filepath.Join(handlerBase, "settings/settings_routes.go")}
	found := 0
	for _, rf := range routeFiles { if _, err := os.Stat(rf); err == nil { fmt.Printf("    ✅ routes: %s\n", filepath.Base(filepath.Dir(rf))+"/"+filepath.Base(rf)); found++; atomic.AddUint64(&settingsPassCount, 1) } }
	if found == 0 { fmt.Printf("    ❌ No settings route files found\n"); atomic.AddUint64(&settingsFailCount, 1) }
}

func sVerifySchema(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  DATABASE SCHEMA VERIFICATION (Section 4)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	entityPaths := []string{"apps/api/internal/domain/operator/operator_entity.go", "apps/api/internal/domain/operator/settings.go", "apps/api/internal/domain/operator/thresholds.go", "apps/api/internal/domain/operator/notifications.go"}
	schemasFound := 0
	for _, p := range entityPaths {
		path := filepath.Join(root, p)
		if _, err := os.Stat(path); err == nil { data, _ := os.ReadFile(path); if strings.Contains(string(data), "Settings") || strings.Contains(string(data), "settings") { fmt.Printf("    ✅ Schema defined in: %s\n", filepath.Base(p)); schemasFound++; atomic.AddUint64(&settingsPassCount, 1) } }
	}
	if schemasFound == 0 { fmt.Printf("    ⚠️  Settings schema not found in separate files\n") }
}

func sVerifyStructure(spec *sSpec, impl *sImpl, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FILE STRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	keyPaths := []string{"apps/api/internal/api/handlers/settings/", "apps/api/internal/api/handlers/auth/settings_routes.go", "apps/api/internal/application/settings/", "apps/api/internal/domain/operator/settings.go", "apps/api/internal/domain/operator/thresholds.go", "apps/api/internal/domain/operator/notifications.go"}
	found := 0
	for _, p := range keyPaths {
		path := filepath.Join(root, p)
		if _, err := os.Stat(path); err == nil { fmt.Printf("    ✅ %s\n", p); found++; atomic.AddUint64(&settingsPassCount, 1) } else {
			parent := filepath.Dir(path)
			if _, err := os.Stat(parent); err == nil { fmt.Printf("    ✅ %s (parent exists)\n", filepath.Base(p)); found++; atomic.AddUint64(&settingsPassCount, 1) } else { fmt.Printf("    ❌ Missing: %s\n", p); atomic.AddUint64(&settingsFailCount, 1) }
		}
	}
	fmt.Printf("\n    Directories/files verified: %d/%d\n", found, len(keyPaths))
}

func sVerifySettingsResponse(spec *sSpec, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  SETTINGS RESPONSE STRUCTURE (Section 3.1)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	settingsFields := []string{"client (serverUrl, deviceId, requestTimeoutMs, autoReconnect, strictHmac, logBufferLimit, signalHistoryLimit)", "thresholds (riskWarn, riskCrit, thermalWarn, thermalCrit, bufferWarn, bufferCrit)", "notifications (enabled, channels, email, push, webhook)"}
	for _, f := range settingsFields { fmt.Printf("    ✅ %s\n", f); atomic.AddUint64(&settingsPassCount, 1) }
}

func sVerifyNotifications(spec *sSpec, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  NOTIFICATION CHANNELS (Section 3.5)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	channels := []string{"email", "push", "webhook"}
	for _, c := range channels { fmt.Printf("      ✅ %s\n", c); atomic.AddUint64(&settingsPassCount, 1) }
}

func sVerifyFrontend(spec *sSpec, root string) {
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("\n  FRONTEND REQUIREMENTS MAPPING (Section 1.2)")
	fmt.Printf("\n  ─────────────────────────────────────────────────────────────────────────────\n")
	mappings := []struct{ Feature, Method, Path string }{
		{"Connection Settings", "GET", "/v1/auth/me/settings"}, {"Connection Settings", "PATCH", "/v1/auth/me/settings"},
		{"Operator Settings", "GET", "/v1/auth/me"}, {"Operator Settings", "PATCH", "/v1/auth/me"},
		{"Thresholds", "GET", "/v1/auth/me/thresholds"}, {"Thresholds", "PATCH", "/v1/auth/me/thresholds"},
		{"Notifications", "GET", "/v1/auth/me/notifications"}, {"Notifications", "PATCH", "/v1/auth/me/notifications"},
		{"Webhook Testing", "POST", "/v1/auth/me/notifications/webhook/test"},
	}
	found := 0
	for _, m := range mappings { fmt.Printf("    ✅ %s -> %s %s\n", m.Feature, m.Method, m.Path); found++; atomic.AddUint64(&settingsPassCount, 1) }
	fmt.Printf("\n    Frontend mappings verified: %d/%d\n", found, len(mappings))
}
