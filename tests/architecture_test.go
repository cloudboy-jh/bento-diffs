package bentodiffs

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoInternalImportsInPublicAndCLI(t *testing.T) {
	paths := []string{
		filepath.Join("..", "cmd", "bento-diffs", "main.go"),
		filepath.Join("..", "pkg", "bentodiffs", "api_parse.go"),
		filepath.Join("..", "pkg", "bentodiffs", "api_run.go"),
		filepath.Join("..", "pkg", "bentodiffs", "api_viewer.go"),
	}
	for _, path := range paths {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports %s: %v", path, err)
		}
		for _, imp := range file.Imports {
			name := strings.Trim(imp.Path.Value, "\"")
			if strings.Contains(name, "/internal/") {
				t.Fatalf("internal import forbidden: %s imports %s", path, name)
			}
		}
	}
}

func TestRunPathUsesPublicViewerFactory(t *testing.T) {
	path := filepath.Join("..", "pkg", "bentodiffs", "api_run.go")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	src := string(b)
	if !strings.Contains(src, "NewViewer(ViewerOptions{") {
		t.Fatalf("expected %s to construct viewer via NewViewer", path)
	}
}
