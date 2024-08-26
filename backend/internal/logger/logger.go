package logger

import (
	"io"

	"golang.org/x/exp/slog"
)

// Logger оборачивает стандартный slog.Logger для предоставления удобного интерфейса
// для логирования в приложении.
type Logger struct {
	*slog.Logger
}

// InitLogger инициализирует новый экземпляр Logger с указанным выходным потоком.
// Эта функция настраивает логгер для записи логов в формате JSON с уровнем логирования Info.
// Аргументы:
//
//	w - io.Writer, который будет использоваться для записи логов (например, файл, stdout).
//
// Возвращает:
//
//	*Logger - новый экземпляр Logger, настроенный для записи логов в формате JSON.
func InitLogger(w io.Writer) *Logger {
	// Создаем опции для обработчика логов, устанавливая уровень логирования на Info.
	options := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// Создаем новый JSON-обработчик для записи логов в указанный выходной поток.
	handler := slog.NewJSONHandler(w, options)

	// Возвращаем новый экземпляр Logger, использующий созданный обработчик.
	return &Logger{Logger: slog.New(handler)}
}
