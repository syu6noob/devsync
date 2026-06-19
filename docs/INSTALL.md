# Installation

## Windows client

Download `devsync-client_<version>_windows_amd64.zip` from GitHub Releases and extract `devsync.exe`.

You can also build it locally:

```powershell
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -trimpath -o devsync.exe ./cmd/devsync
```

## Linux server with Docker

```bash
docker compose up --build -d
```

## Linux server with DEB package

On Ubuntu/Debian, download the `.deb` file from GitHub Releases:

```bash
sudo apt install ./devsync-server_<version>_linux_amd64.deb
sudo nano /etc/devsync-server/devsync-server.env
sudo systemctl restart devsync-server
```

## Linux server from APT repository

After GitHub Pages is enabled and the `APT Repository` workflow has deployed the repository:

```bash
echo "deb [trusted=yes] https://syu6noob.github.io/devsync/apt stable main" | sudo tee /etc/apt/sources.list.d/devsync.list
sudo apt update
sudo apt install devsync-server
```

The repository generated here is unsigned for simplicity. For public production use, add GPG signing later.

## Linux server with RPM package

On RPM-based distributions, download the `.rpm` file from GitHub Releases:

```bash
sudo rpm -i devsync-server_<version>_linux_amd64.rpm
sudo nano /etc/devsync-server/devsync-server.env
sudo systemctl restart devsync-server
```
