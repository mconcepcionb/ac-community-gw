package fakeazerothcore

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	maxAccountName = 17
	maxPassword    = 16
	maxEmail       = 255
)

// execute parses an AzerothCore command string and returns the result text the
// real worldserver would produce.
func (s *Server) execute(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return msgIncorrectSyntax
	}
	switch strings.ToLower(fields[0]) {
	case ".server":
		return s.serverCommand(fields)
	case ".account":
		return s.accountCommand(fields)
	case ".ban":
		return s.banCommand(fields)
	case ".unban":
		return s.unbanCommand(fields)
	default:
		return fmt.Sprintf(msgCommandNotFound, fields[0])
	}
}

func (s *Server) serverCommand(fields []string) string {
	if len(fields) >= 2 && strings.EqualFold(fields[1], "info") {
		return s.serverInfo()
	}
	return msgIncorrectSyntax
}

func (s *Server) serverInfo() string {
	uptime := secsToTimeString(int(s.now().Sub(s.started).Seconds()))
	lines := []string{
		"AzerothCore rev. fake-azerothcore (ac-community-gw dummy)",
		"Connected players: 0. Characters in world: 0.",
		"Connection peak: 0.",
		"Allowed security level: 0",
		fmt.Sprintf(msgServerUptime, uptime),
		"Update time diff: 0ms. Last 1 diffs summary:",
		"|- Mean: 0ms",
		"|- Median: 0ms",
		"|- Percentiles (95, 99, max): 0ms, 0ms, 0ms",
	}
	return strings.Join(lines, "\n")
}

func (s *Server) accountCommand(fields []string) string {
	if len(fields) < 2 {
		return msgIncorrectSyntax
	}
	switch strings.ToLower(fields[1]) {
	case "create":
		return s.accountCreate(fields)
	case "set":
		return s.accountSet(fields)
	default:
		return fmt.Sprintf(msgCommandNotFound, ".account "+strings.ToLower(fields[1]))
	}
}

func (s *Server) accountCreate(fields []string) string {
	if len(fields) < 4 {
		return msgIncorrectSyntax
	}
	name, password := fields[2], fields[3]
	email := ""
	if len(fields) >= 5 {
		email = fields[4]
	}
	if len([]rune(name)) > maxAccountName {
		return msgAccountNameTooLong
	}
	if len([]rune(password)) > maxPassword {
		return msgAccountPassTooLong
	}
	if len([]rune(email)) > maxEmail {
		return msgEmailTooLong
	}

	s.mu.Lock()
	if _, exists := s.getAccount(name); exists {
		s.mu.Unlock()
		return msgAccountAlreadyExists
	}
	s.upsertAccount(Account{Username: normalizeAccountKey(name), Password: password, Email: email})
	s.mu.Unlock()
	s.mysql.upsertAccount(normalizeAccountKey(name), email)
	return fmt.Sprintf(msgAccountCreated, name)
}

func (s *Server) accountSet(fields []string) string {
	if len(fields) < 3 {
		return msgIncorrectSyntax
	}
	switch strings.ToLower(fields[2]) {
	case "password":
		return s.accountSetPassword(fields)
	case "email":
		return s.accountSetEmail(fields)
	case "gmlevel":
		return s.accountSetGMLevel(fields)
	default:
		return fmt.Sprintf(msgCommandNotFound, ".account set "+strings.ToLower(fields[2]))
	}
}

func (s *Server) accountSetPassword(fields []string) string {
	if len(fields) < 6 {
		return msgIncorrectSyntax
	}
	name, password, confirmation := fields[3], fields[4], fields[5]
	if len([]rune(password)) > maxPassword {
		return msgAccountPassTooLong
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	account, ok := s.getAccount(name)
	if !ok {
		return fmt.Sprintf(msgAccountNotExist, name)
	}
	if password != confirmation {
		return msgPasswordsDoNotMatch
	}
	account.Password = password
	return msgPasswordChanged
}

func (s *Server) accountSetEmail(fields []string) string {
	if len(fields) < 5 {
		return msgIncorrectSyntax
	}
	name, email := fields[3], fields[4]
	if len([]rune(email)) > maxEmail {
		return msgEmailTooLong
	}

	s.mu.Lock()
	account, ok := s.getAccount(name)
	if !ok {
		s.mu.Unlock()
		return fmt.Sprintf(msgAccountNotExist, name)
	}
	// The real handler takes an email and a confirmation; the gateway sends a
	// single value, so the confirmation is optional here.
	if len(fields) >= 6 && fields[5] != email {
		s.mu.Unlock()
		return msgEmailsDoNotMatch
	}
	account.Email = email
	s.mu.Unlock()
	s.mysql.setEmail(normalizeAccountKey(name), email)
	return msgEmailChanged
}

func (s *Server) accountSetGMLevel(fields []string) string {
	if len(fields) < 5 {
		return msgIncorrectSyntax
	}
	name := fields[3]
	level, err := strconv.Atoi(fields[4])
	if err != nil {
		return msgBadValue
	}
	if level < -1 || level > 4 {
		return msgBadValue
	}

	s.mu.Lock()
	account, ok := s.getAccount(name)
	if !ok {
		s.mu.Unlock()
		return fmt.Sprintf(msgAccountNotExist, name)
	}
	account.GMLevel = level
	s.mu.Unlock()
	s.mysql.setGMLevel(normalizeAccountKey(name), level)
	return fmt.Sprintf(msgSecurityChanged, name, level)
}

func (s *Server) banCommand(fields []string) string {
	if len(fields) < 4 || !strings.EqualFold(fields[1], "account") {
		return msgIncorrectSyntax
	}
	name, duration := fields[2], fields[3]
	reason := "No reason"
	if len(fields) >= 5 {
		reason = strings.Join(fields[4:], " ")
	}

	seconds, ok := timeStringToSeconds(duration)
	if !ok {
		return msgBadValue
	}

	s.mu.Lock()
	account, ok := s.getAccount(name)
	if !ok {
		s.mu.Unlock()
		return fmt.Sprintf(msgBanNotFound, "account", name)
	}
	account.Banned = true
	account.BanReason = reason
	if seconds <= 0 {
		account.BanDuration = "permanent"
	} else {
		account.BanDuration = secsToTimeString(seconds)
	}
	s.mu.Unlock()

	s.mysql.setBanned(normalizeAccountKey(name), seconds, reason)
	if seconds <= 0 {
		return fmt.Sprintf(msgBanPermanent, name, reason)
	}
	return fmt.Sprintf(msgBanTemporary, name, secsToTimeString(seconds), reason)
}

func (s *Server) unbanCommand(fields []string) string {
	if len(fields) < 3 || !strings.EqualFold(fields[1], "account") {
		return msgIncorrectSyntax
	}
	name := fields[2]

	s.mu.Lock()
	account, ok := s.getAccount(name)
	if !ok || !account.Banned {
		s.mu.Unlock()
		return fmt.Sprintf(msgUnbanError, name)
	}
	account.Banned = false
	account.BanDuration = ""
	account.BanReason = ""
	s.mu.Unlock()
	s.mysql.clearBan(normalizeAccountKey(name))
	return fmt.Sprintf(msgUnbanned, name)
}
