// Package architecture contains tests that enforce the import boundaries of
// the modular monolith.
package architecture

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go/format"
)

type kind int

const (
	kindOther kind = iota
	kindCore
	kindPlugin
	kindAdapter
)

// TestImportBoundaries walks every Go file in the module, classifies it and
// fails on forbidden imports. It also fails if it inspected nothing, so it
// cannot silently become a no-op.
func TestImportBoundaries(t *testing.T) {
	root := moduleRoot(t)
	modulePath := modulePath(t, root)

	var violations []string
	inspected := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "bin", "node_modules", "web", "docs", "api", "scripts":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel := toSlash(t, root, path)
		fileKind, owner, classified := classify(rel)
		if !classified {
			violations = append(violations, fmt.Sprintf("unclassified package %s (add it to the architecture test)", rel))
			return nil
		}
		inspected++

		for _, imported := range parseImports(t, path) {
			if !strings.HasPrefix(imported, modulePath+"/") {
				continue
			}
			sub := strings.TrimPrefix(imported, modulePath+"/")

			switch {
			case fileKind == kindCore && strings.HasPrefix(sub, "internal/plugins/"):
				violations = append(violations, rel+" imports plugin "+sub)
			case fileKind == kindCore && strings.HasPrefix(sub, "internal/adapters/"):
				violations = append(violations, rel+" imports adapter "+sub)
			case fileKind == kindAdapter && strings.HasPrefix(sub, "internal/plugins/"):
				violations = append(violations, rel+" imports plugin "+sub)
			case fileKind == kindPlugin && strings.HasPrefix(sub, "internal/adapters/"):
				violations = append(violations, rel+" imports adapter "+sub)
			case fileKind == kindPlugin && strings.HasPrefix(sub, "internal/plugins/"):
				if sibling := pluginName(sub); sibling != owner {
					violations = append(violations, rel+" imports sibling plugin "+sub)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if inspected == 0 {
		t.Fatal("architecture test inspected no Go files; the boundary rules are not being enforced")
	}
	if len(violations) > 0 {
		t.Fatalf("architecture violations:\n  %s", strings.Join(violations, "\n  "))
	}
}

func TestSourceIsFormatted(t *testing.T) {
	root := moduleRoot(t)
	skipDirs := map[string]bool{".git": true, "bin": true, "node_modules": true}

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skipDirs[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		formatted, err := format.Source(source)
		if err != nil {
			t.Errorf("parse error in %s: %v", path, err)
			return nil
		}
		if !bytes.Equal(source, formatted) {
			t.Errorf("file is not gofmt-ed: %s", toSlash(t, root, path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

// modulePath reads the module path from go.mod instead of hardcoding it.
func modulePath(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	t.Fatal("module path not found in go.mod")
	return ""
}

// classify returns the kind and owner of a file and whether its package is
// explicitly recognized.
func classify(rel string) (kind, string, bool) {
	switch {
	case strings.HasPrefix(rel, "internal/core/"):
		return kindCore, "", true
	case strings.HasPrefix(rel, "internal/adapters/"):
		return kindAdapter, "", true
	case strings.HasPrefix(rel, "internal/plugins/"):
		return kindPlugin, pluginName(rel), true
	case strings.HasPrefix(rel, "cmd/"):
		return kindOther, "", true
	case strings.HasPrefix(rel, "internal/architecture/"):
		return kindOther, "", true
	case strings.HasPrefix(rel, "internal/fake/"):
		return kindOther, "", true
	case strings.HasPrefix(rel, "migrations/"):
		return kindOther, "", true
	default:
		return kindOther, "", false
	}
}

func pluginName(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "plugins" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func parseImports(t *testing.T, path string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	imports := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		imports = append(imports, strings.Trim(spec.Path.Value, `"`))
	}
	return imports
}

func toSlash(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	return filepath.ToSlash(rel)
}
