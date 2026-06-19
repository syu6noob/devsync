#!/bin/sh
set -e

if command -v systemctl >/dev/null 2>&1; then
  systemctl stop devsync-server.service >/dev/null 2>&1 || true
  systemctl disable devsync-server.service >/dev/null 2>&1 || true
fi
