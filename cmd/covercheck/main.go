// Command covercheck reports Go coverage over an explicit measured set and
// enforces a ratcheting floor.
//
// It reads one or more Go coverage profiles (the output of
// `go test -coverprofile`), excludes the path globs listed in an ignore file,
// prints a per-package table and the total for the measured set, and exits
// non-zero when the total is below the floor in coverage.floor.
//
// It is intentionally dependency-free and cross-platform: the filtering happens
// in Go, not with grep/awk, so it behaves the same on Windows, macOS and Linux.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "covercheck:", err)
		os.Exit(1)
	}
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type pkgStats struct {
	pkg     string
	total   int64
	covered int64
}

type coverageBlock struct {
	stmts int64
	count int64
}

type summary struct {
	Total             float64          `json:"total"`
	Floor             *float64         `json:"floor,omitempty"`
	Measured          []packageSummary `json:"measured"`
	Excluded          []packageSummary `json:"excluded"`
	TotalStatements   int64            `json:"total_statements"`
	CoveredStatements int64            `json:"covered_statements"`
}

type packageSummary struct {
	Package    string  `json:"package"`
	Statements int64   `json:"statements"`
	Covered    int64   `json:"covered"`
	Coverage   float64 `json:"coverage"`
}

func run(args []string, out io.Writer) error {
	var profiles stringList
	var ignorePath, floorPath string
	var asJSON bool

	fs := flag.NewFlagSet("covercheck", flag.ContinueOnError)
	fs.Var(&profiles, "profile", "coverage profile to read (repeatable)")
	fs.StringVar(&ignorePath, "ignore", "coverage.ignore", "path to the excluded-path list")
	fs.StringVar(&floorPath, "floor", "", "path to a file holding the minimum total percentage")
	fs.BoolVar(&asJSON, "json", false, "print a JSON summary instead of a table")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(profiles) == 0 {
		return errors.New("at least one -profile is required")
	}

	modulePrefix, err := readModulePath()
	if err != nil {
		return err
	}

	blocks := map[string]*coverageBlock{}
	for _, profile := range profiles {
		if err := parseProfile(profile, blocks); err != nil {
			return err
		}
	}

	patterns, err := readIgnore(ignorePath)
	if err != nil {
		return err
	}
	isIgnored := compileIgnore(patterns)

	measured, excluded := split(aggregate(blocks, modulePrefix), isIgnored)
	total, covered := totals(measured)
	measuredPercent := percent(covered, total)

	var floor *float64
	if floorPath != "" {
		value, err := readFloor(floorPath)
		if err != nil {
			return err
		}
		floor = &value
	}

	if asJSON {
		if err := writeJSON(out, summary{
			Total:             measuredPercent,
			Floor:             floor,
			Measured:          summarise(measured),
			Excluded:          summarise(excluded),
			TotalStatements:   total,
			CoveredStatements: covered,
		}); err != nil {
			return err
		}
	} else {
		writeTable(out, measured, excluded, total, covered, measuredPercent)
	}

	if floor != nil && measuredPercent < *floor {
		return fmt.Errorf("measured coverage %.1f%% is below the floor %.1f%%", measuredPercent, *floor)
	}
	return nil
}

func parseProfile(profile string, blocks map[string]*coverageBlock) error {
	file, err := os.Open(profile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		loc, stmts, count, err := parseLine(line)
		if err != nil {
			return fmt.Errorf("%s: %w", profile, err)
		}
		block := blocks[loc]
		if block == nil {
			blocks[loc] = &coverageBlock{stmts: stmts, count: count}
			continue
		}
		if count > block.count {
			block.count = count
		}
	}
	return scanner.Err()
}

// parseLine parses a profile line "<loc> <statements> <count>". The location is
// "<file>:<startLine>.<startCol>,<endLine>.<endCol>".
func parseLine(line string) (string, int64, int64, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return "", 0, 0, fmt.Errorf("unexpected profile line %q", line)
	}
	stmts, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("statements in %q: %w", line, err)
	}
	count, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("count in %q: %w", line, err)
	}
	return fields[0], stmts, count, nil
}

func aggregate(blocks map[string]*coverageBlock, modulePrefix string) map[string]*pkgStats {
	stats := map[string]*pkgStats{}
	for loc, block := range blocks {
		rel := loc
		if i := strings.LastIndex(rel, ":"); i >= 0 {
			rel = rel[:i]
		}
		if modulePrefix != "" {
			rel = strings.TrimPrefix(rel, modulePrefix+"/")
		}
		pkg := path.Dir(filepath.ToSlash(rel))

		s := stats[pkg]
		if s == nil {
			s = &pkgStats{pkg: pkg}
			stats[pkg] = s
		}
		s.total += block.stmts
		if block.count > 0 {
			s.covered += block.stmts
		}
	}
	return stats
}

func split(stats map[string]*pkgStats, isIgnored func(string) bool) (measured, excluded []*pkgStats) {
	for _, s := range stats {
		if isIgnored(s.pkg) {
			excluded = append(excluded, s)
		} else {
			measured = append(measured, s)
		}
	}
	sortStats(measured)
	sortStats(excluded)
	return measured, excluded
}

func sortStats(stats []*pkgStats) {
	sort.Slice(stats, func(i, j int) bool { return stats[i].pkg < stats[j].pkg })
}

func totals(stats []*pkgStats) (total, covered int64) {
	for _, s := range stats {
		total += s.total
		covered += s.covered
	}
	return total, covered
}

func percent(covered, total int64) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(covered) / float64(total)
}

func summarise(stats []*pkgStats) []packageSummary {
	out := make([]packageSummary, 0, len(stats))
	for _, s := range stats {
		out = append(out, packageSummary{
			Package:    s.pkg,
			Statements: s.total,
			Covered:    s.covered,
			Coverage:   percent(s.covered, s.total),
		})
	}
	return out
}

func writeJSON(out io.Writer, value summary) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeTable(out io.Writer, measured, excluded []*pkgStats, total, covered int64, measuredPercent float64) {
	_, _ = fmt.Fprintf(out, "%-60s %14s %10s\n", "PACKAGE", "COVERED/TOTAL", "COVERAGE")
	for _, s := range measured {
		_, _ = fmt.Fprintf(out, "%-60s %14s %9.1f%%\n", s.pkg, formatFraction(s.covered, s.total), percent(s.covered, s.total))
	}
	_, _ = fmt.Fprintf(out, "%-60s %14s %9.1f%%\n", "TOTAL (measured set)", formatFraction(covered, total), measuredPercent)

	if len(excluded) > 0 {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintf(out, "excluded from the denominator:\n")
		for _, s := range excluded {
			_, _ = fmt.Fprintf(out, "  %-58s %14s %9.1f%%\n", s.pkg, formatFraction(s.covered, s.total), percent(s.covered, s.total))
		}
	}
}

func formatFraction(covered, total int64) string {
	return fmt.Sprintf("%d/%d", covered, total)
}

func readModulePath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
				}
			}
			return "", errors.New("module path not found in go.mod")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}

func readIgnore(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, nil
}

func compileIgnore(patterns []string) func(string) bool {
	var regexps []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := globToRegexp(pattern)
		if err != nil {
			continue
		}
		regexps = append(regexps, re)
	}
	return func(pkg string) bool {
		for _, re := range regexps {
			if re.MatchString(pkg) {
				return true
			}
		}
		return false
	}
}

// globToRegexp converts an ignore glob into a regexp anchored at the start that
// matches the path or any path below it. `**` crosses separators, `*` and `?`
// do not.
func globToRegexp(pattern string) (*regexp.Regexp, error) {
	p := strings.TrimSuffix(filepath.ToSlash(pattern), "/")
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(p); i++ {
		switch c := p[i]; c {
		case '*':
			if i+1 < len(p) && p[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("(/.*)?$")
	return regexp.Compile(b.String())
}

func readFloor(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(string(data)), "%"))
	floor, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", path, err)
	}
	return floor, nil
}
