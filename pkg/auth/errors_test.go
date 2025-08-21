package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthErrorStructure(t *testing.T) {
	t.Run("NewAuthError", func(t *testing.T) {
		err := NewAuthError(ErrCodeInvalidCredentials, "Invalid username or password")
		
		if err.Code != ErrCodeInvalidCredentials {
			t.Errorf("Expected code %s, got %s", ErrCodeInvalidCredentials, err.Code)
		}
		if err.Message != "Invalid username or password" {
			t.Errorf("Expected message 'Invalid username or password', got %s", err.Message)
		}
		if err.Details != "" {
			t.Errorf("Expected empty details, got %s", err.Details)
		}
	})

	t.Run("NewAuthErrorWithDetails", func(t *testing.T) {
		err := NewAuthErrorWithDetails(ErrCodeWeakPassword, "Password too weak", "Must contain uppercase, lowercase, and numbers")
		
		if err.Code != ErrCodeWeakPassword {
			t.Errorf("Expected code %s, got %s", ErrCodeWeakPassword, err.Code)
		}
		if err.Message != "Password too weak" {
			t.Errorf("Expected message 'Password too weak', got %s", err.Message)
		}
		if err.Details != "Must contain uppercase, lowercase, and numbers" {
			t.Errorf("Expected details 'Must contain uppercase, lowercase, and numbers', got %s", err.Details)
		}
	})

	t.Run("Error method", func(t *testing.T) {
		err := NewAuthError(ErrCodeUserNotFound, "User not found")
		expected := "USER_NOT_FOUND: User not found"
		if err.Error() != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
		}

		errWithDetails := NewAuthErrorWithDetails(ErrCodeValidationError, "Validation failed", "Invalid email format")
		expectedWithDetails := "VALIDATION_ERROR: Validation failed (Invalid email format)"
		if errWithDetails.Error() != expectedWithDetails {
			t.Errorf("Expected error string '%s', got '%s'", expectedWithDetails, errWithDetails.Error())
		}
	})

	t.Run("Is method", func(t *testing.T) {
		err1 := NewAuthError(ErrCodeInvalidCredentials, "Invalid credentials")
		err2 := NewAuthError(ErrCodeInvalidCredentials, "Different message")
		err3 := NewAuthError(ErrCodeUserNotFound, "User not found")
		genericErr := errors.New("generic error")

		if !err1.Is(err2) {
			t.Error("Expected err1.Is(err2) to be true (same code)")
		}
		if err1.Is(err3) {
			t.Error("Expected err1.Is(err3) to be false (different code)")
		}
		if err1.Is(genericErr) {
			t.Error("Expected err1.Is(genericErr) to be false (different type)")
		}
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Run("ErrInvalidCredentials", func(t *testing.T) {
		err := ErrInvalidCredentials()
		if err.Code != ErrCodeInvalidCredentials {
			t.Errorf("Expected code %s, got %s", ErrCodeInvalidCredentials, err.Code)
		}
		if err.Message != "Invalid username or password" {
			t.Errorf("Expected standard message, got %s", err.Message)
		}
	})

	t.Run("ErrUserNotFound", func(t *testing.T) {
		err := ErrUserNotFound()
		if err.Code != ErrCodeUserNotFound {
			t.Errorf("Expected code %s, got %s", ErrCodeUserNotFound, err.Code)
		}
	})

	t.Run("ErrUserExists", func(t *testing.T) {
		err := ErrUserExists("email")
		if err.Code != ErrCodeUserExists {
			t.Errorf("Expected code %s, got %s", ErrCodeUserExists, err.Code)
		}
		if !strings.Contains(err.Details, "email") {
			t.Errorf("Expected details to contain 'email', got %s", err.Details)
		}
	})

	t.Run("ErrWeakPassword", func(t *testing.T) {
		requirements := "Must be at least 8 characters"
		err := ErrWeakPassword(requirements)
		if err.Code != ErrCodeWeakPassword {
			t.Errorf("Expected code %s, got %s", ErrCodeWeakPassword, err.Code)
		}
		if err.Details != requirements {
			t.Errorf("Expected details '%s', got %s", requirements, err.Details)
		}
	})
}

func TestWrapError(t *testing.T) {
	t.Run("WrapError with nil", func(t *testing.T) {
		result := WrapError(nil, ErrCodeDatabaseError, "Database failed")
		if result != nil {
			t.Error("Expected nil when wrapping nil error")
		}
	})

	t.Run("WrapError with AuthError", func(t *testing.T) {
		originalErr := NewAuthError(ErrCodeUserNotFound, "User not found")
		result := WrapError(originalErr, ErrCodeDatabaseError, "Database failed")
		
		// Should return the original AuthError unchanged
		if result != originalErr {
			t.Error("Expected original AuthError to be returned unchanged")
		}
	})

	t.Run("WrapError with generic error", func(t *testing.T) {
		genericErr := errors.New("some database error")
		result := WrapError(genericErr, ErrCodeDatabaseError, "Database failed")
		
		if result.Code != ErrCodeDatabaseError {
			t.Errorf("Expected code %s, got %s", ErrCodeDatabaseError, result.Code)
		}
		if result.Message != "Database failed" {
			t.Errorf("Expected message 'Database failed', got %s", result.Message)
		}
		if result.Details != "Internal error occurred" {
			t.Errorf("Expected generic details, got %s", result.Details)
		}
	})

	t.Run("WrapDatabaseError", func(t *testing.T) {
		genericErr := errors.New("connection failed")
		result := WrapDatabaseError(genericErr)
		
		if result.Code != ErrCodeDatabaseError {
			t.Errorf("Expected code %s, got %s", ErrCodeDatabaseError, result.Code)
		}
		if result.Message != "Database operation failed" {
			t.Errorf("Expected standard database message, got %s", result.Message)
		}
	})
}

func TestHTTPErrorHandling(t *testing.T) {
	t.Run("WriteErrorResponse with AuthError", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := NewAuthErrorWithDetails(ErrCodeInvalidCredentials, "Invalid credentials", "Username or password incorrect")
		
		WriteErrorResponse(w, err, http.StatusUnauthorized)
		
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
		
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
		
		body := w.Body.String()
		if !strings.Contains(body, ErrCodeInvalidCredentials) {
			t.Errorf("Expected response to contain error code, got: %s", body)
		}
		if !strings.Contains(body, "Invalid credentials") {
			t.Errorf("Expected response to contain error message, got: %s", body)
		}
	})

	t.Run("WriteErrorResponse with generic error", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("some internal error")
		
		WriteErrorResponse(w, err, http.StatusInternalServerError)
		
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
		
		body := w.Body.String()
		if !strings.Contains(body, ErrCodeInternalError) {
			t.Errorf("Expected response to contain generic error code, got: %s", body)
		}
		if !strings.Contains(body, "An internal error occurred") {
			t.Errorf("Expected response to contain generic message, got: %s", body)
		}
	})

	t.Run("WriteJSONError", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := NewAuthError(ErrCodeUserNotFound, "User not found")
		
		WriteJSONError(w, err)
		
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestGetHTTPStatusFromError(t *testing.T) {
	testCases := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{"InvalidCredentials", NewAuthError(ErrCodeInvalidCredentials, "Invalid"), http.StatusUnauthorized},
		{"InvalidToken", NewAuthError(ErrCodeInvalidToken, "Invalid"), http.StatusUnauthorized},
		{"TokenExpired", NewAuthError(ErrCodeTokenExpired, "Expired"), http.StatusUnauthorized},
		{"TokenRevoked", NewAuthError(ErrCodeTokenRevoked, "Revoked"), http.StatusUnauthorized},
		{"MissingToken", NewAuthError(ErrCodeMissingToken, "Missing"), http.StatusUnauthorized},
		{"UserNotFound", NewAuthError(ErrCodeUserNotFound, "Not found"), http.StatusNotFound},
		{"UserExists", NewAuthError(ErrCodeUserExists, "Exists"), http.StatusConflict},
		{"UserInactive", NewAuthError(ErrCodeUserInactive, "Inactive"), http.StatusForbidden},
		{"PermissionDenied", NewAuthError(ErrCodePermissionDenied, "Denied"), http.StatusForbidden},
		{"WeakPassword", NewAuthError(ErrCodeWeakPassword, "Weak"), http.StatusBadRequest},
		{"ValidationError", NewAuthError(ErrCodeValidationError, "Invalid"), http.StatusBadRequest},
		{"RateLimitExceeded", NewAuthError(ErrCodeRateLimitExceeded, "Rate limit"), http.StatusTooManyRequests},
		{"DatabaseError", NewAuthError(ErrCodeDatabaseError, "DB error"), http.StatusInternalServerError},
		{"InternalError", NewAuthError(ErrCodeInternalError, "Internal"), http.StatusInternalServerError},
		{"GenericError", errors.New("generic"), http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status := getHTTPStatusFromError(tc.err)
			if status != tc.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tc.expectedStatus, tc.name, status)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are defined and non-empty
	constants := []string{
		ErrCodeInvalidCredentials,
		ErrCodeInvalidToken,
		ErrCodeTokenExpired,
		ErrCodeTokenRevoked,
		ErrCodeMissingToken,
		ErrCodeMalformedToken,
		ErrCodeUserExists,
		ErrCodeUserNotFound,
		ErrCodeUserInactive,
		ErrCodeUserDeleted,
		ErrCodeWeakPassword,
		ErrCodePasswordMismatch,
		ErrCodeInvalidResetToken,
		ErrCodeResetTokenExpired,
		ErrCodeDatabaseError,
		ErrCodeStorageError,
		ErrCodeConnectionError,
		ErrCodeMigrationError,
		ErrCodeConfigError,
		ErrCodeInvalidConfig,
		ErrCodeMissingConfig,
		ErrCodeInternalError,
		ErrCodeValidationError,
		ErrCodePermissionDenied,
		ErrCodeRateLimitExceeded,
	}

	for _, constant := range constants {
		if constant == "" {
			t.Errorf("Error constant is empty")
		}
		if !strings.Contains(constant, "_") {
			t.Errorf("Error constant %s should follow UPPER_CASE format", constant)
		}
	}
}