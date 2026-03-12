#!/usr/bin/env bash
# 02_create_project_task.sh — create a project and task, print IDs to .smoke_env

BASE="http://localhost:8080/api"
ENV_FILE="$(dirname "$0")/.smoke_env"

echo "==> Creating project"
PROJECT_JSON=$(curl -sf -X POST "$BASE/projects" \
  -H "Content-Type: application/json" \
  -d '{"name":"Smoke Test","description":"Backend smoke test","repo_url":"https://github.com/test/repo"}')
echo "$PROJECT_JSON" | jq .
PROJECT_ID=$(echo "$PROJECT_JSON" | jq -r '.id')

if [[ "$PROJECT_ID" == "null" || -z "$PROJECT_ID" ]]; then
  echo "ERROR: failed to create project" >&2
  exit 1
fi

echo ""
echo "==> Creating task"
TASK_JSON=$(curl -sf -X POST "$BASE/projects/$PROJECT_ID/tasks" \
  -H "Content-Type: application/json" \
  -d '{"title":"Write a hello world function","description":"Create a Python function called greet(name) that returns a greeting string. Include a docstring and a usage example.","autonomy_level":"supervised"}')
echo "$TASK_JSON" | jq .
TASK_ID=$(echo "$TASK_JSON" | jq -r '.id')

if [[ "$TASK_ID" == "null" || -z "$TASK_ID" ]]; then
  echo "ERROR: failed to create task" >&2
  exit 1
fi

echo ""
echo "==> Verifying task is in backlog"
curl -sf "$BASE/projects/$PROJECT_ID/tasks" | jq '[.[] | {id, title, status}]'

echo ""
echo "PROJECT_ID=$PROJECT_ID" > "$ENV_FILE"
echo "TASK_ID=$TASK_ID" >> "$ENV_FILE"
echo "Saved to $ENV_FILE"
echo "  PROJECT_ID=$PROJECT_ID"
echo "  TASK_ID=$TASK_ID"
