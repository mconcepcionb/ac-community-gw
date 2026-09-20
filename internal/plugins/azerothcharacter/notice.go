package azerothcharacter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/notice"
)

// NoticeService is the capability name published by this plugin.
const NoticeService = "azeroth.notice"

// Send implements notice.Service. It verifies that the character belongs to
// the account, then mails the text with the configured notice item attached
// (AzerothCore mail requires an enclosure).
func (p *Plugin) Send(ctx context.Context, req notice.Request) (string, error) {
	recipient := strings.TrimSpace(req.Character)
	if !characterNamePattern.MatchString(recipient) {
		return "", fmt.Errorf("%w: invalid character name", notice.ErrInvalidRequest)
	}
	if p.noticeItemID <= 0 {
		return "", notice.ErrNotConfigured
	}
	if p.characters == nil {
		return "", fmt.Errorf("%w: character database not configured", azerothdb.ErrUnavailable)
	}

	character, err := p.characters.FindCharacter(ctx, recipient)
	if errors.Is(err, azerothdb.ErrCharacterNotFound) {
		return "", notice.ErrCharacterNotFound
	}
	if err != nil {
		return "", err
	}
	if character.AccountID != req.AccountID {
		return "", notice.ErrNotOwner
	}

	subject, body := sanitizeMailText(req.Subject), sanitizeMailText(req.Body)
	if subject == "" {
		subject = "Notice"
	}
	if body == "" {
		body = " "
	}
	command := buildSendItems(recipient, subject, body, []MailItem{{ID: p.noticeItemID, Count: 1}})
	return p.executor.Execute(ctx, command)
}
