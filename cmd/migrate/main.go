// Command migrate applies or rolls back the goose SQL migrations.
//
// Usage:
//
//	migrate        apply all pending migrations
//	migrate up     same as above
//	migrate down   roll back the most recent migration
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"github.com/georgifotev1/remy/internal/platform/config"
	"github.com/georgifotev1/remy/internal/platform/log"
	"github.com/georgifotev1/remy/migrations"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}
	if direction != "up" && direction != "down" {
		return fmt.Errorf("unknown direction %q (want up or down)", direction)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Version = version

	logger := log.New(os.Stderr, cfg)

	sqlDB, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("migration locker: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("migration provider: %w", err)
	}
	if direction == "down" {
		return down(ctx, provider, logger)
	}
	return up(ctx, provider, logger)
}

func up(ctx context.Context, provider *goose.Provider, logger *slog.Logger) error {
	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	if len(results) == 0 {
		logger.Info("migrations up to date")
		return nil
	}
	for _, result := range results {
		logger.Info("migration applied",
			"migration_version", result.Source.Version,
			"path", result.Source.Path,
			"duration", result.Duration.String(),
		)
	}
	return nil
}

func down(ctx context.Context, provider *goose.Provider, logger *slog.Logger) error {
	result, err := provider.Down(ctx)
	if errors.Is(err, goose.ErrNoNextVersion) {
		logger.Info("no migrations to roll back")
		return nil
	}
	if err != nil {
		return fmt.Errorf("roll back migration: %w", err)
	}
	logger.Info("migration rolled back",
		"migration_version", result.Source.Version,
		"path", result.Source.Path,
		"duration", result.Duration.String(),
	)
	return nil
}
