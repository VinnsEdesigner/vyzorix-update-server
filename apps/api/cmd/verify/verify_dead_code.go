package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type deadCodeBaseline struct {
	Allowed []string `json:"allowed"`
	Version int      `json:"version"`
}

const deadCodeBaselineVersion = 2

type deadDecl struct {
	id   string
	kind string
	pos  token.Position
}

func loadTypedPackages() ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax |
			packages.NeedModule,
		Dir:   filepath.Join(getRoot(), "apps", "api"),
		Tests: false,
	}
	return packages.Load(cfg, "./internal/...")
}

func collectDeclarations(pkgs []*packages.Package) map[string]deadDecl {
	decls := map[string]deadDecl{}

	for _, pkg := range pkgs {
		if pkg.Types == nil || pkg.TypesInfo == nil {
			continue
		}
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			if !isAnalyzable(name) {
				continue
			}
			obj := scope.Lookup(name)
			key := objectKey(pkg.PkgPath, obj)
			decls[key] = deadDecl{
				id:   key,
				pos:  packagePosition(pkg, obj),
				kind: objectKind(obj),
			}
		}

		for ident, obj := range pkg.TypesInfo.Defs {
			fn, ok := obj.(*types.Func)
			if !ok || fn == nil {
				continue
			}
			sig, ok := fn.Type().(*types.Signature)
			if !ok || sig.Recv() == nil {
				continue
			}
			if !isAnalyzable(fn.Name()) {
				continue
			}
			key := methodKey(pkg.PkgPath, fn)
			decls[key] = deadDecl{
				id:   key,
				pos:  pkg.Fset.Position(ident.Pos()),
				kind: "method",
			}
		}
	}
	return decls
}

func isAnalyzable(name string) bool {
	if name == "" || name == "_" {
		return false
	}
	if name == "main" || name == "init" {
		return false
	}
	r := []rune(name)[0]
	return r >= 'a' && r <= 'z'
}

func objectKey(pkgPath string, obj types.Object) string {
	return pkgPath + "." + obj.Name()
}

func methodKey(_ string, fn *types.Func) string {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return fn.Name()
	}
	recv := sig.Recv().Type()
	if ptr, isPtr := recv.(*types.Pointer); isPtr {
		recv = ptr.Elem()
	}
	return recv.String() + "." + fn.Name()
}

func objectKind(obj types.Object) string {
	switch obj.(type) {
	case *types.Func:
		return "func"
	case *types.TypeName:
		return "type"
	case *types.Var:
		return "var"
	case *types.Const:
		return "const"
	default:
		return "other"
	}
}

func packagePosition(pkg *packages.Package, obj types.Object) token.Position {
	if obj != nil && obj.Pos().IsValid() {
		return pkg.Fset.Position(obj.Pos())
	}
	return token.Position{Filename: pkg.PkgPath}
}

func collectUses(pkgs []*packages.Package) map[string]bool {
	used := map[string]bool{}

	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}
		for _, obj := range pkg.TypesInfo.Uses {
			if obj == nil || obj.Pkg() == nil {
				continue
			}
			key := objectKey(obj.Pkg().Path(), obj)
			used[key] = true
			if fn, ok := obj.(*types.Func); ok {
				if sig, ok := fn.Type().(*types.Signature); ok && sig.Recv() != nil {
					used[methodKey(obj.Pkg().Path(), fn)] = true
				}
			}
		}

		for _, sel := range pkg.TypesInfo.Selections {
			obj := sel.Obj()
			if fn, ok := obj.(*types.Func); ok && fn.Pkg() != nil {
				used[methodKey(fn.Pkg().Path(), fn)] = true
				used[objectKey(fn.Pkg().Path(), fn)] = true
			}
		}
	}
	return used
}

func reachableFromEntryPoints(pkgs []*packages.Package, used map[string]bool) {
	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil || pkg.Syntax == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					if obj, ok := pkg.TypesInfo.Uses[id]; ok && obj != nil && obj.Pkg() != nil {
						used[objectKey(obj.Pkg().Path(), obj)] = true
					}
				}
				return true
			})
		}
	}
}

func verifyDeadCode() bool {
	pkgs, err := loadTypedPackages()
	if err != nil {
		fmt.Printf("    [ERROR] loading packages: %v\n", err)
		return false
	}
	if packages.PrintErrors(pkgs) > 0 {
		fmt.Printf("    [FAIL] the tree does not type-check; fix build errors first\n")
		return false
	}

	decls := collectDeclarations(pkgs)
	used := collectUses(pkgs)
	reachableFromEntryPoints(pkgs, used)

	allowed := map[string]bool{}
	for _, id := range loadDeadCodeBaseline() {
		allowed[id] = true
	}

	var dead []deadDecl
	for id, d := range decls {
		if used[id] || allowed[id] {
			continue
		}
		dead = append(dead, d)
	}
	sort.Slice(dead, func(i, j int) bool {
		if dead[i].pos.Filename != dead[j].pos.Filename {
			return dead[i].pos.Filename < dead[j].pos.Filename
		}
		return dead[i].pos.Line < dead[j].pos.Line
	})

	if len(dead) == 0 {
		fmt.Printf("    [OK] %d unexported declaration(s) type-checked, none unreferenced\n", len(decls))
		return true
	}

	for _, d := range dead {
		fmt.Printf("      %s:%d  %s (%s)\n", displayPath(d.pos.Filename), d.pos.Line, shortID(d.id), d.kind)
	}
	fmt.Printf("    [FAIL] %d unreferenced unexported declaration(s)\n", len(dead))
	fmt.Printf("    Remove them, or freeze with `make deadcode-baseline`.\n")
	return false
}

func displayPath(path string) string {
	root := getRoot()
	if rel, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return path
}

func shortID(id string) string {
	const prefix = "github.com/VinnsEdesigner/vyzorix/apps/api/"
	return strings.TrimPrefix(id, prefix)
}

func loadDeadCodeBaseline() []string {
	path := filepath.Join(getRoot(), "apps", "api", "config", "deadcode-baseline.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var b deadCodeBaseline
	if jsonUnmarshal(raw, &b) != nil {
		return nil
	}
	return b.Allowed
}

func exportDeadCodeBaseline() error {
	pkgs, err := loadTypedPackages()
	if err != nil {
		return err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("the tree does not type-check")
	}

	decls := collectDeclarations(pkgs)
	used := collectUses(pkgs)
	reachableFromEntryPoints(pkgs, used)

	allowed := []string{}
	for id := range decls {
		if !used[id] {
			allowed = append(allowed, id)
		}
	}
	sort.Strings(allowed)

	out, err := jsonMarshalIndent(deadCodeBaseline{
		Version: deadCodeBaselineVersion,
		Allowed: allowed,
	})
	if err != nil {
		return err
	}

	dir := filepath.Join(getRoot(), "apps", "api", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "deadcode-baseline.json"), out, 0o644)
}
