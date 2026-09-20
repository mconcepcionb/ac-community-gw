package fakeazerothcore

import (
	"fmt"
	"strings"
)

// timeStringToSeconds mirrors AzerothCore's TimeStringToSecs for the common
// units (d, h, m, s). A bare number is treated as seconds. It returns false for
// any input containing an unknown unit, so a malformed duration is rejected
// rather than silently treated as permanent.
func timeStringToSeconds(value string) (int, bool) {
	total := 0
	current := 0
	hasDigits := false
	hasUnit := false
	for _, r := range value {
		if r >= '0' && r <= '9' {
			current = current*10 + int(r-'0')
			hasDigits = true
			continue
		}
		switch r {
		case 'd', 'D':
			total += current * 86400
		case 'h', 'H':
			total += current * 3600
		case 'm', 'M':
			total += current * 60
		case 's', 'S':
			total += current
		default:
			return 0, false
		}
		current = 0
		hasDigits = false
		hasUnit = true
	}
	if hasDigits {
		total += current
	}
	if !hasUnit && !hasDigits {
		return 0, false
	}
	return total, true
}

// secsToTimeString mirrors the shape of AzerothCore's secsToTimeString, e.g.
// "1 Day(s) 2 Hour(s)".
func secsToTimeString(total int) string {
	if total <= 0 {
		return "0 Second(s)"
	}
	days := total / 86400
	total %= 86400
	hours := total / 3600
	total %= 3600
	minutes := total / 60
	seconds := total % 60

	parts := make([]string, 0, 4)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d Day(s)", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d Hour(s)", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d Minute(s)", minutes))
	}
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%d Second(s)", seconds))
	}
	return strings.Join(parts, " ")
}
