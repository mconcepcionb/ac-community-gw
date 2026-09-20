package azerothmysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

const characterColumns = `
       c.guid,
       c.account,
       c.name,
       c.race,
       c.class,
       c.gender,
       c.level,
       c.online,
       c.logout_time,
       c.totaltime,
       c.money,
       c.arenaPoints,
       COALESCE(g.name, '') AS guild_name`

const listCharactersQuery = `SELECT ` + characterColumns + `
FROM characters c
LEFT JOIN guild_member gm ON gm.guid = c.guid
LEFT JOIN guild g ON g.guildid = gm.guildid
WHERE c.account = ?
  AND (? = '' OR c.name LIKE ?)
ORDER BY c.level DESC, c.name ASC
LIMIT ? OFFSET ?`

const findCharacterQuery = `SELECT ` + characterColumns + `
FROM characters c
LEFT JOIN guild_member gm ON gm.guid = c.guid
LEFT JOIN guild g ON g.guildid = gm.guildid
WHERE c.name = ?
LIMIT 1`

// CharacterStore is a read-only AzerothCore character database client.
type CharacterStore struct {
	db *sql.DB
}

var _ azerothdb.CharacterReader = (*CharacterStore)(nil)

// NewCharacterStore opens and verifies a connection pool to the character DB.
func NewCharacterStore(cfg Config) (*CharacterStore, error) {
	db, err := openPool(cfg.DSN, cfg, "characters")
	if err != nil {
		return nil, err
	}
	return &CharacterStore{db: db}, nil
}

// ListCharacters implements azerothdb.CharacterReader.
func (c *CharacterStore) ListCharacters(ctx context.Context, query azerothdb.CharacterQuery) ([]azerothdb.Character, error) {
	limit := clampLimit(query.Limit)
	offset := clampOffset(query.Offset)
	filter := strings.TrimSpace(query.Filter)
	pattern := "%" + escapeLike(filter) + "%"

	rows, err := c.db.QueryContext(ctx, listCharactersQuery, query.AccountID, filter, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("azerothmysql: list characters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	characters := make([]azerothdb.Character, 0, limit)
	for rows.Next() {
		character, err := scanCharacter(rows)
		if err != nil {
			return nil, fmt.Errorf("azerothmysql: scan character: %w", err)
		}
		characters = append(characters, character)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azerothmysql: iterate characters: %w", err)
	}
	return characters, nil
}

// FindCharacter implements azerothdb.CharacterReader.
func (c *CharacterStore) FindCharacter(ctx context.Context, name string) (azerothdb.Character, error) {
	row := c.db.QueryRowContext(ctx, findCharacterQuery, strings.TrimSpace(name))
	character, err := scanCharacter(row)
	if errors.Is(err, sql.ErrNoRows) {
		return azerothdb.Character{}, azerothdb.ErrCharacterNotFound
	}
	if err != nil {
		return azerothdb.Character{}, fmt.Errorf("azerothmysql: find character: %w", err)
	}
	return character, nil
}

// TopCharacters implements azerothdb.CharacterReader.
func (c *CharacterStore) TopCharacters(
	ctx context.Context,
	board string,
	limit, offset int,
) ([]azerothdb.Character, error) {
	rows, err := c.db.QueryContext(ctx, leaderboardQuery(board), clampLimit(limit), clampOffset(offset))
	if err != nil {
		return nil, fmt.Errorf("azerothmysql: top characters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	characters := make([]azerothdb.Character, 0)
	for rows.Next() {
		character, err := scanCharacter(rows)
		if err != nil {
			return nil, fmt.Errorf("azerothmysql: scan character: %w", err)
		}
		characters = append(characters, character)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("azerothmysql: iterate characters: %w", err)
	}
	return characters, nil
}

// leaderboardQuery builds the ordering for a board. The board name is validated
// by the caller, so it never reaches the query as free text.
func leaderboardQuery(board string) string {
	order := "c.level DESC, c.name ASC"
	switch board {
	case azerothdb.BoardWealth:
		order = "c.money DESC, c.name ASC"
	case azerothdb.BoardPlaytime:
		order = "c.totaltime DESC, c.name ASC"
	case azerothdb.BoardPvP:
		order = "c.arenaPoints DESC, c.name ASC"
	}
	return `SELECT ` + characterColumns + `
FROM characters c
LEFT JOIN guild_member gm ON gm.guid = c.guid
LEFT JOIN guild g ON g.guildid = gm.guildid
ORDER BY ` + order + `
LIMIT ? OFFSET ?`
}

// Close closes the pool.
func (c *CharacterStore) Close() error { return c.db.Close() }

func scanCharacter(row rowScanner) (azerothdb.Character, error) {
	var character azerothdb.Character
	if err := row.Scan(
		&character.GUID,
		&character.AccountID,
		&character.Name,
		&character.Race,
		&character.Class,
		&character.Gender,
		&character.Level,
		&character.Online,
		&character.LogoutTime,
		&character.TotalTime,
		&character.Money,
		&character.ArenaPoints,
		&character.GuildName,
	); err != nil {
		return azerothdb.Character{}, err
	}
	return character, nil
}
