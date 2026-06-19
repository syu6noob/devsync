# DevSync

DevSync is a small development-file synchronization tool written in Go.

It synchronizes files between Windows clients and a Linux server, with Git-like ignore rules using `.devsyncignore`.

> Current status: MVP. Use it for experiments first.

## Features

- File-based synchronization
- SHA-256 based change detection
- `.devsyncignore` support
- `status`, `push`, `pull`, `sync`
- Conflict detection
- Conflict copy generation
- Linux server
- Docker server runtime
- GitHub Actions cross-builds

## Repository

This project assumes:

```text
github.com/syu6noob/devsync
```

## Build locally

```bash
go test ./...
go build -o bin/devsync ./cmd/devsync
go build -o bin/devsync-server ./cmd/devsync-server
```

or:

```bash
make build
```

## Run the server locally

```bash
./bin/devsync-server -addr :8080 -data ./data -token secret
```

Environment variables are also supported:

```bash
DEVSYNC_ADDR=:8080 DEVSYNC_DATA_DIR=./data DEVSYNC_TOKEN=secret ./bin/devsync-server
```

## Run the server with Docker

```bash
docker compose up --build -d
```

Default compose settings:

- port: `8080`
- data volume: `devsync-data`
- token: `change-me`

Change the token before exposing the server.

## Client usage

Initialize a workspace:

```bash
devsync init --workspace my-project --client pc-a --server http://localhost:8080 --token secret
```

Check local state:

```bash
devsync status
```

Synchronize:

```bash
devsync sync
```

Push only:

```bash
devsync push
```

Pull only:

```bash
devsync pull
```

## Ignore rules

Create `.devsyncignore` in the workspace root.

Example:

```gitignore
.devsync/
.git/
node_modules/
dist/
build/
.env
.env.*
*.log
```

## GitHub Actions

This repository includes:

- CI workflow
- cross-platform build workflow artifacts
- tag-based release workflow
- Docker image build workflow

Create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds:

- Linux amd64 / arm64
- Windows amd64 / arm64
- macOS amd64 / arm64

## Smoke test

```bash
scripts/smoke-test.sh
```

## Security notes

- Use HTTPS in production, for example by putting the server behind Caddy or Nginx.
- Change `DEVSYNC_TOKEN` before exposing the server.
- Do not commit `.devsync/config.json` if it contains a real token.

## License

No license has been selected yet. Add a license before publishing as open source.
