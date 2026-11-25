package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// envLoader reads env vars and accumulates parse errors.
type envLoader struct {
	errs []error
}

func (l *envLoader) str(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// secret is like str, but in required mode (production) the variable
// must be set explicitly, so dev defaults never leak into prod.
func (l *envLoader) secret(key, devDefault string, required bool) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	if required {
		l.errs = append(l.errs, fmt.Errorf("%s is required", key))
		return ""
	}
	return devDefault
}

func (l *envLoader) int(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: invalid integer %q", key, v))
		return def
	}
	return n
}

func (l *envLoader) duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: invalid duration %q (use e.g. 5s, 2m)", key, v))
		return def
	}
	return d
}

func (l *envLoader) list(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (l *envLoader) level(key string, def slog.Level) slog.Level {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(v)); err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: invalid level %q (debug, info, warn, error)", key, v))
		return def
	}
	return lvl
}
