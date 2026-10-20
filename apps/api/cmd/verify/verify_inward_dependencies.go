package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var layerRank = map[string]int{
	"domain":         1,
	"application":    2,
	"infrastructure": 3,
	"ws":             3,
	"audit":          3,
	"api":            4,
}

var governedLayerNames = []string{"domain", "application", "infrastructure", "ws", "audit", "api"}

type ruleViolation struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type architectureBaseline struct {
	Allowed map[string]int `json:"allowed"`
	Version int            `json:"version"`
}

const architectureBaselineVersion = 1

const modulePrefix = "github.com/VinnsEdesigner/vyzorix/apps/api/internal/"

func internalLayer(importPath string) (string, int, bool) {
	if !strings.HasPrefix(importPath, modulePrefix) {
		return "", 0, false
	}
	rest := importPath[len(modulePrefix):]
	name := rest
	if idx := strings.IndexByte(rest, '/'); idx >= 0 {
		name = rest[:idx]
	}
	rank, ok := layerRank[name]
	if !ok {
		return "", 0, false
	}
	return name, rank, true
}

func verifyArchitectureBoundaries() bool {
	root := filepath.Join(getRoot(), "apps", "api", "internal")
	fset := token.NewFileSet()

	var violations []ruleViolation
	parseFailures := 0

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, string(os.PathSeparator)+"wire"+string(os.PathSeparator)) {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			fmt.Printf("    [ERROR] cannot parse %s: %v\n", path, perr)
			parseFailures++
			return nil
		}

		fromPath := importPathForFile(path)
		fromLayer, fromRank, ok := internalLayer(fromPath)
		if !ok {
			return nil
		}
		for _, imp := range file.Imports {
			target, uerr := strconv.Unquote(imp.Path.Value)
			if uerr != nil {
				continue
			}
			toLayer, toRank, ok := internalLayer(target)
			if !ok {
				continue
			}
			if toRank > fromRank {
				violations = append(violations, ruleViolation{
					From: fromLayer,
					To:   toLayer,
				})
			}
		}
		return nil
	})
	if walkErr != nil {
		fmt.Printf("    [ERROR] walking internal/: %v\n", walkErr)
		return false
	}

	maxAllowed := currentArchitectureBaseline()
	observed := make(map[string]int)
	for _, v := range violations {
		observed[v.From+" -> "+v.To]++
	}

	keys := make([]string, 0, len(observed))
	for k := range observed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	newViolations := 0
	fixedEdges := 0
	for _, k := range keys {
		allowed := maxAllowed.Allowed[k]
		if observed[k] > allowed {
			newViolations++
			fmt.Printf("      [NEW] %s: %d import(s), baseline allows %d\n", k, observed[k], allowed)
		}
	}
	for k, allowed := range maxAllowed.Allowed {
		if observed[k] < allowed {
			fixedEdges++
			fmt.Printf("      [IMPROVED] %s: %d import(s) now, baseline allowed %d - tighten the baseline\n", k, observed[k], allowed)
		}
	}

	total := len(violations)
	if parseFailures > 0 {
		fmt.Printf("    [FAIL] %d file(s) could not be parsed\n", parseFailures)
		return false
	}
	if newViolations > 0 {
		fmt.Printf("    [FAIL] %d new outward dependency edge(s) beyond the baseline\n", newViolations)
		fmt.Printf("    Rule: domain <- application <- infrastructure/ws/audit <- api.\n")
		fmt.Printf("    An inner layer must never import an outer one.\n")
		return false
	}
	if fixedEdges > 0 {
		fmt.Printf("    [FAIL] baseline is looser than reality - %d edge(s) already fixed\n", fixedEdges)
		fmt.Printf("    Lower the counts in config/architecture-baseline.json to lock in the win.\n")
		return false
	}

	fmt.Printf("    [OK] %d layered packages checked, %d tolerated edge(s), 0 new\n", countLayeredPackages(root), total)
	return true
}

func countLayeredPackages(root string) int {
	count := 0
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if !e.IsDir() || !isGovernedLayer(e.Name()) {
			continue
		}
		_ = filepath.WalkDir(filepath.Join(root, e.Name()), func(_ string, d os.DirEntry, err error) error {
			if err == nil && d.IsDir() {
				count++
			}
			return nil
		})
	}
	return count
}

func isGovernedLayer(name string) bool {
	for _, l := range governedLayerNames {
		if l == name {
			return true
		}
	}
	return false
}

func importPathForFile(path string) string {
	rel := filepath.ToSlash(path)
	marker := "/apps/api/internal/"
	idx := strings.Index(rel, marker)
	if idx < 0 {
		return ""
	}
	return modulePrefix + rel[idx+len(marker):]
}

func currentArchitectureBaseline() architectureBaseline {
	baseline := architectureBaseline{
		Version: architectureBaselineVersion,
		Allowed: map[string]int{},
	}
	path := filepath.Join(getRoot(), "apps", "api", "config", "architecture-baseline.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return baseline
	}
	var loaded architectureBaseline
	if err := json.Unmarshal(raw, &loaded); err != nil {
		fmt.Printf("    [WARN] %s is unreadable (%v); treating baseline as empty\n", path, err)
		return baseline
	}
	if loaded.Allowed == nil {
		loaded.Allowed = map[string]int{}
	}
	return loaded
}

func exportArchitectureBaseline() error {
	root := filepath.Join(getRoot(), "apps", "api", "internal")
	fset := token.NewFileSet()
	observed := map[string]int{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		if strings.Contains(path, string(os.PathSeparator)+"wire"+string(os.PathSeparator)) {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		fromLayer, fromRank, ok := internalLayer(importPathForFile(path))
		if !ok {
			return nil
		}
		for _, imp := range file.Imports {
			target, uerr := strconv.Unquote(imp.Path.Value)
			if uerr != nil {
				continue
			}
			toLayer, toRank, ok := internalLayer(target)
			if !ok {
				continue
			}
			if toRank > fromRank {
				observed[fromLayer+" -> "+toLayer]++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	baseline := architectureBaseline{
		Version: architectureBaselineVersion,
		Allowed: observed,
	}
	out, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	dir := filepath.Join(getRoot(), "apps", "api", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "architecture-baseline.json"), out, 0o644)
}
