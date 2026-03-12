package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/accursedgalaxy/forge/internal/logs"
)

// LogsHandler serves POST /api/logs (browser log ingestion) and
// GET /api/logs/stream (SSE tail for the dev log viewer).
type LogsHandler struct {
	broadcaster *logs.Broadcaster
}

// browserLogEntry is the shape sent by the frontend logger.
type browserLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	Meta      any    `json:"meta,omitempty"`
}

// Ingest accepts a JSON array of log entries from the browser and
// writes each one to slog with source=browser.
func (h *LogsHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	var entries []browserLogEntry
	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		writeError(w, http.StatusBadRequest, "invalid log entries")
		return
	}

	for _, e := range entries {
		level := slog.LevelInfo
		switch e.Level {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}

		attrs := []slog.Attr{
			slog.String("source", "browser"),
			slog.String("component", e.Component),
		}
		if e.Meta != nil {
			attrs = append(attrs, slog.Any("meta", e.Meta))
		}

		slog.LogAttrs(r.Context(), level, e.Message, attrs...)
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stream sends a SSE stream of log lines to the client.
// Recent buffered lines are flushed first, then live lines follow.
func (h *LogsHandler) Stream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Flush recent history immediately
	for _, line := range h.broadcaster.Recent(200) {
		fmt.Fprintf(w, "data: %s\n\n", line)
	}
	flusher.Flush()

	ctx := r.Context()
	ch, cancel := h.broadcaster.Subscribe(ctx)
	defer cancel()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
