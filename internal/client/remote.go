package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/syu6noob/devsync/internal/core"
)

type Remote struct {
	Config Config
	HTTP   *http.Client
}

func NewRemote(c Config) Remote {
	return Remote{Config: c, HTTP: http.DefaultClient}
}

func (r Remote) endpoint(path string, q url.Values) string {
	base := strings.TrimRight(r.Config.ServerURL, "/")
	if q != nil {
		return base + path + "?" + q.Encode()
	}
	return base + path
}

func (r Remote) do(req *http.Request) (*http.Response, error) {
	if r.Config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Config.Token)
	}
	return r.HTTP.Do(req)
}

func (r Remote) Plan(local map[string]core.FileMeta, base map[string]core.FileMeta) (core.SyncPlan, error) {
	body := core.SyncRequest{
		WorkspaceID: r.Config.WorkspaceID,
		ClientID:    r.Config.ClientID,
		Files:       local,
		Base:        base,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return core.SyncPlan{}, err
	}
	url := r.endpoint(fmt.Sprintf("/api/v1/workspaces/%s/plan", url.PathEscape(r.Config.WorkspaceID)), nil)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return core.SyncPlan{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := r.do(req)
	if err != nil {
		return core.SyncPlan{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(res.Body)
		return core.SyncPlan{}, fmt.Errorf("server returned %s: %s", res.Status, strings.TrimSpace(string(msg)))
	}
	var plan core.SyncPlan
	return plan, json.NewDecoder(res.Body).Decode(&plan)
}

func (r Remote) FetchManifest() (core.WorkspaceManifest, error) {
	url := r.endpoint(fmt.Sprintf("/api/v1/workspaces/%s/manifest", url.PathEscape(r.Config.WorkspaceID)), nil)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return core.WorkspaceManifest{}, err
	}
	res, err := r.do(req)
	if err != nil {
		return core.WorkspaceManifest{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(res.Body)
		return core.WorkspaceManifest{}, fmt.Errorf("server returned %s: %s", res.Status, strings.TrimSpace(string(msg)))
	}
	var m core.WorkspaceManifest
	return m, json.NewDecoder(res.Body).Decode(&m)
}

func (r Remote) Upload(root, rel string) error {
	abs, err := core.SafeJoin(root, rel)
	if err != nil {
		return err
	}
	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer f.Close()
	q := url.Values{"path": {rel}, "client_id": {r.Config.ClientID}}
	u := r.endpoint(fmt.Sprintf("/api/v1/workspaces/%s/file", url.PathEscape(r.Config.WorkspaceID)), q)
	req, err := http.NewRequest(http.MethodPut, u, f)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := r.do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(res.Body)
		return fmt.Errorf("upload failed for %s: %s: %s", rel, res.Status, strings.TrimSpace(string(msg)))
	}
	return nil
}

func (r Remote) Download(root, rel string) error {
	abs, err := core.SafeJoin(root, rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	q := url.Values{"path": {rel}}
	u := r.endpoint(fmt.Sprintf("/api/v1/workspaces/%s/file", url.PathEscape(r.Config.WorkspaceID)), q)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	res, err := r.do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(res.Body)
		return fmt.Errorf("download failed for %s: %s: %s", rel, res.Status, strings.TrimSpace(string(msg)))
	}
	tmp := abs + ".devsync-tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, res.Body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	return os.Rename(tmp, abs)
}

func (r Remote) DeleteRemote(rel string) error {
	q := url.Values{"path": {rel}, "client_id": {r.Config.ClientID}}
	u := r.endpoint(fmt.Sprintf("/api/v1/workspaces/%s/file", url.PathEscape(r.Config.WorkspaceID)), q)
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	res, err := r.do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(res.Body)
		return fmt.Errorf("remote delete failed for %s: %s: %s", rel, res.Status, strings.TrimSpace(string(msg)))
	}
	return nil
}
