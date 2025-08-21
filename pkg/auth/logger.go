package auth

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// Logger provides structured logging for authentication operations
type Logger struct {
	level  LogLevel
	output io.Writer
	logger *log.Logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Component string                 `json:"component"`
	UserID    string                 `json:"user_id,omitempty"`
	Username  string                 `json:"username,omitempty"`
	Event     string                 `json:"event,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Duration  string                 `json:"duration,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new logger with the specified level and output
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	return &Logger{
		level:  level,
		output: output,
		logger: log.New(output, "", 0), // No prefix, we'll handle formatting
	}
}

// NewDefaultLogger creates a logger with default settings (INFO level, stdout)
func NewDefaultLogger() *Logger {
	return NewLogger(LogLevelInfo, os.Stdout)
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// IsEnabled checks if a log level is enabled
func (l *Logger) IsEnabled(level LogLevel) bool {
	return level >= l.level
}

// log writes a structured log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.IsEnabled(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Component: "go-auth",
		Fields:    fields,
	}

	// Extract common fields from the fields map
	if fields != nil {
		if userID, ok := fields["user_id"].(string); ok {
			entry.UserID = userID
			delete(fields, "user_id")
		}
		if username, ok := fields["username"].(string); ok {
			entry.Username = username
			delete(fields, "username")
		}
		if event, ok := fields["event"].(string); ok {
			entry.Event = event
			delete(fields, "event")
		}
		if ip, ok := fields["ip"].(string); ok {
			entry.IP = ip
			delete(fields, "ip")
		}
		if userAgent, ok := fields["user_agent"].(string); ok {
			entry.UserAgent = userAgent
			delete(fields, "user_agent")
		}
		if duration, ok := fields["duration"].(time.Duration); ok {
			entry.Duration = duration.String()
			delete(fields, "duration")
		}
		if err, ok := fields["error"].(error); ok {
			entry.Error = err.Error()
			delete(fields, "error")
		}
		if errStr, ok := fields["error"].(string); ok {
			entry.Error = errStr
			delete(fields, "error")
		}
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.logger.Printf("[%s] %s: %s", level.String(), message, err.Error())
		return
	}

	l.logger.Println(string(jsonData))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelDebug, message, f)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelInfo, message, f)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelWarn, message, f)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelError, message, f)
}

// WithFields creates a new logger with default fields
func (l *Logger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// FieldLogger is a logger with pre-set fields
type FieldLogger struct {
	logger *Logger
	fields map[string]interface{}
}

// Debug logs a debug message with pre-set fields
func (fl *FieldLogger) Debug(message string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.Debug(message, fields)
}

// Info logs an info message with pre-set fields
func (fl *FieldLogger) Info(message string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.Info(message, fields)
}

// Warn logs a warning message with pre-set fields
func (fl *FieldLogger) Warn(message string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.Warn(message, fields)
}

// Error logs an error message with pre-set fields
func (fl *FieldLogger) Error(message string, additionalFields ...map[string]interface{}) {
	fields := fl.mergeFields(additionalFields...)
	fl.logger.Error(message, fields)
}

// mergeFields merges the pre-set fields with additional fields
func (fl *FieldLogger) mergeFields(additionalFields ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	// Copy pre-set fields
	for k, v := range fl.fields {
		merged[k] = v
	}
	
	// Add additional fields (they override pre-set fields)
	if len(additionalFields) > 0 && additionalFields[0] != nil {
		for k, v := range additionalFields[0] {
			merged[k] = v
		}
	}
	
	return merged
}

// AuthEventLogger provides specialized logging for authentication events
type AuthEventLogger struct {
	logger *Logger
}

// NewAuthEventLogger creates a new authentication event logger
func NewAuthEventLogger(logger *Logger) *AuthEventLogger {
	return &AuthEventLogger{
		logger: logger,
	}
}

// LogRegistration logs a user registration event
func (ael *AuthEventLogger) LogRegistration(userID, username, email, ip, userAgent string, success bool, err error) {
	fields := map[string]interface{}{
		"event":      "user_registration",
		"user_id":    userID,
		"username":   username,
		"email":      email,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
	}

	if err != nil {
		fields["error"] = err
		ael.logger.Error("User registration failed", fields)
	} else {
		ael.logger.Info("User registered successfully", fields)
	}
}

// LogLogin logs a user login event
func (ael *AuthEventLogger) LogLogin(userID, username, ip, userAgent string, success bool, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"event":      "user_login",
		"user_id":    userID,
		"username":   username,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
		"duration":   duration,
	}

	if err != nil {
		fields["error"] = err
		ael.logger.Warn("User login failed", fields)
	} else {
		ael.logger.Info("User logged in successfully", fields)
	}
}

// LogTokenRefresh logs a token refresh event
func (ael *AuthEventLogger) LogTokenRefresh(userID, ip, userAgent string, success bool, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"event":      "token_refresh",
		"user_id":    userID,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
		"duration":   duration,
	}

	if err != nil {
		fields["error"] = err
		ael.logger.Warn("Token refresh failed", fields)
	} else {
		ael.logger.Info("Token refreshed successfully", fields)
	}
}

// LogTokenRevocation logs a token revocation event
func (ael *AuthEventLogger) LogTokenRevocation(userID, tokenID, ip, userAgent string, success bool, err error) {
	fields := map[string]interface{}{
		"event":      "token_revocation",
		"user_id":    userID,
		"token_id":   tokenID,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
	}

	if err != nil {
		fields["error"] = err
		ael.logger.Error("Token revocation failed", fields)
	} else {
		ael.logger.Info("Token revoked successfully", fields)
	}
}

// LogPasswordChange logs a password change event
func (ael *AuthEventLogger) LogPasswordChange(userID, username, ip, userAgent string, success bool, err error) {
	fields := map[string]interface{}{
		"event":      "password_change",
		"user_id":    userID,
		"username":   username,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
	}

	if err != nil {
		fields["error"] = err
		ael.logger.Error("Password change failed", fields)
	} else {
		ael.logger.Info("Password changed successfully", fields)
	}
}

// LogPasswordReset logs a password reset event
func (ael *AuthEventLogger) LogPasswordReset(userID, username, ip, userAgent string, success bool, err error) {
	fields := map[string]interface{}{
		"event":      "password_reset",
		"user_id":    userID,
		"username":   username,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
	}

	if err != nil {
		fields["error"] = err
		ael.logger.Error("Password reset failed", fields)
	} else {
		ael.logger.Info("Password reset successfully", fields)
	}
}

// LogTokenValidation logs a token validation event
func (ael *AuthEventLogger) LogTokenValidation(userID, ip, userAgent string, success bool, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"event":      "token_validation",
		"user_id":    userID,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
		"duration":   duration,
	}

	if err != nil {
		fields["error"] = err
		ael.logger.Debug("Token validation failed", fields)
	} else {
		ael.logger.Debug("Token validated successfully", fields)
	}
}