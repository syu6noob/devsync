package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syu6noob/devsync/internal/core"
	"github.com/syu6noob/devsync/internal/ignore"
)

type Mode string

const (
	ModeSync Mode = "sync"
	ModePush Mode = "push"
	ModePull Mode = "pull"
)

func Status(root string) error {
	cfg, err := LoadConfig(root)
	if err != nil {
		return err
	}
	st, err := LoadState(root)
	if err != nil {
		return err
	}
	matcher, err := ignore.Load(filepath.Join(root, ".devsyncignore"), ignore.DefaultRules())
	if err != nil {
		return err
	}
	scan, err := Scan(root, matcher)
	if err != nil {
		return err
	}
	fmt.Printf("Workspace: %s\nClient: %s\nServer: %s\n\n", cfg.WorkspaceID, cfg.ClientID, cfg.ServerURL)
	printLocalDiff(scan.Files, st.Files)
	fmt.Printf("\nIgnored entries: %d\n", scan.IgnoredCount)
	return nil
}

func printLocalDiff(local, base map[string]core.FileMeta) {
	added, modified, deleted := []string{}, []string{}, []string{}
	for p, f := range local {
		b, ok := base[p]
		if !ok || b.IsDeleted {
			added = append(added, p)
			continue
		}
		if b.SHA256 != f.SHA256 {
			modified = append(modified, p)
		}
	}
	for p, b := range base {
		if b.IsDeleted {
			continue
		}
		if _, ok := local[p]; !ok {
			deleted = append(deleted, p)
		}
	}
	for _, v := range []struct {
		name  string
		items []string
	}{
		{"added", added}, {"modified", modified}, {"deleted", deleted},
	} {
		if len(v.items) == 0 {
			continue
		}
		fmt.Println(v.name + ":")
		for _, p := range v.items {
			fmt.Println("  " + p)
		}
	}
	if len(added) == 0 && len(modified) == 0 && len(deleted) == 0 {
		fmt.Println("No local changes.")
	}
}

func Run(root string, mode Mode) error {
	cfg, err := LoadConfig(root)
	if err != nil {
		return err
	}
	st, err := LoadState(root)
	if err != nil {
		return err
	}
	matcher, err := ignore.Load(filepath.Join(root, ".devsyncignore"), ignore.DefaultRules())
	if err != nil {
		return err
	}
	scan, err := Scan(root, matcher)
	if err != nil {
		return err
	}

	remote := NewRemote(cfg)
	plan, err := remote.Plan(scan.Files, st.Files)
	if err != nil {
		return err
	}

	if len(plan.Operations) == 0 {
		fmt.Println("Already up to date.")
		return updateStateFromServer(root, cfg, remote)
	}

	for _, op := range plan.Operations {
		switch op.Action {
		case "upload":
			if mode == ModePull {
				continue
			}
			fmt.Println("upload:", op.Path)
			if err := remote.Upload(root, op.Path); err != nil {
				return err
			}
		case "download":
			if mode == ModePush {
				continue
			}
			fmt.Println("download:", op.Path)
			if err := remote.Download(root, op.Path); err != nil {
				return err
			}
		case "delete_remote":
			if mode == ModePull {
				continue
			}
			fmt.Println("delete remote:", op.Path)
			if err := remote.DeleteRemote(op.Path); err != nil {
				return err
			}
		case "conflict":
			if mode == ModePush {
				fmt.Println("conflict:", op.Path, "- skipped in push mode")
				continue
			}
			fmt.Println("conflict:", op.Path, "- keeping local copy and downloading server version")
			if err := keepConflictCopy(root, cfg.ClientID, op.Path); err != nil {
				return err
			}
			if err := remote.Download(root, op.Path); err != nil {
				return err
			}
		default:
			// noop or unknown future operation
		}
	}

	return updateStateFromServer(root, cfg, remote)
}

func updateStateFromServer(root string, cfg Config, remote Remote) error {
	manifest, err := remote.FetchManifest()
	if err != nil {
		return err
	}
	return SaveState(root, State{
		WorkspaceID: cfg.WorkspaceID,
		ClientID:    cfg.ClientID,
		Files:       manifest.Files,
	})
}

func keepConflictCopy(root, clientID, rel string) error {
	abs, err := core.SafeJoin(root, rel)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil
	}
	stamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(abs)
	base := strings.TrimSuffix(abs, ext)
	conflictPath := fmt.Sprintf("%s.conflict-%s-%s%s", base, sanitizeClientID(clientID), stamp, ext)
	input, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	return os.WriteFile(conflictPath, input, 0o644)
}

func sanitizeClientID(s string) string {
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	return s
}
