package logger

import (
	"io"

	"golang.org/x/exp/slog"
)

type Logger struct {
	*slog.Logger
}

func InitLogger(w io.Writer) *Logger {
	options := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(w, options)
	return &Logger{Logger: slog.New(handler)}
}
