package azerothinfo

import "testing"

func TestParseServerInfo(t *testing.T) {
	output := "AzerothCore rev. fake\n" +
		"Connected players: 3. Characters in world: 5.\n" +
		"Connection peak: 10.\n" +
		"Allowed security level: 0\n" +
		"Server uptime: 1 Hour(s) 2 Minute(s)\n" +
		"Update time diff: 0ms. Last 1 diffs summary:"

	info := parseServerInfo(output)
	if info.Version != "AzerothCore rev. fake" {
		t.Fatalf("version = %q", info.Version)
	}
	if info.ConnectedPlayers != 3 || info.CharactersInWorld != 5 {
		t.Fatalf("players = %d/%d", info.ConnectedPlayers, info.CharactersInWorld)
	}
	if info.ConnectionPeak != 10 {
		t.Fatalf("peak = %d", info.ConnectionPeak)
	}
	if info.Uptime != "1 Hour(s) 2 Minute(s)" {
		t.Fatalf("uptime = %q", info.Uptime)
	}
}

func TestParseServerInfoQueue(t *testing.T) {
	info := parseServerInfo("Connected players: 1. Characters in world: 2. Queue: 7.")
	if info.Queue != 7 {
		t.Fatalf("queue = %d", info.Queue)
	}
}

func TestParseServerInfoGarbage(t *testing.T) {
	info := parseServerInfo("not a valid output")
	if info.ConnectedPlayers != 0 || info.Uptime != "" {
		t.Fatalf("info = %+v", info)
	}
}
