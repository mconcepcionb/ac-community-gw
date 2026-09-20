package azerothcore

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrUnsafeCommand is returned by ValidateCommand when a command string is not
// safe to hand to AzerothCore.
var ErrUnsafeCommand = errors.New("azerothcore: unsafe command")

// Validation errors returned by the Safe* constructors.
var (
	// ErrInvalidIdentifier is returned when an account/character identifier is
	// not safe to embed in a console command.
	ErrInvalidIdentifier = errors.New("azerothcore: invalid identifier")
	// ErrInvalidPassword is returned when a password is not safe to embed in a
	// console command.
	ErrInvalidPassword = errors.New("azerothcore: invalid password")
	// ErrInvalidEmail is returned when an email is not safe to embed in a
	// console command.
	ErrInvalidEmail = errors.New("azerothcore: invalid email")
	// ErrInvalidDuration is returned when a ban/mute duration is malformed.
	ErrInvalidDuration = errors.New("azerothcore: invalid duration")
)

const (
	// MaxAccountPasswordLength matches the AzerothCore account password limit.
	MaxAccountPasswordLength = 16
	// MaxCommandTextLength caps free-text command arguments.
	MaxCommandTextLength = 200
	// MaxIdentifierLength caps account/character identifiers.
	MaxIdentifierLength = 32
	// MaxEmailLength caps email arguments.
	MaxEmailLength = 254
)

// identifierPattern accepts only characters that cannot alter a console
// command: letters, digits and underscore.
var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,32}$`)

// passwordPattern is deliberately a whitelist: it excludes whitespace, quotes,
// backslashes and shell-ish metacharacters so a password cannot terminate or
// extend the command argument list.
var passwordPattern = regexp.MustCompile(`^[A-Za-z0-9!@#$%^&*()_+\-=\[\]{}:,.?/]{1,16}$`)

// emailPattern is a conservative email grammar without whitespace or command
// metacharacters.
var emailPattern = regexp.MustCompile(`^[A-Za-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)+$`)

// durationPattern is a bounded number followed by a single AzerothCore unit
// (seconds, minutes, hours, days, weeks).
var durationPattern = regexp.MustCompile(`^\d{1,4}[smhdw]$`)

// SafeIdentifier validates an account or character identifier.
func SafeIdentifier(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if !identifierPattern.MatchString(trimmed) {
		return "", fmt.Errorf("%w: must be 1-%d of [A-Za-z0-9_]", ErrInvalidIdentifier, MaxIdentifierLength)
	}
	return trimmed, nil
}

// SafeAccountPassword validates an AzerothCore account password.
func SafeAccountPassword(value string) (string, error) {
	if len([]rune(value)) > MaxAccountPasswordLength || !passwordPattern.MatchString(value) {
		return "", fmt.Errorf("%w: must be 1-%d safe characters", ErrInvalidPassword, MaxAccountPasswordLength)
	}
	return value, nil
}

// SafeEmail validates an email address used in a console command.
func SafeEmail(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > MaxEmailLength || !emailPattern.MatchString(trimmed) {
		return "", fmt.Errorf("%w: not a valid email address", ErrInvalidEmail)
	}
	return trimmed, nil
}

// SafeBanDuration validates a ban or mute duration such as `1d`, `12h`, `30m`.
// An empty value is invalid here; callers apply their own default first.
func SafeBanDuration(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if !durationPattern.MatchString(trimmed) {
		return "", fmt.Errorf("%w: expected <number><s|m|h|d|w>", ErrInvalidDuration)
	}
	return trimmed, nil
}

// SafeQuotedText sanitizes free text embedded in a quoted command argument.
func SafeQuotedText(value string) string {
	replacer := strings.NewReplacer(`"`, "", `\`, "", "\r", " ", "\n", " ")
	cleaned := strings.TrimSpace(replacer.Replace(value))
	if len(cleaned) > MaxCommandTextLength {
		cleaned = cleaned[:MaxCommandTextLength]
	}
	return cleaned
}

// ValidateCommand rejects commands that could smuggle a second console command.
// It is a transport-level, defense-in-depth guard; plugins must still validate
// every dynamic value before building a command.
func ValidateCommand(command string) error {
	if strings.ContainsAny(command, "\r\n") {
		return fmt.Errorf("%w: command contains a line break", ErrUnsafeCommand)
	}
	return nil
}

// GuardExecutor wraps a CommandExecutor and refuses to transport commands that
// contain line breaks.
func GuardExecutor(inner CommandExecutor) CommandExecutor {
	if inner == nil {
		return nil
	}
	return guardedExecutor{inner: inner}
}

type guardedExecutor struct {
	inner CommandExecutor
}

func (g guardedExecutor) Execute(ctx context.Context, command string) (string, error) {
	if err := ValidateCommand(command); err != nil {
		return "", err
	}
	return g.inner.Execute(ctx, command)
}
