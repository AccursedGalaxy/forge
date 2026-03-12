#!/usr/bin/env bash
# 04_approve.sh — check plan output, approve, and stream execution SSE
# Expects .smoke_env from scripts 02+03 to be present.

BASE="http://localhost:8080/api"
ENV_FILE="$(dirname "$0")/.smoke_env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "ERROR: $ENV_FILE not found — run previous scripts first" >&2
  exit 1
fi
source "$ENV_FILE"

echo "==> Session status before approval"
curl -sf "$BASE/sessions/$SESSION_ID" | jq '{status, plan_steps}'

echo ""
STATUS=$(curl -sf "$BASE/sessions/$SESSION_ID" | jq -r '.status')
if [[ "$STATUS" != "awaiting_approval" ]]; then
  echo "WARNING: session status is '$STATUS', expected 'awaiting_approval'"
  echo "         Plan may still be running — check 03_plan_session.sh output"
  read -p "Approve anyway? [y/N] " yn
  [[ "$yn" != "y" ]] && exit 0
fi

echo ""
echo "==> Approving plan"
curl -sf -X POST "$BASE/sessions/$SESSION_ID/approve" | jq .

echo ""
echo "==> Streaming execution SSE (Ctrl+C when done)"
curl -N "$BASE/sessions/$SESSION_ID/stream"
