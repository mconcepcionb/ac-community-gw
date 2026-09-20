package azerothcore

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSafeIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"plain", "Alice", "Alice", false},
		{"underscore", "my_account_1", "my_account_1", false},
		{"trimmed", "  Bob  ", "Bob", false},
		{"empty", "", "", true},
		{"space", "foo bar", "", true},
		{"newline", "foo\n.server shutdown", "", true},
		{"semicolon", "foo;bar", "", true},
		{"quote", `foo"bar`, "", true},
		{"too long", strings.Repeat("a", 33), "", true},
		{"unicode", "üser", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeIdentifier(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				if !errors.Is(err, ErrInvalidIdentifier) {
					t.Fatalf("error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestSafeAccountPassword(t *testing.T) {
	valid := []string{"secret", "P@ssw0rd!", "a", strings.Repeat("x", 16)}
	for _, in := range valid {
		if _, err := SafeAccountPassword(in); err != nil {
			t.Errorf("SafeAccountPassword(%q) error = %v", in, err)
		}
	}
	invalid := []string{"", "with space", "new\nline", `quote"here`, "back\\slash", strings.Repeat("x", 17), "semi;colon"}
	for _, in := range invalid {
		if _, err := SafeAccountPassword(in); !errors.Is(err, ErrInvalidPassword) {
			t.Errorf("SafeAccountPassword(%q) error = %v, want ErrInvalidPassword", in, err)
		}
	}
}

func TestSafeEmail(t *testing.T) {
	valid := []string{"alice@example.com", "a.b+tag@sub.example.co.uk"}
	for _, in := range valid {
		if _, err := SafeEmail(in); err != nil {
			t.Errorf("SafeEmail(%q) error = %v", in, err)
		}
	}
	invalid := []string{"", "not-an-email", "a b@example.com", "a@example", "a\n@example.com"}
	for _, in := range invalid {
		if _, err := SafeEmail(in); !errors.Is(err, ErrInvalidEmail) {
			t.Errorf("SafeEmail(%q) error = %v, want ErrInvalidEmail", in, err)
		}
	}
}

func TestSafeQuotedText(t *testing.T) {
	if got := SafeQuotedText(`hello "world"`); got != "hello world" {
		t.Fatalf("got %q", got)
	}
	if got := SafeQuotedText("line1\nline2\r\n"); strings.ContainsAny(got, "\r\n\"\\") {
		t.Fatalf("unsafe text survived: %q", got)
	}
	long := strings.Repeat("a", 500)
	if got := SafeQuotedText(long); len(got) != MaxCommandTextLength {
		t.Fatalf("len = %d", len(got))
	}
}

func TestSafeBanDuration(t *testing.T) {
	valid := []string{"1d", "12h", "30m", "45s", "2w", "9999d"}
	for _, in := range valid {
		if _, err := SafeBanDuration(in); err != nil {
			t.Errorf("SafeBanDuration(%q) error = %v", in, err)
		}
	}
	invalid := []string{"", "1", "d", "1x", "1d; .server shutdown", "1d2h", "-1d", "99999d"}
	for _, in := range invalid {
		if _, err := SafeBanDuration(in); !errors.Is(err, ErrInvalidDuration) {
			t.Errorf("SafeBanDuration(%q) error = %v, want ErrInvalidDuration", in, err)
		}
	}
}

func TestValidateCommand(t *testing.T) {
	if err := ValidateCommand(".account create A p p@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, bad := range []string{"a\nb", "a\rb", ".x\n.server shutdown"} {
		if err := ValidateCommand(bad); !errors.Is(err, ErrUnsafeCommand) {
			t.Fatalf("ValidateCommand(%q) = %v", bad, err)
		}
	}
}

type recordingExecutor struct{ called bool }

func (r *recordingExecutor) Execute(context.Context, string) (string, error) {
	r.called = true
	return "ok", nil
}

func TestGuardExecutor(t *testing.T) {
	inner := &recordingExecutor{}
	guard := GuardExecutor(inner)
	if _, err := guard.Execute(context.Background(), ".server info"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inner.called {
		t.Fatal("inner executor not called")
	}

	inner2 := &recordingExecutor{}
	guard2 := GuardExecutor(inner2)
	if _, err := guard2.Execute(context.Background(), ".server info\n.server shutdown"); !errors.Is(err, ErrUnsafeCommand) {
		t.Fatalf("error = %v", err)
	}
	if inner2.called {
		t.Fatal("unsafe command reached the inner executor")
	}
}
