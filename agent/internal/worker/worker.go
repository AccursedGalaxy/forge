// Package worker wraps the asynq server for background job processing.
package worker

import (
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/accursedgalaxy/forge/internal/orchestrator"
)

// Server wraps an asynq.Server with a simple Start/Stop API.
type Server struct {
	srv *asynq.Server
}

// New creates a Worker Server connected to Redis at the given address.
// ShutdownTimeout is set to 5 minutes to allow long-running sessions to complete gracefully.
func New(redisOpt asynq.RedisClientOpt) *Server {
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency:     10,
		ShutdownTimeout: 5 * time.Minute,
		Logger:          &asynqLogger{},
	})
	return &Server{srv: srv}
}

// Start runs the asynq server with the provided mux. Blocks until Stop is called.
func (s *Server) Start(mux *asynq.ServeMux) error {
	slog.Info("worker: starting")
	return s.srv.Run(mux)
}

// Stop gracefully shuts down the worker server.
func (s *Server) Stop() {
	slog.Info("worker: stopping")
	s.srv.Shutdown()
}

// RegisterHandlers wires all worker task handlers into the asynq mux.
func RegisterHandlers(mux *asynq.ServeMux, orch *orchestrator.Orchestrator) {
	plan := &planSessionHandler{orch: orch}
	execute := &executeSessionHandler{orch: orch}
	resume := &resumeSessionHandler{orch: orch}

	mux.HandleFunc(TypePlanSession, plan.ProcessTask)
	mux.HandleFunc(TypeExecuteSession, execute.ProcessTask)
	mux.HandleFunc(TypeResumeSession, resume.ProcessTask)
}

// asynqLogger adapts asynq's logger interface to slog.
type asynqLogger struct{}

func (l *asynqLogger) Debug(args ...interface{}) { slog.Debug("asynq", "msg", args) }
func (l *asynqLogger) Info(args ...interface{})  { slog.Info("asynq", "msg", args) }
func (l *asynqLogger) Warn(args ...interface{})  { slog.Warn("asynq", "msg", args) }
func (l *asynqLogger) Error(args ...interface{}) { slog.Error("asynq", "msg", args) }
func (l *asynqLogger) Fatal(args ...interface{}) { slog.Error("asynq fatal", "msg", args) }
