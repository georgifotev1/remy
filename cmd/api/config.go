package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const (
	envDevelopment = "development"
	envProduction  = "production"
)

type config struct {
	env    string
	server struct {
		addr              string
		readHeaderTimeout time.Duration
		readTimeout       time.Duration
		writeTimeout      time.Duration
		idleTimeout       time.Duration
		shutdownTimeout   time.Duration
	}
	db struct {
		url             string
		maxConns        int32
		minConns        int32
		maxConnLifetime time.Duration
	}
	cors struct {
		allowedOrigins []string
	}
	log struct {
		level  slog.Level
		format string // "json" or "text"
	}
}

func loadConfig() (config, error) {
	load := &envLoader{}

	var cfg config

	cfg.env = load.str("APP_ENV", envDevelopment)
	isProd := cfg.env == envProduction

	cfg.server.addr = normalizeAddr(load.str("ADDR", ":8080"))
	cfg.server.readHeaderTimeout = load.duration("SERVER_READ_HEADER_TIMEOUT", 5*time.Second)
	cfg.server.readTimeout = load.duration("SERVER_READ_TIMEOUT", 10*time.Second)
	cfg.server.writeTimeout = load.duration("SERVER_WRITE_TIMEOUT", 10*time.Second)
	cfg.server.idleTimeout = load.duration("SERVER_IDLE_TIMEOUT", 60*time.Second)
	cfg.server.shutdownTimeout = load.duration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second)

	cfg.db.url = load.secret("DATABASE_URL", "postgres://remy:remy@localhost:5432/remy?sslmode=disable", isProd)
	cfg.db.maxConns = int32(load.int("DB_MAX_CONNS", 10))
	cfg.db.minConns = int32(load.int("DB_MIN_CONNS", 0))
	cfg.db.maxConnLifetime = load.duration("DB_MAX_CONN_LIFETIME", time.Hour)

	cfg.cors.allowedOrigins = load.list("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"})

	cfg.log.level = load.level("LOG_LEVEL", slog.LevelInfo)
	cfg.log.format = load.str("LOG_FORMAT", "json")

	if err := errors.Join(append(load.errs, cfg.validate())...); err != nil {
		return config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

func (c config) isProduction() bool { return c.env == envProduction }

func (c config) validate() error {
	var errs []error

	if c.env != envDevelopment && c.env != envProduction {
		errs = append(errs, fmt.Errorf("APP_ENV must be %q or %q, got %q", envDevelopment, envProduction, c.env))
	}
	if c.log.format != "json" && c.log.format != "text" {
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be \"json\" or \"text\", got %q", c.log.format))
	}
	if c.db.maxConns < 1 {
		errs = append(errs, errors.New("DB_MAX_CONNS must be >= 1"))
	}
	if c.db.minConns < 0 || c.db.minConns > c.db.maxConns {
		errs = append(errs, errors.New("DB_MIN_CONNS must be between 0 and DB_MAX_CONNS"))
	}
	if c.isProduction() {
		for _, o := range c.cors.allowedOrigins {
			if o == "*" {
				errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS must not contain \"*\" in production"))
			}
		}
	}
	return errors.Join(errs...)
}

// normalizeAddr turns "8080" into ":8080" so http.Server accepts it.
func normalizeAddr(a string) string {
	if !strings.Contains(a, ":") {
		return ":" + a
	}
	return a
}
