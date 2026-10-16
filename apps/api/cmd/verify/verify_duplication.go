package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	duplicationWindowTokens = 60
	duplicationMinLines     = 12
)

type duplicateBlock struct {
	Hash  string   `json:"hash"`
	Files []string `json:"files"`
	Lines int      `json:"lines"`
}

//nolint:govet // fieldalignment: field order chosen for readability, not packing
type duplicationBaseline struct {
	Version int      `json:"version"`
	Allowed []string `json:"allowed"`
}

const duplicationBaselineVersion = 1

var duplicationSkipDirs = map[string]bool{
	"node_modules": true, "dist": true, "build": true, "vendor": true,
	".git": true, "coverage": true, "wire": true, "generated": true,
}

var duplicationExts = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
}

func normalizeTokenStream(src string) []string {
	var tokens []string
	var b strings.Builder
	inLine := false
	inBlock := false
	inStr := byte(0)

	flush := func() {
		if b.Len() > 0 {
			tokens = append(tokens, strings.ToLower(b.String()))
			b.Reset()
		}
	}

	for i := 0; i < len(src); i++ {
		c := src[i]
		if inLine {
			if c == '\n' {
				inLine = false
			}
			continue
		}
		if inBlock {
			if c == '*' && i+1 < len(src) && src[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if inStr != 0 {
			if c == '\\' {
				i++
				continue
			}
			if c == inStr {
				inStr = 0
			}
			continue
		}
		if c == '/' && i+1 < len(src) {
			switch src[i+1] {
			case '/':
				flush()
				inLine = true
				i++
				continue
			case '*':
				flush()
				inBlock = true
				i++
				continue
			}
		}
		if c == '"' || c == '\'' || c == '`' {
			flush()
			inStr = c
			continue
		}
		if isIdentByte(c) {
			b.WriteByte(c)
			continue
		}
		flush()
		if !isSpaceByte(c) {
			tokens = append(tokens, string(c))
		}
	}
	flush()
	return tokens
}

func isIdentByte(c byte) bool {
	return c == '_' || c == '$' ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}

func hashTokens(tokens []string) string {
	h := sha256.New()
	for _, t := range tokens {
		h.Write([]byte(t))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func verifyDuplication() bool {
	root := getRoot()
	seen := map[string][]string{}

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Propagate: a gate that silently skips unreadable input can be
			// bypassed by making a file unreadable.
			return err
		}
		if d.IsDir() {
			if duplicationSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !duplicationExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		tokens := normalizeTokenStream(string(src))
		if len(tokens) < duplicationWindowTokens*2 {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		seenHashes := map[string]bool{}
		for i := 0; i+duplicationWindowTokens <= len(tokens); i += duplicationWindowTokens {
			h := hashTokens(tokens[i : i+duplicationWindowTokens])
			if seenHashes[h] {
				continue
			}
			seenHashes[h] = true
			seen[h] = append(seen[h], rel)
		}
		return nil
	})
	if walkErr != nil {
		fmt.Printf("    [ERROR] walking tree: %v\n", walkErr)
		return false
	}

	allowed := map[string]bool{}
	for _, h := range loadDuplicationBaseline() {
		allowed[h] = true
	}

	var findings []duplicateBlock
	for h, files := range seen {
		if len(files) < 2 {
			continue
		}
		uniq := uniqueStrings(files)
		if len(uniq) < 2 {
			continue
		}
		if allowed[h] {
			continue
		}
		findings = append(findings, duplicateBlock{
			Hash:  h,
			Files: uniq,
			Lines: duplicationMinLines,
		})
	}

	sort.Slice(findings, func(i, j int) bool { return findings[i].Hash < findings[j].Hash })

	if len(findings) == 0 {
		fmt.Printf("    [OK] %d duplicated window(s) tolerated, none new\n", len(allowed))
		return true
	}

	for _, f := range findings {
		fmt.Printf("      %s  %s\n", f.Hash, strings.Join(f.Files, "  |  "))
	}
	fmt.Printf("    [FAIL] %d new duplicated block(s)\n", len(findings))
	fmt.Printf("    Extract the shared logic, or add the hash to config/duplication-baseline.json.\n")
	return false
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func loadDuplicationBaseline() []string {
	path := filepath.Join(getRoot(), "apps", "api", "config", "duplication-baseline.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var b duplicationBaseline
	if json.Unmarshal(raw, &b) != nil {
		return nil
	}
	return b.Allowed
}

func exportDuplicationBaseline() error {
	root := getRoot()
	seen := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if duplicationSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !duplicationExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		tokens := normalizeTokenStream(string(src))
		if len(tokens) < duplicationWindowTokens*2 {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		for i := 0; i+duplicationWindowTokens <= len(tokens); i += duplicationWindowTokens {
			h := hashTokens(tokens[i : i+duplicationWindowTokens])
			seen[h] = append(seen[h], rel)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var allowed []string
	for h, files := range seen {
		if len(uniqueStrings(files)) >= 2 {
			allowed = append(allowed, h)
		}
	}
	sort.Strings(allowed)

	out, err := json.MarshalIndent(duplicationBaseline{
		Version: duplicationBaselineVersion,
		Allowed: allowed,
	}, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	dir := filepath.Join(root, "apps", "api", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "duplication-baseline.json"), out, 0o644)
}
