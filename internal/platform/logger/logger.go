package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/georgifotev1/remy/internal/platform/config"
)

func New(cfg config.Config) *slog.Logger {
	return NewWithWriter(os.Stdout, cfg)
}

func NewWithWriter(w io.Writer, cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     cfg.Log.Level,
		AddSource: !cfg.IsProduction(),
	}

	var h slog.Handler
	if cfg.Log.Format == "text" {
		h = slog.NewTextHandler(w, opts)
	} else {
		h = slog.NewJSONHandler(w, opts)
	}

	return slog.New(h).With("env", cfg.Env)
}
