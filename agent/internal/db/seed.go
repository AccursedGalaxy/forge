package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

// DefaultUserID is the hardcoded UUID for the single local user.
var DefaultUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// EnsureDefaultUser creates the default local user if it doesn't exist.
// Idempotent — safe to call on every startup.
func EnsureDefaultUser(ctx context.Context, q *Queries) error {
	_, err := q.GetDefaultUser(ctx)
	if err == nil {
		slog.Info("default user already exists")
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = q.CreateUser(ctx, CreateUserParams{
		ID:   DefaultUserID,
		Name: "Local User",
	})
	if err != nil {
		return err
	}
	slog.Info("default user ensured", "id", DefaultUserID)
	return nil
}

// EnsureDefaultProvider creates the default Claude provider if none exist.
// Idempotent — safe to call on every startup.
func EnsureDefaultProvider(ctx context.Context, q *Queries) error {
	_, err := q.GetDefaultProvider(ctx)
	if err == nil {
		slog.Info("default provider already exists")
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	cfg, _ := json.Marshal(map[string]string{})
	_, err = q.CreateProvider(ctx, CreateProviderParams{
		Name:         "claude",
		ProviderType: "claude",
		Config:       json.RawMessage(cfg),
		IsDefault:    true,
	})
	if err != nil {
		return err
	}
	slog.Info("default provider ensured", "name", "claude")
	return nil
}
