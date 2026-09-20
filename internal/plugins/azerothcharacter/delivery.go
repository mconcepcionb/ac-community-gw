package azerothcharacter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
)

// DeliveryService is the capability name published by this plugin.
const DeliveryService = "azeroth.character.delivery"

// Deliver implements delivery.Service. It verifies that the character belongs
// to the account before sending anything.
func (p *Plugin) Deliver(ctx context.Context, req delivery.Request) (string, error) {
	recipient := strings.TrimSpace(req.Character)
	if !characterNamePattern.MatchString(recipient) {
		return "", fmt.Errorf("%w: invalid character name", delivery.ErrInvalidRequest)
	}
	if len(req.Items) == 0 && req.Money <= 0 {
		return "", fmt.Errorf("%w: empty delivery", delivery.ErrInvalidRequest)
	}
	if p.characters == nil {
		return "", fmt.Errorf("%w: character database not configured", azerothdb.ErrUnavailable)
	}

	character, err := p.characters.FindCharacter(ctx, recipient)
	if errors.Is(err, azerothdb.ErrCharacterNotFound) {
		return "", delivery.ErrCharacterNotFound
	}
	if err != nil {
		return "", err
	}
	if character.AccountID != req.AccountID {
		return "", delivery.ErrNotOwner
	}

	subject, body := sanitizeMailText(req.Subject), sanitizeMailText(req.Body)
	if subject == "" {
		subject = "Reward"
	}
	if body == "" {
		body = " "
	}

	results := make([]string, 0, 2)
	if len(req.Items) > 0 {
		items := make([]MailItem, 0, len(req.Items))
		for _, item := range req.Items {
			if item.ID <= 0 || item.Count <= 0 {
				return "", fmt.Errorf("%w: invalid item", delivery.ErrInvalidRequest)
			}
			items = append(items, MailItem{ID: item.ID, Count: item.Count})
		}
		output, err := p.executor.Execute(ctx, buildSendItems(recipient, subject, body, items))
		if err != nil {
			return "", err
		}
		results = append(results, output)
	}
	if req.Money > 0 {
		output, err := p.executor.Execute(ctx, buildSendMoney(recipient, subject, body, req.Money))
		if err != nil {
			return "", err
		}
		results = append(results, output)
	}
	return strings.Join(results, "\n"), nil
}
