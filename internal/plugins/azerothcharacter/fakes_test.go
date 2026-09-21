package azerothcharacter

import (
	"context"
	"strings"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

type fakeExecutor struct {
	last   string
	result string
	err    error
}

func (f *fakeExecutor) Execute(_ context.Context, command string) (string, error) {
	f.last = command
	return f.result, f.err
}

type fakeCharacters struct {
	characters    []azerothdb.Character
	last          azerothdb.CharacterQuery
	findErr       error
	listErr       error
	topErr        error
	topCalls      int
	lastTopLimit  int
	lastTopOffset int
	equipment     []azerothdb.Equipment
	equipmentErr  error
}

func (f *fakeCharacters) ListCharacters(_ context.Context, query azerothdb.CharacterQuery) ([]azerothdb.Character, error) {
	f.last = query
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.characters, nil
}

func (f *fakeCharacters) CountCharacters(_ context.Context, query azerothdb.CharacterQuery) (int, error) {
	f.last = query
	if f.listErr != nil {
		return 0, f.listErr
	}
	return len(f.characters), nil
}

func (f *fakeCharacters) Equipment(context.Context, int64) ([]azerothdb.Equipment, error) {
	return f.equipment, f.equipmentErr
}

func (f *fakeCharacters) TopCharacters(_ context.Context, _ string, limit, offset int) ([]azerothdb.Character, error) {
	f.topCalls++
	f.lastTopLimit, f.lastTopOffset = limit, offset
	if f.topErr != nil {
		return nil, f.topErr
	}
	return f.characters, nil
}

func (f *fakeCharacters) FindCharacter(_ context.Context, name string) (azerothdb.Character, error) {
	if f.findErr != nil {
		return azerothdb.Character{}, f.findErr
	}
	for _, character := range f.characters {
		if strings.EqualFold(character.Name, name) {
			return character, nil
		}
	}
	return azerothdb.Character{}, azerothdb.ErrCharacterNotFound
}

type fakeAccounts struct {
	account azerothdb.Account
	found   bool
}

func (f *fakeAccounts) ListAccounts(context.Context, azerothdb.Query) ([]azerothdb.Account, error) {
	return nil, nil
}

func (f *fakeAccounts) FindAccountByUsername(context.Context, string) (azerothdb.Account, error) {
	if !f.found {
		return azerothdb.Account{}, azerothdb.ErrAccountNotFound
	}
	return f.account, nil
}

type fakeDirectory struct {
	username  string
	accountID *int64
	err       error
}

func (f *fakeDirectory) LinkedAccount(context.Context, string) (string, *int64, error) {
	if f.err != nil {
		return "", nil, f.err
	}
	return f.username, f.accountID, nil
}
