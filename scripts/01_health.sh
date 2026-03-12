#!/usr/bin/env bash
# 01_health.sh — verify the backend is up

BASE="http://localhost:8080/api"

echo "==> Health check"
curl -sf "$BASE/health" | jq .

echo ""
echo "==> Registered providers"
curl -sf "$BASE/providers" | jq .
