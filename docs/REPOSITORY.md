# Repository guide

## Repository name

This project assumes the following module path:

```text
github.com/syu6noob/devsync
```

Create the repository as:

```bash
gh repo create syu6noob/devsync --public --source=. --remote=origin --push
```

or push manually:

```bash
git init
git add .
git commit -m "Initial DevSync MVP"
git branch -M main
git remote add origin git@github.com:syu6noob/devsync.git
git push -u origin main
```

## Local build

```bash
make build
```

## Server with Docker

```bash
docker compose up --build -d
```

The server listens on port `8080` and stores data in the `devsync-data` Docker volume.

## GitHub Actions

- `.github/workflows/ci.yml`
  - runs `gofmt`, `go test`, normal builds, Docker build, and cross-builds.
  - uploads executable archives as workflow artifacts.

- `.github/workflows/release.yml`
  - runs on `v*` tags.
  - builds Linux, Windows, and macOS binaries.
  - creates a GitHub Release and uploads executable archives.

- `.github/workflows/docker.yml`
  - builds the server Docker image.
  - on version tags, pushes `ghcr.io/syu6noob/devsync-server:<tag>` and `latest`.

## Release

```bash
git tag v0.1.0
git push origin v0.1.0
```
