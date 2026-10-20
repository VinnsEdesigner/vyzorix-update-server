package apierrors

import (
	"go/parser"
	"go/token"
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
//
// Imports are parsed rather than substring-matched on the raw file, so a comment
// or string that merely mentions the package name does not fail the build.
func TestNoDependencyCycle(t *testing.T) {
	root := "../../.." // -> apps/api/internal
	forbidden := []string{
		"internal/api/apierrors",
		"internal/api/responses",
	}

	fset := token.NewFileSet()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Only inspect domain/ and application/ — those layers must not depend on
		// the HTTP adapter.
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if !strings.HasPrefix(rel, "domain/") && !strings.HasPrefix(rel, "application/") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, imp := range file.Imports {
			target := strings.Trim(imp.Path.Value, `"`)
			for _, f := range forbidden {
				if strings.Contains(target, f) {
					t.Errorf("%s imports %q; domain/application must not depend on the HTTP layer", rel, target)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
