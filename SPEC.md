# FORGE — Project Specification

> Multi-agent AI coding orchestration platform.
> Open source, self-hostable, built for developers who ship.

-----

## What It Is

FORGE wraps AI coding agents in a structured project management layer. Developers define tasks on a kanban board, set an autonomy level per task, and FORGE orchestrates the agent loop — plan, approve, execute, review. The human controls how much the agent does unsupervised.

The gap FORGE fills: existing agent frameworks nail automation but fail at the human interface. Intervention surfaces are terrible. FORGE owns that gap.

-----

## Architecture

```
React + Vite (SPA)
  ├── REST API calls ────────────► Go Backend (Chi)
  └── SSE stream ────────────────►     │
                                       ├── PostgreSQL  (all persistent state)
                                       ├── Redis       (job queue + SSE pub/sub)
                                       ├── claude-cli  (spawned processes)
                                       └── Anthropic SDK (direct LLM calls)

Auth (v1):
  None — single-user local installs need no auth.
  Optional: FORGE_SECRET_KEY env var, Go middleware checks header if set.
  Stub in internal/auth/ ready for Clerk/SSO when team features land.
```

No Convex. No Supabase. The Go binary owns everything stateful.

-----

## Tech Stack

|Layer          |Technology                                                     |
|---------------|---------------------------------------------------------------|
|Frontend       |React 18 + Vite + TypeScript                                   |
|Backend        |Go + Chi router                                                |
|Database       |PostgreSQL 16 + pgvector                                       |
|Queue          |Redis + Asynq                                                  |
|Auth           |None (v1) — optional secret key header for API protection      |
|Agent (primary)|claude-cli via `os/exec`                                       |
|Agent (other)  |Anthropic Go SDK, provider interface for plugins               |
|Self-hosted    |docker-compose — Postgres + Redis + Go binary + static frontend|
|Cloud install  |`npx forge-init` scaffolds docker-compose + .env               |

-----

## Repo Structure

```
forge/
  web/                        # React + Vite frontend
    src/
      components/
        ui/                   # primitives: Button, Badge, Input, Card, Modal
        forge/                # features: TaskCard, KanbanBoard, SessionStream, Sidebar
      pages/                  # route-level components
      hooks/                  # useSession, useProjects, useTasks, useStream
      lib/                    # API client, SSE client, auth helpers
    styles/
      tokens.css              # CSS variables (colors, spacing, typography)
      globals.css

  agent/                      # Go backend
    cmd/forge/
      main.go                 # entrypoint
    internal/
      api/                    # Chi router, all HTTP handlers
      auth/                   # auth middleware stub (no-op v1, Clerk-ready for v2)
      db/                     # sqlc generated queries + migrations
      runner/                 # claude-cli spawn, stream-json parsing, process manager
      provider/               # AgentProvider interface + claude impl + plugin stubs
      stream/                 # SSE broadcaster, Redis pub/sub fan-out
      worker/                 # Asynq job handlers for async sessions
      llm/                    # Direct Anthropic SDK calls (summarization, embeddings)
      context/                # pgvector retrieval for pre-session context injection
    migrations/               # SQL migration files

  docker-compose.yml
  Makefile
  docs/
    SPEC.md                   # this file
    DESIGN.md                 # UI design system
```

-----

## Data Model

### users

```sql
id           UUID PRIMARY KEY DEFAULT gen_random_uuid()
name         TEXT NOT NULL
email        TEXT
created_at   TIMESTAMPTZ DEFAULT now()
```

> v1: single implicit user, created on first launch. No login flow.
> v2+: multi-user with Clerk JWT, SSO, org membership.

### projects

```sql
id           UUID PRIMARY KEY
owner_id     UUID REFERENCES users
name         TEXT NOT NULL
description  TEXT
repo_url     TEXT NOT NULL           -- any git remote
created_at   TIMESTAMPTZ
```

### tasks

```sql
id             UUID PRIMARY KEY
project_id     UUID REFERENCES projects
title          TEXT NOT NULL
description    TEXT
status         TEXT CHECK (status IN ('backlog','planned','in_progress','review','done'))
autonomy_level TEXT CHECK (autonomy_level IN ('supervised','checkpoint','autonomous'))
position       INTEGER                -- sort order within status column
created_at     TIMESTAMPTZ
```

### sessions

```sql
id               UUID PRIMARY KEY
task_id          UUID REFERENCES tasks
project_id       UUID REFERENCES projects
session_type     TEXT CHECK (session_type IN ('plan','execute'))
status           TEXT CHECK (status IN ('pending','planning','awaiting_approval','running','paused','done','error'))
claude_session_id TEXT                -- claude-cli --resume ID
plan_steps       JSONB               -- structured steps from plan session
claude_notes     TEXT                -- post-session Haiku summary
error_message    TEXT
started_at       TIMESTAMPTZ
completed_at     TIMESTAMPTZ
created_at       TIMESTAMPTZ
```

### context_chunks

```sql
id           UUID PRIMARY KEY
project_id   UUID REFERENCES projects
session_id   UUID REFERENCES sessions
chunk_type   TEXT CHECK (chunk_type IN ('session_output','code_diff','task_note'))
content      TEXT
embedding    vector(1536)            -- pgvector
created_at   TIMESTAMPTZ
```

### agent_providers

```sql
id            UUID PRIMARY KEY
name          TEXT NOT NULL
provider_type TEXT CHECK (provider_type IN ('claude','codex','ollama'))
bin_path      TEXT                   -- override binary path
is_default    BOOLEAN
config        JSONB
```

-----

## Agent Provider Interface

All providers implement this Go interface. Claude Code is the default. Other providers register as plugins — no dynamic linking, just registered structs.

```go
type AgentProvider interface {
    Run(ctx context.Context, task Task, opts RunOptions) (<-chan Event, error)
    Resume(ctx context.Context, sessionID string, prompt string) (<-chan Event, error)
    Interrupt(sessionID string) error
    Capabilities() ProviderCaps
}
```

The Go runner is stateless — it spawns processes, streams output, writes results back to Postgres. All state lives in the DB.

-----

## Core Flows

### 1. Task execution — supervised

1. User creates task, sets `autonomy_level = supervised`
1. Frontend calls `POST /api/sessions` → Go creates session record, enqueues plan job
1. Go runner spawns `claude-cli -p <prompt> --output-format stream-json`
1. Stream events forwarded to SSE endpoint → frontend renders live
1. Plan completes → session status → `awaiting_approval`
1. Frontend renders plan steps, user approves or rejects
1. On approval: `POST /api/sessions/:id/execute` → Go runs execution session with plan as context
1. On non-zero exit: Go classifies failure, generates recovery prompt suggestion
1. On success: Go calls Anthropic SDK (Haiku) to summarize → stores `claude_notes`, moves task to `review`

### 2. Task execution — autonomous

1. User creates task, sets `autonomy_level = autonomous`
1. `POST /api/sessions` → job enqueued in Redis via Asynq
1. Worker picks up job, runs full plan + execute loop without UI intervention
1. On completion or error: task status updated, SSE event broadcast to any connected clients

### 3. Interrupt and resume

1. User clicks Interrupt in UI → `POST /api/sessions/:id/interrupt`
1. Go sends SIGTERM to claude-cli process, waits for clean exit
1. Last N stream events returned as context summary
1. User edits correction prompt, submits → `POST /api/sessions/:id/resume`
1. Go resumes via `--resume <claude_session_id>` with correction prepended to prompt

### 4. Context retrieval

1. Post-session: Go calls Anthropic embeddings API on session output + diffs
1. Stores vectors in pgvector
1. Pre-execution: top-K relevant chunks retrieved, injected into task prompt
1. Users can browse and delete context chunks per project

-----

## API Surface

```
All routes optionally protected by FORGE_SECRET_KEY header check (if env var is set).
No auth in default local install.

Projects
  GET    /api/projects
  POST   /api/projects
  GET    /api/projects/:id
  PATCH  /api/projects/:id
  DELETE /api/projects/:id

Tasks
  GET    /api/projects/:id/tasks
  POST   /api/projects/:id/tasks
  PATCH  /api/tasks/:id
  DELETE /api/tasks/:id
  PATCH  /api/tasks/:id/reorder

Sessions
  GET    /api/tasks/:id/sessions
  POST   /api/sessions                    -- create + enqueue
  GET    /api/sessions/:id
  POST   /api/sessions/:id/approve        -- approve plan, trigger execution
  POST   /api/sessions/:id/interrupt
  POST   /api/sessions/:id/resume

Stream
  GET    /api/sessions/:id/stream         -- SSE endpoint

Context
  GET    /api/projects/:id/context
  DELETE /api/context/:id

Providers
  GET    /api/providers
  POST   /api/providers
  PATCH  /api/providers/:id
```

-----

## Feature Scope

|Feature                          |v1 (OSS)|Paid Cloud|
|---------------------------------|--------|----------|
|Task board (kanban)              |✓       |✓         |
|Claude Code orchestration        |✓       |✓         |
|Plan approval gate               |✓       |✓         |
|Interrupt + resume               |✓       |✓         |
|Session history + stream log     |✓       |✓         |
|Project context memory (pgvector)|✓       |✓         |
|Any git remote                   |✓       |✓         |
|Self-hosted docker install       |✓       |✓         |
|Agent provider plugin interface  |✓       |✓         |
|Single-user, no auth required    |✓       |✓         |
|Optional API secret key          |✓       |✓         |
|Team / org management            |—       |✓         |
|Multi-user collaboration         |—       |✓         |
|SSO (SAML, OIDC)                 |—       |✓         |
|Usage analytics dashboard        |—       |✓         |
|Managed cloud hosting            |—       |✓         |
|Audit logs                       |—       |✓         |

-----

## Open Core Model

License: Apache 2.0.

**Rule:** agent orchestration, task board, session history, context memory — never paywalled. These are the product. Paywalling them invites forks.

**Paid tier charges for:** hosted cloud, team/org management, SSO, audit logs, analytics, support SLA.

**Install story:**

- `docker compose up` — full self-hosted (Postgres + Redis + Go binary + frontend)
- `npx forge-init` — scaffolds docker-compose.yml + .env, then `docker compose up`
- Single Go binary with embedded frontend assets for minimal installs

-----

## v1 Milestones

|Milestone          |Scope                                                                                         |
|-------------------|----------------------------------------------------------------------------------------------|
|M1 Foundation      |Go scaffold, Postgres schema, sqlc, Chi router, optional secret key middleware, docker-compose|
|M2 Agent Runner    |claude-cli spawn, stream-json parsing, SSE broadcaster, Redis pub/sub, session persistence    |
|M3 Task Board      |REST API for projects/tasks, React kanban UI, task CRUD, status transitions                   |
|M4 Plan Gate       |Plan session flow, structured step rendering, approve/reject, execution trigger               |
|M5 Interrupt/Resume|SIGTERM handling, context capture, correction prompt UI, –resume flow                         |
|M6 Context Memory  |pgvector setup, post-session embedding, pre-execution retrieval, context browser UI           |
|M7 Install Story   |npx forge-init, docker-compose polish, single binary build, README                            |
|M8 v1 Release      |Docs site, GitHub Actions CI/CD, public launch                                                |

-----

## Out of Scope for v1

- Team / org features
- Slack / Discord / email notifications
- Self-improving agent feedback loops
- Fine-tuned models
- Mobile UI
- Non-coding task support
