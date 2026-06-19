package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/syu6noob/devsync/internal/core"
	"github.com/syu6noob/devsync/internal/server"
)

func main() {
	addr := flag.String("addr", envOrDefault("DEVSYNC_ADDR", ":8080"), "listen address")
	dataDir := flag.String("data", envOrDefault("DEVSYNC_DATA_DIR", "./data"), "data directory")
	token := flag.String("token", os.Getenv("DEVSYNC_TOKEN"), "api token; empty disables auth")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("devsync-server %s commit=%s date=%s\n", core.Version, core.Commit, core.Date)
		return
	}

	s := server.New(*dataDir, *token)
	fmt.Println("devsync-server listening on", *addr)
	fmt.Println("data directory:", *dataDir)
	if *token == "" {
		fmt.Println("warning: token auth is disabled")
	}
	log.Fatal(http.ListenAndServe(*addr, s.Handler()))
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
