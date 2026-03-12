#!/usr/bin/env bash
# 06_cleanup.sh — delete the smoke test project (cascades to tasks + sessions)

BASE="http://localhost:8080/api"
ENV_FILE="$(dirname "$0")/.smoke_env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "ERROR: $ENV_FILE not found" >&2
  exit 1
fi
source "$ENV_FILE"

echo "==> Deleting project $PROJECT_ID (cascades to tasks + sessions)"
STATUS=$(curl -sf -o /dev/null -w "%{http_code}" -X DELETE "$BASE/projects/$PROJECT_ID")
if [[ "$STATUS" == "204" ]]; then
  echo "OK: project deleted"
else
  echo "ERROR: got HTTP $STATUS" >&2
  exit 1
fi

rm -f "$ENV_FILE"
echo "Cleaned up $ENV_FILE"
