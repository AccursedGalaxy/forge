// Package llm provides Anthropic API integration for plan review, summarization,
// and context-aware prompt construction.
package llm

import (
	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// NewClient constructs an Anthropic API client using the provided API key.
// If apiKey is empty, the client will attempt to use the ANTHROPIC_API_KEY
// environment variable automatically (SDK default behaviour).
func NewClient(apiKey string) *anthropic.Client {
	opts := []option.RequestOption{}
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	client := anthropic.NewClient(opts...)
	return &client
}
