package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/rs/zerolog"
)

type BentoLogWriter struct {
	logger zerolog.Logger
}

func NewBentoLogWriter(logger zerolog.Logger) io.Writer {
	return &BentoLogWriter{logger: logger}
}

func (w *BentoLogWriter) Write(p []byte) (n int, err error) {
	line := string(p)

	fields := parseLogfmt(line)

	level, hasLevel := fields["level"]
	msg, hasMsg := fields["msg"]

	if !hasLevel {
		level = "info"
	}
	if !hasMsg {
		msg = line
	}

	delete(fields, "level")
	delete(fields, "msg")

	var event *zerolog.Event
	switch level {
	case "debug":
		event = w.logger.Debug()
	case "info":
		event = w.logger.Info()
	case "warn", "warning":
		event = w.logger.Warn()
	case "error":
		event = w.logger.Error()
	default:
		event = w.logger.Info()
	}

	for key, value := range fields {
		event.Str(key, value)
	}

	event.Msg(msg)
	return len(p), nil
}

func ConfigureBentoLogger(level string, staticFields map[string]any) io.Writer {
	var logLevel zerolog.Level
	switch level {
	case "DEBUG", "debug":
		logLevel = zerolog.DebugLevel
	case "INFO", "info":
		logLevel = zerolog.InfoLevel
	case "WARN", "warn", "WARNING", "warning":
		logLevel = zerolog.WarnLevel
	case "ERROR", "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)

	if len(staticFields) > 0 {
		ctx := logger.With()
		for key, value := range staticFields {
			ctx = ctx.Interface(key, value)
		}
		logger = ctx.Logger()
	}

	return NewBentoLogWriter(logger)
}

func ParseBentoLogLevel(level string) string {
	switch level {
	case "NONE", "none", "silent":
		return "ERROR"
	case "FATAL", "fatal":
		return "ERROR"
	case "ERROR", "error":
		return "ERROR"
	case "WARN", "warn", "WARNING", "warning":
		return "WARN"
	case "INFO", "info":
		return "INFO"
	case "DEBUG", "debug":
		return "DEBUG"
	case "TRACE", "trace":
		return "DEBUG"
	default:
		return "INFO"
	}
}

func parseLogfmt(line string) map[string]string {
	fields := make(map[string]string)

	var key string
	var inQuotes bool
	var currentToken string

	for i := 0; i < len(line); i++ {
		ch := line[i]

		switch {
		case ch == '=' && !inQuotes && key == "":
			key = currentToken
			currentToken = ""
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ' ' && !inQuotes:
			if key != "" {
				fields[key] = currentToken
				key = ""
				currentToken = ""
			} else if currentToken != "" {
				currentToken = ""
			}
		case ch == '\n' || ch == '\r':
		default:
			currentToken += string(ch)
		}
	}

	if key != "" && currentToken != "" {
		fields[key] = currentToken
	}

	return fields
}

func CreateBentoContextLogger(baseLogger zerolog.Logger, componentName string, fields map[string]any) zerolog.Logger {
	ctx := baseLogger.With().Str("@service", "qaynaq").Str("component", componentName)

	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}

	return ctx.Logger()
}

type slogHandler struct {
	logger zerolog.Logger
	attrs  []slog.Attr
	groups []string
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	var zLevel zerolog.Level
	switch level {
	case slog.LevelDebug:
		zLevel = zerolog.DebugLevel
	case slog.LevelInfo:
		zLevel = zerolog.InfoLevel
	case slog.LevelWarn:
		zLevel = zerolog.WarnLevel
	case slog.LevelError:
		zLevel = zerolog.ErrorLevel
	default:
		zLevel = zerolog.InfoLevel
	}
	return h.logger.GetLevel() <= zLevel
}

func (h *slogHandler) Handle(_ context.Context, record slog.Record) error {
	var event *zerolog.Event

	switch record.Level {
	case slog.LevelDebug:
		event = h.logger.Debug()
	case slog.LevelInfo:
		event = h.logger.Info()
	case slog.LevelWarn:
		event = h.logger.Warn()
	case slog.LevelError:
		event = h.logger.Error()
	default:
		event = h.logger.Info()
	}

	for _, attr := range h.attrs {
		event = event.Interface(attr.Key, attr.Value.Any())
	}

	record.Attrs(func(attr slog.Attr) bool {
		event = event.Interface(attr.Key, attr.Value.Any())
		return true
	})

	event.Msg(record.Message)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &slogHandler{
		logger: h.logger,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &slogHandler{
		logger: h.logger,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

func NewSlogLogger(level string, staticFields map[string]any) *slog.Logger {
	var logLevel zerolog.Level
	switch level {
	case "DEBUG", "debug":
		logLevel = zerolog.DebugLevel
	case "INFO", "info":
		logLevel = zerolog.InfoLevel
	case "WARN", "warn", "WARNING", "warning":
		logLevel = zerolog.WarnLevel
	case "ERROR", "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)

	if len(staticFields) > 0 {
		ctx := logger.With()
		for key, value := range staticFields {
			ctx = ctx.Interface(key, value)
		}
		logger = ctx.Logger()
	}

	handler := &slogHandler{
		logger: logger,
		attrs:  []slog.Attr{},
		groups: []string{},
	}

	return slog.New(handler)
}
