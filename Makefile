.PHONY: install dev dev-api dev-web dev-tmux \
        build build-api build-web \
        generate migrate lint \
        services docker down logs clean help

# ── Colours ──────────────────────────────────────────────────────────────────
CYAN  := \033[36m
RESET := \033[0m

# ── Dev ──────────────────────────────────────────────────────────────────────

## install: Install frontend dependencies via bun
install:
	cd web && bun install

## dev: Start backend + frontend concurrently (Ctrl-C stops both)
dev:
	@printf "$(CYAN)▶ starting FORGE dev servers…$(RESET)\n"
	@trap 'kill 0' SIGINT; \
		( if command -v air &>/dev/null; then \
		    cd agent && air; \
		  else \
		    cd agent && go run ./cmd/forge; \
		  fi ) & \
		( cd web && bun run dev ) & \
		wait

## dev-api: Backend only (air hot-reload if installed, else go run)
dev-api:
	@if command -v air &>/dev/null; then \
	    cd agent && air; \
	  else \
	    printf "$(CYAN)▶ air not found — using go run (run: go install github.com/air-verse/air@latest)$(RESET)\n"; \
	    cd agent && go run ./cmd/forge; \
	  fi

## dev-web: Frontend only
dev-web:
	cd web && bun run dev

## dev-tmux: Open a tmux session: api / web / shell windows
dev-tmux:
	@tmux new-session -d -s forge -n api \
	    'if command -v air &>/dev/null; then cd agent && air; else cd agent && go run ./cmd/forge; fi' 2>/dev/null || true
	@tmux new-window -t forge -n web   'cd web && bun run dev'
	@tmux new-window -t forge -n shell
	@tmux select-window -t forge:api
	@tmux attach-session -t forge

# ── Build ─────────────────────────────────────────────────────────────────────

## build: Compile Go binary + production frontend bundle
build: build-api build-web

## build-api: Compile Go binary → agent/bin/forge
build-api:
	cd agent && go build -o bin/forge ./cmd/forge

## build-web: Type-check + Vite production bundle → web/dist
build-web:
	cd web && bun run build

# ── Code generation + DB ──────────────────────────────────────────────────────

## generate: Re-run sqlc code generation
generate:
	cd agent/internal/db && sqlc generate

## migrate: Run database migrations (requires DATABASE_URL)
migrate:
	cd agent && go run ./cmd/forge migrate

# ── Quality ───────────────────────────────────────────────────────────────────

## lint: Run ESLint on the frontend
lint:
	cd web && bun run lint

# ── Docker ────────────────────────────────────────────────────────────────────

## services: Start only Postgres + Redis in the background
services:
	docker compose up -d postgres redis

## docker: Full stack build — Postgres + Redis + app
docker:
	docker compose up --build

## down: Stop and remove containers
down:
	docker compose down

## logs: Tail docker compose output
logs:
	docker compose logs -f

# ── Misc ──────────────────────────────────────────────────────────────────────

## clean: Remove build artefacts
clean:
	rm -rf agent/bin web/dist

## help: List all available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /' | column -t -s ':'
