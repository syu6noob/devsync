#!/usr/bin/env bash
set -euo pipefail

tmp=$(mktemp -d)
cleanup() {
  if [[ -n "${server_pid:-}" ]]; then
    kill "$server_pid" 2>/dev/null || true
  fi
  rm -rf "$tmp"
}
trap cleanup EXIT

root=$(pwd)
server_bin="$root/bin/devsync-server"
client_bin="$root/bin/devsync"

mkdir -p "$root/bin"
go build -o "$server_bin" ./cmd/devsync-server
go build -o "$client_bin" ./cmd/devsync

"$server_bin" -addr 127.0.0.1:18080 -data "$tmp/server" -token secret >"$tmp/server.log" 2>&1 &
server_pid=$!

for i in {1..30}; do
  if curl -fsS http://127.0.0.1:18080/healthz >/dev/null; then
    break
  fi
  sleep 0.2
done

mkdir -p "$tmp/a" "$tmp/b"
(
  cd "$tmp/a"
  "$client_bin" init --workspace smoke --client a --server http://127.0.0.1:18080 --token secret
  echo hello > README.md
  "$client_bin" sync
)
(
  cd "$tmp/b"
  "$client_bin" init --workspace smoke --client b --server http://127.0.0.1:18080 --token secret
  "$client_bin" sync
  test "$(cat README.md)" = "hello"
)

echo "smoke test passed"
