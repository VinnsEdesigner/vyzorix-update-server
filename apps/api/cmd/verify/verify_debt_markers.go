package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//nolint:govet // fieldalignment: field order chosen for readability, not packing
type debtMarker struct {
	Marker string
	File   string
	Line   int
	Text   string
}

var markersRequiringTicket = []string{"TODO", "FIXME", "HACK", "XXX"}

func hasTicketReference(text string) bool {
	if strings.Contains(text, "#") {
		return true
	}
	if strings.Contains(text, "http://") || strings.Contains(text, "https://") {
		return true
	}
	for i := 1; i+1 < len(text); i++ {
		if text[i] != '-' || !isUpperLetter(text[i-1]) || !isDigit(text[i+1]) {
			continue
		}
		j := i - 1
		for j >= 0 && (isUpperLetter(text[j]) || isDigit(text[j])) {
			j--
		}
		if i-j >= 3 {
			return true
		}
	}
	return false
}

func startsWithMarker(upperText, marker string) bool {
	if !strings.HasPrefix(upperText, marker) {
		return false
	}
	rest := upperText[len(marker):]
	if rest == "" {
		return true
	}
	switch rest[0] {
	case ':', ' ', '\t', '(', '-', ',', '.', '!', '/', '\n':
		return true
	default:
		return false
	}
}

func isUpperLetter(b byte) bool { return b >= 'A' && b <= 'Z' }
func isDigit(b byte) bool       { return b >= '0' && b <= '9' }

func verifyDebtMarkers() bool {
	root := filepath.Join(getRoot(), "apps", "api")
	fset := token.NewFileSet()

	var missingTicket []debtMarker
	scanned := 0

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", "wire", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		file, perr := parser.ParseFile(fset, path, src, parser.ParseComments)
		if perr != nil {
			return perr
		}
		scanned++
		rel, _ := filepath.Rel(getRoot(), path)

		for _, group := range file.Comments {
			for _, comment := range group.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
				upper := strings.ToUpper(text)
				for _, marker := range markersRequiringTicket {
					if !startsWithMarker(upper, marker) {
						continue
					}
					if !hasTicketReference(text) {
						missingTicket = append(missingTicket, debtMarker{
							Marker: marker,
							File:   rel,
							Line:   fset.Position(comment.Pos()).Line,
							Text:   text,
						})
					}
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		fmt.Printf("    [ERROR] walking apps/api: %v\n", walkErr)
		return false
	}

	if len(missingTicket) == 0 {
		fmt.Printf("    [OK] %d Go files scanned, every debt marker cites a ticket\n", scanned)
		return true
	}

	sort.Slice(missingTicket, func(i, j int) bool {
		if missingTicket[i].File != missingTicket[j].File {
			return missingTicket[i].File < missingTicket[j].File
		}
		return missingTicket[i].Line < missingTicket[j].Line
	})
	for _, m := range missingTicket {
		fmt.Printf("      %s:%d  %s\n", m.File, m.Line, truncateForReport(m.Text, 72))
	}
	fmt.Printf("    [FAIL] %d debt marker(s) cite no ticket\n", len(missingTicket))
	fmt.Printf("    Add a reference (issue #123, ABC-456, or a URL) so the debt is actionable.\n")
	return false
}

func truncateForReport(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit-1] + "..."
}

func verifyNoOrphanedFiles() bool {
	root := filepath.Join(getRoot(), "apps", "api")
	orphans := 0

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		entries, rerr := os.ReadDir(filepath.Dir(path))
		if rerr != nil {
			return rerr
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
				return nil
			}
		}
		orphans++
		return nil
	})
	if err != nil {
		fmt.Printf("    [ERROR] walking apps/api: %v\n", err)
		return false
	}
	if orphans > 0 {
		fmt.Printf("    [FAIL] %d Go file(s) have no package sibling\n", orphans)
		return false
	}
	fmt.Printf("    [OK] no orphaned Go files\n")
	return true
}
