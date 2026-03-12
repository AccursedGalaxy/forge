// Package provider defines the AgentProvider interface and shared types used
// by all agent backend implementations (Claude, Codex, Ollama, etc.).
package provider

import "context"

// Task is the unit of work passed to a provider for execution.
type Task struct {
	ID            string
	Title         string
	Description   string
	ProjectID     string
	AutonomyLevel string // "supervised" | "checkpoint" | "autonomous"
}

// RunOptions configures a single provider invocation.
type RunOptions struct {
	SessionType   string   // "plan" | "execute"
	PlanSteps     []string // populated for execute sessions
	ContextChunks []string // pre-session context from pgvector retrieval
}

// EventType classifies a streaming event from a provider.
type EventType string

const (
	EventTypeText     EventType = "text"
	EventTypeThinking EventType = "thinking"
	EventTypeTool     EventType = "tool"
	EventTypeError    EventType = "error"
	EventTypeDone     EventType = "done"
)

// Event is a single streamed event from a provider run.
type Event struct {
	Type    EventType
	Content string
	Meta    map[string]string // optional key/value metadata (e.g. tool name, session_id)
}

// ProviderCaps describes what an agent provider supports.
type ProviderCaps struct {
	Name         string
	CanPlan      bool
	CanResume    bool
	CanInterrupt bool
}

// AgentProvider is the interface all agent backends must implement.
// Claude Code is the default provider; others register as named alternatives.
type AgentProvider interface {
	// Run starts a new agent session and returns a channel of streaming events.
	Run(ctx context.Context, task Task, opts RunOptions) (<-chan Event, error)

	// Resume continues a previously interrupted session.
	// sessionID is the provider-native session identifier (e.g. claude --resume ID).
	Resume(ctx context.Context, sessionID string, prompt string) (<-chan Event, error)

	// Interrupt signals the running session to stop cleanly (e.g. SIGTERM).
	Interrupt(sessionID string) error

	// Capabilities returns metadata describing what this provider supports.
	Capabilities() ProviderCaps
}
