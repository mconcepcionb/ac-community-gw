// Package repository implements the PostgreSQL persistence owned by the
// identity-discord plugin: sessions, OAuth states, community users and role
// mappings. It is the only place that translates between the core auth types
// and the sqlc-generated queries.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	identitydiscordrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/identitydiscord/repository/generated"
)

// ErrStateNotFound is returned when an OAuth state is unknown or expired.
var ErrStateNotFound = errors.New("repository: oauth state not found")

// ErrUserNotFound is returned when a community user does not exist.
var ErrUserNotFound = errors.New("repository: community user not found")

// Store implements the identity-discord persistence operations.
type Store struct {
	db *sql.DB
	q  *identitydiscordrepo.Queries
}

var _ auth.Store = (*Store)(nil)

// New creates a store over an open database pool.
func New(db *sql.DB) *Store {
	return &Store{db: db, q: identitydiscordrepo.New(db)}
}

// Create implements auth.Store. The session ID is expected to already be the
// hash of the cookie token.
func (s *Store) Create(ctx context.Context, session auth.Session) error {
	_, err := s.q.CreateSession(ctx, identitydiscordrepo.CreateSessionParams{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
		Roles:     nonNilRoles(session.Roles),
	})
	if err != nil {
		return fmt.Errorf("repository: create session: %w", err)
	}
	return nil
}

// Get implements auth.Store.
func (s *Store) Get(ctx context.Context, id string) (auth.Session, error) {
	row, err := s.q.GetSession(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	if err != nil {
		return auth.Session{}, fmt.Errorf("repository: get session: %w", err)
	}
	return auth.Session{
		ID:        row.ID,
		UserID:    row.UserID,
		DiscordID: row.DiscordID,
		Roles:     row.Roles,
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

// Delete implements auth.Store.
func (s *Store) Delete(ctx context.Context, id string) error {
	return s.q.DeleteSession(ctx, id)
}

// DeleteExpiredSessions removes sessions past their expiry.
func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	return s.q.DeleteExpiredSessions(ctx)
}

// CreateOAuthState persists a short-lived OAuth state.
func (s *Store) CreateOAuthState(ctx context.Context, state, codeVerifier, returnTo string, expiresAt time.Time) error {
	return s.q.CreateOAuthState(ctx, identitydiscordrepo.CreateOAuthStateParams{
		State:        state,
		CodeVerifier: codeVerifier,
		ReturnTo:     returnTo,
		ExpiresAt:    expiresAt,
	})
}

// ConsumeOAuthState deletes and returns an OAuth state in a single statement.
func (s *Store) ConsumeOAuthState(ctx context.Context, state string) (string, string, error) {
	row, err := s.q.ConsumeOAuthState(ctx, state)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrStateNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("repository: consume oauth state: %w", err)
	}
	return row.CodeVerifier, row.ReturnTo, nil
}

// DeleteExpiredOAuthStates removes OAuth states past their expiry.
func (s *Store) DeleteExpiredOAuthStates(ctx context.Context) error {
	return s.q.DeleteExpiredOAuthStates(ctx)
}

// Provision upserts the community user and its Discord identity in one
// transaction and reports whether the user was created.
func (s *Store) Provision(ctx context.Context, user auth.DiscordUser) (uuid.UUID, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("repository: begin provision: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := s.q.WithTx(tx)
	row, err := q.UpsertCommunityUser(ctx, identitydiscordrepo.UpsertCommunityUserParams{
		ID:          uuid.New(),
		DiscordID:   user.ID,
		DisplayName: displayName(user),
	})
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("repository: upsert community user: %w", err)
	}
	if _, err := q.UpsertDiscordIdentity(ctx, identitydiscordrepo.UpsertDiscordIdentityParams{
		UserID:     row.ID,
		DiscordID:  user.ID,
		Username:   user.Username,
		GlobalName: user.GlobalName,
		Avatar:     user.Avatar,
	}); err != nil {
		return uuid.Nil, false, fmt.Errorf("repository: upsert discord identity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return uuid.Nil, false, fmt.Errorf("repository: commit provision: %w", err)
	}
	return row.ID, row.Inserted, nil
}

// UserIDByDiscordID resolves a Discord id to a community user id.
func (s *Store) UserIDByDiscordID(ctx context.Context, discordID string) (uuid.UUID, bool, error) {
	user, err := s.q.GetCommunityUserByDiscordID(ctx, discordID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("repository: get community user: %w", err)
	}
	return user.ID, true, nil
}

// ListUsers searches community users by id, username, global name or display
// name. An empty filter lists everyone.
func (s *Store) ListUsers(ctx context.Context, filter string, limit, offset int) ([]userdir.User, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	pattern := "%" + escapeLikePattern(filter) + "%"
	rows, err := s.q.ListCommunityUsers(ctx, identitydiscordrepo.ListCommunityUsersParams{
		Column1:   filter,
		DiscordID: pattern,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("repository: list community users: %w", err)
	}
	users := make([]userdir.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, userdir.User{
			ID:          row.ID,
			DiscordID:   row.DiscordID,
			Username:    row.Username,
			GlobalName:  row.GlobalName,
			DisplayName: row.DisplayName,
			CreatedAt:   row.CreatedAt,
		})
	}
	return users, nil
}

// ResolveUserByName returns exact (case-insensitive) matches on display name,
// Discord username or global name, plus an exact Discord id match.
func (s *Store) ResolveUserByName(ctx context.Context, name string) ([]userdir.User, error) {
	rows, err := s.q.ResolveUserByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, fmt.Errorf("repository: resolve user by name: %w", err)
	}
	users := make([]userdir.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, userdir.User{
			ID:          row.ID,
			DiscordID:   row.DiscordID,
			Username:    row.Username,
			GlobalName:  row.GlobalName,
			DisplayName: row.DisplayName,
			CreatedAt:   row.CreatedAt,
		})
	}
	return users, nil
}

// UserProfile returns a community user together with its Discord profile.
func (s *Store) UserProfile(ctx context.Context, userID uuid.UUID) (userdir.User, error) {
	row, err := s.q.GetUserProfile(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return userdir.User{}, ErrUserNotFound
	}
	if err != nil {
		return userdir.User{}, fmt.Errorf("repository: get user profile: %w", err)
	}
	return userdir.User{
		ID:          row.ID,
		DiscordID:   row.DiscordID,
		Username:    row.Username,
		GlobalName:  row.GlobalName,
		DisplayName: row.DisplayName,
		Avatar:      row.Avatar,
		CreatedAt:   row.CreatedAt,
	}, nil
}

// MappedRoles returns the internal roles mapped from Discord role ids.
func (s *Store) MappedRoles(ctx context.Context, discordRoleIDs []string) ([]string, error) {
	if len(discordRoleIDs) == 0 {
		return nil, nil
	}
	roles, err := s.q.ListInternalRolesForDiscordRoles(ctx, discordRoleIDs)
	if err != nil {
		return nil, fmt.Errorf("repository: list mapped roles: %w", err)
	}
	return roles, nil
}

// UserRoles returns the last known effective roles of a community user.
func (s *Store) UserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	user, err := s.q.GetCommunityUserByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get community user: %w", err)
	}
	return user.Roles, nil
}

// UpdateUserRoles stores the effective roles of a community user.
func (s *Store) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roles []string) error {
	err := s.q.UpdateCommunityUserRoles(ctx, identitydiscordrepo.UpdateCommunityUserRolesParams{
		ID:      userID,
		Column2: nonNilRoles(roles),
	})
	if err != nil {
		return fmt.Errorf("repository: update user roles: %w", err)
	}
	return nil
}

// ListRolePermissions returns every role -> permission grant.
func (s *Store) ListRolePermissions(ctx context.Context) ([]identitydiscordrepo.RolePermission, error) {
	rows, err := s.q.ListRolePermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: list role permissions: %w", err)
	}
	return rows, nil
}

// SyncPermissions upserts the registered permission definitions.
func (s *Store) SyncPermissions(ctx context.Context, defs []permissions.Definition) error {
	for _, def := range defs {
		if err := s.q.UpsertPermission(ctx, identitydiscordrepo.UpsertPermissionParams{
			Name:        string(def.Name),
			Description: def.Description,
			Owner:       def.Owner,
		}); err != nil {
			return fmt.Errorf("repository: upsert permission %s: %w", def.Name, err)
		}
	}
	return nil
}

// UpsertRole ensures an internal role exists.
func (s *Store) UpsertRole(ctx context.Context, role string) error {
	if err := s.q.UpsertRole(ctx, role); err != nil {
		return fmt.Errorf("repository: upsert role: %w", err)
	}
	return nil
}

// GrantRolePermission grants a permission to a role.
func (s *Store) GrantRolePermission(ctx context.Context, role, permission string) error {
	if err := s.q.GrantRolePermission(ctx, identitydiscordrepo.GrantRolePermissionParams{
		Role:       role,
		Permission: permission,
	}); err != nil {
		return fmt.Errorf("repository: grant role permission: %w", err)
	}
	return nil
}

// RevokeRolePermission revokes a permission from a role.
func (s *Store) RevokeRolePermission(ctx context.Context, role, permission string) error {
	if err := s.q.RevokeRolePermission(ctx, identitydiscordrepo.RevokeRolePermissionParams{
		Role:       role,
		Permission: permission,
	}); err != nil {
		return fmt.Errorf("repository: revoke role permission: %w", err)
	}
	return nil
}

// ListDiscordRoleMappings returns every Discord role -> internal role mapping.
func (s *Store) ListDiscordRoleMappings(ctx context.Context) ([]identitydiscordrepo.DiscordRoleMapping, error) {
	rows, err := s.q.ListDiscordRoleMappings(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: list discord role mappings: %w", err)
	}
	return rows, nil
}

// UpsertDiscordRoleMapping maps a Discord role to an internal role.
func (s *Store) UpsertDiscordRoleMapping(ctx context.Context, discordRoleID, role string) error {
	if err := s.q.UpsertDiscordRoleMapping(ctx, identitydiscordrepo.UpsertDiscordRoleMappingParams{
		DiscordRoleID: discordRoleID,
		Role:          role,
	}); err != nil {
		return fmt.Errorf("repository: upsert discord role mapping: %w", err)
	}
	return nil
}

// DeleteDiscordRoleMapping removes a Discord role mapping.
func (s *Store) DeleteDiscordRoleMapping(ctx context.Context, discordRoleID string) error {
	return s.q.DeleteDiscordRoleMapping(ctx, discordRoleID)
}

func displayName(user auth.DiscordUser) string {
	switch {
	case user.GlobalName != "":
		return user.GlobalName
	case user.Username != "":
		return user.Username
	default:
		return user.ID
	}
}

// nonNilRoles guarantees a non-nil slice: pq.Array encodes a nil slice as SQL
// NULL, which violates the NOT NULL constraint on the roles columns.
func nonNilRoles(roles []string) []string {
	if roles == nil {
		return []string{}
	}
	return roles
}

// escapeLikePattern escapes LIKE/ILIKE metacharacters so user input is matched
// literally (PostgreSQL uses backslash as the default escape character).
func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
