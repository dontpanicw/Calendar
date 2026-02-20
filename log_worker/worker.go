package log_worker

import (
	"context"
	"fmt"
	"log/slog"
)

type Logger struct {
	logChannel chan string
}

func NewLogger() *Logger {
	return &Logger{
		logChannel: make(chan string, 100),
	}
}

// Публичный методы для записи логов из хендлеров
func (l *Logger) Write(message string) {
	select {
	case l.logChannel <- message:
	default:
		// Канал переполнен - логируем ошибку, но не блокируем хендлер
		slog.Error("Log channel full, discarding message", "message", message)
	}
}
func (l *Logger) Writef(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	select {
	case l.logChannel <- message:
	default:
		// При переполнении используем fallback
		slog.Error("Log channel full", "format", format, "args", fmt.Sprint(args...))
	}
}

func (l *Logger) Log(ctx context.Context) {
	for {
		select {
		case message := <-l.logChannel:
			slog.Info(message)
		case <-ctx.Done():
			slog.Info("log_worker stopped")
			return
		}
	}
}
