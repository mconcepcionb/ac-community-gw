package azerothdb

import (
	"context"
	"errors"
	"testing"
)

func TestValidLeaderboard(t *testing.T) {
	for _, board := range []string{BoardProgression, BoardWealth, BoardPlaytime, BoardPvP} {
		if !ValidLeaderboard(board) {
			t.Errorf("ValidLeaderboard(%q) = false, want true", board)
		}
	}
	if ValidLeaderboard("bogus") || ValidLeaderboard("") {
		t.Error("an unknown board must be rejected")
	}
}

// TestUnavailableReadersWrapErrUnavailable calls every method of the
// Unavailable* readers with and without an underlying cause, so both branches
// of the "cause present" guard are covered.
func TestUnavailableReadersWrapErrUnavailable(t *testing.T) {
	ctx := context.Background()
	dialErr := errors.New("dial tcp: refused")

	calls := []struct {
		name string
		call func(underlying error) error
	}{
		{"Unavailable.ListAccounts", func(e error) error {
			_, err := (Unavailable{Err: e}).ListAccounts(ctx, Query{})
			return err
		}},
		{"Unavailable.FindAccountByUsername", func(e error) error {
			_, err := (Unavailable{Err: e}).FindAccountByUsername(ctx, "thrall")
			return err
		}},
		{"UnavailableCharacters.ListCharacters", func(e error) error {
			_, err := (UnavailableCharacters{Err: e}).ListCharacters(ctx, CharacterQuery{})
			return err
		}},
		{"UnavailableCharacters.CountCharacters", func(e error) error {
			_, err := (UnavailableCharacters{Err: e}).CountCharacters(ctx, CharacterQuery{})
			return err
		}},
		{"UnavailableCharacters.Equipment", func(e error) error {
			_, err := (UnavailableCharacters{Err: e}).Equipment(ctx, 1)
			return err
		}},
		{"UnavailableCharacters.FindCharacter", func(e error) error {
			_, err := (UnavailableCharacters{Err: e}).FindCharacter(ctx, "thrall")
			return err
		}},
		{"UnavailableCharacters.TopCharacters", func(e error) error {
			_, err := (UnavailableCharacters{Err: e}).TopCharacters(ctx, BoardPvP, 10, 0)
			return err
		}},
		{"UnavailableItems.ListItems", func(e error) error {
			_, err := (UnavailableItems{Err: e}).ListItems(ctx, ItemQuery{})
			return err
		}},
		{"UnavailableItems.FindItem", func(e error) error {
			_, err := (UnavailableItems{Err: e}).FindItem(ctx, 1)
			return err
		}},
	}

	for _, tc := range calls {
		for _, underlying := range []error{nil, dialErr} {
			if err := tc.call(underlying); !errors.Is(err, ErrUnavailable) {
				t.Errorf("%s(cause=%v) = %v, want it to wrap ErrUnavailable", tc.name, underlying, err)
			}
		}
	}
}

func TestSentinelsAreDistinct(t *testing.T) {
	pairs := [][2]error{
		{ErrUnavailable, ErrAccountNotFound},
		{ErrUnavailable, ErrCharacterNotFound},
		{ErrUnavailable, ErrItemNotFound},
		{ErrAccountNotFound, ErrCharacterNotFound},
	}
	for _, pair := range pairs {
		if errors.Is(pair[0], pair[1]) || errors.Is(pair[1], pair[0]) {
			t.Fatalf("%v and %v must be distinct", pair[0], pair[1])
		}
	}
}
