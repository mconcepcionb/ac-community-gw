package azerothcharacter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
)

var (
	errInvalidRecipient = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_recipient", "character name is invalid")
	errEmptyDelivery = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"empty_delivery", "provide items or money")
	errInvalidItem = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_item", "each item needs a positive id and count")
	errMailFailed = httpapi.NewAPIError(http.StatusBadGateway,
		"mail_failed", "AzerothCore rejected the delivery")
)

// characterNamePattern keeps the recipient safe for the command string.
var characterNamePattern = regexp.MustCompile(`^[A-Za-z]{2,32}$`)

// MailItem is a single item stack to deliver.
type MailItem struct {
	ID    int `json:"id"`
	Count int `json:"count"`
} // @name AzerothMailItem

// SendMailRequest is the body of POST /api/v1/azeroth/mail.
type SendMailRequest struct {
	Character string     `json:"character"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body"`
	Money     int64      `json:"money"`
	Items     []MailItem `json:"items"`
} // @name AzerothSendMailRequest

// SendMailResponse is the body of POST /api/v1/azeroth/mail.
type SendMailResponse struct {
	Recipient string   `json:"recipient"`
	Results   []string `json:"results"`
} // @name AzerothSendMailResponse

// handleSendMail handles POST /api/v1/azeroth/mail.
//
//	@Summary		Send in-game mail
//	@Description	Delivers items and/or money to a character. Requires the azeroth.mail.send permission.
//	@Tags			azeroth-character
//	@ID				azeroth.mail.send
//	@Accept			json
//	@Produce		json
//	@Param			request	body	SendMailRequest	true	"delivery request"
//	@Success		200	{object}	SendMailResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/mail [post]
func (p *Plugin) handleSendMail(w http.ResponseWriter, r *http.Request) {
	var req SendMailRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	recipient, err := validateMailRequest(req)
	if err != nil {
		httpapi.WriteError(w, r, err)
		return
	}

	if p.characters == nil {
		httpapi.WriteError(w, r, errCharacterDBNotConfigured)
		return
	}
	if p.directory == nil {
		httpapi.WriteError(w, r, errDirectoryUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	_, accountID, err := p.directory.LinkedAccount(r.Context(), principal.UserID.String())
	if err != nil || accountID == nil {
		httpapi.WriteError(w, r, errAccountNotLinked)
		return
	}
	character, err := p.characters.FindCharacter(r.Context(), recipient)
	if errors.Is(err, azerothdb.ErrCharacterNotFound) {
		httpapi.WriteError(w, r, errCharacterNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errCharacterDBUnavailable)
		return
	}
	if character.AccountID != *accountID {
		httpapi.WriteError(w, r, errNotOwner)
		return
	}

	results, err := p.executeDelivery(r.Context(), recipient, req)
	if err != nil {
		p.recordMail(r, recipient, PermissionMailSend, "azeroth.mail.send", audit.ResultFailure)
		httpapi.WriteError(w, r, errMailFailed)
		return
	}
	p.recordMail(r, recipient, PermissionMailSend, "azeroth.mail.send", audit.ResultSuccess)
	httpapi.WriteJSON(w, http.StatusOK, SendMailResponse{Recipient: recipient, Results: results})
}

// handleAdminMail handles POST /api/v1/admin/characters/{name}/mail. It sends
// mail to any character without an ownership check.
//
//	@Summary		Send in-game mail as staff
//	@Description	Delivers items and/or money to any character. Requires the azeroth.admin.mail.send permission.
//	@Tags			azeroth-character
//	@ID				azeroth.admin.characters.mail
//	@Accept			json
//	@Produce		json
//	@Param			name	path	string			true	"character name"
//	@Param			request	body	SendMailRequest	true	"delivery request"
//	@Success		200	{object}	SendMailResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/characters/{name}/mail [post]
func (p *Plugin) handleAdminMail(w http.ResponseWriter, r *http.Request) {
	var req SendMailRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.Character = strings.TrimSpace(r.PathValue("name"))
	recipient, err := validateMailRequest(req)
	if err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	if p.characters == nil {
		httpapi.WriteError(w, r, errCharacterDBNotConfigured)
		return
	}
	if _, err := p.characters.FindCharacter(r.Context(), recipient); errors.Is(err, azerothdb.ErrCharacterNotFound) {
		httpapi.WriteError(w, r, errCharacterNotFound)
		return
	} else if err != nil {
		httpapi.WriteError(w, r, errCharacterDBUnavailable)
		return
	}

	results, err := p.executeDelivery(r.Context(), recipient, req)
	if err != nil {
		p.recordMail(r, recipient, PermissionAdminMailSend, "azeroth.admin.mail.send", audit.ResultFailure)
		httpapi.WriteError(w, r, errMailFailed)
		return
	}
	p.recordMail(r, recipient, PermissionAdminMailSend, "azeroth.admin.mail.send", audit.ResultSuccess)
	httpapi.WriteJSON(w, http.StatusOK, SendMailResponse{Recipient: recipient, Results: results})
}

// validateMailRequest checks the recipient and delivery, returning the trimmed
// recipient name.
func validateMailRequest(req SendMailRequest) (string, error) {
	recipient := strings.TrimSpace(req.Character)
	if !characterNamePattern.MatchString(recipient) {
		return "", errInvalidRecipient
	}
	if len(req.Items) == 0 && req.Money <= 0 {
		return "", errEmptyDelivery
	}
	for _, item := range req.Items {
		if item.ID <= 0 || item.Count <= 0 {
			return "", errInvalidItem
		}
	}
	return recipient, nil
}

// executeDelivery builds and runs the .send commands for a validated request.
func (p *Plugin) executeDelivery(ctx context.Context, recipient string, req SendMailRequest) ([]string, error) {
	subject, body := sanitizeMailText(req.Subject), sanitizeMailText(req.Body)
	if subject == "" {
		subject = "Reward"
	}
	if body == "" {
		body = " "
	}

	results := make([]string, 0, 2)
	if len(req.Items) > 0 {
		output, err := p.executor.Execute(ctx, buildSendItems(recipient, subject, body, req.Items))
		if err != nil {
			return nil, err
		}
		results = append(results, output)
	}
	if req.Money > 0 {
		output, err := p.executor.Execute(ctx, buildSendMoney(recipient, subject, body, req.Money))
		if err != nil {
			return nil, err
		}
		results = append(results, output)
	}
	return results, nil
}

func buildSendItems(recipient, subject, body string, items []MailItem) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%d:%d", item.ID, item.Count))
	}
	return fmt.Sprintf(`.send items %s "%s" "%s" %s`, recipient, subject, body, strings.Join(parts, " "))
}

func buildSendMoney(recipient, subject, body string, money int64) string {
	return fmt.Sprintf(`.send money %s "%s" "%s" %d`, recipient, subject, body, money)
}

// sanitizeMailText removes characters that could break out of the quoted
// command argument and caps the length.
func sanitizeMailText(value string) string {
	replacer := strings.NewReplacer(`"`, "", "\r", " ", "\n", " ")
	cleaned := strings.TrimSpace(replacer.Replace(value))
	if len(cleaned) > 200 {
		cleaned = cleaned[:200]
	}
	return cleaned
}

func (p *Plugin) recordMail(
	r *http.Request,
	recipient string,
	permission permissions.Permission,
	action string,
	result audit.Result,
) {
	principal, _ := auth.PrincipalFromContext(r.Context())
	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         action,
		Permission:     string(permission),
		TargetType:     "character",
		TargetID:       recipient,
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}
