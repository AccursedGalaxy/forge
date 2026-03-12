package stream

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const heartbeatInterval = 15 * time.Second

// SSEHandler returns an http.HandlerFunc that streams events for a session via SSE.
// sessionID is extracted by the caller from the URL path parameter.
func (b *Broadcaster) SSEHandler(sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
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

		ctx := r.Context()
		events := b.Subscribe(ctx, sessionID)

		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		slog.Info("sse: client connected", "session_id", sessionID)
		defer slog.Info("sse: client disconnected", "session_id", sessionID)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-events:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case <-ticker.C:
				// SSE comment heartbeat keeps the connection alive
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}
}
