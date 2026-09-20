package azerothinfo

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	connectedPlayersPattern = regexp.MustCompile(`Connected players: (\d+)\. Characters in world: (\d+)\.(?: Queue: (\d+)\.)?`)
	connectionPeakPattern   = regexp.MustCompile(`Connection peak: (\d+)\.`)
	uptimePattern           = regexp.MustCompile(`Server uptime: (.+)`)
)

// parseServerInfo extracts the useful fields from the `.server info` output.
// Unparseable fields are left at their zero value.
func parseServerInfo(output string) ServerInfo {
	info := ServerInfo{Output: output}
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		info.Version = strings.TrimSpace(lines[0])
	}
	for _, line := range lines {
		if match := connectedPlayersPattern.FindStringSubmatch(line); match != nil {
			info.ConnectedPlayers = atoi(match[1])
			info.CharactersInWorld = atoi(match[2])
			if match[3] != "" {
				info.Queue = atoi(match[3])
			}
		}
		if match := connectionPeakPattern.FindStringSubmatch(line); match != nil {
			info.ConnectionPeak = atoi(match[1])
		}
		if match := uptimePattern.FindStringSubmatch(line); match != nil {
			info.Uptime = strings.TrimSpace(match[1])
		}
	}
	return info
}

func atoi(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
