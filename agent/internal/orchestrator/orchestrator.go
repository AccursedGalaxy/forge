package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/sqlc-dev/pqtype"

	appcontext "github.com/accursedgalaxy/forge/internal/context"
	"github.com/accursedgalaxy/forge/internal/db"
	"github.com/accursedgalaxy/forge/internal/llm"
	"github.com/accursedgalaxy/forge/internal/provider"
	"github.com/accursedgalaxy/forge/internal/runner"
	"github.com/accursedgalaxy/forge/internal/stream"
)

// Orchestrator coordinates the plan → approve → execute session lifecycle.
type Orchestrator struct {
	queries   *db.Queries
	broadcaster *stream.Broadcaster
	manager   *runner.Manager
	summarizer *llm.Summarizer
	retriever  *appcontext.Retriever
	indexer    *appcontext.Indexer
}

// New creates an Orchestrator wired with all required dependencies.
func New(
	queries *db.Queries,
	broadcaster *stream.Broadcaster,
	manager *runner.Manager,
	summarizer *llm.Summarizer,
	retriever *appcontext.Retriever,
	indexer *appcontext.Indexer,
) *Orchestrator {
	return &Orchestrator{
		queries:    queries,
		broadcaster: broadcaster,
		manager:    manager,
		summarizer: summarizer,
		retriever:  retriever,
		indexer:    indexer,
	}
}

// PlanSession runs the planning phase for sessionID.
// It spawns claude-cli in read-only mode, collects output, reviews it with the LLM,
// stores validated plan steps, and transitions the session to awaiting_approval.
func (o *Orchestrator) PlanSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := o.queries.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("orchestrator: get session: %w", err)
	}

	task, err := o.queries.GetTask(ctx, session.TaskID)
	if err != nil {
		return fmt.Errorf("orchestrator: get task: %w", err)
	}

	project, err := o.queries.GetProject(ctx, session.ProjectID)
	if err != nil {
		return fmt.Errorf("orchestrator: get project: %w", err)
	}

	// 1. Transition to planning.
	o.updateStatus(ctx, sessionID, "planning")
	o.publish(ctx, sessionID, "claude:start", map[string]string{"phase": "plan"})

	// 2. Retrieve context (non-fatal).
	contextChunks := o.retrieveContext(ctx, session.ProjectID, task.Description)

	// 3. Build refined prompt via LLM (falls back to basic prompt on failure).
	prompt := o.summarizer.BuildContextPrompt(ctx, task, project, contextChunks)

	// 4. Spawn claude-cli in read-only planning mode.
	events, err := o.manager.Start(ctx, runner.SpawnOptions{
		SessionID:    sessionID.String(),
		Prompt:       prompt,
		AllowedTools: PlanToolset,
	})
	if err != nil {
		o.setError(ctx, sessionID, fmt.Sprintf("failed to start agent: %v", err))
		return fmt.Errorf("orchestrator: start runner: %w", err)
	}

	// 5. Stream events to SSE; accumulate text output; capture claude session ID.
	rawPlanOutput, claudeSessionID := o.consumeEvents(ctx, sessionID, events)

	// Store claude session ID if we got one.
	if claudeSessionID != "" {
		if _, err := o.queries.UpdateSessionClaudeID(ctx, db.UpdateSessionClaudeIDParams{
			ID:              sessionID,
			ClaudeSessionID: sql.NullString{String: claudeSessionID, Valid: true},
		}); err != nil {
			slog.Warn("orchestrator: update claude session ID failed", "err", err)
		}
	}

	// 6. Review plan with LLM → structured plan steps.
	steps := o.summarizer.ReviewPlan(ctx, rawPlanOutput, task)
	planStepsJSON, err := json.Marshal(steps)
	if err != nil {
		planStepsJSON, _ = json.Marshal(llm.FallbackPlanStep(rawPlanOutput))
	}

	if _, err := o.queries.UpdateSessionPlanSteps(ctx, db.UpdateSessionPlanStepsParams{
		ID:        sessionID,
		PlanSteps: pqtype.NullRawMessage{RawMessage: planStepsJSON, Valid: true},
	}); err != nil {
		slog.Warn("orchestrator: update plan steps failed", "err", err)
	}

	// 7. Transition to awaiting_approval.
	o.updateStatus(ctx, sessionID, "awaiting_approval")
	o.publish(ctx, sessionID, "claude:done", map[string]any{
		"phase":      "plan",
		"plan_steps": steps,
	})

	slog.Info("orchestrator: planning complete", "session_id", sessionID, "steps", len(steps))
	return nil
}

// ExecuteSession runs the execution phase for an approved session.
func (o *Orchestrator) ExecuteSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := o.queries.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("orchestrator: get session: %w", err)
	}

	task, err := o.queries.GetTask(ctx, session.TaskID)
	if err != nil {
		return fmt.Errorf("orchestrator: get task: %w", err)
	}

	project, err := o.queries.GetProject(ctx, session.ProjectID)
	if err != nil {
		return fmt.Errorf("orchestrator: get project: %w", err)
	}

	// 1. Transition to running.
	o.updateStatus(ctx, sessionID, "running")
	o.publish(ctx, sessionID, "claude:start", map[string]string{"phase": "execute"})

	// 2. Retrieve context and parse plan steps (non-fatal on both).
	contextChunks := o.retrieveContext(ctx, session.ProjectID, task.Description)
	steps := parsePlanSteps(session.PlanSteps)

	// 3. Build execute prompt (pure Go, no LLM).
	prompt := buildExecutePrompt(task, project, steps, contextChunks)

	// 4. Spawn claude-cli with full permissions.
	events, err := o.manager.Start(ctx, runner.SpawnOptions{
		SessionID:    sessionID.String(),
		Prompt:       prompt,
		AllowedTools: nil, // nil → --dangerously-skip-permissions
	})
	if err != nil {
		o.setError(ctx, sessionID, fmt.Sprintf("failed to start agent: %v", err))
		return fmt.Errorf("orchestrator: start runner: %w", err)
	}

	// 5. Stream events; accumulate full output.
	fullOutput, claudeSessionID := o.consumeEvents(ctx, sessionID, events)

	if claudeSessionID != "" {
		if _, err := o.queries.UpdateSessionClaudeID(ctx, db.UpdateSessionClaudeIDParams{
			ID:              sessionID,
			ClaudeSessionID: sql.NullString{String: claudeSessionID, Valid: true},
		}); err != nil {
			slog.Warn("orchestrator: update claude session ID failed", "err", err)
		}
	}

	// 6. Summarize output (Haiku, non-fatal).
	notes := o.summarizer.Summarize(ctx, fullOutput, task)

	// 7. Mark session done.
	o.updateStatus(ctx, sessionID, "done")

	// 8. Mark task as review.
	if _, err := o.queries.UpdateTask(ctx, db.UpdateTaskParams{
		Status: sql.NullString{String: "review", Valid: true},
		ID:     task.ID,
	}); err != nil {
		slog.Warn("orchestrator: update task status failed", "err", err)
	}

	o.publish(ctx, sessionID, "claude:done", map[string]any{
		"phase": "execute",
		"notes": notes,
	})

	// 9. Index output asynchronously — never blocks SSE.
	go o.indexer.Index(context.Background(), session.ProjectID, sessionID.String(), fullOutput)

	slog.Info("orchestrator: execution complete", "session_id", sessionID)
	return nil
}

// ResumeSession resumes a paused session with a correction prompt.
func (o *Orchestrator) ResumeSession(ctx context.Context, sessionID uuid.UUID, correctionPrompt string) error {
	session, err := o.queries.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("orchestrator: get session: %w", err)
	}

	task, err := o.queries.GetTask(ctx, session.TaskID)
	if err != nil {
		return fmt.Errorf("orchestrator: get task: %w", err)
	}

	steps := parsePlanSteps(session.PlanSteps)
	prompt := buildResumePrompt(correctionPrompt, task, steps)

	claudeSessionID := ""
	if session.ClaudeSessionID.Valid {
		claudeSessionID = session.ClaudeSessionID.String
	}

	// 1. Transition to running.
	o.updateStatus(ctx, sessionID, "running")
	o.publish(ctx, sessionID, "claude:start", map[string]string{"phase": "resume"})

	// 2. Spawn with --resume if we have a claude session ID.
	events, err := o.manager.Start(ctx, runner.SpawnOptions{
		SessionID:      sessionID.String(),
		Prompt:         prompt,
		ClaudeResumeID: claudeSessionID,
		AllowedTools:   nil, // full permissions for execution
	})
	if err != nil {
		o.setError(ctx, sessionID, fmt.Sprintf("failed to resume agent: %v", err))
		return fmt.Errorf("orchestrator: start runner: %w", err)
	}

	fullOutput, newClaudeID := o.consumeEvents(ctx, sessionID, events)

	if newClaudeID != "" && newClaudeID != claudeSessionID {
		if _, err := o.queries.UpdateSessionClaudeID(ctx, db.UpdateSessionClaudeIDParams{
			ID:              sessionID,
			ClaudeSessionID: sql.NullString{String: newClaudeID, Valid: true},
		}); err != nil {
			slog.Warn("orchestrator: update claude session ID failed", "err", err)
		}
	}

	notes := o.summarizer.Summarize(ctx, fullOutput, task)

	o.updateStatus(ctx, sessionID, "done")
	if _, err := o.queries.UpdateTask(ctx, db.UpdateTaskParams{
		Status: sql.NullString{String: "review", Valid: true},
		ID:     task.ID,
	}); err != nil {
		slog.Warn("orchestrator: update task status failed", "err", err)
	}

	o.publish(ctx, sessionID, "claude:done", map[string]any{
		"phase": "resume",
		"notes": notes,
	})

	go o.indexer.Index(context.Background(), session.ProjectID, sessionID.String(), fullOutput)

	slog.Info("orchestrator: resume complete", "session_id", sessionID)
	return nil
}

// consumeEvents drains the event channel, publishing each event as SSE and accumulating
// text output. Returns (rawTextOutput, claudeSessionID).
func (o *Orchestrator) consumeEvents(ctx context.Context, sessionID uuid.UUID, events <-chan provider.Event) (string, string) {
	var textBuf []byte
	var claudeSessionID string

	for evt := range events {
		switch evt.Type {
		case provider.EventTypeSystem:
			if id, ok := evt.Meta["session_id"]; ok && id != "" {
				claudeSessionID = id
			}
			// Don't forward system events to the client.
			continue

		case provider.EventTypeText:
			textBuf = append(textBuf, []byte(evt.Content)...)

		case provider.EventTypeDone:
			// The result event may also carry a session_id.
			if id, ok := evt.Meta["session_id"]; ok && id != "" && claudeSessionID == "" {
				claudeSessionID = id
			}
		}

		o.publish(ctx, sessionID, "claude:stream", map[string]string{
			"type":    string(evt.Type),
			"content": evt.Content,
		})
	}

	return string(textBuf), claudeSessionID
}

// retrieveContext attempts pgvector retrieval; returns nil on failure (non-fatal).
func (o *Orchestrator) retrieveContext(ctx context.Context, projectID uuid.UUID, query string) []string {
	// Embedder returns empty vector when no API is configured.
	// Retriever returns nil immediately for empty vectors.
	return o.retriever.TopK(ctx, projectID, pgvector.Vector{}, 5)
}

// updateStatus sets session status and publishes an SSE event.
func (o *Orchestrator) updateStatus(ctx context.Context, sessionID uuid.UUID, status string) {
	if _, err := o.queries.UpdateSessionStatus(ctx, db.UpdateSessionStatusParams{
		ID:     sessionID,
		Status: status,
	}); err != nil {
		slog.Error("orchestrator: update status failed", "session_id", sessionID, "status", status, "err", err)
	}
	o.publish(ctx, sessionID, "session:status", map[string]string{"status": status})
}

// setError marks the session as errored and publishes an SSE event.
func (o *Orchestrator) setError(ctx context.Context, sessionID uuid.UUID, msg string) {
	if _, err := o.queries.UpdateSessionError(ctx, db.UpdateSessionErrorParams{
		ID:    sessionID,
		Error: sql.NullString{String: msg, Valid: true},
	}); err != nil {
		slog.Error("orchestrator: set error failed", "session_id", sessionID, "err", err)
	}
	o.publish(ctx, sessionID, "session:status", map[string]string{"status": "error", "error": msg})
}

// publish is a fire-and-forget SSE publish; errors are logged but not returned.
func (o *Orchestrator) publish(ctx context.Context, sessionID uuid.UUID, eventType string, data any) {
	if err := o.broadcaster.Publish(ctx, sessionID.String(), eventType, data); err != nil {
		slog.Warn("orchestrator: SSE publish failed", "event", eventType, "err", err)
	}
}

// parsePlanSteps decodes the session's plan_steps JSONB field.
// Returns an empty slice on any error.
func parsePlanSteps(raw pqtype.NullRawMessage) []llm.PlanStep {
	if !raw.Valid || len(raw.RawMessage) == 0 {
		return nil
	}
	var steps []llm.PlanStep
	if err := json.Unmarshal(raw.RawMessage, &steps); err != nil {
		slog.Warn("orchestrator: parse plan steps failed", "err", err)
		return nil
	}
	return steps
}
