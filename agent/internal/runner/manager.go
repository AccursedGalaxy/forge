package runner

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/accursedgalaxy/forge/internal/provider"
)

// Manager is a thread-safe registry of live claude-cli processes keyed by Forge session ID.
type Manager struct {
	binPath string

	mu        sync.Mutex
	processes map[string]*os.Process
}

// NewManager creates a Manager that spawns claude-cli via binPath.
// Pass an empty string to resolve "claude" from PATH.
func NewManager(binPath string) *Manager {
	if binPath == "" {
		binPath = "claude"
	}
	return &Manager{
		binPath:   binPath,
		processes: make(map[string]*os.Process),
	}
}

// Start spawns a claude-cli subprocess for the given session and returns a channel of events.
// The channel is closed when the process exits. The process is removed from the registry
// automatically on exit.
func (m *Manager) Start(ctx context.Context, opts SpawnOptions) (<-chan provider.Event, error) {
	proc, rawEvents, err := spawn(ctx, m.binPath, opts)
	if err != nil {
		return nil, fmt.Errorf("manager: spawn: %w", err)
	}

	m.mu.Lock()
	m.processes[opts.SessionID] = proc
	m.mu.Unlock()

	wrapped := make(chan provider.Event, 64)
	go func() {
		defer close(wrapped)
		defer func() {
			m.mu.Lock()
			delete(m.processes, opts.SessionID)
			m.mu.Unlock()
		}()
		for evt := range rawEvents {
			wrapped <- evt
		}
	}()

	return wrapped, nil
}

// Interrupt sends SIGTERM to the running process for the given session ID.
// Returns an error if no process is registered for that session.
func (m *Manager) Interrupt(sessionID string) error {
	m.mu.Lock()
	proc, ok := m.processes[sessionID]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("manager: no running process for session %s", sessionID)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("manager: SIGTERM: %w", err)
	}
	return nil
}
