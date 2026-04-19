#!/usr/bin/env bash
set -e

trap 'kill 0' EXIT

export ADMIN_TOKEN="${ADMIN_TOKEN:-admin}"

cd internal/web/frontend && npm run dev &
cd "$OLDPWD" && go run . &

sleep 2

echo ""
echo "========================================"
echo "  ADMIN_TOKEN: $ADMIN_TOKEN"
echo "  Backend:     http://localhost:8080"
echo "  Dashboard:   http://localhost:5173/dashboard"
echo "========================================"
echo ""

wait
