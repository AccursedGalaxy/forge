// Package claude implements the AgentProvider interface for Claude Code (claude-cli).
package claude

import (
	"context"
	"log/slog"

	"github.com/accursedgalaxy/forge/internal/provider"
	"github.com/accursedgalaxy/forge/internal/runner"
)

// Provider is the Claude Code agent provider.
type Provider struct {
	manager *runner.Manager
}

// New creates a Claude provider. Pass an empty string for binPath to use
// the claude binary from PATH.
func New(binPath string) *Provider {
	return &Provider{
		manager: runner.NewManager(binPath),
	}
}

// Run starts a new Claude Code session.
func (p *Provider) Run(ctx context.Context, task provider.Task, opts provider.RunOptions) (<-chan provider.Event, error) {
	slog.Info("claude: Run", "task_id", task.ID, "session_type", opts.SessionType)

	var allowedTools []string
	if opts.SessionType == "plan" {
		allowedTools = []string{"Glob", "Grep", "Read", "Bash", "LS"}
	}

	prompt := task.Description
	if len(opts.ContextChunks) > 0 {
		prompt += "\n\n(context available)"
	}

	return p.manager.Start(ctx, runner.SpawnOptions{
		SessionID:    task.ID,
		Prompt:       prompt,
		AllowedTools: allowedTools,
	})
}

// Resume continues an interrupted Claude session.
func (p *Provider) Resume(ctx context.Context, sessionID string, prompt string) (<-chan provider.Event, error) {
	slog.Info("claude: Resume", "session_id", sessionID)

	return p.manager.Start(ctx, runner.SpawnOptions{
		SessionID: sessionID,
		Prompt:    prompt,
		// ClaudeResumeID is the claude native session ID; we use sessionID as a
		// best-effort fallback here since the provider interface doesn't carry it.
		// The orchestrator uses ResumeSession directly and passes the correct ID.
		ClaudeResumeID: sessionID,
		AllowedTools:   nil,
	})
}

// Interrupt sends SIGTERM to the running claude-cli process for this session.
func (p *Provider) Interrupt(sessionID string) error {
	return p.manager.Interrupt(sessionID)
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
