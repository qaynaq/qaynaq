package logger

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	logger                    zerolog.Logger
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	LogLevel                  logger.LogLevel
}

func NewGormLogger(zerologger zerolog.Logger) *GormLogger {
	return &GormLogger{
		logger:                    zerologger,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  logger.Warn,
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Info {
		l.logger.Info().Msgf(msg, data...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		l.logger.Warn().Msgf(msg, data...)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		l.logger.Error().Msgf(msg, data...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && l.IgnoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		l.logger.Error().
			Err(err).
			Str("sql", sql).
			Int64("rows", rows).
			Dur("elapsed", elapsed).
			Msg("database query error")
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		l.logger.Warn().
			Str("sql", sql).
			Int64("rows", rows).
			Dur("elapsed", elapsed).
			Str("threshold", l.SlowThreshold.String()).
			Msg("slow query")
	case l.LogLevel >= logger.Info:
		l.logger.Debug().
			Str("sql", sql).
			Int64("rows", rows).
			Dur("elapsed", elapsed).
			Msg("database query")
	}
}
