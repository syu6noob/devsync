package core

// FileMeta is the metadata used to compare files between clients and the server.
type FileMeta struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	MTime     int64  `json:"mtime"`
	SHA256    string `json:"sha256"`
	Version   int64  `json:"version"`
	IsDeleted bool   `json:"is_deleted"`
	UpdatedBy string `json:"updated_by"`
	UpdatedAt int64  `json:"updated_at"`
}

type SyncRequest struct {
	WorkspaceID string              `json:"workspace_id"`
	ClientID    string              `json:"client_id"`
	Files       map[string]FileMeta `json:"files"`
	Base        map[string]FileMeta `json:"base"`
}

type Operation struct {
	Action string   `json:"action"` // upload, download, delete_remote, conflict, noop
	Path   string   `json:"path"`
	Meta   FileMeta `json:"meta"`
	Reason string   `json:"reason"`
}

type SyncPlan struct {
	Operations []Operation `json:"operations"`
}

type WorkspaceManifest struct {
	WorkspaceID string              `json:"workspace_id"`
	Files       map[string]FileMeta `json:"files"`
}
