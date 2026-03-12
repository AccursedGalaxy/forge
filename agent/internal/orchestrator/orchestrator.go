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
	queries     *db.Queries
	broadcaster *stream.Broadcaster
	manager     *runner.Manager
	summarizer  *llm.Summarizer
	retriever   *appcontext.Retriever
	indexer     *appcontext.Indexer
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
		queries:     queries,
		broadcaster: broadcaster,
		manager:     manager,
		summarizer:  summarizer,
		retriever:   retriever,
		indexer:     indexer,
	}
}

// PlanSession runs the planning phase for sessionID.
// It spawns claude-cli in read-only mode, collects output, reviews it with the LLM,
// stores validated plan steps, and transitions the session to awaiting_approval.
func (o *Orchestrator) PlanSession(ctx context.Context, sessionID uuid.UUID) error {
	log := slog.With("session_id", sessionID, "phase", "plan")
	log.Info("orchestrator: starting plan phase")

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

	log = log.With("task_id", task.ID, "task_title", task.Title, "project", project.Name)
	log.Info("orchestrator: loaded task and project")

	// 1. Transition to planning.
	o.updateStatus(ctx, sessionID, "planning")
	o.publish(ctx, sessionID, "claude:start", map[string]string{"phase": "plan"})

	// 2. Retrieve relevant context from previous sessions (non-fatal).
	log.Info("orchestrator: retrieving context chunks")
	contextChunks := o.retrieveContext(ctx, session.ProjectID, task.Description)
	log.Info("orchestrator: context retrieved", "chunks", len(contextChunks))

	// 3. Build refined prompt via LLM (falls back to basic prompt on failure).
	log.Info("orchestrator: building context prompt via LLM")
	prompt := o.summarizer.BuildContextPrompt(ctx, task, project, contextChunks)
	log.Info("orchestrator: prompt ready", "prompt_bytes", len(prompt))

	// 4. Spawn claude-cli in read-only planning mode.
	log.Info("orchestrator: spawning claude-cli", "allowed_tools", PlanToolset, "work_dir", project.LocalPath)
	events, err := o.manager.Start(ctx, runner.SpawnOptions{
		SessionID:    sessionID.String(),
		Prompt:       prompt,
		WorkDir:      project.LocalPath,
		AllowedTools: PlanToolset,
	})
	if err != nil {
		o.setError(ctx, sessionID, fmt.Sprintf("failed to start agent: %v", err))
		return fmt.Errorf("orchestrator: start runner: %w", err)
	}

	// 5. Stream events to SSE; accumulate text output; capture claude session ID.
	log.Info("orchestrator: consuming claude-cli events")
	rawPlanOutput, claudeSessionID := o.consumeEvents(ctx, sessionID, events, log)
	log.Info("orchestrator: claude-cli finished", "output_bytes", len(rawPlanOutput), "claude_session_id", claudeSessionID)

	// Store claude session ID if we got one.
	if claudeSessionID != "" {
		if _, err := o.queries.UpdateSessionClaudeID(ctx, db.UpdateSessionClaudeIDParams{
			ID:              sessionID,
			ClaudeSessionID: sql.NullString{String: claudeSessionID, Valid: true},
		}); err != nil {
			log.Warn("orchestrator: failed to save claude session ID", "err", err)
		}
	}

	// 6. Review plan output with LLM → structured plan steps.
	log.Info("orchestrator: reviewing plan output with LLM")
	steps := o.summarizer.ReviewPlan(ctx, rawPlanOutput, task)
	log.Info("orchestrator: plan steps extracted", "steps", len(steps))
	for i, s := range steps {
		log.Info("orchestrator: plan step", "index", s.Index, "description", s.Description)
		_ = i
	}

	planStepsJSON, err := json.Marshal(steps)
	if err != nil {
		planStepsJSON, _ = json.Marshal(llm.FallbackPlanStep(rawPlanOutput))
		log.Warn("orchestrator: failed to marshal plan steps, using fallback", "err", err)
	}

	if _, err := o.queries.UpdateSessionPlanSteps(ctx, db.UpdateSessionPlanStepsParams{
		ID:        sessionID,
		PlanSteps: pqtype.NullRawMessage{RawMessage: planStepsJSON, Valid: true},
	}); err != nil {
		log.Warn("orchestrator: failed to save plan steps", "err", err)
	}

	// 7. Transition to awaiting_approval.
	o.updateStatus(ctx, sessionID, "awaiting_approval")
	o.publish(ctx, sessionID, "claude:done", map[string]any{
		"phase":      "plan",
		"plan_steps": steps,
	})

	log.Info("orchestrator: plan phase complete, awaiting user approval", "steps", len(steps))
	return nil
}

// ExecuteSession runs the execution phase for an approved session.
func (o *Orchestrator) ExecuteSession(ctx context.Context, sessionID uuid.UUID) error {
	log := slog.With("session_id", sessionID, "phase", "execute")
	log.Info("orchestrator: starting execute phase")

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

	log = log.With("task_id", task.ID, "task_title", task.Title, "project", project.Name)

	// Fix #3: mark this session as session_type="execute" now that we're executing.
	if _, err := o.queries.UpdateSessionType(ctx, sessionID, "execute"); err != nil {
		log.Warn("orchestrator: failed to update session type to execute", "err", err)
	}

	// 1. Transition to running.
	o.updateStatus(ctx, sessionID, "running")
	o.publish(ctx, sessionID, "claude:start", map[string]string{"phase": "execute"})

	// 2. Retrieve context and parse plan steps (non-fatal on both).
	log.Info("orchestrator: retrieving context chunks")
	contextChunks := o.retrieveContext(ctx, session.ProjectID, task.Description)
	log.Info("orchestrator: context retrieved", "chunks", len(contextChunks))

	steps := parsePlanSteps(session.PlanSteps)
	log.Info("orchestrator: executing against plan", "steps", len(steps))
	for _, s := range steps {
		log.Info("orchestrator: plan step to execute", "index", s.Index, "description", s.Description)
	}

	// 3. Build execute prompt (pure Go, no LLM call).
	prompt := buildExecutePrompt(task, project, steps, contextChunks)
	log.Info("orchestrator: execute prompt built", "prompt_bytes", len(prompt))

	// 4. Spawn claude-cli with full permissions.
	log.Info("orchestrator: spawning claude-cli with full permissions (dangerously-skip-permissions)", "work_dir", project.LocalPath)
	events, err := o.manager.Start(ctx, runner.SpawnOptions{
		SessionID:    sessionID.String(),
		Prompt:       prompt,
		WorkDir:      project.LocalPath,
		AllowedTools: nil, // nil → --dangerously-skip-permissions
	})
	if err != nil {
		o.setError(ctx, sessionID, fmt.Sprintf("failed to start agent: %v", err))
		return fmt.Errorf("orchestrator: start runner: %w", err)
	}

	// 5. Stream events; accumulate full output.
	log.Info("orchestrator: consuming claude-cli events")
	fullOutput, claudeSessionID := o.consumeEvents(ctx, sessionID, events, log)
	log.Info("orchestrator: claude-cli finished", "output_bytes", len(fullOutput), "claude_session_id", claudeSessionID)

	if claudeSessionID != "" {
		if _, err := o.queries.UpdateSessionClaudeID(ctx, db.UpdateSessionClaudeIDParams{
			ID:              sessionID,
			ClaudeSessionID: sql.NullString{String: claudeSessionID, Valid: true},
		}); err != nil {
			log.Warn("orchestrator: failed to save claude session ID", "err", err)
		}
	}

	// 6. Summarize output (Haiku, non-fatal).
	log.Info("orchestrator: summarizing execution output via LLM")
	notes := o.summarizer.Summarize(ctx, fullOutput, task)
	log.Info("orchestrator: summary ready", "notes_bytes", len(notes))

	// Fix #2: mark all plan steps as completed.
	completedSteps := markStepsCompleted(steps)
	if completedJSON, err := json.Marshal(completedSteps); err == nil {
		if _, err := o.queries.UpdateSessionPlanSteps(ctx, db.UpdateSessionPlanStepsParams{
			ID:        sessionID,
			PlanSteps: pqtype.NullRawMessage{RawMessage: completedJSON, Valid: true},
		}); err != nil {
			log.Warn("orchestrator: failed to save completed plan steps", "err", err)
		}
	}

	// Fix #1: use CompleteSession to set completed_at alongside status=done.
	if _, err := o.queries.CompleteSession(ctx, sessionID); err != nil {
		log.Warn("orchestrator: failed to complete session", "err", err)
		// Fall back to plain status update so the session isn't left in a bad state.
		o.updateStatus(ctx, sessionID, "done")
	} else {
		o.publish(ctx, sessionID, "session:status", map[string]string{"status": "done"})
	}

	// 8. Auto-advance task to review.
	log.Info("orchestrator: advancing task status to review")
	if _, err := o.queries.UpdateTask(ctx, db.UpdateTaskParams{
		Status: sql.NullString{String: "review", Valid: true},
		ID:     task.ID,
	}); err != nil {
		log.Warn("orchestrator: failed to advance task to review", "err", err)
	}

	o.publish(ctx, sessionID, "claude:done", map[string]any{
		"phase": "execute",
		"notes": notes,
	})

	// 9. Index output asynchronously — never blocks SSE.
	log.Info("orchestrator: indexing session output asynchronously")
	go o.indexer.Index(context.Background(), session.ProjectID, sessionID.String(), fullOutput)

	log.Info("orchestrator: execute phase complete, task moved to review")
	return nil
}

// ResumeSession resumes a paused session with a correction prompt.
func (o *Orchestrator) ResumeSession(ctx context.Context, sessionID uuid.UUID, correctionPrompt string) error {
	log := slog.With("session_id", sessionID, "phase", "resume")
	log.Info("orchestrator: starting resume phase", "has_correction", correctionPrompt != "")

	session, err := o.queries.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("orchestrator: get session: %w", err)
	}

	task, err := o.queries.GetTask(ctx, session.TaskID)
	if err != nil {
		return fmt.Errorf("orchestrator: get task: %w", err)
	}

	log = log.With("task_id", task.ID, "task_title", task.Title)

	steps := parsePlanSteps(session.PlanSteps)
	prompt := buildResumePrompt(correctionPrompt, task, steps)
	log.Info("orchestrator: resume prompt built", "prompt_bytes", len(prompt))

	claudeSessionID := ""
	if session.ClaudeSessionID.Valid {
		claudeSessionID = session.ClaudeSessionID.String
		log.Info("orchestrator: resuming existing claude session", "claude_session_id", claudeSessionID)
	} else {
		log.Info("orchestrator: no previous claude session ID, starting fresh")
	}

	// 1. Transition to running.
	o.updateStatus(ctx, sessionID, "running")
	o.publish(ctx, sessionID, "claude:start", map[string]string{"phase": "resume"})

	// 2. Spawn with --resume if we have a claude session ID.
	project, err := o.queries.GetProject(ctx, session.ProjectID)
	if err != nil {
		return fmt.Errorf("orchestrator: get project: %w", err)
	}
	log.Info("orchestrator: spawning claude-cli for resume", "work_dir", project.LocalPath)
	events, err := o.manager.Start(ctx, runner.SpawnOptions{
		SessionID:      sessionID.String(),
		Prompt:         prompt,
		WorkDir:        project.LocalPath,
		ClaudeResumeID: claudeSessionID,
		AllowedTools:   nil, // full permissions for execution
	})
	if err != nil {
		o.setError(ctx, sessionID, fmt.Sprintf("failed to resume agent: %v", err))
		return fmt.Errorf("orchestrator: start runner: %w", err)
	}

	fullOutput, newClaudeID := o.consumeEvents(ctx, sessionID, events, log)
	log.Info("orchestrator: claude-cli finished", "output_bytes", len(fullOutput), "new_claude_session_id", newClaudeID)

	if newClaudeID != "" && newClaudeID != claudeSessionID {
		if _, err := o.queries.UpdateSessionClaudeID(ctx, db.UpdateSessionClaudeIDParams{
			ID:              sessionID,
			ClaudeSessionID: sql.NullString{String: newClaudeID, Valid: true},
		}); err != nil {
			log.Warn("orchestrator: failed to save new claude session ID", "err", err)
		}
	}

	log.Info("orchestrator: summarizing resume output")
	notes := o.summarizer.Summarize(ctx, fullOutput, task)

	// Mark steps completed and complete the session.
	completedSteps := markStepsCompleted(steps)
	if completedJSON, err := json.Marshal(completedSteps); err == nil {
		if _, err := o.queries.UpdateSessionPlanSteps(ctx, db.UpdateSessionPlanStepsParams{
			ID:        sessionID,
			PlanSteps: pqtype.NullRawMessage{RawMessage: completedJSON, Valid: true},
		}); err != nil {
			log.Warn("orchestrator: failed to save completed plan steps", "err", err)
		}
	}

	if _, err := o.queries.CompleteSession(ctx, sessionID); err != nil {
		log.Warn("orchestrator: failed to complete session", "err", err)
		o.updateStatus(ctx, sessionID, "done")
	} else {
		o.publish(ctx, sessionID, "session:status", map[string]string{"status": "done"})
	}

	if _, err := o.queries.UpdateTask(ctx, db.UpdateTaskParams{
		Status: sql.NullString{String: "review", Valid: true},
		ID:     task.ID,
	}); err != nil {
		log.Warn("orchestrator: failed to advance task to review", "err", err)
	}

	o.publish(ctx, sessionID, "claude:done", map[string]any{
		"phase": "resume",
		"notes": notes,
	})

	go o.indexer.Index(context.Background(), session.ProjectID, sessionID.String(), fullOutput)

	log.Info("orchestrator: resume phase complete, task moved to review")
	return nil
}

// consumeEvents drains the event channel, publishing each event as SSE and accumulating
// text output. Returns (rawTextOutput, claudeSessionID).
func (o *Orchestrator) consumeEvents(ctx context.Context, sessionID uuid.UUID, events <-chan provider.Event, log *slog.Logger) (string, string) {
	var textBuf []byte
	var claudeSessionID string

	for evt := range events {
		switch evt.Type {
		case provider.EventTypeSystem:
			if id, ok := evt.Meta["session_id"]; ok && id != "" {
				claudeSessionID = id
				log.Info("orchestrator: captured claude session ID", "claude_session_id", id)
			}
			// Don't forward system events to the client.
			continue

		case provider.EventTypeText:
			textBuf = append(textBuf, []byte(evt.Content)...)
			log.Debug("orchestrator: text chunk", "bytes", len(evt.Content))

		case provider.EventTypeDone:
			if id, ok := evt.Meta["session_id"]; ok && id != "" && claudeSessionID == "" {
				claudeSessionID = id
				log.Info("orchestrator: captured claude session ID from done event", "claude_session_id", id)
			}
			log.Info("orchestrator: claude-cli signalled done")

		default:
			log.Debug("orchestrator: event", "type", evt.Type, "content_bytes", len(evt.Content))
		}

		o.publish(ctx, sessionID, "claude:stream", map[string]string{
			"type":    string(evt.Type),
			"content": evt.Content,
		})
	}

	return string(textBuf), claudeSessionID
}

// markStepsCompleted returns a copy of the steps slice with all completed flags set to true.
func markStepsCompleted(steps []llm.PlanStep) []llm.PlanStep {
	out := make([]llm.PlanStep, len(steps))
	for i, s := range steps {
		s.Completed = true
		out[i] = s
	}
	return out
}

// retrieveContext attempts pgvector retrieval; returns nil on failure (non-fatal).
func (o *Orchestrator) retrieveContext(ctx context.Context, projectID uuid.UUID, query string) []string {
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
	slog.Error("orchestrator: session error", "session_id", sessionID, "error", msg)
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
