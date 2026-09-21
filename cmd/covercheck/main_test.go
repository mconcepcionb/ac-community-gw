package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlobToRegexp(t *testing.T) {
	cases := []struct {
		pattern string
		pkg     string
		want    bool
	}{
		{"internal/plugins/*/repository/generated", "internal/plugins/apikeys/repository/generated", true},
		{"internal/plugins/*/repository/generated", "internal/plugins/apikeys/repository/generated/sub", true},
		{"internal/plugins/*/repository/generated", "internal/plugins/apikeys/repository", false},
		{"cmd/", "cmd/covercheck", true},
		{"internal/architecture", "internal/architecture", true},
		{"internal/fake/azerothcore", "internal/fake/azerothcore", true},
		{"internal/fake/azerothcore", "internal/core/audit", false},
	}
	for _, tc := range cases {
		re, err := globToRegexp(tc.pattern)
		if err != nil {
			t.Fatalf("globToRegexp(%q): %v", tc.pattern, err)
		}
		if got := re.MatchString(tc.pkg); got != tc.want {
			t.Errorf("pattern %q against %q = %v, want %v", tc.pattern, tc.pkg, got, tc.want)
		}
	}
}

func TestParseLine(t *testing.T) {
	loc, stmts, count, err := parseLine("example.com/x/a.go:1.2,3.4 7 1")
	if err != nil {
		t.Fatalf("parseLine: %v", err)
	}
	if loc != "example.com/x/a.go:1.2,3.4" || stmts != 7 || count != 1 {
		t.Fatalf("parseLine = %q,%d,%d", loc, stmts, count)
	}

	if _, _, _, err := parseLine("not a profile line"); err == nil {
		t.Fatal("expected an error for a malformed line")
	}
	if _, _, _, err := parseLine("a.go:1.2,3.4 x 1"); err == nil {
		t.Fatal("expected an error for a non-numeric statement count")
	}
}

func TestRunMeasuresAndExcludes(t *testing.T) {
	dir := t.TempDir()
	profile := writeFile(t, dir, "coverage.txt", strings.Join([]string{
		"mode: set",
		"github.com/mconcepcionb/ac-community-gw/internal/core/config/config.go:1.1,2.1 4 1",
		"github.com/mconcepcionb/ac-community-gw/internal/core/config/config.go:3.1,4.1 4 0",
		"github.com/mconcepcionb/ac-community-gw/internal/plugins/apikeys/repository/generated/db.go:1.1,2.1 6 1",
	}, "\n"))
	ignore := writeFile(t, dir, "coverage.ignore", "internal/plugins/*/repository/generated\n")

	var out bytes.Buffer
	if err := run([]string{"-profile", profile, "-ignore", ignore}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "internal/core/config") {
		t.Errorf("measured package missing from output:\n%s", got)
	}
	if !strings.Contains(got, "TOTAL (measured set)") {
		t.Errorf("total missing from output:\n%s", got)
	}
	if !strings.Contains(got, "excluded from the denominator") {
		t.Errorf("excluded section missing from output:\n%s", got)
	}
	// 4 of 8 measured statements are covered; the generated package must not
	// change the denominator.
	if !strings.Contains(got, "50.0%") {
		t.Errorf("expected 50.0%% measured coverage:\n%s", got)
	}
}

func TestRunFloor(t *testing.T) {
	dir := t.TempDir()
	profile := writeFile(t, dir, "coverage.txt", strings.Join([]string{
		"mode: set",
		"github.com/mconcepcionb/ac-community-gw/internal/core/config/config.go:1.1,2.1 4 1",
		"github.com/mconcepcionb/ac-community-gw/internal/core/config/config.go:3.1,4.1 4 0",
	}, "\n"))
	ignore := writeFile(t, dir, "coverage.ignore", "")

	failing := writeFile(t, dir, "floor-high.txt", "60\n")
	if err := run([]string{"-profile", profile, "-ignore", ignore, "-floor", failing}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected the floor check to fail for 50% coverage against a 60% floor")
	}

	passing := writeFile(t, dir, "floor-low.txt", "40%\n")
	if err := run([]string{"-profile", profile, "-ignore", ignore, "-floor", passing}, &bytes.Buffer{}); err != nil {
		t.Fatalf("expected the floor check to pass: %v", err)
	}
}

func TestRunMergesProfiles(t *testing.T) {
	dir := t.TempDir()
	first := writeFile(t, dir, "a.txt", strings.Join([]string{
		"mode: set",
		"github.com/mconcepcionb/ac-community-gw/internal/core/config/config.go:1.1,2.1 4 0",
	}, "\n"))
	second := writeFile(t, dir, "b.txt", strings.Join([]string{
		"mode: set",
		"github.com/mconcepcionb/ac-community-gw/internal/core/config/config.go:1.1,2.1 4 1",
	}, "\n"))
	ignore := writeFile(t, dir, "coverage.ignore", "")

	var out bytes.Buffer
	if err := run([]string{"-profile", first, "-profile", second, "-ignore", ignore}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), "100.0%") {
		t.Errorf("merging the same block should take the max count, got:\n%s", out.String())
	}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
