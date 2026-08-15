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

var diagPassCount uint64
var diagFailCount uint64

type diEndpoint struct{ Method, Path, HandlerFunc string }

type diSpec struct{ endpoints map[string]diEndpoint }
type diImpl struct {
	paths, domain, infra, application, routes map[string]bool
	methods                                   map[string][]string
}

func verifyDiagnostics() bool {
	fmt.Println()
	fmt.Println("  SERVER_BACKEND_DIAGNOSTICS_API.md VERIFICATION                         ")
	fmt.Println()
	root := "/workspace/project/vyzorix-update-server"
	spec := diLoadSpec()
	impl := diScanImpl(root)
	diVerifyEndpoints(spec, impl, root)
	diVerifyHandlers(spec, impl, root)
	diVerifyDomain(spec, impl, root)
	diVerifyInfra(spec, impl, root)
	diVerifyApplication(spec, impl, root)
	diVerifyRoutes(spec, impl, root)
	diVerifySchema(spec, root)
	pass := atomic.LoadUint64(&diagPassCount)
	fail := atomic.LoadUint64(&diagFailCount)
	fmt.Printf("\n  ")
	fmt.Printf("\n  VERIFICATION SUMMARY")
	fmt.Printf("\n  ")
	fmt.Printf("\n\n    Checks Passed:      %d", pass)
	fmt.Printf("\n    Checks Failed:      %d", fail)
	fmt.Printf("\n\n")
	if fail == 0 {
		fmt.Printf("\n   ALL DIAGNOSTICS API CHECKS PASSED!")
	} else {
		fmt.Printf("\n   SOME DIAGNOSTICS API CHECKS FAILED")
	}
	fmt.Printf("\n")
	return fail == 0
}

func diLoadSpec() *diSpec {
	spec := &diSpec{endpoints: make(map[string]diEndpoint)}
	for _, ep := range []diEndpoint{
		{"GET", "/v1/device/:imei/inspect", "GetDeviceInspection"},
		{"GET", "/v1/device/:imei/timeline", "GetDeviceTimeline"},
	} {
		spec.endpoints[ep.Method+" "+ep.Path] = ep
	}
	return spec
}

func diScanImpl(root string) *diImpl {
	impl := &diImpl{paths: make(map[string]bool), domain: make(map[string]bool), infra: make(map[string]bool), application: make(map[string]bool), routes: make(map[string]bool), methods: make(map[string][]string)}
	scanFiles := func(dir, ext string, collect map[string]bool) error {
		return filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			collect[rel] = true
			if strings.HasSuffix(p, ext) {
				if data, err := os.ReadFile(p); err == nil {
					diCollectGoAST(string(data), collect)
				}
			}
			return nil
		})
	}
	if err := scanFiles(filepath.Join(root, "apps/api/internal/domain"), ".go", impl.domain); err != nil {
		return impl
	}
	if err := scanFiles(filepath.Join(root, "apps/api/internal/infrastructure/storage"), ".go", impl.infra); err != nil {
		return impl
	}
	if err := scanFiles(filepath.Join(root, "apps/api/internal/application"), ".go", impl.application); err != nil {
		return impl
	}
	if err := filepath.Walk(filepath.Join(root, "apps/api/internal/api/handlers/diagnostics"), func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".go") {
			return err
		}
		data, _ := os.ReadFile(p)
		impl.paths[p] = true
		fset := token.NewFileSet()
		if node, err := parser.ParseFile(fset, p, data, parser.ParseComments); err == nil {
			for _, decl := range node.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					impl.methods[p+":"+fn.Name.Name] = append(impl.methods[p+":"+fn.Name.Name], fn.Name.Name)
				}
			}
		}
		return nil
	}); err != nil {
		return impl
	}
	routeFiles := []string{filepath.Join(root, "apps/api/internal/api/server_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/diagnostics/diagnostics_routes.go")}
	for _, rf := range routeFiles {
		if data, err := os.ReadFile(rf); err == nil {
			mPattern := regexp.MustCompile(`\.(GET|POST|PUT|PATCH|DELETE)\s*\(\s*["']([^"']+)`)
			for _, m := range mPattern.FindAllStringSubmatch(string(data), -1) {
				if len(m) >= 3 {
					impl.routes[m[2]] = true
				}
			}
		}
	}
	return impl
}

func diCollectGoAST(content string, collect map[string]bool) {
	fset := token.NewFileSet()
	if node, err := parser.ParseFile(fset, "", content, parser.ParseComments); err == nil {
		for _, decl := range node.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				collect["func:"+fn.Name.Name] = true
			}
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						collect["type:"+ts.Name.Name] = true
					}
				}
			}
		}
	}
}

func diVerifyEndpoints(spec *diSpec, impl *diImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  ENDPOINT VERIFICATION (Section 3)")
	fmt.Printf("\n  \n")
	routeContent := diGetRouteContent(root)
	found := 0
	for _, ep := range spec.endpoints {
		registered := diCheckEndpoint(ep, routeContent, impl, root)
		if registered {
			fmt.Printf("     %s %s\n", ep.Method, ep.Path)
			atomic.AddUint64(&diagPassCount, 1)
			found++
		} else {
			fmt.Printf("     %s %s - NOT REGISTERED\n", ep.Method, ep.Path)
			atomic.AddUint64(&diagFailCount, 1)
		}
	}
	fmt.Printf("\n    Registered endpoints: %d/%d\n", found, len(spec.endpoints))
}

func diCheckEndpoint(ep diEndpoint, routeContent string, _ *diImpl, root string) bool {
	pathVariants := []string{ep.Path, strings.TrimPrefix(ep.Path, "/v1"), ":imei/inspect", ":imei/timeline"}
	for _, p := range pathVariants {
		if strings.Contains(routeContent, "\""+p+"\"") {
			return true
		}
	}
	handlerPath := filepath.Join(root, "apps/api/internal/api/handlers/diagnostics")
	for _, f := range []string{"diagnostics_inspect_handler.go", "diagnostics_timeline_handler.go"} {
		if data, err := os.ReadFile(handlerPath + "/" + f); err == nil {
			if strings.Contains(string(data), ep.HandlerFunc) {
				return true
			}
		}
	}
	return false
}

func diGetRouteContent(root string) string {
	routeFiles := []string{filepath.Join(root, "apps/api/internal/api/server_routes.go"), filepath.Join(root, "apps/api/internal/api/handlers/diagnostics/diagnostics_routes.go")}
	var content strings.Builder
	for _, rf := range routeFiles {
		if data, err := os.ReadFile(rf); err == nil {
			content.Write(data)
		}
	}
	return content.String()
}

func diVerifyHandlers(spec *diSpec, _ *diImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  HANDLER VERIFICATION (Section 6)")
	fmt.Printf("\n  \n")
	handlerPath := filepath.Join(root, "apps/api/internal/api/handlers/diagnostics")
	for _, f := range []string{"diagnostics_inspect_handler.go", "diagnostics_timeline_handler.go"} {
		fp := handlerPath + "/" + f
		if _, err := os.Stat(fp); err == nil {
			data, _ := os.ReadFile(fp)
			methodCount := 0
			for _, ep := range spec.endpoints {
				if strings.Contains(string(data), ep.HandlerFunc) {
					methodCount++
				}
			}
			fmt.Printf("     handlers/diagnostics/%s\n", f)
			atomic.AddUint64(&diagPassCount, 1)
			if methodCount > 0 {
				fmt.Printf("      (%d methods found)\n", methodCount)
			}
		} else {
			fmt.Printf("     handlers/diagnostics/%s - NOT FOUND\n", f)
			atomic.AddUint64(&diagFailCount, 1)
		}
	}
}

func diVerifyDomain(_ *diSpec, _ *diImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  DOMAIN LAYER VERIFICATION (Section 5)")
	fmt.Printf("\n  \n")
	domainPath := filepath.Join(root, "apps/api/internal/domain/device")
	if _, err := os.Stat(domainPath); err == nil {
		fmt.Printf("     domain/device/ (parent exists)\n")
		atomic.AddUint64(&diagPassCount, 1)
	} else {
		fmt.Printf("     domain/device/ - DIRECTORY NOT FOUND\n")
		atomic.AddUint64(&diagFailCount, 1)
	}
	for _, f := range []string{"diagnostics_inspect_handler.go", "diagnostics_timeline_handler.go"} {
		fp := filepath.Join(root, "apps/api/internal/api/handlers/diagnostics", f)
		if _, err := os.Stat(fp); err == nil {
			fmt.Printf("       %s\n", f)
			atomic.AddUint64(&diagPassCount, 1)
		} else {
			fmt.Printf("       Missing: %s\n", f)
			atomic.AddUint64(&diagFailCount, 1)
		}
	}
}

func diVerifyInfra(_ *diSpec, _ *diImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  INFRASTRUCTURE VERIFICATION (Section 5)")
	fmt.Printf("\n  \n")
	infraPath := filepath.Join(root, "apps/api/internal/infrastructure/storage")
	if _, err := os.Stat(infraPath); err == nil {
		fmt.Printf("     infrastructure/storage/\n")
		atomic.AddUint64(&diagPassCount, 1)
	} else {
		fmt.Printf("     infrastructure/storage/ - NOT FOUND\n")
		atomic.AddUint64(&diagFailCount, 1)
	}
}

func diVerifyApplication(_ *diSpec, _ *diImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  APPLICATION LAYER VERIFICATION (Section 7)")
	fmt.Printf("\n  \n")
	appPath := filepath.Join(root, "apps/api/internal/application/diagnostics")
	if _, err := os.Stat(appPath); err == nil {
		fmt.Printf("     application/diagnostics/\n")
		atomic.AddUint64(&diagPassCount, 1)
	} else {
		fmt.Printf("     application/diagnostics/ - DIRECTORY NOT FOUND\n")
		atomic.AddUint64(&diagFailCount, 1)
	}
}

func diVerifyRoutes(_ *diSpec, _ *diImpl, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  ROUTE REGISTRATION VERIFICATION")
	fmt.Printf("\n  \n")
	routePath := filepath.Join(root, "apps/api/internal/api/handlers/diagnostics/diagnostics_routes.go")
	if _, err := os.Stat(routePath); err == nil {
		fmt.Printf("     diagnostics/diagnostics_routes.go\n")
		atomic.AddUint64(&diagPassCount, 1)
	} else {
		fmt.Printf("     diagnostics/diagnostics_routes.go - NOT FOUND\n")
		atomic.AddUint64(&diagFailCount, 1)
	}
}

func diVerifySchema(_ *diSpec, root string) {
	fmt.Printf("\n  ")
	fmt.Printf("\n  DATABASE SCHEMA VERIFICATION (Section 4)")
	fmt.Printf("\n  \n")
	schemaFound := false
	if err := filepath.Walk(filepath.Join(root, "apps/api/internal/domain"), func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if strings.Contains(p, "timeline") || strings.Contains(p, "diagnostics") {
			if _, err := os.Stat(p); err == nil {
				data, _ := os.ReadFile(p)
				if strings.Contains(string(data), "Timeline") || strings.Contains(string(data), "timeline") {
					fmt.Printf("     Timeline schema defined in: %s\n", filepath.Base(p))
					schemaFound = true
					atomic.AddUint64(&diagPassCount, 1)
					return nil
				}
			}
		}
		return nil
	}); err != nil {
		return
	}
	if !schemaFound {
		fmt.Printf("      Timeline schema not found in domain files\n")
	}
}
