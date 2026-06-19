package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ServerURL   string `json:"server_url"`
	WorkspaceID string `json:"workspace_id"`
	ClientID    string `json:"client_id"`
	Token       string `json:"token"`
}

func DevsyncDir(root string) string { return filepath.Join(root, ".devsync") }
func ConfigPath(root string) string { return filepath.Join(DevsyncDir(root), "config.json") }
func StatePath(root string) string  { return filepath.Join(DevsyncDir(root), "state.json") }

func LoadConfig(root string) (Config, error) {
	b, err := os.ReadFile(ConfigPath(root))
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.ServerURL == "" || c.WorkspaceID == "" || c.ClientID == "" {
		return Config{}, errors.New("config.json requires server_url, workspace_id and client_id")
	}
	return c, nil
}

func SaveConfig(root string, c Config) error {
	if err := os.MkdirAll(DevsyncDir(root), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(root), b, 0o600)
}

func FindWorkspaceRoot(start string) (string, error) {
	cur, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(cur, ".devsync")); err == nil {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf(".devsync directory not found")
		}
		cur = parent
	}
}
