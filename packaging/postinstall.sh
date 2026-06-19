#!/bin/sh
set -e

if command -v useradd >/dev/null 2>&1; then
  if ! id devsync >/dev/null 2>&1; then
    useradd --system --home-dir /var/lib/devsync --shell /usr/sbin/nologin devsync || true
  fi
fi

mkdir -p /var/lib/devsync /etc/devsync-server
chown -R devsync:devsync /var/lib/devsync 2>/dev/null || true

if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload || true
  systemctl enable devsync-server.service || true
fi
