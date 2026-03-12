package db

import (
	"context"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	pgxstdlib "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending up migrations embedded in the binary.
// It is idempotent — running it multiple times is safe.
func RunMigrations(ctx context.Context, connString string) error {
	// Open a *sql.DB via pgx stdlib adapter (migrate needs database/sql)
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("db: parse conn string: %w", err)
	}
	sqlDB := pgxstdlib.OpenDB(*cfg)
	defer sqlDB.Close()

	// Ping to verify connectivity before running migrations
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db: ping failed: %w", err)
	}

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("db: create migration source: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("db: create migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "forge", driver)
	if err != nil {
		return fmt.Errorf("db: create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("db: run migrations: %w", err)
	}

	return nil
}
