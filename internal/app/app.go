package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/georgifotev1/remy/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	pool   *pgxpool.Pool
	srv    *http.Server
}

func New(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) *App {
	app := &App{cfg: cfg, logger: logger, pool: pool}

	app.srv = &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           app.routes(),
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	return app
}

func (app *App) Run(ctx context.Context) error {
	app.logger.Info("starting server", "addr", app.cfg.Server.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := app.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
		app.logger.Info("shutting down", "timeout", app.cfg.Server.ShutdownTimeout)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := app.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}
	app.logger.Info("server stopped")
	return nil
}
