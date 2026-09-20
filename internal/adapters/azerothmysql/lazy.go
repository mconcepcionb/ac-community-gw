package azerothmysql

import (
	"context"
	"fmt"
	"sync"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

// The lazy readers dial their database on first use and keep the pool for later
// calls. database/sql reconnects a live pool, so they also recover when the
// database restarts. They let the gateway start before its read databases are
// reachable and pick them up once they are.

// LazyAccountReader is an AccountReader that connects on demand.
type LazyAccountReader struct {
	cfg    Config
	mu     sync.Mutex
	inner  azerothdb.AccountReader
	closer func()
}

var _ azerothdb.AccountReader = (*LazyAccountReader)(nil)

// NewLazyAccountReader creates a reader that dials the login database lazily.
func NewLazyAccountReader(cfg Config) *LazyAccountReader { return &LazyAccountReader{cfg: cfg} }

func (l *LazyAccountReader) get() (azerothdb.AccountReader, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.inner != nil {
		return l.inner, nil
	}
	client, err := New(l.cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", azerothdb.ErrUnavailable, err)
	}
	l.inner = client
	l.closer = func() { _ = client.Close() }
	return client, nil
}

// ListAccounts implements azerothdb.AccountReader.
func (l *LazyAccountReader) ListAccounts(ctx context.Context, query azerothdb.Query) ([]azerothdb.Account, error) {
	reader, err := l.get()
	if err != nil {
		return nil, err
	}
	return reader.ListAccounts(ctx, query)
}

// FindAccountByUsername implements azerothdb.AccountReader.
func (l *LazyAccountReader) FindAccountByUsername(ctx context.Context, username string) (azerothdb.Account, error) {
	reader, err := l.get()
	if err != nil {
		return azerothdb.Account{}, err
	}
	return reader.FindAccountByUsername(ctx, username)
}

// Close closes the pool if one was opened.
func (l *LazyAccountReader) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer != nil {
		l.closer()
		l.closer = nil
		l.inner = nil
	}
	return nil
}

// LazyCharacterStore is a CharacterReader that connects on demand.
type LazyCharacterStore struct {
	cfg    Config
	mu     sync.Mutex
	inner  azerothdb.CharacterReader
	closer func()
}

var _ azerothdb.CharacterReader = (*LazyCharacterStore)(nil)

// NewLazyCharacterStore creates a reader that dials the character DB lazily.
func NewLazyCharacterStore(cfg Config) *LazyCharacterStore { return &LazyCharacterStore{cfg: cfg} }

func (l *LazyCharacterStore) get() (azerothdb.CharacterReader, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.inner != nil {
		return l.inner, nil
	}
	store, err := NewCharacterStore(l.cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", azerothdb.ErrUnavailable, err)
	}
	l.inner = store
	l.closer = func() { _ = store.Close() }
	return store, nil
}

// ListCharacters implements azerothdb.CharacterReader.
func (l *LazyCharacterStore) ListCharacters(ctx context.Context, query azerothdb.CharacterQuery) ([]azerothdb.Character, error) {
	reader, err := l.get()
	if err != nil {
		return nil, err
	}
	return reader.ListCharacters(ctx, query)
}

// FindCharacter implements azerothdb.CharacterReader.
func (l *LazyCharacterStore) FindCharacter(ctx context.Context, name string) (azerothdb.Character, error) {
	reader, err := l.get()
	if err != nil {
		return azerothdb.Character{}, err
	}
	return reader.FindCharacter(ctx, name)
}

// TopCharacters implements azerothdb.CharacterReader.
func (l *LazyCharacterStore) TopCharacters(ctx context.Context, board string, limit, offset int) ([]azerothdb.Character, error) {
	reader, err := l.get()
	if err != nil {
		return nil, err
	}
	return reader.TopCharacters(ctx, board, limit, offset)
}

// Close closes the pool if one was opened.
func (l *LazyCharacterStore) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer != nil {
		l.closer()
		l.closer = nil
		l.inner = nil
	}
	return nil
}

// LazyItemStore is an ItemReader that connects on demand.
type LazyItemStore struct {
	cfg    Config
	mu     sync.Mutex
	inner  azerothdb.ItemReader
	closer func()
}

var _ azerothdb.ItemReader = (*LazyItemStore)(nil)

// NewLazyItemStore creates a reader that dials the world DB lazily.
func NewLazyItemStore(cfg Config) *LazyItemStore { return &LazyItemStore{cfg: cfg} }

func (l *LazyItemStore) get() (azerothdb.ItemReader, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.inner != nil {
		return l.inner, nil
	}
	store, err := NewItemStore(l.cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", azerothdb.ErrUnavailable, err)
	}
	l.inner = store
	l.closer = func() { _ = store.Close() }
	return store, nil
}

// ListItems implements azerothdb.ItemReader.
func (l *LazyItemStore) ListItems(ctx context.Context, query azerothdb.ItemQuery) ([]azerothdb.Item, error) {
	reader, err := l.get()
	if err != nil {
		return nil, err
	}
	return reader.ListItems(ctx, query)
}

// FindItem implements azerothdb.ItemReader.
func (l *LazyItemStore) FindItem(ctx context.Context, entry int64) (azerothdb.Item, error) {
	reader, err := l.get()
	if err != nil {
		return azerothdb.Item{}, err
	}
	return reader.FindItem(ctx, entry)
}

// Close closes the pool if one was opened.
func (l *LazyItemStore) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer != nil {
		l.closer()
		l.closer = nil
		l.inner = nil
	}
	return nil
}
