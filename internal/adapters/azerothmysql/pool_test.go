package azerothmysql

import "testing"

func TestEscapeLike(t *testing.T) {
	cases := map[string]string{
		"abc":   "abc",
		"50%":   `50\%`,
		"a_b":   `a\_b`,
		`a\b`:   `a\\b`,
		"100%_": `100\%\_`,
	}
	for input, want := range cases {
		if got := escapeLike(input); got != want {
			t.Fatalf("escapeLike(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestClampLimit(t *testing.T) {
	cases := map[int]int{
		-1:   defaultLimit,
		0:    defaultLimit,
		10:   10,
		1000: maxLimit,
	}
	for input, want := range cases {
		if got := clampLimit(input); got != want {
			t.Fatalf("clampLimit(%d) = %d, want %d", input, got, want)
		}
	}
}

func TestClampOffset(t *testing.T) {
	cases := map[int]int{
		-1:         0,
		0:          0,
		5:          5,
		10_000_000: maxOffset,
	}
	for input, want := range cases {
		if got := clampOffset(input); got != want {
			t.Fatalf("clampOffset(%d) = %d, want %d", input, got, want)
		}
	}
}

func TestOpenPoolRejectsInvalidDSN(t *testing.T) {
	if _, err := openPool("", Config{}, "login"); err == nil {
		t.Fatal("expected error for empty dsn")
	}
	if _, err := openPool("not a dsn", Config{}, "login"); err == nil {
		t.Fatal("expected error for malformed dsn")
	}
	if _, err := openPool("user:pass@tcp(localhost:3306)/db?multiStatements=true", Config{}, "login"); err == nil {
		t.Fatal("expected error for multiStatements dsn")
	}
}
