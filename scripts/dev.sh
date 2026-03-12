#!/usr/bin/env bash
# Starts the Go API server and Vite frontend concurrently.
# Called by `make dev`; must be run from the project root.

API_PID=""
WEB_PID=""

cleanup() {
    trap - EXIT INT TERM
    [ -n "$API_PID" ] && kill "$API_PID" 2>/dev/null
    [ -n "$WEB_PID" ] && kill "$WEB_PID" 2>/dev/null
    wait 2>/dev/null
    exit 0   # always exit 0 — Ctrl-C is normal for a dev server
}
trap cleanup EXIT INT TERM

if command -v air &>/dev/null; then
    (cd agent && air) &
else
    (cd agent && go run ./cmd/forge) &
fi
API_PID=$!

(cd web && bun run dev) &
WEB_PID=$!

wait "$API_PID" "$WEB_PID"
