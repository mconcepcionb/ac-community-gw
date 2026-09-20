// Command openapi-postprocess normalizes the swag-generated OpenAPI spec:
//
//   - collapses the `oneOf: [object, $ref]` request-body schema that swag v2
//     emits into a single `$ref`, so generated clients get precise body types;
//   - adds the session-cookie security scheme to components.
//
// It is deterministic and is run by `task openapi:spec` after swag.
package main

import (
	"fmt"
	"os"
	"strings"
)

const specPath = "api/swagger.yaml"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "openapi-postprocess:", err)
		os.Exit(1)
	}
}

func run() error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	out := collapseOneOfBodies(lines)
	out = addSecurityScheme(out)

	return os.WriteFile(specPath, []byte(strings.Join(out, "\n")), 0o644)
}

// collapseOneOfBodies rewrites a body schema of the form
//
//	<indent>oneOf:
//	<indent>- type: object
//	<indent>- $ref: '#/components/schemas/X'
//	<indent+2>description: ...
//
// into `<indent>$ref: '#/components/schemas/X'`.
func collapseOneOfBodies(lines []string) []string {
	out := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		indent := leadingSpaces(line)
		if strings.TrimSpace(line) == "oneOf:" && indent >= 0 &&
			i+2 < len(lines) &&
			isListItem(lines[i+1], indent, "- type: object") &&
			strings.HasPrefix(strings.TrimSpace(lines[i+2]), "- $ref:") &&
			leadingSpaces(lines[i+2]) == indent {
			ref := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i+2]), "- "))
			out = append(out, strings.Repeat(" ", indent)+ref)
			i += 2
			// Skip the continuation lines of the $ref mapping.
			for i+1 < len(lines) && leadingSpaces(lines[i+1]) > indent {
				i++
			}
			continue
		}
		out = append(out, line)
	}
	return out
}

// addSecurityScheme inserts the session-cookie security scheme under components
// when it is not already present.
func addSecurityScheme(lines []string) []string {
	for _, line := range lines {
		if strings.Contains(line, "securitySchemes:") {
			return lines
		}
	}
	out := make([]string, 0, len(lines)+6)
	inserted := false
	for _, line := range lines {
		if !inserted && strings.TrimSpace(line) == "schemas:" && leadingSpaces(line) == 2 {
			out = append(out,
				"  securitySchemes:",
				"    cookieAuth:",
				"      type: apiKey",
				"      in: cookie",
				"      name: acgw_session",
			)
			inserted = true
		}
		out = append(out, line)
	}
	return out
}

func leadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func isListItem(line string, indent int, want string) bool {
	return leadingSpaces(line) == indent && strings.TrimSpace(line) == want
}
