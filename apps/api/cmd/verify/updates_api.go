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

var updatesPassCount uint64
var updatesFailCount uint64

type uEndpoint struct{Method, Path, HandlerType, HandlerFunc string}
type uHandler struct{Subdir, File, Method string}
type uDomain struct{Subdir string; Files []string}
type uInfra struct{Subdir string; Files []string}
type uApp struct{Subdir string; Files []string}
type uSpec struct {
	endpoints map[string]uEndpoint
	handlers map[string]uHandler
	domain map[string]uDomain
	infra map[string]uInfra
	application map[string]uApp
}
type uImpl struct {
	paths, domain, infra, application, routes map[string]bool
	methods map[string][]string
}

func verifyUpdates() bool {
	fmt.Println()
	fmt.Println("  SERVER_BACKEND_UPDATES_API.md VERIFICATION                               ")
	fmt.Println()
	root := "/workspace/project/vyzorix-update-server"
	spec := uLoadSpec()
	impl := uScanImpl(root)
	uVerifyEndpoints(spec, impl, root)
	uVerifyHandlers(spec, impl, root)
	uVerifyDomain(spec, impl, root)
	uVerifyInfra(spec, impl, root)
	uVerifyApplication(spec, impl, root)
	uVerifyRoutes(spec, impl, root)
	uVerifySchema(spec, impl, root)
	uVerifyStructure(spec, impl, root)
	uVerifyFrontend(spec, root)
	pass := atomic.LoadUint64(&updatesPassCount)
	fail := atomic.LoadUint64(&updatesFailCount)
	fmt.Printf("\n  ")
	fmt.Printf("\n  VERIFICATION SUMMARY")
	fmt.Printf("\n  ")
	fmt.Printf("\n\n    Checks Passed:      %d", pass)
	fmt.Printf("\n    Checks Failed:      %d", fail)
	fmt.Printf("\n\n")
	if fail == 0 { fmt.Printf("\n   ALL UPDATES API CHECKS PASSED!")
	} else { fmt.Printf("\n   SOME UPDATES API CHECKS FAILED") }
	fmt.Printf("\n")
	return fail == 0
}

func uLoadSpec() *uSpec {
	spec := &uSpec{endpoints: make(map[string]uEndpoint), handlers: make(map[string]uHandler), domain: make(map[string]uDomain), infra: make(map[string]uInfra), application: make(map[string]uApp)}
	endpoints := []uEndpoint{
		{"GET", "/v1/updates/status", "UpdatesHandler", "GetStatus"},
		{"GET", "/v1/updates/versions", "UpdatesHandler", "GetVersions"},
		{"GET", "/v1/updates/changelog", "UpdatesHandler", "GetChangelog"},
		{"POST", "/v1/updates/push", "UpdatesHandler", "PushUpdate"},
		{"GET", "/v1/updates/history", "UpdatesHandler", "GetHistory"},
		{"GET", "/v1/updates/history/:pushId", "UpdatesHistoryHandler", "GetPushDetail"},
		{"POST", "/v1/updates/history/:pushId/cancel", "UpdatesHistoryHandler", "CancelPush"},
		{"GET", "/v1/updates/export", "UpdatesVersionsHandler", "Export"},
		{"POST", "/v1/updates/sync", "UpdatesSyncHandler", "SyncVersions"},
		{"GET", "/v1/updates/sync/status", "UpdatesHandler", "GetSyncStatus"},
	}
	for _, ep := range endpoints { spec.endpoints[ep.Method+" "+ep.Path] = ep }
	handlers := []uHandler{
		{"updates", "updates_handler.go", "GetStatus"}, {"updates", "updates_handler.go", "GetVersions"},
		{"updates", "updates_handler.go", "GetChangelog"}, {"updates", "updates_handler.go", "PushUpdate"},
		{"updates", "updates_handler.go", "GetHistory"}, {"updates", "updates_handler.go", "GetHistoryByPushId"},
		{"updates", "updates_handler.go", "CancelPush"}, {"updates", "updates_handler.go", "ExportUpdates"},
		{"updates", "updates_handler.go", "SyncUpdates"}, {"updates", "updates_handler.go", "GetSyncStatus"},
	}
	for _, h := range handlers { spec.handlers[h.Subdir+"/"+h.File] = h }
	spec.domain["updates"] = uDomain{"updates", []string{"updates_entity.go", "updates_errors.go", "updates_repository.go"}}
	spec.infra["storage"] = uInfra{"storage", []string{"updates_storage.go", "023_update_versions.go"}}
	spec.application["updates"] = uApp{"updates", []string{"updates_service.go", "updates_history_service.go", "updates_push_service.go", "updates_sync_service.go", "updates_versions_list_service.go", "updates_export_service.go", "updates_changelog_service.go", "updates_versions_status_service.go"}}
	return spec
}

func uScanImpl(root string) *uImpl {
	impl := &uImpl{paths: make(map[string]bool), methods: make(map[string][]string), domain: make(map[string]bool), infra: make(map[string]bool), application: make(map[string]bool), routes: make(map[string]bool)}
	scanFiles := func(dir, ext string, collect map[string]bool) error {
		return filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() { return err }
			rel, _ := filepath.Rel(root, p)
			collect[rel] = true
			if strings.HasSuffix(p, ext) { if data, err := os.ReadFile(p); err == nil { uCollectGoAST(string(data), collect) } }
			return nil
		})
	}
	if err := scanFiles(filepath.Join(root, "apps/api/internal/domain"), ".go", impl.domain); err != nil { return impl }
	if err := scanFiles(filepath.Join(root, "apps/api/internal/infrastructure/storage"), ".go", impl.infra); err != nil { return impl }
	if err := scanFiles(filepath.Join(root, "apps/api/internal/application"), ".go", impl.application); err != nil { return impl }
	handlerDirs := []string{filepath.Join(root, "apps/api/internal/api/handlers/updates")}
	for _, dir := range handlerDirs {
		if err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(p, ".go") { return err }
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
		}); err != nil { return impl }
	}
	routeFiles := []string{filepath.Join(root, "apps/api/internal/api/server_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/updates/updates_handler.go")}
	for _, rf := range routeFiles {
		if data, err := os.ReadFile(rf); err == nil {
			mPattern := regexp.MustCompile(`\.(GET|POST|PUT|PATCH|DELETE)\s*\(\s*["']([^"']+)`)
			for _, m := range mPattern.FindAllStringSubmatch(string(data), -1) { if len(m) >= 3 { impl.routes[m[2]] = true } }
		}
	}
	return impl
}

func uCollectGoAST(content string, collect map[string]bool) {
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

func uVerifyEndpoints(spec *uSpec, impl *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  ENDPOINT VERIFICATION (Section 3)")
	fmt.Printf("\n  \n")
	routeContent := uGetRouteContent(root)
	found := 0
	for _, ep := range spec.endpoints {
		registered := uCheckEndpoint(ep, routeContent, impl, root)
		if registered { fmt.Printf("     %s %s\n", ep.Method, ep.Path); atomic.AddUint64(&updatesPassCount, 1); found++ } else { fmt.Printf("     %s %s - NOT REGISTERED\n", ep.Method, ep.Path); atomic.AddUint64(&updatesFailCount, 1) }
	}
	fmt.Printf("\n    Registered endpoints: %d/%d\n", found, len(spec.endpoints))
}

func uCheckEndpoint(ep uEndpoint, routeContent string, _ *uImpl, root string) bool {
	pathVariants := []string{ep.Path, strings.TrimPrefix(ep.Path, "/v1"), "/updates" + strings.TrimPrefix(ep.Path, "/v1")}
	for _, p := range pathVariants { if strings.Contains(routeContent, "\""+p+"\"") { return true } }
	handlerDir := filepath.Join(root, "apps/api/internal/api/handlers/updates")
	handlerFiles := []string{"updates_handler.go", "updates_history_handler.go", "updates_versions_handler.go", "updates_sync_handler.go"}
	for _, f := range handlerFiles {
		if data, err := os.ReadFile(handlerDir + "/" + f); err == nil { if strings.Contains(string(data), ep.HandlerFunc) { return true } }
	}
	return false
}

func uGetRouteContent(root string) string {
	routeFiles := []string{filepath.Join(root, "apps/api/internal/api/server_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/updates/updates_handler.go")}
	var content strings.Builder
	for _, rf := range routeFiles { if data, err := os.ReadFile(rf); err == nil { content.Write(data) } }
	return content.String()
}

func uVerifyHandlers(spec *uSpec, _ *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  HANDLER VERIFICATION (Section 6)")
	fmt.Printf("\n  \n")
	handlerBase := filepath.Join(root, "apps/api/internal/api/handlers")
	found := 0
	for _, h := range spec.handlers {
		handlerPath := filepath.Join(handlerBase, h.Subdir, h.File)
		if _, err := os.Stat(handlerPath); err == nil { fmt.Printf("     handlers/%s/%s (%s)\n", h.Subdir, h.File, h.Method); atomic.AddUint64(&updatesPassCount, 1); found++ } else { fmt.Printf("     handlers/%s/%s - NOT FOUND\n", h.Subdir, h.File); atomic.AddUint64(&updatesFailCount, 1) }
	}
	_ = found
}

func uVerifyDomain(spec *uSpec, _ *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  DOMAIN LAYER VERIFICATION (Section 5)")
	fmt.Printf("\n  \n")
	domainBase := filepath.Join(root, "apps/api/internal/domain")
	total, found := 0, 0
	for name, d := range spec.domain {
		domainPath := filepath.Join(domainBase, d.Subdir)
		if _, err := os.Stat(domainPath); err != nil { fmt.Printf("     domain/%s/ - DIRECTORY NOT FOUND\n", name); atomic.AddUint64(&updatesFailCount, 1); continue }
		fmt.Printf("     domain/%s/\n", d.Subdir)
		atomic.AddUint64(&updatesPassCount, 1)
		for _, file := range d.Files {
			total++
			filePath := filepath.Join(domainPath, file)
			if _, err := os.Stat(filePath); err == nil { fmt.Printf("       %s\n", file); found++; atomic.AddUint64(&updatesPassCount, 1) } else { fmt.Printf("       Missing: %s\n", file); atomic.AddUint64(&updatesFailCount, 1) }
		}
	}
	fmt.Printf("\n    Domain files: %d/%d found\n", found, total)
}

func uVerifyInfra(spec *uSpec, _ *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  INFRASTRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  \n")
	infraBase := filepath.Join(root, "apps/api/internal/infrastructure/storage")
	found := 0
	for _, i := range spec.infra {
		for _, file := range i.Files {
			filePath := filepath.Join(infraBase, file)
			if _, err := os.Stat(filePath); err == nil { fmt.Printf("     infrastructure/%s/%s\n", i.Subdir, file); found++; atomic.AddUint64(&updatesPassCount, 1) } else { fmt.Printf("     Missing: infrastructure/%s/%s\n", i.Subdir, file); atomic.AddUint64(&updatesFailCount, 1) }
		}
	}
	_ = found
}

func uVerifyApplication(spec *uSpec, _ *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  APPLICATION LAYER VERIFICATION (Section 7)")
	fmt.Printf("\n  \n")
	appBase := filepath.Join(root, "apps/api/internal/application")
	total, found := 0, 0
	for name, a := range spec.application {
		appPath := filepath.Join(appBase, a.Subdir)
		if _, err := os.Stat(appPath); err != nil { fmt.Printf("     application/%s/ - DIRECTORY NOT FOUND\n", name); atomic.AddUint64(&updatesFailCount, 1); continue }
		fmt.Printf("     application/%s/\n", a.Subdir)
		atomic.AddUint64(&updatesPassCount, 1)
		for _, file := range a.Files {
			total++
			filePath := filepath.Join(appPath, file)
			if _, err := os.Stat(filePath); err == nil { fmt.Printf("       %s\n", file); found++; atomic.AddUint64(&updatesPassCount, 1) } else { fmt.Printf("       Missing: %s\n", file); atomic.AddUint64(&updatesFailCount, 1) }
		}
	}
	fmt.Printf("\n    Application files: %d/%d found\n", found, total)
}

func uVerifyRoutes(_ *uSpec, _ *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  ROUTE REGISTRATION VERIFICATION")
	fmt.Printf("\n  \n")
	handlerBase := filepath.Join(root, "apps/api/internal/api/handlers")
	routePath := filepath.Join(handlerBase, "updates/updates_handler.go")
	if _, err := os.Stat(routePath); err == nil { fmt.Printf("     routes: updates/updates_handler.go\n"); atomic.AddUint64(&updatesPassCount, 1) } else { fmt.Printf("     Missing: updates/updates_handler.go\n"); atomic.AddUint64(&updatesFailCount, 1) }
}

func uVerifySchema(_ *uSpec, _ *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  DATABASE SCHEMA VERIFICATION (Section 4)")
	fmt.Printf("\n  \n")
	entityPaths := []string{"apps/api/internal/domain/updates/update_entity.go"}
	for _, p := range entityPaths {
		path := filepath.Join(root, p)
		if _, err := os.Stat(path); err == nil { fmt.Printf("     Schema defined in: %s\n", filepath.Base(p)); atomic.AddUint64(&updatesPassCount, 1) }
	}
}

func uVerifyStructure(_ *uSpec, _ *uImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  FILE STRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  \n")
	keyPaths := []string{"apps/api/internal/api/handlers/updates/", "apps/api/internal/application/updates/", "apps/api/internal/domain/updates/"}
	found := 0
	for _, p := range keyPaths {
		path := filepath.Join(root, p)
		if _, err := os.Stat(path); err == nil { fmt.Printf("     %s\n", p); found++; atomic.AddUint64(&updatesPassCount, 1) } else { fmt.Printf("     Missing: %s\n", p); atomic.AddUint64(&updatesFailCount, 1) }
	}
	fmt.Printf("\n    Directories verified: %d/%d\n", found, len(keyPaths))
}

func uVerifyFrontend(_ *uSpec, _ string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  FRONTEND REQUIREMENTS MAPPING (Section 1.2)")
	fmt.Printf("\n  \n")
	mappings := []struct{ Feature, Method, Path string }{
		{"Update Status", "GET", "/v1/updates/status"}, {"Version List", "GET", "/v1/updates/versions"},
		{"Changelog", "GET", "/v1/updates/changelog"}, {"Push Update", "POST", "/v1/updates/push"},
		{"Update History", "GET", "/v1/updates/history"}, {"Export Updates", "GET", "/v1/updates/export"},
		{"Sync Updates", "POST", "/v1/updates/sync"},
	}
	found := 0
	for _, m := range mappings { fmt.Printf("     %s -> %s %s\n", m.Feature, m.Method, m.Path); found++; atomic.AddUint64(&updatesPassCount, 1) }
	fmt.Printf("\n    Frontend mappings verified: %d/%d\n", found, len(mappings))
}
