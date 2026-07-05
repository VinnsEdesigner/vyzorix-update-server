package main

import (
	
	"os"
	"path/filepath"
	"strings"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func listDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func containsAuthPath(paths []string, pattern string) bool {
	for _, p := range paths {
		if strings.Contains(p, pattern) {
			return true
		}
	}
	return false
}

func findAuthFile(dir, pattern string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(entry.Name(), pattern) {
			return filepath.Join(dir, entry.Name())
		}
	}
	return ""
}

func countFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	return len(entries)
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}





func formatPath(base, rel string) string {
	return filepath.Join(base, rel)
}

func normalizeEndpointPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	return "/" + path
}

func extractHandlerName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".go")
	name = strings.TrimPrefix(name, "auth_")
	name = strings.TrimPrefix(name, "handler_")
	return strings.Title(name)
}
