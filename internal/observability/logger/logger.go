// Package logger creates the application structured logger.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	defaultServiceName = "go-tweets"
	defaultEnv         = "development"
	defaultVersion     = "dev"
)

type Config struct {
	Service string
	Env     string
	Version string
	Level   string
	Writer  io.Writer
}

func New(cfg Config) *slog.Logger {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	env := defaultString(cfg.Env, defaultEnv)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: parseLevel(cfg.Level, env),
	})

	return slog.New(handler).With(
		slog.String("service", defaultString(cfg.Service, defaultServiceName)),
		slog.String("env", env),
		slog.String("version", defaultString(cfg.Version, defaultVersion)),
	)
}

func defaultString(value string, fallback string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fallback
	}

	return trimmedValue
}

func parseLevel(level string, env string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	if strings.EqualFold(env, "development") || strings.EqualFold(env, "test") {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}
