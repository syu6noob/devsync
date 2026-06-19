APP_NAME := devsync
MODULE := github.com/syu6noob/devsync
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(MODULE)/internal/core.Version=$(VERSION) -X $(MODULE)/internal/core.Commit=$(COMMIT) -X $(MODULE)/internal/core.Date=$(DATE)

.PHONY: all build test fmt clean docker docker-run

all: fmt test build

build:
	mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/devsync ./cmd/devsync
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/devsync-server ./cmd/devsync-server

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
