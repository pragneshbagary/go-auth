package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pragneshbagary/go-auth/pkg/models"
)

func TestMiddleware_Protect(t *testing.T) {
	// Create an in-memory auth instance for testing
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Register a test user
	_, err = auth.Register(RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Login to get tokens
	loginResult, err := auth.Login("testuser", "password123", nil)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	middleware := auth.Middleware()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + loginResult.AccessToken,
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:           "Invalid token format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that checks for user in context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, ok := GetUserFromContext(r.Context())
				if tt.expectUser {
					if !ok {
						t.Error("Expected user in context, but not found")
						return
					}
					if user.Username != "testuser" {
						t.Errorf("Expected username 'testuser', got '%s'", user.Username)
						return
					}
				} else {
					if ok {
						t.Error("Did not expect user in context, but found one")
						return
					}
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with middleware
			protectedHandler := middleware.Protect(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			protectedHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// For unauthorized requests, check error response format
			if tt.expectedStatus == http.StatusUnauthorized {
				var errorResponse HTTPErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				if errorResponse.Code != http.StatusUnauthorized {
					t.Errorf("Expected error code %d, got %d", http.StatusUnauthorized, errorResponse.Code)
				}
			}
		})
	}
}

func TestMiddleware_Optional(t *testing.T) {
	// Create an in-memory auth instance for testing
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Register a test user
	_, err = auth.Register(RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Login to get tokens
	loginResult, err := auth.Login("testuser", "password123", nil)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	middleware := auth.Middleware()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + loginResult.AccessToken,
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectUser:     false,
		},
		{
			name:           "Invalid token format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusOK,
			expectUser:     false,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusOK,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that checks for user in context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, ok := GetUserFromContext(r.Context())
				if tt.expectUser {
					if !ok {
						t.Error("Expected user in context, but not found")
						return
					}
					if user.Username != "testuser" {
						t.Errorf("Expected username 'testuser', got '%s'", user.Username)
						return
					}
				} else {
					if ok {
						t.Error("Did not expect user in context, but found one")
						return
					}
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with middleware
			optionalHandler := middleware.Optional(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/optional", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			optionalHandler.ServeHTTP(rr, req)

			// Check status code (should always be OK for optional middleware)
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	user := &models.UserProfile{
		ID:       "test-id",
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Test with user in context
	ctx := context.WithValue(context.Background(), UserKey, user)
	retrievedUser, ok := GetUserFromContext(ctx)
	if !ok {
		t.Error("Expected to find user in context")
	}
	if retrievedUser.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrievedUser.Username)
	}

	// Test without user in context
	emptyCtx := context.Background()
	_, ok = GetUserFromContext(emptyCtx)
	if ok {
		t.Error("Did not expect to find user in empty context")
	}
}

func TestGetClaimsFromContext(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":      "test-id",
		"username": "testuser",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	// Test with claims in context
	ctx := context.WithValue(context.Background(), ClaimsKey, claims)
	retrievedClaims, ok := GetClaimsFromContext(ctx)
	if !ok {
		t.Error("Expected to find claims in context")
	}
	if retrievedClaims["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrievedClaims["username"])
	}

	// Test without claims in context
	emptyCtx := context.Background()
	_, ok = GetClaimsFromContext(emptyCtx)
	if ok {
		t.Error("Did not expect to find claims in empty context")
	}
}

func TestAuthError(t *testing.T) {
	err := NewAuthError(ErrCodeInvalidToken, "Token is invalid")
	
	if err.Code != ErrCodeInvalidToken {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeInvalidToken, err.Code)
	}
	
	if err.Message != "Token is invalid" {
		t.Errorf("Expected message 'Token is invalid', got '%s'", err.Message)
	}
	
	expectedError := "INVALID_TOKEN: Token is invalid"
	if err.Error() != expectedError {
		t.Errorf("Expected error string '%s', got '%s'", expectedError, err.Error())
	}
}

func TestWriteErrorResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	err := NewAuthError(ErrCodeInvalidToken, "Token is invalid")
	
	WriteErrorResponse(rr, err, http.StatusUnauthorized)
	
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	
	var response HTTPErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response.Error != ErrCodeInvalidToken {
		t.Errorf("Expected error '%s', got '%s'", ErrCodeInvalidToken, response.Error)
	}
	
	if response.Message != "Token is invalid" {
		t.Errorf("Expected message 'Token is invalid', got '%s'", response.Message)
	}
	
	if response.Code != http.StatusUnauthorized {
		t.Errorf("Expected code %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuth_DirectMiddlewareMethods(t *testing.T) {
	// Create an in-memory auth instance for testing
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Test that Auth.Middleware() returns a Middleware instance
	middleware := auth.Middleware()
	if middleware == nil {
		t.Error("Expected middleware instance, got nil")
	}
	if middleware.auth != auth {
		t.Error("Middleware should reference the auth instance")
	}

	// Test direct access methods
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectHandler := auth.Protect(testHandler)
	if protectHandler == nil {
		t.Error("Expected protect handler, got nil")
	}

	optionalHandler := auth.Optional(testHandler)
	if optionalHandler == nil {
		t.Error("Expected optional handler, got nil")
	}
}