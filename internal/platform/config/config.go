package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

type Config struct {
	Env    string
	Server ServerConfig
	DB     DBConfig
	CORS   CORSConfig
	Log    LogConfig
}

type ServerConfig struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

type DBConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type LogConfig struct {
	Level  slog.Level
	Format string // "json" or "text"
}

func Load() (Config, error) {
	load := &loader{}

	env := load.str("APP_ENV", EnvDevelopment)
	isProd := env == EnvProduction

	cfg := Config{
		Env: env,
		Server: ServerConfig{
			Addr:              load.str("ADDR", ":8080"),
			ReadHeaderTimeout: load.duration("SERVER_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:       load.duration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:      load.duration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:       load.duration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:   load.duration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		DB: DBConfig{
			URL:             load.secret("DATABASE_URL", "postgres://remy:remy@localhost:5432/remy?sslmode=disable", isProd),
			MaxConns:        int32(load.int("DB_MAX_CONNS", 10)),
			MinConns:        int32(load.int("DB_MIN_CONNS", 0)),
			MaxConnLifetime: load.duration("DB_MAX_CONN_LIFETIME", time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins: load.list("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		Log: LogConfig{
			Level:  load.level("LOG_LEVEL", slog.LevelInfo),
			Format: load.str("LOG_FORMAT", "json"),
		},
	}

	cfg.Server.Addr = normalizeAddr(cfg.Server.Addr)

	if err := errors.Join(append(load.errs, cfg.validate())...); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

func (c Config) IsProduction() bool { return c.Env == EnvProduction }

func (c Config) validate() error {
	var errs []error

	if c.Env != EnvDevelopment && c.Env != EnvProduction {
		errs = append(errs, fmt.Errorf("APP_ENV must be %q or %q, got %q", EnvDevelopment, EnvProduction, c.Env))
	}
	if c.Log.Format != "json" && c.Log.Format != "text" {
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be \"json\" or \"text\", got %q", c.Log.Format))
	}
	if c.DB.MaxConns < 1 {
		errs = append(errs, errors.New("DB_MAX_CONNS must be >= 1"))
	}
	if c.DB.MinConns < 0 || c.DB.MinConns > c.DB.MaxConns {
		errs = append(errs, errors.New("DB_MIN_CONNS must be between 0 and DB_MAX_CONNS"))
	}
	if c.IsProduction() {
		for _, o := range c.CORS.AllowedOrigins {
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
