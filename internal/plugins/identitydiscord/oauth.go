package identitydiscord

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/events"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

// handleLogin starts the Discord OAuth2 Authorization Code flow.
//
//	@Summary		Start Discord login
//	@Description	Mints a single-use state and PKCE verifier and redirects to Discord. The optional return_to must be a same-site absolute path.
//	@Tags			auth
//	@ID				auth.discord.login
//	@Param			return_to	query	string	false	"same-site absolute path to redirect to after login"
//	@Success		302	"Redirect to Discord"
//	@Failure		429	{object}	httpapi.ErrorResponse
//	@Failure		500	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/auth/discord/login [get]
func (p *Plugin) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !p.ready() {
		p.loginFailed(w, r, p.unavailableError())
		return
	}
	p.inc(metricLoginStarted)

	state, err := newState()
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrInternal)
		return
	}
	verifier, err := newCodeVerifier()
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrInternal)
		return
	}
	returnTo := sanitizeReturnTo(r.URL.Query().Get("return_to"))

	if err := p.states.CreateOAuthState(r.Context(), state, verifier, returnTo, p.now().Add(p.stateTTL)); err != nil {
		p.logger.Error("identity-discord: create oauth state", "error", err)
		httpapi.WriteError(w, r, httpapi.ErrInternal)
		return
	}

	http.Redirect(w, r, p.provider.AuthorizeURL(state, codeChallenge(verifier)), http.StatusFound)
}

// handleCallback completes the Discord OAuth2 flow.
//
//	@Summary		Discord OAuth2 callback
//	@Description	Consumes the single-use state, exchanges the code, provisions the user, synchronises roles and creates a session. Redirects to return_to or the configured post-login URL; returns the principal when neither is set.
//	@Tags			auth
//	@ID				auth.discord.callback
//	@Param			code	query	string	true	"authorization code"
//	@Param			state	query	string	true	"single-use OAuth state"
//	@Success		302	"Redirect to return_to or the post-login URL"
//	@Success		200	{object}	CallbackResponse	"Returned when no redirect is configured"
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		429	{object}	httpapi.ErrorResponse
//	@Failure		500	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/auth/discord/callback [get]
func (p *Plugin) handleCallback(w http.ResponseWriter, r *http.Request) {
	if !p.ready() {
		p.loginFailed(w, r, p.unavailableError())
		return
	}

	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	if code == "" || state == "" {
		p.loginFailed(w, r, httpapi.ErrBadRequest)
		return
	}

	verifier, returnTo, err := p.states.ConsumeOAuthState(r.Context(), state)
	if err != nil {
		p.loginFailed(w, r, httpapi.ErrBadRequest)
		return
	}

	token, err := p.provider.Exchange(r.Context(), code, verifier)
	if err != nil {
		p.logger.Warn("identity-discord: token exchange failed", "error", err)
		p.loginFailed(w, r, httpapi.ErrBadGateway)
		return
	}

	user, err := p.provider.CurrentUser(r.Context(), token.AccessToken)
	if err != nil {
		p.logger.Warn("identity-discord: current user failed", "error", err)
		p.loginFailed(w, r, httpapi.ErrBadGateway)
		return
	}

	userID, created, err := p.provision(r, user)
	if err != nil {
		p.logger.Error("identity-discord: provision failed", "error", err)
		p.loginFailed(w, r, httpapi.ErrInternal)
		return
	}
	if created {
		if err := p.publish(r, UserProvisioned{UserID: userID, DiscordID: user.ID}); err != nil {
			p.logger.Error("identity-discord: publish provisioned", "error", err)
			p.loginFailed(w, r, httpapi.ErrInternal)
			return
		}
	}

	roles, rolesChanged, err := p.syncRoles(r, token.AccessToken, userID)
	if err != nil {
		p.logger.Warn("identity-discord: role sync failed", "error", err)
		p.loginFailed(w, r, httpapi.ErrBadGateway)
		return
	}
	if rolesChanged {
		p.inc(metricRolesChanged)
		if err := p.publish(r, DiscordRolesChanged{UserID: userID, DiscordID: user.ID, Roles: roles}); err != nil {
			p.logger.Error("identity-discord: publish roles changed", "error", err)
		}
	}
	if err := p.publish(r, UserAuthenticated{UserID: userID, DiscordID: user.ID}); err != nil {
		p.logger.Error("identity-discord: publish authenticated", "error", err)
	}

	if old := p.sessions.TokenFromRequest(r); old != "" {
		_ = p.sessions.Revoke(r.Context(), old)
	}

	principal := auth.Principal{UserID: userID, DiscordID: user.ID, Roles: roles}
	sessionToken, err := p.sessions.Create(r.Context(), principal)
	if err != nil {
		p.logger.Error("identity-discord: create session", "error", err)
		p.loginFailed(w, r, httpapi.ErrInternal)
		return
	}
	p.sessions.SetCookie(w, sessionToken, p.now().Add(p.sessions.TTL()))
	p.inc(metricSessionCreated)
	p.inc(metricLoginCompleted)

	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      p.now(),
		ActorID:        userID,
		ActorDiscordID: user.ID,
		Action:         "identity.login",
		TargetType:     "community_user",
		TargetID:       userID.String(),
		Result:         audit.ResultSuccess,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})

	redirect := returnTo
	if redirect == "" {
		redirect = p.postLoginRedirect
	}
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, CallbackResponse{
		UserID:    userID.String(),
		DiscordID: user.ID,
		Roles:     rolesOrEmpty(roles),
	})
}

func (p *Plugin) ready() bool {
	return p.provider != nil && p.states != nil && p.repo != nil && p.sessions != nil
}

func (p *Plugin) publish(r *http.Request, event events.Event) error {
	if p.events == nil {
		return nil
	}
	return p.events.Publish(r.Context(), event)
}

func newState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func newCodeVerifier() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// sanitizeReturnTo accepts only same-site absolute paths; anything else is
// dropped to avoid turning the callback into an open redirect. Backslashes are
// rejected because browsers normalize them to forward slashes, which would turn
// "/\evil.test" into protocol-relative "//evil.test".
func sanitizeReturnTo(value string) string {
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, "\\\r\n") {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "" || parsed.Host != "" || parsed.Opaque != "" || parsed.User != nil {
		return ""
	}
	path := parsed.Path
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\\\r\n") {
		return ""
	}
	return value
}
