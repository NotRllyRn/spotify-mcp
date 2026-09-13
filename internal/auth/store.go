package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	RefreshToken string    `json:"refresh_token"`
	AuthorizedAt time.Time `json:"authorized_at"`
}

type Store struct{ Path string }

func (s Store) Load() (State, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, errors.New("reauthorization_required: run spotify-mcp auth")
		}
		return State{}, fmt.Errorf("read token state: %w", err)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode token state: %w", err)
	}
	if state.RefreshToken == "" || state.AuthorizedAt.IsZero() {
		return State{}, errors.New("invalid token state: run spotify-mcp auth")
	}
	return state, nil
}

func (s Store) Save(state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode token state: %w", err)
	}
	dir := filepath.Dir(s.Path)
	file, err := os.CreateTemp(dir, ".token-*")
	if err != nil {
		return fmt.Errorf("create token state: %w", err)
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("secure token state: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("write token state: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync token state: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close token state: %w", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		return fmt.Errorf("replace token state: %w", err)
	}
	return nil
}

func (s Store) CheckWritable() error {
	file, err := os.CreateTemp(filepath.Dir(s.Path), ".write-check-*")
	if err != nil {
		return fmt.Errorf("TOKEN_PATH parent is not writable: %w", err)
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		return fmt.Errorf("close write check: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove write check: %w", err)
	}
	return nil
}
