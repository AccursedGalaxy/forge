package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"github.com/accursedgalaxy/forge/internal/db"
)

// PlanStep represents a single step extracted from a planning session.
type PlanStep struct {
	Index       int    `json:"index"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Summarizer uses the Anthropic API to review plans and summarize execution output.
type Summarizer struct {
	client *anthropic.Client
}

// NewSummarizer creates a Summarizer with the provided Anthropic client.
func NewSummarizer(client *anthropic.Client) *Summarizer {
	return &Summarizer{client: client}
}

// BuildContextPrompt uses Sonnet 4.6 to build a refined prompt incorporating
// the task, project, and retrieved context chunks.
// On any error, it falls back to a simple plaintext prompt — never blocks execution.
func (s *Summarizer) BuildContextPrompt(ctx context.Context, task db.Task, project db.Project, contextChunks []string) string {
	if s.client == nil {
		return basicPlanPrompt(task, project)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Task: %s\n\nDescription: %s\n\nProject: %s\n%s\n\n",
		task.Title, task.Description, project.Name, project.Description))

	if len(contextChunks) > 0 {
		sb.WriteString("Relevant context from the codebase:\n")
		for i, chunk := range contextChunks {
			sb.WriteString(fmt.Sprintf("\n--- Context %d ---\n%s\n", i+1, chunk))
		}
	}

	userMsg := fmt.Sprintf(`You are preparing a PLANNING prompt for an AI coding agent.
The agent will run in READ-ONLY mode — it cannot create or modify files during planning.
Its job is to explore the codebase and produce a detailed, numbered implementation plan.

Given this task and context, write a prompt that instructs the agent to:
1. Explore the project structure using Glob, Grep, Read, and Bash (for inspection only)
2. Identify the exact files to create or modify and what changes are needed
3. Output a clear, numbered implementation plan as text — NOT code, NOT file writes

%s

Write only the agent prompt, nothing else.`, sb.String())

	msg, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
		},
	})
	if err != nil {
		slog.Warn("llm: BuildContextPrompt failed, using fallback", "err", err)
		return basicPlanPrompt(task, project)
	}

	for _, block := range msg.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text
		}
	}

	return basicPlanPrompt(task, project)
}

// ReviewPlan uses Sonnet 4.6 to parse the raw planning output into structured PlanSteps.
// If the LLM call fails or the response is malformed, a single fallback step is returned.
func (s *Summarizer) ReviewPlan(ctx context.Context, rawOutput string, task db.Task) []PlanStep {
	if s.client == nil || rawOutput == "" {
		return FallbackPlanStep(rawOutput)
	}

	userMsg := fmt.Sprintf(`You reviewed a coding task and produced planning output.
Extract the concrete steps from the plan below and return them as a JSON array.

Task: %s

Planning output:
%s

Return ONLY a JSON array of objects with this exact shape:
[{"index": 0, "description": "step description", "completed": false}, ...]

Start with index 0. Be concise. Do not include explanations outside the JSON.`, task.Title, rawOutput)

	msg, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
		},
	})
	if err != nil {
		slog.Warn("llm: ReviewPlan API call failed, using fallback", "err", err)
		return FallbackPlanStep(rawOutput)
	}

	var responseText string
	for _, block := range msg.Content {
		if block.Type == "text" && block.Text != "" {
			responseText = block.Text
			break
		}
	}

	return validateAndParsePlanSteps(responseText, rawOutput)
}

// Summarize uses Haiku to produce a brief summary of an execution session's output.
// On failure it returns an empty string — this call is non-fatal.
func (s *Summarizer) Summarize(ctx context.Context, fullOutput string, task db.Task) string {
	if s.client == nil || fullOutput == "" {
		return ""
	}

	// Trim to avoid excessive token usage; Haiku has a smaller context window.
	if len(fullOutput) > 8000 {
		fullOutput = fullOutput[:8000] + "\n... [truncated]"
	}

	userMsg := fmt.Sprintf(`Summarize what the AI coding agent accomplished for this task in 2-3 sentences.
Focus on what was changed or created, not on process.

Task: %s

Agent output:
%s`, task.Title, fullOutput)

	msg, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
		},
	})
	if err != nil {
		slog.Warn("llm: Summarize failed (non-fatal)", "err", err)
		return ""
	}

	for _, block := range msg.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text
		}
	}

	return ""
}

// validateAndParsePlanSteps attempts to unmarshal the LLM response as []PlanStep.
// Returns the fallback step if validation fails.
func validateAndParsePlanSteps(responseText, rawOutput string) []PlanStep {
	if responseText == "" {
		return FallbackPlanStep(rawOutput)
	}

	// Extract JSON array — model might include preamble text.
	start := strings.Index(responseText, "[")
	end := strings.LastIndex(responseText, "]")
	if start < 0 || end <= start {
		slog.Warn("llm: ReviewPlan response has no JSON array, using fallback")
		return FallbackPlanStep(rawOutput)
	}
	jsonStr := responseText[start : end+1]

	var steps []PlanStep
	if err := json.Unmarshal([]byte(jsonStr), &steps); err != nil {
		slog.Warn("llm: ReviewPlan JSON unmarshal failed, using fallback", "err", err)
		return FallbackPlanStep(rawOutput)
	}
	if len(steps) == 0 {
		return FallbackPlanStep(rawOutput)
	}

	// Validate: every step must have a non-empty description.
	for _, step := range steps {
		if strings.TrimSpace(step.Description) == "" {
			slog.Warn("llm: ReviewPlan step has empty description, using fallback")
			return FallbackPlanStep(rawOutput)
		}
	}

	return steps
}

// FallbackPlanStep wraps rawOutput as a single PlanStep.
// Exported so the orchestrator can use it directly if needed.
func FallbackPlanStep(rawOutput string) []PlanStep {
	desc := rawOutput
	if desc == "" {
		desc = "Execute the task"
	}
	return []PlanStep{{Index: 0, Description: desc, Completed: false}}
}

// basicPlanPrompt builds a simple prompt without LLM assistance.
func basicPlanPrompt(task db.Task, project db.Project) string {
	return fmt.Sprintf(`You are in PLANNING MODE. You cannot write or modify files.
Use Glob, Grep, Read, and Bash (read-only) to explore the project, then output a numbered implementation plan.
Do NOT attempt to create files — describe what needs to be done instead.

Project: %s
Task: %s
%s`, project.Name, task.Title, task.Description)
}
