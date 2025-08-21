package auth

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	t.Run("NewLogger", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(LogLevelInfo, &buf)

		if logger == nil {
			t.Fatal("Expected logger to be created")
		}

		if logger.level != LogLevelInfo {
			t.Errorf("Expected log level to be Info, got %v", logger.level)
		}
	})

	t.Run("LogLevels", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(LogLevelWarn, &buf)

		// Debug and Info should not be logged
		logger.Debug("debug message")
		logger.Info("info message")

		// Warn and Error should be logged
		logger.Warn("warn message")
		logger.Error("error message")

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		if len(lines) != 2 {
			t.Errorf("Expected 2 log lines, got %d", len(lines))
		}

		// Check that warn message was logged
		var warnEntry LogEntry
		if err := json.Unmarshal([]byte(lines[0]), &warnEntry); err != nil {
			t.Fatalf("Failed to parse warn log entry: %v", err)
		}
		if warnEntry.Level != "WARN" || warnEntry.Message != "warn message" {
			t.Errorf("Unexpected warn log entry: %+v", warnEntry)
		}

		// Check that error message was logged
		var errorEntry LogEntry
		if err := json.Unmarshal([]byte(lines[1]), &errorEntry); err != nil {
			t.Fatalf("Failed to parse error log entry: %v", err)
		}
		if errorEntry.Level != "ERROR" || errorEntry.Message != "error message" {
			t.Errorf("Unexpected error log entry: %+v", errorEntry)
		}
	})

	t.Run("LogWithFields", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(LogLevelInfo, &buf)

		logger.Info("test message", map[string]interface{}{
			"user_id":  "123",
			"username": "testuser",
			"event":    "login",
			"duration": time.Second,
		})

		output := buf.String()
		var entry LogEntry
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
			t.Fatalf("Failed to parse log entry: %v", err)
		}

		if entry.UserID != "123" {
			t.Errorf("Expected user_id to be '123', got '%s'", entry.UserID)
		}
		if entry.Username != "testuser" {
			t.Errorf("Expected username to be 'testuser', got '%s'", entry.Username)
		}
		if entry.Event != "login" {
			t.Errorf("Expected event to be 'login', got '%s'", entry.Event)
		}
		if entry.Duration != "1s" {
			t.Errorf("Expected duration to be '1s', got '%s'", entry.Duration)
		}
	})

	t.Run("FieldLogger", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(LogLevelInfo, &buf)

		fieldLogger := logger.WithFields(map[string]interface{}{
			"component": "auth",
			"user_id":   "123",
		})

		fieldLogger.Info("test message", map[string]interface{}{
			"action": "login",
		})

		output := buf.String()
		var entry LogEntry
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
			t.Fatalf("Failed to parse log entry: %v", err)
		}

		if entry.UserID != "123" {
			t.Errorf("Expected user_id to be '123', got '%s'", entry.UserID)
		}
		if entry.Fields["component"] != "auth" {
			t.Errorf("Expected component field to be 'auth', got '%v'", entry.Fields["component"])
		}
		if entry.Fields["action"] != "login" {
			t.Errorf("Expected action field to be 'login', got '%v'", entry.Fields["action"])
		}
	})
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"DEBUG", LogLevelDebug},
		{"info", LogLevelInfo},
		{"INFO", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"WARN", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"error", LogLevelError},
		{"ERROR", LogLevelError},
		{"invalid", LogLevelInfo}, // Default to info
	}

	for _, test := range tests {
		result := ParseLogLevel(test.input)
		if result != test.expected {
			t.Errorf("ParseLogLevel(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestAuthEventLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelInfo, &buf)
	eventLogger := NewAuthEventLogger(logger)

	t.Run("LogRegistration", func(t *testing.T) {
		buf.Reset()
		eventLogger.LogRegistration("user123", "testuser", "test@example.com", "192.168.1.1", "Mozilla/5.0", true, nil)

		output := buf.String()
		var entry LogEntry
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
			t.Fatalf("Failed to parse log entry: %v", err)
		}

		if entry.Level != "INFO" {
			t.Errorf("Expected level to be INFO, got %s", entry.Level)
		}
		if entry.Event != "user_registration" {
			t.Errorf("Expected event to be 'user_registration', got %s", entry.Event)
		}
		if entry.UserID != "user123" {
			t.Errorf("Expected user_id to be 'user123', got %s", entry.UserID)
		}
		if entry.IP != "192.168.1.1" {
			t.Errorf("Expected IP to be '192.168.1.1', got %s", entry.IP)
		}
	})

	t.Run("LogLogin", func(t *testing.T) {
		buf.Reset()
		eventLogger.LogLogin("user123", "testuser", "192.168.1.1", "Mozilla/5.0", true, time.Millisecond*500, nil)

		output := buf.String()
		var entry LogEntry
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
			t.Fatalf("Failed to parse log entry: %v", err)
		}

		if entry.Event != "user_login" {
			t.Errorf("Expected event to be 'user_login', got %s", entry.Event)
		}
		if entry.Duration != "500ms" {
			t.Errorf("Expected duration to be '500ms', got %s", entry.Duration)
		}
	})

	t.Run("LogTokenRefresh", func(t *testing.T) {
		buf.Reset()
		eventLogger.LogTokenRefresh("user123", "192.168.1.1", "Mozilla/5.0", true, time.Millisecond*100, nil)

		output := buf.String()
		var entry LogEntry
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
			t.Fatalf("Failed to parse log entry: %v", err)
		}

		if entry.Event != "token_refresh" {
			t.Errorf("Expected event to be 'token_refresh', got %s", entry.Event)
		}
	})

	t.Run("LogPasswordChange", func(t *testing.T) {
		buf.Reset()
		eventLogger.LogPasswordChange("user123", "testuser", "192.168.1.1", "Mozilla/5.0", true, nil)

		output := buf.String()
		var entry LogEntry
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
			t.Fatalf("Failed to parse log entry: %v", err)
		}

		if entry.Event != "password_change" {
			t.Errorf("Expected event to be 'password_change', got %s", entry.Event)
		}
	})
}