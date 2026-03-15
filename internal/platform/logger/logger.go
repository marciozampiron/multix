package logger

import (
	"log/slog"
	"os"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, err error, args ...any)
	Fatal(msg string, err error, args ...any)
	With(args ...any) Logger
}

type slogAdapter struct {
	logger *slog.Logger
}

func New(level string) Logger {
	lvl := slog.LevelInfo
	if level == "debug" {
		lvl = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, opts)
	if os.Getenv("MULTIX_LOG_FORMAT") == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return &slogAdapter{logger: slog.New(handler)}
}

func (s *slogAdapter) Debug(msg string, args ...any) { s.logger.Debug(msg, args...) }
func (s *slogAdapter) Info(msg string, args ...any)  { s.logger.Info(msg, args...) }
func (s *slogAdapter) Warn(msg string, args ...any)  { s.logger.Warn(msg, args...) }
func (s *slogAdapter) Error(msg string, err error, args ...any) {
	if err != nil {
		args = append(args, slog.String("error", err.Error()))
	}
	s.logger.Error(msg, args...)
}
func (s *slogAdapter) Fatal(msg string, err error, args ...any) {
	s.Error(msg, err, args...)
	os.Exit(1)
}
func (s *slogAdapter) With(args ...any) Logger {
	return &slogAdapter{logger: s.logger.With(args...)}
}
