package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:              app.config.server.addr,
		Handler:           app.routes(),
		ReadHeaderTimeout: app.config.server.readHeaderTimeout,
		ReadTimeout:       app.config.server.readTimeout,
		WriteTimeout:      app.config.server.writeTimeout,
		IdleTimeout:       app.config.server.idleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	app.logger.Info("starting server", "addr", srv.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
		app.logger.Info("shutting down", "timeout", app.config.server.shutdownTimeout)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.server.shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}
	app.logger.Info("server stopped")
	return nil
}
