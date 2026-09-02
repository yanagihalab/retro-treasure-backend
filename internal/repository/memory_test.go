package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type fakeStateStore struct {
	loadData []byte
	loadErr  error
	saved    []byte
}

func (s *fakeStateStore) Close() error { return nil }

func (s *fakeStateStore) Load(context.Context) ([]byte, error) {
	return s.loadData, s.loadErr
}

func (s *fakeStateStore) Ping(context.Context) error { return nil }

func (s *fakeStateStore) Save(_ context.Context, data []byte) error {
	s.saved = append([]byte(nil), data...)
	return nil
}

func TestLoadPersistentStateImportsLegacyJSONIntoStateStore(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	legacyRepo := NewMemoryRepository()
	legacyRepo.SetPersistencePath(statePath)
	user, _, err := legacyRepo.CreateUser("legacy-player", "hash", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("legacy state file was not written: %v", err)
	}

	store := &fakeStateStore{loadErr: ErrPersistentStateNotFound}
	repo := NewMemoryRepository()
	repo.SetPersistencePath(statePath)
	repo.SetStateStore(store)
	if err := repo.LoadPersistentState(); err != nil {
		t.Fatalf("LoadPersistentState() error = %v", err)
	}
	if len(store.saved) == 0 {
		t.Fatal("legacy JSON was not imported into the state store")
	}
	loaded, err := repo.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	if loaded.Username != "legacy-player" {
		t.Fatalf("loaded username = %q", loaded.Username)
	}
}

func TestLoadPersistentStateUsesStateStoreBeforeLegacyFile(t *testing.T) {
	source := NewMemoryRepository()
	statePath := filepath.Join(t.TempDir(), "database-state.json")
	source.SetPersistencePath(statePath)
	user, _, err := source.CreateUser("database-player", "hash", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	payload, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	store := &fakeStateStore{loadData: payload}
	repo := NewMemoryRepository()
	repo.SetPersistencePath(filepath.Join(t.TempDir(), "missing.json"))
	repo.SetStateStore(store)
	if err := repo.LoadPersistentState(); err != nil {
		t.Fatalf("LoadPersistentState() error = %v", err)
	}
	if _, err := repo.GetUserByID(user.ID); err != nil {
		t.Fatalf("database state was not loaded: %v", err)
	}
}
