// Structured JSON logging for Lambda invocations.
package platform

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func NewLogger() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{Logger: slog.New(handler)}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	if requestID == "" {
		return l
	}
	return &Logger{Logger: l.With("request_id", requestID)}
}

func (l *Logger) LogError(ctx context.Context, msg string, err error, attrs ...any) {
	args := append([]any{"error", err.Error()}, attrs...)
	l.ErrorContext(ctx, msg, args...)
}

func LogValue(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}
