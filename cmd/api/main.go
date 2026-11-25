package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/georgifotev1/remy/internal/app"
	"github.com/georgifotev1/remy/internal/platform/config"
	"github.com/georgifotev1/remy/internal/platform/db"
	"github.com/georgifotev1/remy/internal/platform/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg)
	slog.SetDefault(log)

	log.Info("starting db")
	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	log.Info("db connected")
	defer func() {
		log.Info("closing db")
		pool.Close()
		log.Info("db closed")
	}()

	return app.New(cfg, log, pool).Run(ctx)
}
