// Package orchestrator coordinates the full session lifecycle.
package orchestrator

import (
	"fmt"
	"strings"

	"github.com/accursedgalaxy/forge/internal/db"
	"github.com/accursedgalaxy/forge/internal/llm"
)

// PlanToolset is the set of allowed tools for planning sessions (read-only).
var PlanToolset = []string{"Glob", "Grep", "Read", "Bash", "LS"}

// buildExecutePrompt constructs the execution prompt from plan steps and context.
// This is pure Go — no LLM call.
func buildExecutePrompt(task db.Task, project db.Project, steps []llm.PlanStep, contextChunks []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Project: %s\n", project.Name))
	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n", project.Description))
	}
	sb.WriteString(fmt.Sprintf("\nTask: %s\n%s\n", task.Title, task.Description))

	if len(steps) > 0 {
		sb.WriteString("\nApproved plan to execute:\n")
		for _, step := range steps {
			check := "[ ]"
			if step.Completed {
				check = "[x]"
			}
			sb.WriteString(fmt.Sprintf("  %s %d. %s\n", check, step.Index+1, step.Description))
		}
	}

	if len(contextChunks) > 0 {
		sb.WriteString("\nRelevant context from the codebase:\n")
		for i, chunk := range contextChunks {
			sb.WriteString(fmt.Sprintf("\n--- Context %d ---\n%s\n", i+1, chunk))
		}
	}

	sb.WriteString("\nExecute the plan completely. Mark each step done as you go.")
	return sb.String()
}

// buildResumePrompt constructs a resume prompt with the correction prepended.
// Per the plan: correction comes FIRST, then context.
func buildResumePrompt(correctionPrompt string, task db.Task, steps []llm.PlanStep) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Correction: %s\n\n", correctionPrompt))
	sb.WriteString(fmt.Sprintf("Context:\nTask: %s\n%s\n", task.Title, task.Description))

	if len(steps) > 0 {
		sb.WriteString("\nPlan steps:\n")
		for _, step := range steps {
			check := "[ ]"
			if step.Completed {
				check = "[x]"
			}
			sb.WriteString(fmt.Sprintf("  %s %d. %s\n", check, step.Index+1, step.Description))
		}
	}

	return sb.String()
}
