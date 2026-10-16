package apierrors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoDependencyCycle is the layering guard: apierrors is an adapter that may
// import domain/application, but none of them may import it back. If a domain or
// application package ever gains an apierrors import the architecture has
// inverted and the build would (currently) succeed while the layering silently
// rots, so it is asserted here rather than left to code review.
func TestNoDependencyCycle(t *testing.T) {
	root := "../../.." // -> apps/api/internal
	forbidden := []string{
		"api/apierrors",
		"api/responses",
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Only inspect domain/ and application/ — those layers must not depend on
		// the HTTP adapter.
		rel, _ := filepath.Rel(root, path)
		if !strings.HasPrefix(rel, "domain/") && !strings.HasPrefix(rel, "application/") {
			return nil
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, f := range forbidden {
			if strings.Contains(string(body), f) {
				t.Errorf("%s imports %q; domain/application must not depend on the HTTP layer", rel, f)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
