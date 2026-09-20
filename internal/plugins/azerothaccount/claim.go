package azerothaccount

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/notice"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
)

// noticeService is the capability published by azeroth-character.
const noticeService = "azeroth.notice"

// claimTTL is how long a claim code stays valid.
const claimTTL = 10 * time.Minute

// claimMaxAttempts is the number of wrong codes allowed per claim.
const claimMaxAttempts = 5

// ClaimStore persists pending account claims.
type ClaimStore interface {
	UpsertClaim(ctx context.Context, userID uuid.UUID, accountUsername, codeHash string, expiresAt time.Time) (domain.Claim, error)
	GetClaim(ctx context.Context, userID uuid.UUID) (domain.Claim, error)
	IncrementClaimAttempts(ctx context.Context, userID uuid.UUID) (domain.Claim, error)
	DeleteClaim(ctx context.Context, userID uuid.UUID) error
	ListClaims(ctx context.Context, limit, offset int) ([]domain.Claim, error)
}

var (
	errClaimUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"claim_unavailable", "account claims are unavailable")
	errClaimAccountNotFound = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"account_not_found", "AzerothCore account does not exist")
	errClaimNoPending = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"no_pending_claim", "no pending claim; request a code first")
	errClaimMismatch = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"claim_mismatch", "the code was requested for a different account")
	errClaimExpired = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"claim_expired", "the claim code has expired; request a new one")
	errClaimInvalid = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_code", "the claim code is incorrect")
	errClaimTooMany = httpapi.NewAPIError(http.StatusTooManyRequests,
		"too_many_attempts", "too many incorrect codes; request a new one")
	errClaimCharacterNotFound = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"character_not_found", "character does not exist on that account")
	errClaimNotOwner = httpapi.NewAPIError(http.StatusForbidden,
		"not_owner", "the character does not belong to that account")
	errNoticeNotConfigured = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"notice_not_configured", "in-game notices are not configured")
	errClaimDeliveryFailed = httpapi.NewAPIError(http.StatusBadGateway,
		"notice_failed", "the claim code could not be delivered in game")
)

// StartClaimRequest is the body of POST /api/v1/azeroth/me/account/claim.
type StartClaimRequest struct {
	AccountUsername string `json:"account_username"`
	Character       string `json:"character"`
} // @name AzerothStartClaimRequest

// StartClaimResponse is the body of POST /api/v1/azeroth/me/account/claim.
type StartClaimResponse struct {
	ExpiresAt string `json:"expires_at"`
} // @name AzerothStartClaimResponse

// VerifyClaimRequest is the body of POST /api/v1/azeroth/me/account/claim/verify.
type VerifyClaimRequest struct {
	AccountUsername string `json:"account_username"`
	Code            string `json:"code"`
} // @name AzerothVerifyClaimRequest

// handleStartClaim handles POST /api/v1/azeroth/me/account/claim.
//
//	@Summary		Start claiming an account
//	@Description	Mails a one-time code to a character of the account the user wants to claim. Requires the azeroth.account.self permission.
//	@Tags			azeroth-account
//	@ID				azeroth.me.account.claim.start
//	@Accept			json
//	@Produce		json
//	@Param			request	body	StartClaimRequest	true	"account and character"
//	@Success		200	{object}	StartClaimResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/account/claim [post]
func (p *Plugin) handleStartClaim(w http.ResponseWriter, r *http.Request) {
	if p.links == nil || p.claims == nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	var req StartClaimRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())

	if _, err := p.links.Get(r.Context(), principal.UserID); err == nil {
		httpapi.WriteError(w, r, errAlreadyLinked)
		return
	} else if !errors.Is(err, domain.ErrLinkNotFound) {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	if p.accounts == nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	account, err := p.accounts.FindAccountByUsername(r.Context(), req.AccountUsername)
	if errors.Is(err, azerothdb.ErrAccountNotFound) {
		httpapi.WriteError(w, r, errClaimAccountNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}

	code, err := generateClaimCode()
	if err != nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	expiresAt := time.Now().Add(claimTTL)
	if _, err := p.claims.UpsertClaim(r.Context(), principal.UserID, req.AccountUsername, hashCode(code), expiresAt); err != nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}

	noticeSvc, err := p.resolveNotice()
	if err != nil {
		httpapi.WriteError(w, r, errNoticeNotConfigured)
		return
	}
	_, err = noticeSvc.Send(r.Context(), notice.Request{
		AccountID: account.ID,
		Character: req.Character,
		Subject:   "Account claim code",
		Body:      fmt.Sprintf("Your AzerothCore account claim code is %s (valid for 10 minutes).", code),
	})
	p.recordSelf(r, "azeroth.account.claim.start", req.AccountUsername, err)
	if err != nil {
		writeNoticeError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, StartClaimResponse{
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	})
}

// handleVerifyClaim handles POST /api/v1/azeroth/me/account/claim/verify.
//
//	@Summary		Verify an account claim
//	@Description	Verifies the one-time code and links the account to the user. Requires the azeroth.account.self permission.
//	@Tags			azeroth-account
//	@ID				azeroth.me.account.claim.verify
//	@Accept			json
//	@Produce		json
//	@Param			request	body	VerifyClaimRequest	true	"account and code"
//	@Success		200	{object}	SelfAccountResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		409	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		429	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/me/account/claim/verify [post]
func (p *Plugin) handleVerifyClaim(w http.ResponseWriter, r *http.Request) {
	if p.links == nil || p.claims == nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	var req VerifyClaimRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())

	if _, err := p.links.Get(r.Context(), principal.UserID); err == nil {
		httpapi.WriteError(w, r, errAlreadyLinked)
		return
	} else if !errors.Is(err, domain.ErrLinkNotFound) {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}

	claim, err := p.claims.GetClaim(r.Context(), principal.UserID)
	if errors.Is(err, domain.ErrClaimNotFound) {
		httpapi.WriteError(w, r, errClaimNoPending)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	if claim.AccountUsername != req.AccountUsername {
		httpapi.WriteError(w, r, errClaimMismatch)
		return
	}
	if time.Now().After(claim.ExpiresAt) {
		_ = p.claims.DeleteClaim(r.Context(), principal.UserID)
		httpapi.WriteError(w, r, errClaimExpired)
		return
	}
	if claim.Attempts >= claimMaxAttempts {
		httpapi.WriteError(w, r, errClaimTooMany)
		return
	}
	if subtle.ConstantTimeCompare([]byte(hashCode(req.Code)), []byte(claim.CodeHash)) != 1 {
		_, _ = p.claims.IncrementClaimAttempts(r.Context(), principal.UserID)
		p.recordSelf(r, "azeroth.account.claim.verify", claim.AccountUsername, errors.New("invalid code"))
		httpapi.WriteError(w, r, errClaimInvalid)
		return
	}

	var accountID *int64
	if p.accounts != nil {
		if account, lookupErr := p.accounts.FindAccountByUsername(r.Context(), claim.AccountUsername); lookupErr == nil {
			id := account.ID
			accountID = &id
		}
	}
	link, err := p.links.Upsert(r.Context(), principal.UserID, claim.AccountUsername, accountID)
	if err != nil {
		p.recordSelf(r, "azeroth.account.claim.verify", claim.AccountUsername, err)
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	_ = p.claims.DeleteClaim(r.Context(), principal.UserID)
	p.recordSelf(r, "azeroth.account.claim.verify", claim.AccountUsername, nil)
	httpapi.WriteJSON(w, http.StatusOK, SelfAccountResponse{
		Linked:          true,
		AccountUsername: link.AccountUsername,
		AccountID:       link.AccountID,
	})
}

func (p *Plugin) resolveNotice() (notice.Service, error) {
	if p.registry == nil {
		return nil, notice.ErrNotConfigured
	}
	return services.Consume[notice.Service](p.registry, noticeService)
}

func writeNoticeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, notice.ErrNotConfigured):
		httpapi.WriteError(w, r, errNoticeNotConfigured)
	case errors.Is(err, notice.ErrNotOwner):
		httpapi.WriteError(w, r, errClaimNotOwner)
	case errors.Is(err, notice.ErrCharacterNotFound):
		httpapi.WriteError(w, r, errClaimCharacterNotFound)
	case errors.Is(err, notice.ErrInvalidRequest):
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable.WithDetails(err.Error()))
	default:
		httpapi.WriteError(w, r, errClaimDeliveryFailed)
	}
}

func generateClaimCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// AdminClaim is a pending claim without the secret hash.
type AdminClaim struct {
	UserID          string `json:"user_id"`
	AccountUsername string `json:"account_username"`
	ExpiresAt       string `json:"expires_at"`
	Attempts        int    `json:"attempts"`
} // @name AzerothAdminClaim

// AdminClaimsResponse is the body of GET /api/v1/azeroth/admin/account-claims.
type AdminClaimsResponse struct {
	Claims []AdminClaim `json:"claims"`
} // @name AzerothAdminClaimsResponse

// handleListClaims handles GET /api/v1/azeroth/admin/account-claims.
//
//	@Summary		List account claims
//	@Description	Lists pending account claims (without the code hash). Requires the azeroth.admin.claims.read permission.
//	@Tags			azeroth-account
//	@ID				azeroth.admin.account_claims.list
//	@Produce		json
//	@Success		200	{object}	AdminClaimsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/admin/account-claims [get]
func (p *Plugin) handleListClaims(w http.ResponseWriter, r *http.Request) {
	if p.claims == nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	claims, err := p.claims.ListClaims(r.Context(), 50, 0)
	if err != nil {
		httpapi.WriteError(w, r, errClaimUnavailable)
		return
	}
	items := make([]AdminClaim, 0, len(claims))
	for _, claim := range claims {
		items = append(items, AdminClaim{
			UserID:          claim.UserID.String(),
			AccountUsername: claim.AccountUsername,
			ExpiresAt:       claim.ExpiresAt.UTC().Format(time.RFC3339),
			Attempts:        claim.Attempts,
		})
	}
	httpapi.WriteJSON(w, http.StatusOK, AdminClaimsResponse{Claims: items})
}
