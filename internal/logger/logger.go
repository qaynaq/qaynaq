package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/warpstreamlabs/bento/public/service"
)

type Logger struct {
	zerologger   zerolog.Logger
	bentoLogger  *service.Logger
	useBento     bool
	staticFields map[string]any
}

func NewFromZerolog(level string, staticFields map[string]any) *Logger {
	var logLevel zerolog.Level
	switch level {
	case "DEBUG", "debug":
		logLevel = zerolog.DebugLevel
	case "INFO", "info":
		logLevel = zerolog.InfoLevel
	case "WARN", "warn":
		logLevel = zerolog.WarnLevel
	case "ERROR", "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerologger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)

	if len(staticFields) > 0 {
		ctx := zerologger.With()
		for key, value := range staticFields {
			ctx = ctx.Interface(key, value)
		}
		zerologger = ctx.Logger()
	}

	return &Logger{
		zerologger:   zerologger,
		useBento:     false,
		staticFields: staticFields,
	}
}

func NewFromBento(bentoLogger *service.Logger, staticFields map[string]any) *Logger {
	writer := &bentoZerologWriter{
		staticFields: staticFields,
	}

	zerologger := zerolog.New(writer).With().Timestamp().Logger()

	if len(staticFields) > 0 {
		ctx := zerologger.With()
		for key, value := range staticFields {
			ctx = ctx.Interface(key, value)
		}
		zerologger = ctx.Logger()
	}

	return &Logger{
		zerologger:   zerologger,
		bentoLogger:  bentoLogger,
		useBento:     true,
		staticFields: staticFields,
	}
}

type bentoZerologWriter struct {
	staticFields map[string]any
}

func (w *bentoZerologWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (l *Logger) GetBentoLogger() *service.Logger {
	return l.bentoLogger
}

func (l *Logger) GetZerolog() zerolog.Logger {
	return l.zerologger
}

func (l *Logger) With(fields map[string]any) *Logger {
	ctx := l.zerologger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}

	return &Logger{
		zerologger:   ctx.Logger(),
		bentoLogger:  l.bentoLogger,
		useBento:     l.useBento,
		staticFields: l.staticFields,
	}
}

func (l *Logger) Info(msg string, fields ...any) {
	if l.useBento && l.bentoLogger != nil {
		if len(fields) > 0 {
			l.bentoLogger.With(fields...).Info(msg)
		} else {
			l.bentoLogger.Info(msg)
		}
	} else {
		event := l.zerologger.Info()
		l.addFields(event, fields...)
		event.Msg(msg)
	}
}

func (l *Logger) Warn(msg string, fields ...any) {
	if l.useBento && l.bentoLogger != nil {
		if len(fields) > 0 {
			l.bentoLogger.With(fields...).Warn(msg)
		} else {
			l.bentoLogger.Warn(msg)
		}
	} else {
		event := l.zerologger.Warn()
		l.addFields(event, fields...)
		event.Msg(msg)
	}
}

func (l *Logger) Error(msg string, fields ...any) {
	if l.useBento && l.bentoLogger != nil {
		if len(fields) > 0 {
			l.bentoLogger.With(fields...).Error(msg)
		} else {
			l.bentoLogger.Error(msg)
		}
	} else {
		event := l.zerologger.Error()
		l.addFields(event, fields...)
		event.Msg(msg)
	}
}

func (l *Logger) Debug(msg string, fields ...any) {
	if l.useBento && l.bentoLogger != nil {
		if len(fields) > 0 {
			l.bentoLogger.With(fields...).Debug(msg)
		} else {
			l.bentoLogger.Debug(msg)
		}
	} else {
		event := l.zerologger.Debug()
		l.addFields(event, fields...)
		event.Msg(msg)
	}
}

func (l *Logger) Fatal(msg string, fields ...any) {
	if l.useBento && l.bentoLogger != nil {
		if len(fields) > 0 {
			l.bentoLogger.With(fields...).Error(msg)
		} else {
			l.bentoLogger.Error(msg)
		}
	} else {
		event := l.zerologger.Fatal()
		l.addFields(event, fields...)
		event.Msg(msg)
	}
	os.Exit(1)
}

func (l *Logger) addFields(event *zerolog.Event, fields ...any) {
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			value := fields[i+1]
			event.Interface(key, value)
		}
	}
}
