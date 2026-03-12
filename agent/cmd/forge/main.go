package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxstdlib "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/accursedgalaxy/forge/internal/api"
	"github.com/accursedgalaxy/forge/internal/config"
	"github.com/accursedgalaxy/forge/internal/db"
	"github.com/accursedgalaxy/forge/internal/logs"
	"github.com/accursedgalaxy/forge/internal/provider"
	claudeprovider "github.com/accursedgalaxy/forge/internal/provider/claude"
	"github.com/accursedgalaxy/forge/internal/stream"
	"github.com/accursedgalaxy/forge/internal/worker"
)

func main() {
	cfg := config.Load()

	// ── Logging ──────────────────────────────────────────────────────────────
	logBroadcaster := logs.NewBroadcaster()
	setupLogger(cfg, logBroadcaster)

	slog.Info("starting FORGE",
		"version", "0.1.0",
		"env", cfg.Env,
		"port", cfg.Port,
	)

	ctx := context.Background()

	// ── PostgreSQL ───────────────────────────────────────────────────────────
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db: failed to create pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, cfg.DatabaseURL); err != nil {
		slog.Error("db: migrations failed", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// sqlc queries use database/sql interface via pgx stdlib adapter
	sqlDB := pgxstdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()
	queries := db.New(sqlDB)

	if err := db.EnsureDefaultUser(ctx, queries); err != nil {
		slog.Error("db: seed default user failed", "err", err)
		os.Exit(1)
	}
	if err := db.EnsureDefaultProvider(ctx, queries); err != nil {
		slog.Error("db: seed default provider failed", "err", err)
		os.Exit(1)
	}

	// ── Redis ────────────────────────────────────────────────────────────────
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("redis: invalid URL", "err", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		slog.Error("redis: ping failed", "err", err)
		os.Exit(1)
	}
	slog.Info("redis connected")

	// ── Asynq job queue ──────────────────────────────────────────────────────
	asynqRedis := asynq.RedisClientOpt{Addr: redisOpts.Addr, Password: redisOpts.Password, DB: redisOpts.DB}
	asynqClient := asynq.NewClient(asynqRedis)
	defer asynqClient.Close()

	// ── SSE broadcaster (Redis pub/sub) ──────────────────────────────────────
	broadcaster := stream.NewBroadcaster(redisClient)

	// ── Worker server ────────────────────────────────────────────────────────
	workerServer := worker.New(asynqRedis)
	workerMux := asynq.NewServeMux()
	worker.RegisterHandlers(workerMux, queries, pool, broadcaster)

	go func() {
		if err := workerServer.Start(workerMux); err != nil {
			slog.Error("worker: failed to start", "err", err)
		}
	}()
	defer workerServer.Stop()

	// ── Provider registry ────────────────────────────────────────────────────
	registry := provider.NewRegistry()
	registry.Register("claude", claudeprovider.New(""))
	slog.Info("provider registered", "name", "claude", "default", registry.DefaultName())

	// ── HTTP router ──────────────────────────────────────────────────────────
	router := api.NewRouter(api.Options{
		SecretKey:      cfg.ForgeSecretKey,
		Registry:       registry,
		DB:             queries,
		Pool:           pool,
		RedisClient:    redisClient,
		AsynqClient:    asynqClient,
		Broadcaster:    broadcaster,
		LogBroadcaster: logBroadcaster,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // no timeout for SSE streams
		IdleTimeout:  120 * time.Second,
	}

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	workerServer.Stop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "err", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}

// setupLogger configures the global slog logger.
//
// Output:
//   - stdout: text in dev, JSON in prod
//   - logs/forge.log: always JSON, rotated by lumberjack (100 MB, 7 days, 5 backups)
//
// All records are also published to logBroadcaster for SSE streaming.
func setupLogger(cfg *config.Config, logBroadcaster *logs.Broadcaster) {
	opts := &slog.HandlerOptions{Level: cfg.SlogLevel()}

	// Ensure logs/ directory exists
	_ = os.MkdirAll("logs", 0o755)

	roller := &lumberjack.Logger{
		Filename:   "logs/forge.log",
		MaxSize:    100, // MB
		MaxAge:     7,   // days
		MaxBackups: 5,
		Compress:   true,
	}

	// Primary output writer: stdout + rolling file
	multi := io.MultiWriter(os.Stdout, roller)

	var primaryHandler slog.Handler
	if cfg.IsDev() {
		primaryHandler = slog.NewTextHandler(multi, opts)
	} else {
		primaryHandler = slog.NewJSONHandler(multi, opts)
	}

	// Wrap with the broadcasting handler so SSE clients receive all records
	handler := logs.NewMultiHandler(primaryHandler, logBroadcaster)
	slog.SetDefault(slog.New(handler))
}
