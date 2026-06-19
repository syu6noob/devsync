APP_NAME := devsync
MODULE := github.com/syu6noob/devsync
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(MODULE)/internal/core.Version=$(VERSION) -X $(MODULE)/internal/core.Commit=$(COMMIT) -X $(MODULE)/internal/core.Date=$(DATE)

.PHONY: all build build-windows build-linux-server test fmt clean docker docker-run snapshot

all: fmt test build

build:
	mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/devsync ./cmd/devsync
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/devsync-server ./cmd/devsync-server

build-windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/devsync.exe ./cmd/devsync

build-linux-server:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/devsync-server-linux-amd64 ./cmd/devsync-server

test:
	go test ./...

fmt:
	gofmt -w cmd internal

clean:
	rm -rf bin dist release

docker:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) -t devsync-server:local .

docker-run:
	docker compose up --build

snapshot:
	goreleaser release --snapshot --clean
