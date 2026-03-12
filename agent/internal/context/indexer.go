package context

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/accursedgalaxy/forge/internal/db"
	"github.com/accursedgalaxy/forge/internal/llm"
)

// Indexer chunks session output, embeds it, and stores context chunks for future retrieval.
type Indexer struct {
	queries *db.Queries
	embedder *llm.Embedder
}

// NewIndexer creates an Indexer.
func NewIndexer(queries *db.Queries, embedder *llm.Embedder) *Indexer {
	return &Indexer{queries: queries, embedder: embedder}
}

// Index chunks fullOutput and stores embedded context chunks for the given project.
// This is always called in a background goroutine and never blocks the caller.
// Errors are logged and ignored — indexing failure is non-fatal.
func (idx *Indexer) Index(ctx context.Context, projectID uuid.UUID, sessionID string, fullOutput string) {
	chunks := chunk(fullOutput, 1000, 100)

	stored := 0
	for i, text := range chunks {
		if strings.TrimSpace(text) == "" {
			continue
		}

		embedding := idx.embedder.Embed(ctx, text)
		if len(embedding.Slice()) == 0 {
			// No embedder configured — skip all remaining chunks.
			break
		}

		source := sessionID
		if i > 0 {
			source = sessionID // all chunks share the session as source
		}

		_, err := idx.queries.InsertContextChunk(ctx, db.InsertContextChunkParams{
			ProjectID: projectID,
			Content:   text,
			Source:    source,
			Embedding: embedding,
		})
		if err != nil {
			slog.Warn("indexer: insert chunk failed", "project_id", projectID, "chunk", i, "err", err)
			continue
		}
		stored++
	}

	if stored > 0 {
		slog.Info("indexer: stored context chunks", "project_id", projectID, "count", stored)
	}
}

// chunk splits text into overlapping segments of approximately chunkSize runes
// with overlap runes of overlap between consecutive chunks.
func chunk(text string, chunkSize, overlap int) []string {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}

	var chunks []string
	for start := 0; start < len(runes); start += chunkSize - overlap {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
		if end == len(runes) {
			break
		}
	}
	return chunks
}
