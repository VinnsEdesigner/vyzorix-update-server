package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "arch-baseline":
			if err := exportArchitectureBaseline(); err != nil {
				fmt.Printf("[ERROR] writing architecture baseline: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("[OK] wrote apps/api/config/architecture-baseline.json")
			return
		case "duplication-baseline":
			if err := exportDuplicationBaseline(); err != nil {
				fmt.Printf("[ERROR] writing duplication baseline: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("[OK] wrote apps/api/config/duplication-baseline.json")
			return
		case "deadcode-baseline":
			if err := exportDeadCodeBaseline(); err != nil {
				fmt.Printf("[ERROR] writing dead-code baseline: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("[OK] wrote apps/api/config/deadcode-baseline.json")
			return
		case "arch":
			if !verifyArchitectureBoundaries() {
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			fmt.Println("usage: verify [arch | arch-baseline | duplication-baseline | deadcode-baseline]")
			fmt.Println("  (no argument)         run every gate")
			fmt.Println("  arch                  run only the inward-dependency gate")
			fmt.Println("  arch-baseline         rewrite the frozen dependency baseline")
			fmt.Println("  duplication-baseline  rewrite the frozen duplication baseline")
			fmt.Println("  deadcode-baseline     rewrite the frozen dead-code baseline")
			return
		default:
			fmt.Printf("unknown command %q\n\nusage: verify [arch | arch-baseline]\n", args[0])
			os.Exit(2)
		}
	}

	fmt.Println("VYZORIX structural verification")
	fmt.Printf("  root: %s\n\n", getRoot())

	//nolint:govet // fieldalignment: field order chosen for readability, not packing
	gates := []struct {
		name string
		run  func() bool
	}{
		{"inward dependencies", verifyArchitectureBoundaries},
		{"debt markers cite a ticket", verifyDebtMarkers},
		{"no orphaned source files", verifyNoOrphanedFiles},
		{"no new dead code", verifyDeadCode},
		{"no new duplication", verifyDuplication},
	}

	allPassed := true
	for _, gate := range gates {
		fmt.Printf("  gate: %s\n", gate.name)
		if !gate.run() {
			allPassed = false
		}
		fmt.Println()
	}

	if !allPassed {
		fmt.Println("FAILED")
		os.Exit(1)
	}
	fmt.Println("PASSED")
}

func getRoot() string {
	if env := os.Getenv("VYZORIX_ROOT"); env != "" {
		return env
	}

	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "apps", "api", "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
