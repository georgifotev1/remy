package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/georgifotev1/remy/internal/data"
	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)

	logger.Info("starting db")
	pool, err := openDB(ctx, cfg)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	logger.Info("db connected")
	defer func() {
		logger.Info("closing db")
		pool.Close()
		logger.Info("db closed")
	}()

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(pool),
	}

	return app.serve(ctx)
}

func newLogger(cfg config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     cfg.log.level,
		AddSource: !cfg.isProduction(),
	}

	var h slog.Handler
	if cfg.log.format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(h).With("env", cfg.env)
}

func openDB(ctx context.Context, cfg config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.db.url)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = cfg.db.maxConns
	poolCfg.MinConns = cfg.db.minConns
	poolCfg.MaxConnLifetime = cfg.db.maxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
