// Package api sets up the Chi HTTP router, all middleware, and route registration.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/accursedgalaxy/forge/internal/auth"
	"github.com/accursedgalaxy/forge/internal/db"
	"github.com/accursedgalaxy/forge/internal/logs"
	"github.com/accursedgalaxy/forge/internal/provider"
	"github.com/accursedgalaxy/forge/internal/stream"
)

const version = "0.1.0"

// Options holds the dependencies wired into the router at startup.
type Options struct {
	SecretKey      string
	Registry       *provider.Registry
	DB             *db.Queries
	Pool           *pgxpool.Pool
	RedisClient    *redis.Client
	AsynqClient    *asynq.Client
	Broadcaster    *stream.Broadcaster
	LogBroadcaster *logs.Broadcaster
}

// ErrorResponse is the consistent JSON shape for all error responses.
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewRouter builds and returns the configured Chi router.
func NewRouter(opts Options) *chi.Mux {
	r := chi.NewRouter()

	// ── Core middleware ─────────────────────────────────────────────────────
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger())
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Forge-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// ── Health check — always public ────────────────────────────────────────
	r.Get("/api/health", handleHealth)

	// ── Protected routes — optional FORGE_SECRET_KEY enforcement ───────────
	r.Group(func(r chi.Router) {
		authn := auth.New(opts.SecretKey)
		r.Use(authn.Middleware())

		// ── Project handlers ─────────────────────────────────────────────
		ph := &ProjectHandler{db: opts.DB, pool: opts.Pool}
		r.Get("/api/projects", ph.List)
		r.Post("/api/projects", ph.Create)
		r.Get("/api/projects/{id}", ph.Get)
		r.Patch("/api/projects/{id}", ph.Update)
		r.Delete("/api/projects/{id}", ph.Delete)

		// ── Task handlers ────────────────────────────────────────────────
		th := &TaskHandler{db: opts.DB, pool: opts.Pool}
		r.Get("/api/projects/{id}/tasks", th.List)
		r.Post("/api/projects/{id}/tasks", th.Create)
		r.Patch("/api/tasks/{id}", th.Update)
		r.Delete("/api/tasks/{id}", th.Delete)
		r.Patch("/api/tasks/{id}/reorder", th.Reorder)

		// ── Session handlers ─────────────────────────────────────────────
		sh := &SessionHandler{
			db:          opts.DB,
			pool:        opts.Pool,
			asynqClient: opts.AsynqClient,
			broadcaster: opts.Broadcaster,
		}
		r.Get("/api/tasks/{id}/sessions", sh.List)
		r.Post("/api/sessions", sh.Create)
		r.Get("/api/sessions/{id}", sh.Get)
		r.Post("/api/sessions/{id}/approve", sh.Approve)
		r.Post("/api/sessions/{id}/interrupt", sh.Interrupt)
		r.Post("/api/sessions/{id}/resume", sh.Resume)
		r.Get("/api/sessions/{id}/stream", sh.Stream)

		// ── Context (stub — implemented in Step 3) ────────────────────────
		r.Get("/api/projects/{id}/context", stub("GET /api/projects/:id/context"))
		r.Delete("/api/context/{id}", stub("DELETE /api/context/:id"))

		// ── Provider handlers ─────────────────────────────────────────────
		provh := &ProviderHandler{db: opts.DB, pool: opts.Pool}
		r.Get("/api/providers", provh.List)
		r.Post("/api/providers", provh.Create)
		r.Patch("/api/providers/{id}", provh.Update)

		// ── Log endpoints ─────────────────────────────────────────────────
		lh := &LogsHandler{broadcaster: opts.LogBroadcaster}
		r.Post("/api/logs", lh.Ingest)
		r.Get("/api/logs/stream", lh.Stream)
	})

	// ── Custom error handlers ───────────────────────────────────────────────
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	return r
}

// handleHealth returns a 200 with version info. Always public.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": version,
	})
}

// stub returns a handler that replies 501 Not Implemented.
func stub(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented: "+route)
	}
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a consistent JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// requestLogger is structured slog middleware that logs every request + response.
func requestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			slog.Info("http",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
