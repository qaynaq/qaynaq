package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewGormLogger(t *testing.T) {
	zerologger := zerolog.New(&bytes.Buffer{})
	gormLogger := NewGormLogger(zerologger)

	if gormLogger == nil {
		t.Fatal("Expected non-nil GORM logger")
	}

	if gormLogger.SlowThreshold != 200*time.Millisecond {
		t.Errorf("Expected SlowThreshold 200ms, got %v", gormLogger.SlowThreshold)
	}

	if !gormLogger.IgnoreRecordNotFoundError {
		t.Error("Expected IgnoreRecordNotFoundError to be true")
	}

	if gormLogger.LogLevel != logger.Warn {
		t.Errorf("Expected LogLevel Warn, got %v", gormLogger.LogLevel)
	}
}

func TestGormLoggerLogMode(t *testing.T) {
	zerologger := zerolog.New(&bytes.Buffer{})
	gormLogger := NewGormLogger(zerologger)

	newLogger := gormLogger.LogMode(logger.Info)
	if newLogger == nil {
		t.Fatal("Expected non-nil logger from LogMode")
	}

	gormLoggerTyped, ok := newLogger.(*GormLogger)
	if !ok {
		t.Fatal("Expected logger to be *GormLogger type")
	}

	if gormLoggerTyped.LogLevel != logger.Info {
		t.Errorf("Expected LogLevel Info, got %v", gormLoggerTyped.LogLevel)
	}
}

func TestGormLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)
	gormLogger.LogLevel = logger.Info

	ctx := context.Background()
	gormLogger.Info(ctx, "test info message")

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "info" {
		t.Errorf("Expected level 'info', got '%v'", logData["level"])
	}

	if logData["message"] != "test info message" {
		t.Errorf("Expected message 'test info message', got '%v'", logData["message"])
	}
}

func TestGormLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)

	ctx := context.Background()
	gormLogger.Warn(ctx, "test warn message")

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "warn" {
		t.Errorf("Expected level 'warn', got '%v'", logData["level"])
	}
}

func TestGormLoggerError(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)

	ctx := context.Background()
	gormLogger.Error(ctx, "test error message")

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "error" {
		t.Errorf("Expected level 'error', got '%v'", logData["level"])
	}
}

func TestGormLoggerTraceWithError(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)

	ctx := context.Background()
	begin := time.Now()
	testError := errors.New("database error")

	gormLogger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM users WHERE id = ?", 0
	}, testError)

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "error" {
		t.Errorf("Expected level 'error', got '%v'", logData["level"])
	}

	if logData["message"] != "database query error" {
		t.Errorf("Expected message 'database query error', got '%v'", logData["message"])
	}

	if logData["sql"] == nil {
		t.Error("Expected 'sql' field to be present")
	}

	if logData["rows"] == nil {
		t.Error("Expected 'rows' field to be present")
	}

	if logData["elapsed"] == nil {
		t.Error("Expected 'elapsed' field to be present")
	}
}

func TestGormLoggerTraceIgnoresRecordNotFound(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)
	gormLogger.IgnoreRecordNotFoundError = true

	ctx := context.Background()
	begin := time.Now()

	gormLogger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM users WHERE id = ?", 0
	}, gorm.ErrRecordNotFound)

	if buf.Len() > 0 {
		t.Error("Expected no log output for ErrRecordNotFound")
	}
}

func TestGormLoggerTraceRecordNotFoundWhenNotIgnored(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)
	gormLogger.IgnoreRecordNotFoundError = false

	ctx := context.Background()
	begin := time.Now()

	gormLogger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM users WHERE id = ?", 0
	}, gorm.ErrRecordNotFound)

	if buf.Len() == 0 {
		t.Error("Expected log output when IgnoreRecordNotFoundError is false")
	}
}

func TestGormLoggerTraceSlowQuery(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)
	gormLogger.SlowThreshold = 1 * time.Millisecond

	ctx := context.Background()
	begin := time.Now().Add(-100 * time.Millisecond)

	gormLogger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM large_table", 1000
	}, nil)

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "warn" {
		t.Errorf("Expected level 'warn' for slow query, got '%v'", logData["level"])
	}

	if logData["message"] != "slow query" {
		t.Errorf("Expected message 'slow query', got '%v'", logData["message"])
	}

	if logData["threshold"] == nil {
		t.Error("Expected 'threshold' field to be present")
	}
}

func TestGormLoggerTraceFastQuery(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger().Level(zerolog.DebugLevel)
	gormLogger := NewGormLogger(zerologger)
	gormLogger.LogLevel = logger.Info
	gormLogger.SlowThreshold = 1 * time.Second

	ctx := context.Background()
	begin := time.Now()

	gormLogger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM users LIMIT 10", 10
	}, nil)

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "debug" {
		t.Errorf("Expected level 'debug' for fast query, got '%v'", logData["level"])
	}

	if logData["message"] != "database query" {
		t.Errorf("Expected message 'database query', got '%v'", logData["message"])
	}
}

func TestGormLoggerSilentMode(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	gormLogger := NewGormLogger(zerologger)
	gormLogger.LogLevel = logger.Silent

	ctx := context.Background()

	gormLogger.Info(ctx, "should not log")
	gormLogger.Warn(ctx, "should not log")
	gormLogger.Error(ctx, "should not log")

	gormLogger.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	if buf.Len() > 0 {
		t.Error("Expected no output in Silent mode")
	}
}

func TestGormLoggerLevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		level       logger.LogLevel
		shouldInfo  bool
		shouldWarn  bool
		shouldError bool
	}{
		{"Silent", logger.Silent, false, false, false},
		{"Error", logger.Error, false, false, true},
		{"Warn", logger.Warn, false, true, true},
		{"Info", logger.Info, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			zerologger := zerolog.New(&buf).With().Timestamp().Logger()
			gormLogger := NewGormLogger(zerologger)
			gormLogger.LogLevel = tt.level

			ctx := context.Background()

			buf.Reset()
			gormLogger.Info(ctx, "info")
			infoLogged := buf.Len() > 0

			buf.Reset()
			gormLogger.Warn(ctx, "warn")
			warnLogged := buf.Len() > 0

			buf.Reset()
			gormLogger.Error(ctx, "error")
			errorLogged := buf.Len() > 0

			if infoLogged != tt.shouldInfo {
				t.Errorf("Info logging mismatch: expected %v, got %v", tt.shouldInfo, infoLogged)
			}
			if warnLogged != tt.shouldWarn {
				t.Errorf("Warn logging mismatch: expected %v, got %v", tt.shouldWarn, warnLogged)
			}
			if errorLogged != tt.shouldError {
				t.Errorf("Error logging mismatch: expected %v, got %v", tt.shouldError, errorLogged)
			}
		})
	}
}
