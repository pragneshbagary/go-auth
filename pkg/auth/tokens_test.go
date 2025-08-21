package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/internal/jwtutils"
	"github.com/pragneshbagary/go-auth/internal/storage/memory"
	"github.com/pragneshbagary/go-auth/pkg/models"
)

func setupTokensTest(t *testing.T) (*Tokens, *models.User) {
	// Create in-memory storage
	storageImpl := memory.NewInMemoryStorage()
	
	// Create JWT manager
	jwtManager := jwtutils.NewJWTManager(jwtutils.JWTConfig{
		AccessSecret:    []byte("test-access-secret"),
		RefreshSecret:   []byte("test-refresh-secret"),
		Issuer:          "test-issuer",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		SigningMethod:   HS256,
	})

	// Create test user
	testUser := models.User{
		ID:           "test-user-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	
	err := storageImpl.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tokens := &Tokens{
		jwtManager: jwtManager,
		storage:    storageImpl,
	}

	return tokens, &testUser
}

func TestTokens_Refresh(t *testing.T) {
	tokens, user := setupTokensTest(t)

	// Generate initial refresh token
	refreshToken, err := tokens.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Test successful refresh
	result, err := tokens.Refresh(refreshToken)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
	if refreshToken == result.RefreshToken {
		t.Error("Expected refresh token to be rotated")
	}

	// Verify new access token is valid
	claims, err := tokens.jwtManager.ValidateAccessToken(result.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error validating access token, got: %v", err)
	}
	if claims["sub"] != user.ID {
		t.Errorf("Expected user ID %s, got %v", user.ID, claims["sub"])
	}
	if claims["username"] != user.Username {
		t.Errorf("Expected username %s, got %v", user.Username, claims["username"])
	}
	if claims["email"] != user.Email {
		t.Errorf("Expected email %s, got %v", user.Email, claims["email"])
	}

	// Test that old refresh token is now blacklisted (should fail on second use)
	_, err = tokens.Refresh(refreshToken)
	if err == nil {
		t.Error("Expected error when using revoked refresh token")
	}
}

func TestTokens_Refresh_InvalidToken(t *testing.T) {
	tokens, _ := setupTokensTest(t)

	// Test with invalid token
	_, err := tokens.Refresh("invalid-token")
	if err == nil {
		t.Error("Expected error with invalid token")
	}
}

func TestTokens_Revoke(t *testing.T) {
	tokens, user := setupTokensTest(t)

	// Generate access token
	accessToken, err := tokens.jwtManager.GenerateAccessToken(user.ID, map[string]interface{}{
		"username": user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Verify token is initially valid
	_, err = tokens.Validate(accessToken)
	if err != nil {
		t.Fatalf("Expected token to be valid initially: %v", err)
	}

	// Revoke the token
	err = tokens.Revoke(accessToken)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Verify token is now invalid
	_, err = tokens.Validate(accessToken)
	if err == nil {
		t.Error("Expected error when validating revoked token")
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Errorf("Expected error to mention 'revoked', got: %v", err)
	}
}

func TestTokens_Validate(t *testing.T) {
	tokens, user := setupTokensTest(t)

	// Generate access token
	accessToken, err := tokens.jwtManager.GenerateAccessToken(user.ID, map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	})
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Test successful validation
	validatedUser, err := tokens.Validate(accessToken)
	if err != nil {
		t.Fatalf("Expected no error validating token: %v", err)
	}
	if validatedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, validatedUser.ID)
	}
	if validatedUser.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, validatedUser.Username)
	}
	if validatedUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, validatedUser.Email)
	}
}

func TestTokens_IsValid(t *testing.T) {
	tokens, user := setupTokensTest(t)

	// Generate access token
	accessToken, err := tokens.jwtManager.GenerateAccessToken(user.ID, map[string]interface{}{
		"username": user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Test valid token
	if !tokens.IsValid(accessToken) {
		t.Error("Expected token to be valid")
	}

	// Test invalid token
	if tokens.IsValid("invalid-token") {
		t.Error("Expected invalid token to be invalid")
	}

	// Test revoked token
	err = tokens.Revoke(accessToken)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}
	if tokens.IsValid(accessToken) {
		t.Error("Expected revoked token to be invalid")
	}
}

func TestTokens_ValidateBatch(t *testing.T) {
	tokens, user := setupTokensTest(t)

	// Generate tokens
	validToken, err := tokens.jwtManager.GenerateAccessToken(user.ID, map[string]interface{}{
		"username": user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to generate valid token: %v", err)
	}

	revokedToken, err := tokens.jwtManager.GenerateAccessToken(user.ID, map[string]interface{}{
		"username": user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to generate token to revoke: %v", err)
	}
	err = tokens.Revoke(revokedToken)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Test batch validation
	testTokens := []string{validToken, "invalid-token", revokedToken}
	results := tokens.ValidateBatch(testTokens)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	
	// Valid token
	if !results[0].Valid {
		t.Error("Expected first token to be valid")
	}
	if results[0].User == nil {
		t.Error("Expected user data for valid token")
	}
	if results[0].Claims == nil {
		t.Error("Expected claims for valid token")
	}
	if results[0].Error != "" {
		t.Errorf("Expected no error for valid token, got: %s", results[0].Error)
	}

	// Invalid token
	if results[1].Valid {
		t.Error("Expected second token to be invalid")
	}
	if results[1].User != nil {
		t.Error("Expected no user data for invalid token")
	}
	if results[1].Claims != nil {
		t.Error("Expected no claims for invalid token")
	}
	if results[1].Error == "" {
		t.Error("Expected error for invalid token")
	}

	// Revoked token
	if results[2].Valid {
		t.Error("Expected third token to be invalid")
	}
	if results[2].User != nil {
		t.Error("Expected no user data for revoked token")
	}
	if results[2].Claims != nil {
		t.Error("Expected no claims for revoked token")
	}
	if !strings.Contains(results[2].Error, "revoked") {
		t.Errorf("Expected error to mention 'revoked', got: %s", results[2].Error)
	}
}

func TestTokens_GetSessionInfo(t *testing.T) {
	tokens, user := setupTokensTest(t)

	// Generate access token
	accessToken, err := tokens.jwtManager.GenerateAccessToken(user.ID, map[string]interface{}{
		"username": user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Get session info
	sessionInfo, err := tokens.GetSessionInfo(accessToken)
	if err != nil {
		t.Fatalf("Failed to get session info: %v", err)
	}
	
	if sessionInfo.TokenID == "" {
		t.Error("Expected non-empty token ID")
	}
	if sessionInfo.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, sessionInfo.UserID)
	}
	if sessionInfo.TokenType != "access" {
		t.Errorf("Expected token type 'access', got %s", sessionInfo.TokenType)
	}
	if sessionInfo.IssuedAt.IsZero() {
		t.Error("Expected non-zero issued at time")
	}
	if sessionInfo.ExpiresAt.IsZero() {
		t.Error("Expected non-zero expires at time")
	}
	if !sessionInfo.ExpiresAt.After(sessionInfo.IssuedAt) {
		t.Error("Expected expires at to be after issued at")
	}
}

func TestTokens_CleanupExpired(t *testing.T) {
	tokens, _ := setupTokensTest(t)

	// Test cleanup (should not error)
	err := tokens.CleanupExpired()
	if err != nil {
		t.Errorf("Expected no error from cleanup, got: %v", err)
	}
}

func TestTokens_ListActiveSessions(t *testing.T) {
	tokens, user := setupTokensTest(t)

	// Test listing active sessions (placeholder implementation)
	sessions, err := tokens.ListActiveSessions(user.ID)
	if err != nil {
		t.Fatalf("Failed to list active sessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Error("Expected empty sessions list from placeholder implementation")
	}
}

// Integration test with Auth component
func TestAuth_Tokens_Integration(t *testing.T) {
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Register a user
	user, err := auth.Register(RegisterRequest{
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

	// Test token refresh through Auth component
	refreshResult, err := auth.RefreshToken(loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}
	if refreshResult.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if refreshResult.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}

	// Test token validation through Tokens component
	tokens := auth.Tokens()
	validatedUser, err := tokens.Validate(refreshResult.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	if validatedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, validatedUser.ID)
	}

	// Test token revocation
	err = tokens.Revoke(refreshResult.AccessToken)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Verify token is now invalid
	if tokens.IsValid(refreshResult.AccessToken) {
		t.Error("Expected revoked token to be invalid")
	}
}