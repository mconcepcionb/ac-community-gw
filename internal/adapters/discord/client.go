// Package discord implements the Discord OAuth2 transport.
//
// It is responsible for building authorization URLs, exchanging authorization
// codes for access tokens and reading the user profile and guild member roles.
// It is transport-only: it has no knowledge of sessions, users or roles. It
// never logs tokens, codes or client secrets.
package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
)

const (
	defaultAuthorizeURL = "https://discord.com/oauth2/authorize"
	defaultTokenURL     = "https://discord.com/api/oauth2/token"
	defaultAPIBaseURL   = "https://discord.com/api/v10"
	maxResponseBytes    = 1 << 20
)

var (
	// ErrUpstream is returned for any non-2xx Discord response or decode error.
	// The upstream body is deliberately never included.
	ErrUpstream = errors.New("discord: upstream request failed")
)

// Config configures a Discord OAuth2 client.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthorizeURL string
	TokenURL     string
	APIBaseURL   string
	Scopes       []string
	Timeout      time.Duration
	HTTPClient   *http.Client
}

// Client is a transport-only Discord OAuth2 client.
type Client struct {
	clientID     string
	clientSecret string
	redirectURL  string
	authorizeURL string
	tokenURL     string
	apiBaseURL   string
	scopes       []string
	client       *http.Client
}

var _ auth.DiscordProvider = (*Client)(nil)

// New builds a Discord OAuth2 client.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, errors.New("discord: client id must not be empty")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, errors.New("discord: client secret must not be empty")
	}
	if strings.TrimSpace(cfg.RedirectURL) == "" {
		return nil, errors.New("discord: redirect url must not be empty")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	} else if cfg.Timeout > 0 {
		clone := *httpClient
		if clone.Timeout == 0 {
			clone.Timeout = cfg.Timeout
		}
		httpClient = &clone
	}

	return &Client{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURL:  cfg.RedirectURL,
		authorizeURL: orDefault(cfg.AuthorizeURL, defaultAuthorizeURL),
		tokenURL:     orDefault(cfg.TokenURL, defaultTokenURL),
		apiBaseURL:   strings.TrimRight(orDefault(cfg.APIBaseURL, defaultAPIBaseURL), "/"),
		scopes:       cfg.Scopes,
		client:       httpClient,
	}, nil
}

// AuthorizeURL implements auth.DiscordProvider.
func (c *Client) AuthorizeURL(state, codeChallenge string) string {
	scopes := c.scopes
	if len(scopes) == 0 {
		scopes = []string{"identify"}
	}
	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", c.redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(scopes, " "))
	query.Set("state", state)
	query.Set("prompt", "consent")
	if codeChallenge != "" {
		query.Set("code_challenge", codeChallenge)
		query.Set("code_challenge_method", "S256")
	}
	return c.authorizeURL + "?" + query.Encode()
}

// Exchange implements auth.DiscordProvider.
func (c *Client) Exchange(ctx context.Context, code, codeVerifier string) (auth.DiscordToken, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.redirectURL)
	if codeVerifier != "" {
		form.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return auth.DiscordToken{}, fmt.Errorf("discord: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := c.doJSON(req, &payload); err != nil {
		return auth.DiscordToken{}, err
	}
	if payload.AccessToken == "" {
		return auth.DiscordToken{}, fmt.Errorf("%w: token response missing access_token", ErrUpstream)
	}
	expiresAt := time.Time{}
	if payload.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}
	return auth.DiscordToken{AccessToken: payload.AccessToken, ExpiresAt: expiresAt}, nil
}

// CurrentUser implements auth.DiscordProvider.
func (c *Client) CurrentUser(ctx context.Context, accessToken string) (auth.DiscordUser, error) {
	req, err := c.bearerRequest(ctx, http.MethodGet, c.apiBaseURL+"/users/@me", accessToken)
	if err != nil {
		return auth.DiscordUser{}, err
	}
	var payload struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
		Avatar     string `json:"avatar"`
	}
	if err := c.doJSON(req, &payload); err != nil {
		return auth.DiscordUser{}, err
	}
	if payload.ID == "" {
		return auth.DiscordUser{}, fmt.Errorf("%w: user response missing id", ErrUpstream)
	}
	return auth.DiscordUser{
		ID:         payload.ID,
		Username:   payload.Username,
		GlobalName: payload.GlobalName,
		Avatar:     payload.Avatar,
	}, nil
}

// GuildMemberRoleIDs implements auth.DiscordProvider.
func (c *Client) GuildMemberRoleIDs(ctx context.Context, accessToken, guildID string) ([]string, error) {
	if strings.TrimSpace(guildID) == "" {
		return nil, errors.New("discord: guild id must not be empty")
	}
	req, err := c.bearerRequest(ctx, http.MethodGet, c.apiBaseURL+"/users/@me/guilds/"+url.PathEscape(guildID)+"/member", accessToken)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Roles []string `json:"roles"`
	}
	if err := c.doJSON(req, &payload); err != nil {
		return nil, err
	}
	return payload.Roles, nil
}

func (c *Client) bearerRequest(ctx context.Context, method, target, accessToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return nil, fmt.Errorf("discord: build request: %w", err)
	}
	if accessToken == "" {
		return nil, errors.New("discord: access token must not be empty")
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) doJSON(req *http.Request, target any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("%w: read response: %v", ErrUpstream, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Surface only the OAuth2 error code/description. They never contain
		// tokens, and hiding them entirely makes debugging impossible.
		var oauthErr struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		_ = json.Unmarshal(data, &oauthErr)
		switch {
		case oauthErr.Error != "" && oauthErr.Description != "":
			return fmt.Errorf("%w: status %d: %s: %s", ErrUpstream, resp.StatusCode, oauthErr.Error, oauthErr.Description)
		case oauthErr.Error != "":
			return fmt.Errorf("%w: status %d: %s", ErrUpstream, resp.StatusCode, oauthErr.Error)
		default:
			return fmt.Errorf("%w: status %d", ErrUpstream, resp.StatusCode)
		}
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%w: decode response: %v", ErrUpstream, err)
	}
	return nil
}

func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
