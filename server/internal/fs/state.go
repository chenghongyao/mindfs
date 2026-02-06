package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const stateFileName = "state.json"

// State captures cursor/position info for a managed directory.
type State struct {
	Cursor   string `json:"cursor,omitempty"`
	Position int    `json:"position,omitempty"`
}

// LoadState reads the state file if present.
func LoadState(managedDir string) (State, error) {
	if managedDir == "" {
		return State{}, errors.New("managed dir required")
	}
	path := filepath.Join(managedDir, stateFileName)
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(payload, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

// SaveState writes the state file.
func SaveState(managedDir string, state State) error {
	if managedDir == "" {
		return errors.New("managed dir required")
	}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(managedDir, stateFileName)
	return os.WriteFile(path, payload, 0o644)
}
