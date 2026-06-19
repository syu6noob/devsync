package client

import (
	"encoding/json"
	"os"

	"github.com/syu6noob/devsync/internal/core"
)

type State struct {
	WorkspaceID string                   `json:"workspace_id"`
	ClientID    string                   `json:"client_id"`
	Files       map[string]core.FileMeta `json:"files"`
}

func LoadState(root string) (State, error) {
	b, err := os.ReadFile(StatePath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return State{Files: map[string]core.FileMeta{}}, nil
		}
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	if s.Files == nil {
		s.Files = map[string]core.FileMeta{}
	}
	return s, nil
}

func SaveState(root string, s State) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(root), b, 0o600)
}
