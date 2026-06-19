package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/syu6noob/devsync/internal/core"
)

type WorkspaceState struct {
	WorkspaceID string                   `json:"workspace_id"`
	NextVersion int64                    `json:"next_version"`
	Files       map[string]core.FileMeta `json:"files"`
}

type Storage struct {
	DataDir string
}

func (s Storage) workspaceRoot(id string) string {
	return filepath.Join(s.DataDir, "workspaces", id)
}

func (s Storage) filesRoot(id string) string {
	return filepath.Join(s.workspaceRoot(id), "files")
}

func (s Storage) metaPath(id string) string {
	return filepath.Join(s.workspaceRoot(id), "meta.json")
}

func (s Storage) Load(id string) (WorkspaceState, error) {
	root := s.workspaceRoot(id)
	if err := os.MkdirAll(filepath.Join(root, "files"), 0o755); err != nil {
		return WorkspaceState{}, err
	}
	b, err := os.ReadFile(s.metaPath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return WorkspaceState{WorkspaceID: id, NextVersion: 1, Files: map[string]core.FileMeta{}}, nil
		}
		return WorkspaceState{}, err
	}
	var st WorkspaceState
	if err := json.Unmarshal(b, &st); err != nil {
		return WorkspaceState{}, err
	}
	if st.Files == nil {
		st.Files = map[string]core.FileMeta{}
	}
	if st.NextVersion <= 0 {
		st.NextVersion = 1
	}
	return st, nil
}

func (s Storage) Save(id string, st WorkspaceState) error {
	root := s.workspaceRoot(id)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(id), b, 0o600)
}

func (s Storage) FilePath(workspaceID, rel string) (string, error) {
	return core.SafeJoin(s.filesRoot(workspaceID), rel)
}

func (s Storage) SetFile(workspaceID string, meta core.FileMeta, clientID string) error {
	st, err := s.Load(workspaceID)
	if err != nil {
		return err
	}
	meta.Version = st.NextVersion
	st.NextVersion++
	meta.IsDeleted = false
	meta.UpdatedBy = clientID
	meta.UpdatedAt = time.Now().UnixNano()
	st.Files[meta.Path] = meta
	return s.Save(workspaceID, st)
}

func (s Storage) DeleteFile(workspaceID, rel, clientID string) error {
	st, err := s.Load(workspaceID)
	if err != nil {
		return err
	}
	meta := st.Files[rel]
	meta.Path = rel
	meta.IsDeleted = true
	meta.Version = st.NextVersion
	st.NextVersion++
	meta.UpdatedBy = clientID
	meta.UpdatedAt = time.Now().UnixNano()
	st.Files[rel] = meta
	return s.Save(workspaceID, st)
}
