package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/syu6noob/devsync/internal/client"
	"github.com/syu6noob/devsync/internal/core"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	var err error
	switch cmd {
	case "init":
		err = cmdInit(os.Args[2:])
	case "status":
		err = withRoot(func(root string) error { return client.Status(root) })
	case "sync":
		err = withRoot(func(root string) error { return client.Run(root, client.ModeSync) })
	case "push":
		err = withRoot(func(root string) error { return client.Run(root, client.ModePush) })
	case "pull":
		err = withRoot(func(root string) error { return client.Run(root, client.ModePull) })
	case "version":
		fmt.Printf("devsync %s commit=%s date=%s\n", core.Version, core.Commit, core.Date)
		return
	case "help", "-h", "--help":
		usage()
		return
	default:
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	workspace := fs.String("workspace", filepath.Base(mustGetwd()), "workspace id")
	clientID := fs.String("client", hostname(), "client id")
	serverURL := fs.String("server", "http://localhost:8080", "server url")
	token := fs.String("token", "", "api token")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := mustGetwd()
	cfg := client.Config{ServerURL: *serverURL, WorkspaceID: *workspace, ClientID: *clientID, Token: *token}
	if err := client.SaveConfig(root, cfg); err != nil {
		return err
	}
	ignorePath := filepath.Join(root, ".devsyncignore")
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		content := []byte(`# DevSync ignore rules
.devsync/
.git/
node_modules/
dist/
build/
.venv/
__pycache__/
*.pyc
*.log
.env
.env.*
`)
		if err := os.WriteFile(ignorePath, content, 0o644); err != nil {
			return err
		}
	}
	fmt.Println("initialized devsync workspace")
	fmt.Println("workspace:", *workspace)
	fmt.Println("client:", *clientID)
	fmt.Println("server:", *serverURL)
	return nil
}

func withRoot(fn func(root string) error) error {
	wd := mustGetwd()
	root, err := client.FindWorkspaceRoot(wd)
	if err != nil {
		return err
	}
	return fn(root)
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "client"
	}
	return h
}

func usage() {
	fmt.Print(`devsync - development file sync client

Usage:
  devsync init [--workspace NAME] [--client NAME] [--server URL] [--token TOKEN]
  devsync status
  devsync sync
  devsync push
  devsync pull
  devsync version
`)
}
