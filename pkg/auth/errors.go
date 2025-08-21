package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AuthError represents structured authentication errors with error codes and context.
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error codes for common authentication errors
const (
	// Authentication errors
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken      = "INVALID_TOKEN"
	ErrCodeTokenExpired      = "TOKEN_EXPIRED"
	ErrCodeTokenRevoked      = "TOKEN_REVOKED"
	ErrCodeMissingToken      = "MISSING_TOKEN"
	ErrCodeMalformedToken    = "MALFORMED_TOKEN"
	
	// User management errors
	ErrCodeUserExists        = "USER_EXISTS"
	ErrCodeUserNotFound      = "USER_NOT_FOUND"
	ErrCodeUserInactive      = "USER_INACTIVE"
	ErrCodeUserDeleted       = "USER_DELETED"
	
	// Password errors
	ErrCodeWeakPassword      = "WEAK_PASSWORD"
	ErrCodePasswordMismatch  = "PASSWORD_MISMATCH"
	ErrCodeInvalidResetToken = "INVALID_RESET_TOKEN"
	ErrCodeResetTokenExpired = "RESET_TOKEN_EXPIRED"
	
	// Database and storage errors
	ErrCodeDatabaseError     = "DATABASE_ERROR"
	ErrCodeStorageError      = "STORAGE_ERROR"
	ErrCodeConnectionError   = "CONNECTION_ERROR"
	ErrCodeMigrationError    = "MIGRATION_ERROR"
	
	// Configuration errors
	ErrCodeConfigError       = "CONFIG_ERROR"
	ErrCodeInvalidConfig     = "INVALID_CONFIG"
	ErrCodeMissingConfig     = "MISSING_CONFIG"
	
	// General errors
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeValidationError   = "VALIDATION_ERROR"
	ErrCodePermissionDenied  = "PERMISSION_DENIED"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// NewAuthError creates a new AuthError with the specified code and message.
func NewAuthError(code, message string) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
	}
}

// NewAuthErrorWithDetails creates a new AuthError with code, message, and additional details.
func NewAuthErrorWithDetails(code, message, details string) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Is implements error comparison for Go 1.13+ error handling.
func (e *AuthError) Is(target error) bool {
	if t, ok := target.(*AuthError); ok {
		return e.Code == t.Code
	}
	return false
}

// HTTPErrorResponse represents an HTTP error response structure.
type HTTPErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// WriteErrorResponse writes a structured error response to the HTTP response writer.
// It handles AuthError types specially and provides generic handling for other errors.
func WriteErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var response HTTPErrorResponse

	if authErr, ok := err.(*AuthError); ok {
		response = HTTPErrorResponse{
			Error:   authErr.Code,
			Message: authErr.Message,
			Code:    statusCode,
			Details: authErr.Details,
		}
	} else {
		// For non-AuthError types, provide a generic response without exposing internal details
		response = HTTPErrorResponse{
			Error:   ErrCodeInternalError,
			Message: "An internal error occurred",
			Code:    statusCode,
		}
	}

	json.NewEncoder(w).Encode(response)
}

// WriteJSONError is a convenience function that writes an error response with appropriate HTTP status codes.
func WriteJSONError(w http.ResponseWriter, err error) {
	statusCode := getHTTPStatusFromError(err)
	WriteErrorResponse(w, err, statusCode)
}

// getHTTPStatusFromError maps AuthError codes to appropriate HTTP status codes.
func getHTTPStatusFromError(err error) int {
	if authErr, ok := err.(*AuthError); ok {
		switch authErr.Code {
		case ErrCodeInvalidCredentials, ErrCodeInvalidToken, ErrCodeTokenExpired, 
			 ErrCodeTokenRevoked, ErrCodeMissingToken, ErrCodeMalformedToken:
			return http.StatusUnauthorized
		case ErrCodeUserNotFound, ErrCodeInvalidResetToken:
			return http.StatusNotFound
		case ErrCodeUserExists:
			return http.StatusConflict
		case ErrCodeUserInactive, ErrCodeUserDeleted, ErrCodePermissionDenied:
			return http.StatusForbidden
		case ErrCodeWeakPassword, ErrCodePasswordMismatch, ErrCodeValidationError,
			 ErrCodeInvalidConfig, ErrCodeMissingConfig:
			return http.StatusBadRequest
		case ErrCodeRateLimitExceeded:
			return http.StatusTooManyRequests
		case ErrCodeDatabaseError, ErrCodeStorageError, ErrCodeConnectionError,
			 ErrCodeMigrationError, ErrCodeInternalError:
			return http.StatusInternalServerError
		case ErrCodeConfigError:
			return http.StatusInternalServerError
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// Common error constructors for frequently used errors

// ErrInvalidCredentials creates a standard invalid credentials error.
func ErrInvalidCredentials() *AuthError {
	return NewAuthError(ErrCodeInvalidCredentials, "Invalid username or password")
}

// ErrUserNotFound creates a standard user not found error.
func ErrUserNotFound() *AuthError {
	return NewAuthError(ErrCodeUserNotFound, "User not found")
}

// ErrUserExists creates a standard user already exists error.
func ErrUserExists(identifier string) *AuthError {
	return NewAuthErrorWithDetails(ErrCodeUserExists, "User already exists", 
		fmt.Sprintf("A user with this %s already exists", identifier))
}

// ErrInvalidToken creates a standard invalid token error.
func ErrInvalidToken() *AuthError {
	return NewAuthError(ErrCodeInvalidToken, "Invalid or expired token")
}

// ErrTokenRevoked creates a standard token revoked error.
func ErrTokenRevoked() *AuthError {
	return NewAuthError(ErrCodeTokenRevoked, "Token has been revoked")
}

// ErrMissingToken creates a standard missing token error.
func ErrMissingToken() *AuthError {
	return NewAuthError(ErrCodeMissingToken, "Authorization token is required")
}

// ErrUserInactive creates a standard user inactive error.
func ErrUserInactive() *AuthError {
	return NewAuthError(ErrCodeUserInactive, "User account is inactive")
}

// ErrWeakPassword creates a standard weak password error.
func ErrWeakPassword(requirements string) *AuthError {
	return NewAuthErrorWithDetails(ErrCodeWeakPassword, "Password does not meet requirements", requirements)
}

// ErrDatabaseError creates a database error without exposing internal details.
func ErrDatabaseError() *AuthError {
	return NewAuthError(ErrCodeDatabaseError, "Database operation failed")
}

// ErrConfigError creates a configuration error.
func ErrConfigError(field string) *AuthError {
	return NewAuthErrorWithDetails(ErrCodeConfigError, "Configuration error", 
		fmt.Sprintf("Invalid or missing configuration for: %s", field))
}

// ErrValidationError creates a validation error.
func ErrValidationError(field string) *AuthError {
	return NewAuthErrorWithDetails(ErrCodeValidationError, "Validation failed", 
		fmt.Sprintf("Invalid value for field: %s", field))
}

// WrapError wraps a generic error as an AuthError with the specified code.
// This is useful for converting storage or other errors to structured AuthErrors.
func WrapError(err error, code, message string) *AuthError {
	if err == nil {
		return nil
	}
	
	// If it's already an AuthError, return it as-is
	if authErr, ok := err.(*AuthError); ok {
		return authErr
	}
	
	// Create a new AuthError with the original error as details (but don't expose sensitive info)
	return NewAuthErrorWithDetails(code, message, "Internal error occurred")
}

// WrapDatabaseError wraps database errors as AuthErrors without exposing sensitive details.
func WrapDatabaseError(err error) *AuthError {
	if err == nil {
		return nil
	}
	return WrapError(err, ErrCodeDatabaseError, "Database operation failed")
}

// WrapStorageError wraps storage errors as AuthErrors without exposing sensitive details.
func WrapStorageError(err error) *AuthError {
	if err == nil {
		return nil
	}
	return WrapError(err, ErrCodeStorageError, "Storage operation failed")
}