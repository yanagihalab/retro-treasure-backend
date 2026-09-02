package repository

import (
	"context"
	"errors"
)

var ErrPersistentStateNotFound = errors.New("persistent state not found")

// StateStore persists the mutable game state while master data remains seeded in memory.
type StateStore interface {
	Close() error
	Load(context.Context) ([]byte, error)
	Ping(context.Context) error
	Save(context.Context, []byte) error
}
