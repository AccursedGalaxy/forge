package llm

import (
	"context"
	"log/slog"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/pgvector/pgvector-go"
)

// Embedder generates vector embeddings for text.
// NOTE: Anthropic does not currently expose a public embeddings API.
// This implementation is a stub that returns nil — the indexer handles nil gracefully
// by skipping storage. Swap this for a real embedding provider (e.g. Voyage AI, OpenAI)
// when vector search is needed.
type Embedder struct {
	client *anthropic.Client
}

// NewEmbedder creates an Embedder. The client is unused in the current stub implementation
// but is wired here for future use when Anthropic exposes an embeddings endpoint.
func NewEmbedder(client *anthropic.Client) *Embedder {
	return &Embedder{client: client}
}

// Embed returns a 1536-dimensional embedding vector for the given text.
// Returns nil on failure — callers must treat nil as "no embedding available".
func (e *Embedder) Embed(ctx context.Context, text string) pgvector.Vector {
	// TODO: replace with a real embeddings API call once available.
	// Anthropic does not yet have a public embeddings endpoint.
	// Options: Voyage AI (voyage-3-lite, 1024-dim), OpenAI text-embedding-3-small (1536-dim).
	slog.Debug("embedder: embed called (stub — no embeddings API configured)")
	return pgvector.Vector{}
}
