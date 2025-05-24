package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
	"time"

	"github.com/fatih/color"
)

// Logger - глобальный экземпляр логгера

// Init инициализирует глобальный логгер
func Init() slog.Logger {
	color.NoColor = false // Принудительно включаем цвета

	handler := &ColorHandler{
		colors: map[slog.Level]*color.Color{
			slog.LevelDebug: color.New(color.FgCyan),
			slog.LevelInfo:  color.New(color.FgGreen),
			slog.LevelWarn:  color.New(color.FgYellow),
			slog.LevelError: color.New(color.FgRed),
		},
		defaultColor: color.New(color.FgWhite),
		grayColor:    color.New(color.FgHiBlack),
	}
	var logger *slog.Logger
	logger = slog.New(handler)
	slog.SetDefault(logger)
	return *logger
}

// ColorHandler - кастомный обработчик для slog с цветным выводом
type ColorHandler struct {
	colors       map[slog.Level]*color.Color
	defaultColor *color.Color
	grayColor    *color.Color
}

func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	// Выбираем цвет для уровня логирования
	levelColor, ok := h.colors[r.Level]
	if !ok {
		levelColor = h.defaultColor
	}

	// Форматируем время
	timeStr := r.Time.Format("2006-01-02 15:04:05")

	// Выводим уровень и время
	levelColor.Fprintf(os.Stdout, "%-5s", r.Level.String())
	h.grayColor.Fprintf(os.Stdout, " [%s] ", timeStr)

	// Выводим сообщение
	color.New(color.FgHiWhite).Fprint(os.Stdout, r.Message)

	// Обрабатываем атрибуты
	attrs := make(map[string]interface{})
	r.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})

	// Выводим атрибуты как JSON серым цветом
	if len(attrs) > 0 {
		jsonData, _ := json.MarshalIndent(attrs, "", "  ")
		h.grayColor.Fprintf(os.Stdout, " %s", jsonData)
	}

	fmt.Fprintln(os.Stdout)
	return nil
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return h
}

// HTTPLogMiddleware - middleware для логирования HTTP-запросов
func HTTPLogMiddleware(logger slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем логирование для статики и swagger
		if c.Request.URL.Path == "/" || c.Request.URL.Path == "/static/*filepath" || c.Request.URL.Path == "/swagger/*any" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		// Обрабатываем запрос
		c.Next()

		// Формируем данные для лога
		latency := time.Since(start)
		status := c.Writer.Status()
		logFields := []interface{}{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", latency,
			"client", c.ClientIP(),
		}

		// Логируем в зависимости от статуса
		switch {
		case status >= 500:
			logger.Error("Ошибка сервера", logFields...)
		case status >= 400:
			logger.Warn("Ошибка клиента", logFields...)
		default:
			logger.Info("Обработан запрос", logFields...)
		}
	}
}
