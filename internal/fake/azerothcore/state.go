package fakeazerothcore

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Account is the mutable state the double tracks per AzerothCore account.
type Account struct {
	Username string `json:"username"`
	// Password is never serialized: the double does not authenticate account
	// passwords and must not expose them through /state or the dashboard.
	Password    string `json:"-"`
	Email       string `json:"email"`
	GMLevel     int    `json:"gm_level"`
	Banned      bool   `json:"banned"`
	BanDuration string `json:"ban_duration,omitempty"`
	BanReason   string `json:"ban_reason,omitempty"`
}

// Seed is the JSON document accepted by LoadSeed and returned by Snapshot.
type Seed struct {
	Accounts []Account `json:"accounts"`
}

// LoadSeed reads an initial account set from a JSON file.
func LoadSeed(path string) ([]Account, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("fakeazerothcore: read seed: %w", err)
	}
	var seed Seed
	if err := json.Unmarshal(data, &seed); err != nil {
		return nil, fmt.Errorf("fakeazerothcore: parse seed: %w", err)
	}
	return seed.Accounts, nil
}

// Snapshot returns the accounts sorted by username.
func (s *Server) Snapshot() Seed {
	s.mu.Lock()
	defer s.mu.Unlock()
	accounts := make([]Account, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, *account)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].Username < accounts[j].Username })
	return Seed{Accounts: accounts}
}

func (s *Server) getAccount(name string) (*Account, bool) {
	account, ok := s.accounts[normalizeAccountKey(name)]
	return account, ok
}

func (s *Server) upsertAccount(account Account) {
	s.accounts[normalizeAccountKey(account.Username)] = &account
}

func normalizeAccountKey(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}
