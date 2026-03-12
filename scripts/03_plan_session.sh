#!/usr/bin/env bash
# 03_plan_session.sh — create a plan session and stream its SSE output
# Expects .smoke_env from script 02 to be present.

BASE="http://localhost:8080/api"
ENV_FILE="$(dirname "$0")/.smoke_env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "ERROR: $ENV_FILE not found — run 02_create_project_task.sh first" >&2
  exit 1
fi
source "$ENV_FILE"

echo "==> Creating plan session (project=$PROJECT_ID, task=$TASK_ID)"
SESSION_JSON=$(curl -sf -X POST "$BASE/sessions" \
  -H "Content-Type: application/json" \
  -d "{\"task_id\":\"$TASK_ID\",\"project_id\":\"$PROJECT_ID\",\"session_type\":\"plan\"}")
echo "$SESSION_JSON" | jq .
SESSION_ID=$(echo "$SESSION_JSON" | jq -r '.id')

if [[ "$SESSION_ID" == "null" || -z "$SESSION_ID" ]]; then
  echo "ERROR: failed to create session" >&2
  exit 1
fi

echo "SESSION_ID=$SESSION_ID" >> "$ENV_FILE"
echo ""
echo "==> Streaming SSE (Ctrl+C to stop, then run 04_approve.sh)"
echo "    Session ID: $SESSION_ID"
echo ""
curl -N "$BASE/sessions/$SESSION_ID/stream"
