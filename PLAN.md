# FORGE — Implementation Plan

> Master checklist. Work top to bottom. Don't start a step until the previous is checked off.
> Reference `docs/SPEC.md` for architecture decisions. Reference `docs/DESIGN.md` for all UI decisions.

-----

## STEP 1 — Project Baseline

> Goal: working skeleton — frontend renders, backend serves, logs flow from both.

### 1.1 Frontend — Component Foundation

- [x] Init React + Vite + TypeScript in `web/`
- [x] Install Tailwind CSS, configure `tailwind.config.ts`
- [x] Create `web/styles/tokens.css` — all CSS variables from DESIGN.md (colors, spacing, typography, radius, motion)
- [x] Create `web/styles/globals.css` — reset, base styles, font imports (Geist, Geist Mono, Instrument Serif)
- [x] Configure fonts via `@font-face` or CDN import

**Primitive components** (`web/src/components/ui/`)

- [x] `Button.tsx` — variants: primary, ghost, danger; sizes: sm, md, lg
- [x] `Badge.tsx` — variants: status, autonomy, count; reads from token colors
- [x] `Input.tsx` — with label and error props
- [x] `Card.tsx` — with optional hoverable state
- [x] `Modal.tsx` — backdrop blur, escape to close
- [x] `Skeleton.tsx` — shimmer animation for loading states

**Feature components** (`web/src/components/forge/`)

- [x] `Sidebar.tsx` — fixed 240px, project switcher, nav items, user section
- [x] `TopBar.tsx` — sticky, breadcrumb, action buttons
- [x] `TaskCard.tsx` — title, description, autonomy badge, session status dot + pulse
- [x] `KanbanColumn.tsx` — header with status badge + count, task list, empty state
- [x] `KanbanBoard.tsx` — 5 columns, horizontal scroll
- [x] `SessionStream.tsx` — fixed right panel, slide-in, stream line rendering, interrupt button
- [x] `ProjectSwitcher.tsx` — dropdown with projects
- [x] `PlanApprovalPanel.tsx` — numbered plan steps, approve/reject buttons ⚠️ added beyond original plan

**Pages** (`web/src/pages/`)

- [x] `LandingPage.tsx` — hero (Instrument Serif italic), how it works, features grid, pricing, footer
- [x] `DashboardPage.tsx` — KanbanBoard + SessionStream panel, full session workflow
- [x] `AppShell.tsx` — layout wrapper: Sidebar + TopBar + content area
- [x] `ContextPage.tsx` — route registered at `/dashboard/context`
- [x] `SettingsPage.tsx` — route registered at `/dashboard/settings`
- [x] `LogsPage.tsx` — route registered at `/dashboard/logs`

**Routing + wiring**

- [x] Set up React Router — `/` → LandingPage, `/dashboard` → AppShell + DashboardPage
- [x] Mock data skipped — went straight to real API hooks ⚠️ no `mockData.ts`

-----

### 1.2 Backend — Go Framework

- [x] Init Go module in `agent/` — `go mod init github.com/accursedgalaxy/forge`
- [x] Install dependencies: `chi`, `golang-migrate`, `sqlc`, `asynq`, `go-redis`, `anthropic-go`
- [x] `cmd/forge/main.go` — entrypoint: loads config, wires all dependencies, starts server + worker
- [x] `internal/config/` — config struct, loads from env vars + `.env` file

**Router + middleware** (`internal/api/`)

- [x] Chi router setup with base middleware: RequestID, RealIP, Logger, Recoverer, CORS
- [x] `internal/auth/middleware.go` — no-op stub v1; reads `FORGE_SECRET_KEY` env var, checks `X-Forge-Key` header if set; interface ready for Clerk v2
- [x] Health check: `GET /api/health` → `{ "status": "ok", "version": "0.1.0" }`
- [x] 404 and 500 error handlers with consistent JSON response shape: `{ "error": "message" }`

**Handler stubs** — routes registered, return `501 Not Implemented` with TODO comment

- [x] Projects: GET /api/projects, POST, GET /:id, PATCH /:id, DELETE /:id
- [x] Tasks: GET /api/projects/:id/tasks, POST, PATCH /api/tasks/:id, DELETE, PATCH reorder
- [x] Sessions: GET /api/tasks/:id/sessions, POST /api/sessions, GET /:id, POST approve, interrupt, resume
- [x] Stream: GET /api/sessions/:id/stream
- [x] Context: GET /api/projects/:id/context, DELETE /api/context/:id
- [x] Providers: GET /api/providers, POST, PATCH /:id

**Provider interface** (`internal/provider/`)

- [x] `provider.go` — define `AgentProvider` interface (Run, Resume, Interrupt, Capabilities)
- [x] `claude/claude.go` — Claude provider struct, fully implemented, wraps runner.Manager
- [x] `registry.go` — provider registry, register claude as default

**Confirm**

- [x] `go build ./...` succeeds with zero errors
- [x] `curl localhost:8080/api/health` returns 200

-----

### 1.3 Unified Logging System

> Rolling logs, structured, searchable from both frontend and backend. Easy to access in dev and prod.

**Backend logging** (`internal/logs/`)

- [x] Use `slog` (stdlib) with JSON handler in production, text handler in dev — `MultiHandler` wraps primary handler + SSE broadcaster
- [x] Log levels: DEBUG, INFO, WARN, ERROR — controlled by `LOG_LEVEL` env var
- [x] Structured fields on every log line + request-specific fields
- [x] Request logger middleware: logs every request in + out with duration
- [x] Rolling log file: `logs/forge.log` — lumberjack (100 MB max, 7-day retention, 5 backups, gzip)
- [x] Logs also stream to stdout — `io.MultiWriter(os.Stdout, roller)` in `main.go`

**Frontend logging** (`web/src/lib/logger.ts`)

- [x] Thin logger wrapper: `log.info()`, `log.warn()`, `log.error()`, `log.debug()`
- [x] In prod: buffers logs and ships to `POST /api/logs`
- [x] Frontend log format: `{ timestamp, level, component, message, meta? }`
- [x] Backend `POST /api/logs` endpoint: ingests `[]browserLogEntry`, writes via slog with `source=browser`

**Log viewer** (`web/src/components/forge/LogViewer.tsx`)

- [x] `GET /api/logs/stream` — SSE endpoint: flushes last 200 buffered lines, then tails live; 15s heartbeat
- [x] Log viewer panel: level filter buttons (ALL/DEBUG/INFO/WARN/ERROR), auto-scroll (pauses on scroll-up), level badges, attrs with tooltip, max 2000 lines, clear button
- [x] Route `/dashboard/logs` registered; "Logs" nav item in `Sidebar.tsx`

-----

## STEP 2 — Backend Hardening

> Goal: real database and queue connected, all API stubs replaced with working handlers.

### 2.1 PostgreSQL Setup

- [x] Add Postgres service to `docker-compose.yml` with persistent volume
- [x] Write migration files using `golang-migrate` format ⚠️ located at `agent/internal/db/migrations/`, not `agent/migrations/` as planned
  - [x] `001_create_users.sql`
  - [x] `002_create_projects.sql`
  - [x] `003_create_tasks.sql`
  - [x] `004_create_sessions.sql`
  - [x] `005_create_context_chunks.sql` — with pgvector extension (`CREATE EXTENSION IF NOT EXISTS vector`)
  - [x] `006_create_agent_providers.sql`
- [x] Run migrations on startup — `internal/db/migrate.go` (auto-migrate, idempotent)
- [x] `internal/db/` — sqlc config + generated queries for all tables
- [x] Write sqlc queries for: CRUD on all entities, session status updates, task reorder, context chunk insert + vector search
- [x] On first launch: auto-create default user record (single-user v1 mode)
- [x] On first launch: seed default Claude provider record

### 2.2 Redis + Queue Setup

- [x] Add Redis service to `docker-compose.yml`
- [x] `internal/worker/` — Asynq client + server setup (10 concurrent workers, 5min shutdown timeout)
- [x] Define job types: `TypePlanSession`, `TypeExecuteSession`, `TypeResumeSession` ⚠️ job types differ slightly from original plan
- [x] Worker server starts alongside HTTP server in `main.go`
- [x] `internal/stream/` — SSE broadcaster using Redis pub/sub for fan-out across connections
- [x] SSE manager: handles client connect/disconnect, heartbeat every 15s, clean shutdown

### 2.3 Working API Handlers

Replace all 501 stubs with real implementations:

**Projects**

- [x] `GET /api/projects` — list all projects, ordered by created_at desc
- [x] `POST /api/projects` — validate, insert, return created
- [x] `GET /api/projects/:id` — fetch with task counts by status
- [x] `PATCH /api/projects/:id` — partial update (name, description, repo_url)
- [x] `DELETE /api/projects/:id` — cascade delete tasks + sessions

**Tasks**

- [x] `GET /api/projects/:id/tasks` — list tasks, grouped by status, ordered by position
- [x] `POST /api/projects/:id/tasks` — insert with default status=backlog, position=last
- [x] `PATCH /api/tasks/:id` — update title, description, autonomy_level, status
- [x] `DELETE /api/tasks/:id` — delete task + sessions
- [x] `PATCH /api/tasks/:id/reorder` — update position + status, reorder siblings

**Sessions**

- [x] `GET /api/tasks/:id/sessions` — list sessions for task, ordered by created_at desc
- [x] `POST /api/sessions` — create session record, enqueue plan job, return session
- [x] `GET /api/sessions/:id` — fetch session with plan_steps
- [x] `POST /api/sessions/:id/approve` — validate status=awaiting_approval, enqueue execute job
- [x] `POST /api/sessions/:id/interrupt` — send interrupt signal to runner, update status=paused
- [x] `POST /api/sessions/:id/resume` — validate paused, enqueue resume job with correction prompt

**Providers**

- [x] `GET /api/providers` — list registered providers
- [x] `POST /api/providers` — register new provider config
- [x] `PATCH /api/providers/:id` — update provider config, set default

-----

## STEP 3 — Agent Runner

> Goal: claude-cli runs, streams back to frontend, full plan/execute/interrupt/resume loop works end to end.

### 3.1 Process Manager

- [x] `internal/runner/manager.go` — thread-safe map of active session PIDs; Start(), Interrupt()
- [x] `internal/runner/process.go` — spawns claude-cli via `os/exec`
  - [x] Builds arg list: `--print --output-format stream-json --include-partial-messages --verbose`
  - [x] Strips `CLAUDECODE` from env (prevents refusal inside parent Claude session)
  - [x] For resume: prepends `--resume <claude_session_id>`
  - [x] For plan sessions: `--allowedTools Glob,Grep,Read,Bash,LS`
  - [x] For execute sessions: `--dangerously-skip-permissions`
- [x] `internal/runner/parser.go` — parses stream-json line by line
  - [x] `system/init` → extract session_id, store in DB
  - [x] `assistant` messages with `content_block_delta` → route text/thinking/tool events
  - [x] `result` → capture final output (EventTypeDone / EventTypeError)
  - [x] stderr → forward as error event

### 3.2 Orchestrator ⚠️ new package not in original plan

- [x] `internal/orchestrator/orchestrator.go` — session lifecycle engine; wired by main.go
  - [x] `PlanSession()` — transitions to "planning", builds context-enhanced prompt (Sonnet 4.6), spawns claude-cli with read-only tools, extracts plan steps via LLM (Sonnet 4.6 JSON), stores plan_steps, transitions to "awaiting_approval", emits `claude:done`
  - [x] `ExecuteSession()` — transitions to "running", builds execute prompt with plan steps, spawns claude-cli with full permissions, summarizes output (Haiku), marks session "done" + task "review", emits `claude:done`, indexes output async
  - [x] `ResumeSession()` — prepends correction prompt, resumes via `--resume <claudeSessionID>`, full execution loop
  - [x] `Interrupt()` — delegates to runner.Manager.Interrupt()
  - [x] On exit code 0 → Haiku summary stored as `claude_notes`, task moves to "review"
  - [x] On non-zero exit → error classified, `error_message` stored, status → "error"
- [x] `internal/orchestrator/prompt.go` — prompt builders for plan, execute, resume phases

### 3.3 SSE Stream

- [x] `GET /api/sessions/:id/stream` — SSE handler registered and functional
- [x] Sets headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `X-Accel-Buffering: no`
- [x] Subscribes to Redis pub/sub channel for session
- [x] Sends heartbeat comment every 15s
- [x] Cleans up on client disconnect
- [x] Event types emitted to frontend:
  - [x] `claude:start` — session began, includes session_id and phase
  - [x] `claude:stream` — text chunk, includes type (text/thinking/tool) and content
  - [x] `claude:done` — session complete, includes claude_notes and plan_steps
  - [x] `claude:error` — session failed, includes error_message
  - [x] `session:status` — status change broadcast

### 3.4 Asynq Job Handlers

- [x] `worker/handler_plan.go` — handles TypePlanSession → calls `orchestrator.PlanSession()`
- [x] `worker/handler_execute.go` — handles TypeExecuteSession → calls `orchestrator.ExecuteSession()`
- [x] `worker/handler_resume.go` — handles TypeResumeSession → calls `orchestrator.ResumeSession()`

### 3.5 LLM Integration

- [x] `internal/llm/client.go` — Anthropic SDK wrapper, initialized with ANTHROPIC_API_KEY
- [x] `internal/llm/summarizer.go` — BuildContextPrompt() Sonnet 4.6, ReviewPlan() Sonnet 4.6 JSON extraction, Summarize() Haiku; all methods fall back gracefully on error
- [ ] `internal/llm/embedder.go` — **STUB**: returns empty vector ⚠️ Anthropic has no public embeddings API; needs Voyage AI or OpenAI as provider
- [x] `internal/context/retriever.go` — pgvector similarity search, returns top-K chunks; no-ops gracefully when embedding is empty
- [x] `internal/context/indexer.go` — post-session: chunks output (1000-char chunks, 100-char overlap), embeds, stores; skips silently when embedder returns empty vector

**Confirm**

- [ ] Create a task, POST /api/sessions → claude-cli spawns (visible in process list)
- [ ] SSE stream receives events, logs show stream-json parsing working
- [ ] Plan session completes → status flips to awaiting_approval
- [ ] Approve → execution session runs
- [ ] Interrupt works: process terminates cleanly, status → paused
- [ ] Resume works: –resume flag used, session continues

-----

## STEP 4 — Frontend Wiring

> Goal: frontend talks to real backend. No more mock data. Full user flow works end to end.

### 4.1 API Client

- [x] `web/src/lib/api.ts` — typed fetch wrapper
  - [x] Base URL from `VITE_API_URL` env var (defaults to `http://localhost:8080`)
  - [x] Attaches `X-Forge-Key` header if `VITE_FORGE_KEY` is set
  - [x] Returns typed responses, throws typed errors
  - [x] Retry on network failure (3x, exponential backoff: 200ms / 400ms / 800ms)
- [x] `web/src/lib/sse.ts` — SSE client
  - [x] Connects to `/api/sessions/:id/stream`
  - [x] Typed event handlers: `claude:start`, `claude:stream`, `claude:done`, `claude:error`, `session:status`
  - [x] Auto-reconnect on disconnect (max 5 attempts, exponential backoff)
  - [x] Cleanup on unmount

### 4.2 Hooks

- [x] `useProjects.ts` — fetch project list, create project, delete project
- [x] `useTasks.ts` — fetch tasks by project, create/update/delete/reorder
- [x] `useSession.ts` — fetch session, approve, interrupt, resume
- [x] `useStream.ts` — SSE connection, accumulates stream lines, extracts plan_steps, exposes status
- [x] `useProviders.ts` — fetch and configure agent providers
- [x] `useContext.ts` — fetch and delete context chunks ⚠️ added beyond original plan

### 4.3 Page Wiring

- [x] `DashboardPage` — real project list in sidebar, real task board, full session workflow
- [ ] `KanbanBoard` — drag-to-reorder calls PATCH reorder endpoint ⚠️ UI only, not wired
- [x] `TaskCard` — clicking opens task detail / session panel
- [x] `SessionStream` — connects to real SSE stream, renders live output with auto-scroll
- [x] Plan approval UI — `PlanApprovalPanel.tsx` renders plan_steps, approve/reject buttons call API
- [ ] Error state: session failed → show error message + recovery suggestion with retry button
- [ ] Loading states: Skeleton components while data fetches
- [ ] Empty states: no projects, no tasks, no sessions — each has a clear empty state with action

### 4.4 Context Browser

- [x] `ContextPage.tsx` — route registered at `/dashboard/context`
- [ ] Verify: shows chunk type, content preview, created date — context API endpoints still return 501
- [ ] `GET /api/projects/:id/context` — implement handler (currently 501)
- [ ] `DELETE /api/context/:id` — implement handler (currently 501)

### 4.5 Settings Page

- [x] `SettingsPage.tsx` — route registered at `/dashboard/settings`
- [ ] Verify: provider config, API key input, danger zone fully wired

### 4.6 Install Story

- [x] `docker-compose.yml` — Postgres + Redis with healthchecks and volumes
- [ ] Add Go binary service to `docker-compose.yml` ⚠️ missing — backend must be run separately
- [x] `Makefile` targets: `make dev`, `make build`, `make migrate`, `make docker`
- [ ] `npx forge-init` script — scaffolds docker-compose.yml + .env.example in current dir
- [x] `.env.example` — all required env vars documented with descriptions
- [ ] Go binary embeds compiled frontend assets via `//go:embed`

**Confirm**

- [ ] Full flow: create project → create task → run agent → watch stream → approve plan → execution completes → task moves to review
- [ ] Interrupt and resume work from UI
- [ ] `docker compose up` from cold start works with zero manual steps
- [ ] `npx forge-init && docker compose up` works in a fresh directory

-----

## STEP 5 — Remaining Work

> Items not yet done, grouped by priority.

### 5.1 Embeddings Provider (blocks context memory feature)

- [ ] Choose embeddings provider: Voyage AI (`voyage-3-lite`) or OpenAI (`text-embedding-3-small`)
- [ ] Implement `internal/llm/embedder.go` with chosen provider
- [ ] Wire VOYAGE_API_KEY or OPENAI_API_KEY into config + .env.example
- [ ] Verify: post-session indexing stores vectors, pre-execution retrieval injects context

### 5.2 Context API Endpoints

- [ ] `GET /api/projects/:id/context` — list context chunks for project (type, content preview, created_at)
- [ ] `DELETE /api/context/:id` — delete single chunk

### 5.3 Frontend Polish

- [ ] KanbanBoard drag-to-reorder → wire to PATCH reorder endpoint
- [ ] Error state for failed sessions: error message + recovery suggestion + retry button
- [ ] Loading states: Skeleton components while data fetches
- [ ] Empty states: no projects, no tasks, no sessions — clear empty state with action CTA

### 5.4 Single-Binary Deploy

- [ ] Go binary embeds compiled frontend via `//go:embed`
- [ ] Add Go binary service to `docker-compose.yml`
- [ ] `npx forge-init` scaffolds docker-compose.yml + .env in fresh directory
- [ ] Cold-start smoke test: `docker compose up` from zero → full app running

### 5.5 End-to-End Smoke Test

- [ ] Create project → create task → POST /api/sessions → claude-cli spawns
- [ ] SSE stream delivers events to frontend in real time
- [ ] Plan completes → awaiting_approval → approve → execution runs → task → review
- [ ] Interrupt: SIGTERM sent, status → paused
- [ ] Resume: --resume flag used, session continues
