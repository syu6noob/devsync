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
- Windows client binary
- Linux server binary
- Docker server runtime
- GitHub Actions cross-builds
- GoReleaser releases
- `.deb` / `.rpm` packages for the Linux server
- Optional APT repository generated with GitHub Pages

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

## Build Windows client locally

From Linux/macOS/WSL:

```bash
make build-windows
```

or directly:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o devsync.exe ./cmd/devsync
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
- Windows client build workflow
- tag-based GoReleaser release workflow
- Docker image build workflow
- APT repository deployment workflow

### Build `devsync.exe` from GitHub Actions

Open the `Build Windows Client` workflow and run it manually, or push to `main`.
The artifact is uploaded as `devsync-windows-amd64-exe`.

### Create a release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds:

- `devsync.exe` for Windows amd64 / arm64
- `devsync` for Linux amd64 / arm64
- `devsync` for macOS amd64 / arm64
- `devsync-server` for Linux amd64 / arm64
- `devsync-server` `.deb` packages
- `devsync-server` `.rpm` packages
- checksums

## Install Linux server from packages

See [docs/INSTALL.md](docs/INSTALL.md).

For Ubuntu/Debian, after GitHub Pages is enabled and the APT workflow has deployed:

```bash
echo "deb [trusted=yes] https://syu6noob.github.io/devsync/apt stable main" | sudo tee /etc/apt/sources.list.d/devsync.list
sudo apt update
sudo apt install devsync-server
```

The generated APT repository is unsigned for simplicity. Use it for private/testing usage first, or add GPG signing before public production use.

## Smoke test

```bash
scripts/smoke-test.sh
```

## Security notes

- Use HTTPS in production, for example by putting the server behind Caddy or Nginx.
- Change `DEVSYNC_TOKEN` before exposing the server.
- Do not commit `.devsync/config.json` if it contains a real token.
- Sign APT repository metadata before treating the package repository as production-ready.

## License

No license has been selected yet. Add a license before publishing as open source.
