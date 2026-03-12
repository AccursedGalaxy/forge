// Package claude implements the AgentProvider interface for Claude Code (claude-cli).
//
// v1: stub implementation — Run/Resume return a done event immediately.
// Step 3 replaces this with real os/exec spawning of claude-cli with
// --output-format stream-json, SIGTERM interrupt handling, and --resume support.
package claude

import (
	"context"
	"log/slog"

	"github.com/accursedgalaxy/forge/internal/provider"
)

// Provider is the Claude Code agent provider.
type Provider struct {
	// binPath is the path to the claude-cli binary.
	// Defaults to "claude" (resolved from PATH).
	binPath string
}

// New creates a Claude provider. Pass an empty string for binPath to use
// the claude binary from PATH.
func New(binPath string) *Provider {
	if binPath == "" {
		binPath = "claude"
	}
	return &Provider{binPath: binPath}
}

// Run starts a new Claude Code session (stub — Step 3 implements real spawning).
func (p *Provider) Run(ctx context.Context, task provider.Task, opts provider.RunOptions) (<-chan provider.Event, error) {
	// TODO(step3): spawn claude-cli with -p <prompt> --output-format stream-json
	//   --include-partial-messages --verbose
	//   plan sessions: --allowedTools Bash,Glob,Grep,Read,LS
	//   execute sessions: --dangerously-skip-permissions
	//   strip CLAUDECODE env var to avoid refusal inside parent Claude session
	slog.Info("claude: Run called (stub)", "task_id", task.ID, "session_type", opts.SessionType)

	ch := make(chan provider.Event, 1)
	go func() {
		defer close(ch)
		ch <- provider.Event{
			Type:    provider.EventTypeDone,
			Content: "stub: claude provider not yet implemented — coming in Step 3",
		}
	}()
	return ch, nil
}

// Resume continues an interrupted Claude session (stub — Step 3 implements real --resume).
func (p *Provider) Resume(ctx context.Context, sessionID string, prompt string) (<-chan provider.Event, error) {
	// TODO(step3): spawn claude-cli with --resume <sessionID>, prepend correction prompt
	slog.Info("claude: Resume called (stub)", "session_id", sessionID)

	ch := make(chan provider.Event, 1)
	go func() {
		defer close(ch)
		ch <- provider.Event{
			Type:    provider.EventTypeDone,
			Content: "stub: claude resume not yet implemented — coming in Step 3",
		}
	}()
	return ch, nil
}

// Interrupt sends a stop signal to the running claude-cli process (stub).
func (p *Provider) Interrupt(sessionID string) error {
	// TODO(step3): send SIGTERM to the tracked os/exec process for this sessionID
	slog.Info("claude: Interrupt called (stub)", "session_id", sessionID)
	return nil
}

// Capabilities describes what the Claude provider supports.
func (p *Provider) Capabilities() provider.ProviderCaps {
	return provider.ProviderCaps{
		Name:         "claude",
		CanPlan:      true,
		CanResume:    true,
		CanInterrupt: true,
	}
}
