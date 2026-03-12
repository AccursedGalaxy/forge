#!/usr/bin/env bash
# 05_review_and_done.sh — move task through review → done and show final state

BASE="http://localhost:8080/api"
ENV_FILE="$(dirname "$0")/.smoke_env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "ERROR: $ENV_FILE not found — run previous scripts first" >&2
  exit 1
fi
source "$ENV_FILE"

echo "==> Session final state"
curl -sf "$BASE/sessions/$SESSION_ID" | jq '{status, plan_steps, error}'

echo ""
echo "==> Moving task to 'review'"
curl -sf -X PATCH "$BASE/tasks/$TASK_ID" \
  -H "Content-Type: application/json" \
  -d '{"status":"review"}' | jq '{id, title, status}'

echo ""
echo "==> Moving task to 'done'"
curl -sf -X PATCH "$BASE/tasks/$TASK_ID" \
  -H "Content-Type: application/json" \
  -d '{"status":"done"}' | jq '{id, title, status}'

echo ""
echo "==> All sessions for task"
curl -sf "$BASE/tasks/$TASK_ID/sessions" | jq '[.[] | {id, session_type, status, plan_steps}]'

echo ""
echo "==> All tasks in project"
curl -sf "$BASE/projects/$PROJECT_ID/tasks" | jq '[.[] | {title, status, position}]'
