package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoRawBubblesImportsInComposition(t *testing.T) {
	paths := []string{"main.go"}
	modelFiles, err := filepath.Glob(filepath.Join("internal", "model", "*.go"))
	if err != nil {
		t.Fatalf("glob model files: %v", err)
	}
	paths = append(paths, modelFiles...)

	for _, path := range paths {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports %s: %v", path, err)
		}
		for _, imp := range file.Imports {
			name := strings.Trim(imp.Path.Value, "\"")
			if strings.HasPrefix(name, "charm.land/bubbles/v2") && name != "charm.land/bubbles/v2/spinner" {
				t.Fatalf("raw bubbles import not allowed in app composition: %s imports %s", path, name)
			}
		}
	}
}

func TestNoThemeCurrentThemeInPageViews(t *testing.T) {
	modelFiles, err := filepath.Glob(filepath.Join("internal", "model", "*.go"))
	if err != nil {
		t.Fatalf("glob model files: %v", err)
	}

	for _, path := range modelFiles {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Name == nil || fn.Name.Name != "View" {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkg, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if pkg.Name == "theme" && sel.Sel != nil && sel.Sel.Name == "CurrentTheme" {
					t.Fatalf("theme.CurrentTheme() is not allowed inside View(): %s", path)
				}
				return true
			})
		}
	}
}
