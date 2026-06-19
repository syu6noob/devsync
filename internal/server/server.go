package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/syu6noob/devsync/internal/core"
)

type Server struct {
	Storage Storage
	Token   string
	mu      sync.Mutex
}

func New(dataDir, token string) *Server {
	return &Server{Storage: Storage{DataDir: dataDir}, Token: token}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/api/v1/workspaces/", s.auth(s.workspaceHandler))
	return mux
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Token != "" {
			got := r.Header.Get("Authorization")
			want := "Bearer " + s.Token
			if got != want {
				core.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) workspaceHandler(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/workspaces/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" {
		core.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	workspaceID, err := url.PathUnescape(parts[0])
	if err != nil || workspaceID == "" || strings.Contains(workspaceID, "/") || strings.Contains(workspaceID, "..") {
		core.WriteError(w, http.StatusBadRequest, "invalid workspace id")
		return
	}
	action := parts[1]

	switch {
	case action == "manifest" && r.Method == http.MethodGet:
		s.handleManifest(w, r, workspaceID)
	case action == "plan" && r.Method == http.MethodPost:
		s.handlePlan(w, r, workspaceID)
	case action == "file" && r.Method == http.MethodPut:
		s.handleUpload(w, r, workspaceID)
	case action == "file" && r.Method == http.MethodGet:
		s.handleDownload(w, r, workspaceID)
	case action == "file" && r.Method == http.MethodDelete:
		s.handleDelete(w, r, workspaceID)
	default:
		core.WriteError(w, http.StatusNotFound, "not found")
	}
}

func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request, workspaceID string) {
	s.mu.Lock()
	st, err := s.Storage.Load(workspaceID)
	s.mu.Unlock()
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	active := map[string]core.FileMeta{}
	for p, m := range st.Files {
		if !m.IsDeleted {
			active[p] = m
		}
	}
	core.WriteJSON(w, http.StatusOK, core.WorkspaceManifest{WorkspaceID: workspaceID, Files: active})
}

func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request, workspaceID string) {
	var req core.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.WorkspaceID != workspaceID {
		core.WriteError(w, http.StatusBadRequest, "workspace id mismatch")
		return
	}
	for p := range req.Files {
		if !core.IsSafeRelativePath(p) {
			core.WriteError(w, http.StatusBadRequest, "unsafe path in manifest: "+p)
			return
		}
	}
	for p := range req.Base {
		if !core.IsSafeRelativePath(p) {
			core.WriteError(w, http.StatusBadRequest, "unsafe path in base: "+p)
			return
		}
	}
	s.mu.Lock()
	st, err := s.Storage.Load(workspaceID)
	s.mu.Unlock()
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	plan := BuildPlan(req, st.Files)
	core.WriteJSON(w, http.StatusOK, plan)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request, workspaceID string) {
	rel := core.NormalizeRelativePath(r.URL.Query().Get("path"))
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = "unknown"
	}
	if !core.IsSafeRelativePath(rel) {
		core.WriteError(w, http.StatusBadRequest, "invalid path")
		return
	}
	abs, err := s.Storage.FilePath(workspaceID, rel)
	if err != nil {
		core.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tmp := abs + ".uploading"
	out, err := os.Create(tmp)
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n, copyErr := io.Copy(out, r.Body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		core.WriteError(w, http.StatusInternalServerError, copyErr.Error())
		return
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		core.WriteError(w, http.StatusInternalServerError, closeErr.Error())
		return
	}
	if err := os.Rename(tmp, abs); err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	hash, err := core.SHA256File(abs)
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	meta := core.FileMeta{Path: rel, Size: n, MTime: info.ModTime().UnixNano(), SHA256: hash}
	s.mu.Lock()
	err = s.Storage.SetFile(workspaceID, meta, clientID)
	s.mu.Unlock()
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	core.WriteJSON(w, http.StatusOK, map[string]string{"status": "uploaded"})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request, workspaceID string) {
	rel := core.NormalizeRelativePath(r.URL.Query().Get("path"))
	if !core.IsSafeRelativePath(rel) {
		core.WriteError(w, http.StatusBadRequest, "invalid path")
		return
	}
	s.mu.Lock()
	st, err := s.Storage.Load(workspaceID)
	s.mu.Unlock()
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	m, ok := st.Files[rel]
	if !ok || m.IsDeleted {
		core.WriteError(w, http.StatusNotFound, "file not found")
		return
	}
	abs, err := s.Storage.FilePath(workspaceID, rel)
	if err != nil {
		core.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Devsync-Sha256", m.SHA256)
	http.ServeFile(w, r, abs)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request, workspaceID string) {
	rel := core.NormalizeRelativePath(r.URL.Query().Get("path"))
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = "unknown"
	}
	if !core.IsSafeRelativePath(rel) {
		core.WriteError(w, http.StatusBadRequest, "invalid path")
		return
	}
	abs, err := s.Storage.FilePath(workspaceID, rel)
	if err == nil {
		_ = os.Remove(abs)
	}
	s.mu.Lock()
	err = s.Storage.DeleteFile(workspaceID, rel, clientID)
	s.mu.Unlock()
	if err != nil {
		core.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	core.WriteJSON(w, http.StatusOK, map[string]string{"status": fmt.Sprintf("deleted %s", rel)})
}
