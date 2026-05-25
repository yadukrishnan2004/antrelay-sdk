package logger

import (
	"io"
	"log/slog"
	"os"
)

type Config struct{
	Level slog.Level

	Format string

	Output io.Writer
}

func New(cfg Config) *slog.Logger{
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	var handler slog.Handler

	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(cfg.Output, opts)
	default:
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return  slog.New(handler)
}

func NewFileLogger(path string, format string) (*slog.Logger,error) {
	f,err :=os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644)
	if err != nil {
		return nil,err
	}

	return New(Config{
		Level: slog.LevelInfo,
		Format: format,
		Output: f,
	}),nil
}

func NewMultiLogger(path string) (*slog.Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

		multi := io.MultiWriter(os.Stdout, f)

	return New(Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: multi,
	}), nil
}