// Package runner handles spawning and managing claude-cli subprocesses.
package runner

import (
	"encoding/json"

	"github.com/accursedgalaxy/forge/internal/provider"
)

// claudeStreamLine is a single JSONL line from claude-cli --output-format stream-json.
type claudeStreamLine struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`

	// system/init fields
	SessionID string   `json:"session_id"`
	CWD       string   `json:"cwd"`
	Tools     []string `json:"tools"`

	// assistant message
	Message *claudeMessage `json:"message"`

	// tool_use fields
	ToolID    string          `json:"id"`
	ToolName  string          `json:"name"`
	ToolInput json.RawMessage `json:"input"`

	// tool_result fields
	ToolUseID string `json:"tool_use_id"`

	// result fields
	Result  string  `json:"result"`
	Error   string  `json:"error"`
	CostUSD float64 `json:"cost_usd"`
}

type claudeMessage struct {
	ID         string               `json:"id"`
	Role       string               `json:"role"`
	Content    []claudeContentBlock `json:"content"`
	Model      string               `json:"model"`
	StopReason string               `json:"stop_reason"`
}

type claudeContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Thinking string `json:"thinking"`
	Name     string `json:"name"` // for tool_use blocks
	ID       string `json:"id"`
}

// ParseLine decodes one JSONL line into a provider.Event.
// Returns (event, true) when the line produces a meaningful event, (zero, false) otherwise.
func ParseLine(line []byte) (provider.Event, bool) {
	if len(line) == 0 {
		return provider.Event{}, false
	}

	var raw claudeStreamLine
	if err := json.Unmarshal(line, &raw); err != nil {
		return provider.Event{}, false
	}

	switch raw.Type {
	case "system":
		if raw.Subtype == "init" && raw.SessionID != "" {
			return provider.Event{
				Type: provider.EventTypeSystem,
				Meta: map[string]string{"session_id": raw.SessionID},
			}, true
		}

	case "assistant":
		if raw.Message == nil {
			return provider.Event{}, false
		}
		var text, thinking string
		for _, block := range raw.Message.Content {
			switch block.Type {
			case "text":
				text += block.Text
			case "thinking":
				thinking += block.Thinking
			case "tool_use":
				// emit a tool event for each tool_use block in the message
				inputStr := "{}"
				if block.ID != "" {
					// name is in block.Name for inline tool_use blocks
				}
				return provider.Event{
					Type:    provider.EventTypeTool,
					Content: block.Name,
					Meta:    map[string]string{"id": block.ID, "input": inputStr},
				}, true
			}
		}
		if thinking != "" {
			return provider.Event{
				Type:    provider.EventTypeThinking,
				Content: thinking,
			}, true
		}
		if text != "" {
			return provider.Event{
				Type:    provider.EventTypeText,
				Content: text,
			}, true
		}

	case "tool_use":
		inputStr := string(raw.ToolInput)
		if inputStr == "" {
			inputStr = "{}"
		}
		return provider.Event{
			Type:    provider.EventTypeTool,
			Content: raw.ToolName,
			Meta:    map[string]string{"id": raw.ToolID, "input": inputStr},
		}, true

	case "result":
		if raw.Subtype == "error" {
			return provider.Event{
				Type:    provider.EventTypeError,
				Content: raw.Error,
				Meta:    map[string]string{"session_id": raw.SessionID},
			}, true
		}
		// success or other subtypes
		return provider.Event{
			Type:    provider.EventTypeDone,
			Content: raw.Result,
			Meta:    map[string]string{"session_id": raw.SessionID},
		}, true
	}

	return provider.Event{}, false
}
