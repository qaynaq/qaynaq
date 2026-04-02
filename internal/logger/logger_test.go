package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewFromZerologWithNilStaticFields(t *testing.T) {
	log := NewFromZerolog("INFO", nil)

	if log == nil {
		t.Fatal("Expected non-nil logger")
	}

	if log.useBento {
		t.Error("Expected useBento to be false")
	}

	if len(log.staticFields) != 0 {
		t.Errorf("Expected empty static fields, got %d", len(log.staticFields))
	}
}

func TestNewFromZerologWithEmptyStaticFields(t *testing.T) {
	staticFields := make(map[string]any)
	log := NewFromZerolog("INFO", staticFields)

	if log == nil {
		t.Fatal("Expected non-nil logger")
	}

	if len(log.staticFields) != 0 {
		t.Errorf("Expected empty static fields, got %d", len(log.staticFields))
	}
}

func TestNewFromZerologWithStaticFields(t *testing.T) {
	staticFields := map[string]any{
		"service":     "qaynaq",
		"environment": "test",
		"version":     "1.0.0",
	}

	log := NewFromZerolog("INFO", staticFields)

	if log == nil {
		t.Fatal("Expected non-nil logger")
	}

	if len(log.staticFields) != len(staticFields) {
		t.Errorf("Expected %d static fields, got %d", len(staticFields), len(log.staticFields))
	}

	for key, expectedValue := range staticFields {
		if actualValue, ok := log.staticFields[key]; !ok {
			t.Errorf("Expected static field '%s' to be present", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected static field '%s' = '%v', got '%v'", key, expectedValue, actualValue)
		}
	}
}

func TestNewFromZerologLogLevels(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		shouldDebug bool
		shouldInfo  bool
		shouldWarn  bool
		shouldError bool
	}{
		{"DEBUG level", "DEBUG", true, true, true, true},
		{"INFO level", "INFO", false, true, true, true},
		{"WARN level", "WARN", false, false, true, true},
		{"ERROR level", "ERROR", false, false, false, true},
		{"Unknown level defaults to INFO", "INVALID", false, true, true, true},
		{"Empty level defaults to INFO", "", false, true, true, true},
		{"Lowercase debug", "debug", true, true, true, true},
		{"Lowercase info", "info", false, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			log := NewFromZerolog(tt.level, nil)
			if log == nil {
				t.Fatal("Expected non-nil logger")
			}

			log.zerologger = log.zerologger.Output(&buf)

			buf.Reset()
			log.Debug("debug message")
			debugLogged := buf.Len() > 0

			buf.Reset()
			log.Info("info message")
			infoLogged := buf.Len() > 0

			buf.Reset()
			log.Warn("warn message")
			warnLogged := buf.Len() > 0

			buf.Reset()
			log.Error("error message")
			errorLogged := buf.Len() > 0

			if debugLogged != tt.shouldDebug {
				t.Errorf("Debug logging mismatch: expected %v, got %v", tt.shouldDebug, debugLogged)
			}
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

func TestLoggerInfoWithFields(t *testing.T) {
	var buf bytes.Buffer

	log := NewFromZerolog("INFO", nil)
	log.zerologger = log.zerologger.Output(&buf)

	log.Info("test message", "key1", "value1", "key2", 123)

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "info" {
		t.Errorf("Expected level 'info', got '%v'", logData["level"])
	}

	if logData["message"] != "test message" {
		t.Errorf("Expected message 'test message', got '%v'", logData["message"])
	}

	if logData["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got '%v'", logData["key1"])
	}

	if logData["key2"] != float64(123) {
		t.Errorf("Expected key2 '123', got '%v'", logData["key2"])
	}
}

func TestLoggerWarnWithFields(t *testing.T) {
	var buf bytes.Buffer

	log := NewFromZerolog("WARN", nil)
	log.zerologger = log.zerologger.Output(&buf)

	log.Warn("warning message", "reason", "test")

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "warn" {
		t.Errorf("Expected level 'warn', got '%v'", logData["level"])
	}

	if logData["message"] != "warning message" {
		t.Errorf("Expected message 'warning message', got '%v'", logData["message"])
	}

	if logData["reason"] != "test" {
		t.Errorf("Expected reason 'test', got '%v'", logData["reason"])
	}
}

func TestLoggerErrorWithFields(t *testing.T) {
	var buf bytes.Buffer

	log := NewFromZerolog("ERROR", nil)
	log.zerologger = log.zerologger.Output(&buf)

	log.Error("error occurred", "error_code", 500, "error_msg", "internal error")

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "error" {
		t.Errorf("Expected level 'error', got '%v'", logData["level"])
	}

	if logData["message"] != "error occurred" {
		t.Errorf("Expected message 'error occurred', got '%v'", logData["message"])
	}

	if logData["error_code"] != float64(500) {
		t.Errorf("Expected error_code '500', got '%v'", logData["error_code"])
	}
}

func TestLoggerDebugWithFields(t *testing.T) {
	var buf bytes.Buffer

	log := NewFromZerolog("DEBUG", nil)
	log.zerologger = log.zerologger.Output(&buf)

	log.Debug("debug info", "trace_id", "abc123")

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "debug" {
		t.Errorf("Expected level 'debug', got '%v'", logData["level"])
	}

	if logData["message"] != "debug info" {
		t.Errorf("Expected message 'debug info', got '%v'", logData["message"])
	}

	if logData["trace_id"] != "abc123" {
		t.Errorf("Expected trace_id 'abc123', got '%v'", logData["trace_id"])
	}
}

func TestLoggerWithStaticFields(t *testing.T) {
	var buf bytes.Buffer

	staticFields := map[string]any{
		"@service": "qaynaq",
		"env":      "test",
	}

	log := NewFromZerolog("INFO", staticFields)
	log.zerologger = log.zerologger.Output(&buf)

	log.Info("test message", "extra", "value")

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["@service"] != "qaynaq" {
		t.Errorf("Expected static field '@service' = 'qaynaq', got '%v'", logData["@service"])
	}

	if logData["env"] != "test" {
		t.Errorf("Expected static field 'env' = 'test', got '%v'", logData["env"])
	}

	if logData["extra"] != "value" {
		t.Errorf("Expected dynamic field 'extra' = 'value', got '%v'", logData["extra"])
	}

	if logData["message"] != "test message" {
		t.Errorf("Expected message 'test message', got '%v'", logData["message"])
	}
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer

	log := NewFromZerolog("INFO", nil)
	log.zerologger = log.zerologger.Output(&buf)

	childLog := log.With(map[string]any{
		"request_id": "req-123",
		"user_id":    456,
	})

	childLog.Info("child log message")

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["request_id"] != "req-123" {
		t.Errorf("Expected request_id 'req-123', got '%v'", logData["request_id"])
	}

	if logData["user_id"] != float64(456) {
		t.Errorf("Expected user_id '456', got '%v'", logData["user_id"])
	}

	if logData["message"] != "child log message" {
		t.Errorf("Expected message 'child log message', got '%v'", logData["message"])
	}
}

func TestGetZerolog(t *testing.T) {
	log := NewFromZerolog("INFO", nil)

	zerologger := log.GetZerolog()
	if zerologger.GetLevel() != zerolog.InfoLevel {
		t.Error("Expected zerolog logger to be at INFO level")
	}
}

func TestGetBentoLogger(t *testing.T) {
	log := NewFromZerolog("INFO", nil)

	bentoLogger := log.GetBentoLogger()
	if bentoLogger != nil {
		t.Error("Expected nil bento logger for zerolog-based logger")
	}
}

func TestNewFromBentoWithStaticFields(t *testing.T) {
	staticFields := map[string]any{
		"component": "mysql-binlog",
		"instance":  "db-001",
	}

	log := NewFromBento(nil, staticFields)

	if log == nil {
		t.Fatal("Expected non-nil logger")
	}

	if !log.useBento {
		t.Error("Expected useBento to be true")
	}

	if len(log.staticFields) != len(staticFields) {
		t.Errorf("Expected %d static fields, got %d", len(staticFields), len(log.staticFields))
	}
}

func TestStaticFieldsPreservation(t *testing.T) {
	staticFields := map[string]any{
		"@service":    "qaynaq",
		"environment": "production",
		"version":     "1.0.0",
		"team":        "platform",
	}

	log := NewFromZerolog("INFO", staticFields)

	if len(log.staticFields) != len(staticFields) {
		t.Errorf("Expected %d static fields, got %d", len(staticFields), len(log.staticFields))
	}

	for key, expectedValue := range staticFields {
		if actualValue, ok := log.staticFields[key]; !ok {
			t.Errorf("Static field '%s' is missing", key)
		} else if actualValue != expectedValue {
			t.Errorf("Static field '%s': expected '%v', got '%v'", key, expectedValue, actualValue)
		}
	}
}

func TestParseLogfmt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "simple key-value",
			input: `level=info msg="test message"`,
			expected: map[string]string{
				"level": "info",
				"msg":   "test message",
			},
		},
		{
			name:  "multiple fields",
			input: `level=info msg="Connected to store" @service=qaynaq label=shopify`,
			expected: map[string]string{
				"level":    "info",
				"msg":      "Connected to store",
				"@service": "qaynaq",
				"label":    "shopify",
			},
		},
		{
			name:  "with path",
			input: `level=info msg="Resource type: products" path=root.input`,
			expected: map[string]string{
				"level": "info",
				"msg":   "Resource type: products",
				"path":  "root.input",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogfmt(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d fields, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("Expected field '%s' to be present", key)
				} else if actualValue != expectedValue {
					t.Errorf("Field '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestBentoLogWriter(t *testing.T) {
	var buf bytes.Buffer
	zerologger := zerolog.New(&buf).With().Timestamp().Logger()
	writer := NewBentoLogWriter(zerologger)

	bentoLog := `level=info msg="Connected to Shopify store" @service=qaynaq label=shopify_products`

	_, err := writer.Write([]byte(bentoLog))
	if err != nil {
		t.Fatalf("Failed to write log: %v", err)
	}

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "info" {
		t.Errorf("Expected level 'info', got '%v'", logData["level"])
	}

	if logData["message"] != "Connected to Shopify store" {
		t.Errorf("Expected message 'Connected to Shopify store', got '%v'", logData["message"])
	}

	if logData["@service"] != "qaynaq" {
		t.Errorf("Expected @service 'qaynaq', got '%v'", logData["@service"])
	}

	if logData["label"] != "shopify_products" {
		t.Errorf("Expected label 'shopify_products', got '%v'", logData["label"])
	}
}

func TestAddFieldsOddNumber(t *testing.T) {
	var buf bytes.Buffer

	log := NewFromZerolog("INFO", nil)
	log.zerologger = log.zerologger.Output(&buf)

	log.Info("test", "key1", "value1", "key2")

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got '%v'", logData["key1"])
	}

	if _, exists := logData["key2"]; exists {
		t.Error("Expected key2 to not be present (odd number of fields)")
	}
}

func TestLoggerWithComplexTypes(t *testing.T) {
	var buf bytes.Buffer

	log := NewFromZerolog("INFO", nil)
	log.zerologger = log.zerologger.Output(&buf)

	complexData := map[string]any{
		"nested": "data",
		"count":  42,
	}

	log.Info("complex test", "data", complexData, "list", []string{"a", "b", "c"})

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["message"] != "complex test" {
		t.Errorf("Expected message 'complex test', got '%v'", logData["message"])
	}

	if logData["data"] == nil {
		t.Error("Expected 'data' field to be present")
	}

	if logData["list"] == nil {
		t.Error("Expected 'list' field to be present")
	}
}

func TestNewSlogLogger(t *testing.T) {
	var buf bytes.Buffer

	slogLogger := NewSlogLogger("INFO", map[string]any{
		"@service":  "airturct",
		"stream_id": 123,
	})

	if slogLogger == nil {
		t.Fatal("Expected non-nil slog logger")
	}

	zerologLogger := zerolog.New(&buf).With().Timestamp().Logger()
	handler := &slogHandler{
		logger: zerologLogger,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
	slogLogger = slog.New(handler)

	slogLogger.Info("test message", "key", "value")

	var logData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logData["level"] != "info" {
		t.Errorf("Expected level 'info', got '%v'", logData["level"])
	}

	if logData["message"] != "test message" {
		t.Errorf("Expected message 'test message', got '%v'", logData["message"])
	}

	if logData["key"] != "value" {
		t.Errorf("Expected key 'value', got '%v'", logData["key"])
	}
}

func TestSlogHandlerLevels(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		shouldDebug bool
		shouldInfo  bool
		shouldWarn  bool
		shouldError bool
	}{
		{"DEBUG level", "DEBUG", true, true, true, true},
		{"INFO level", "INFO", false, true, true, true},
		{"WARN level", "WARN", false, false, true, true},
		{"ERROR level", "ERROR", false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			zerologLogger := zerolog.New(&buf).With().Timestamp().Logger()

			var zLevel zerolog.Level
			switch tt.level {
			case "DEBUG":
				zLevel = zerolog.DebugLevel
			case "INFO":
				zLevel = zerolog.InfoLevel
			case "WARN":
				zLevel = zerolog.WarnLevel
			case "ERROR":
				zLevel = zerolog.ErrorLevel
			}
			zerologLogger = zerologLogger.Level(zLevel)

			handler := &slogHandler{
				logger: zerologLogger,
				attrs:  []slog.Attr{},
				groups: []string{},
			}
			slogLogger := slog.New(handler)

			buf.Reset()
			slogLogger.Debug("debug message")
			debugLogged := buf.Len() > 0

			buf.Reset()
			slogLogger.Info("info message")
			infoLogged := buf.Len() > 0

			buf.Reset()
			slogLogger.Warn("warn message")
			warnLogged := buf.Len() > 0

			buf.Reset()
			slogLogger.Error("error message")
			errorLogged := buf.Len() > 0

			if debugLogged != tt.shouldDebug {
				t.Errorf("Debug logging mismatch: expected %v, got %v", tt.shouldDebug, debugLogged)
			}
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
