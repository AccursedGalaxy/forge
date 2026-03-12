// Package context provides context retrieval and indexing for Forge sessions.
package context

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"

	"github.com/accursedgalaxy/forge/internal/db"
)

// Retriever performs vector similarity search against stored context chunks.
type Retriever struct {
	queries *db.Queries
}

// NewRetriever creates a Retriever backed by the provided sqlc queries.
func NewRetriever(queries *db.Queries) *Retriever {
	return &Retriever{queries: queries}
}

// TopK returns the k most similar context chunk contents for the given project and query vector.
// If embedding is empty (no embedding configured), it returns nil immediately.
// Errors are logged and a nil slice is returned — callers must handle nil gracefully.
func (r *Retriever) TopK(ctx context.Context, projectID uuid.UUID, embedding pgvector.Vector, k int) []string {
	if len(embedding.Slice()) == 0 {
		return nil
	}

	rows, err := r.queries.SearchSimilarChunks(ctx, db.SearchSimilarChunksParams{
		ProjectID: projectID,
		Embedding: embedding,
		Limit:     int32(k),
	})
	if err != nil {
		slog.Warn("context: TopK query failed", "project_id", projectID, "err", err)
		return nil
	}

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Content)
	}
	return out
}
