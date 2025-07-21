package jwtutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestManager creates a standard JWTManager for testing.
func setupTestManager(t *testing.T) TokenManager {
	cfg := JWTConfig{
		AccessSecret:    []byte("test-access-secret"),
		RefreshSecret:   []byte("test-refresh-secret"),
		Issuer:          "test-issuer",
		AccessTokenTTL:  5 * time.Minute,
		RefreshTokenTTL: 1 * time.Hour,
		SigningMethod:   HS256,
	}
	return NewJWTManager(cfg)
}

// TestGenerateAndValidateTokens covers the happy path for token creation and validation.
func TestGenerateAndValidateTokens(t *testing.T) {
	tm := setupTestManager(t)
	userID := "user-123"
	customClaims := map[string]any{"role": "admin", "scope": "read:all"}

	// 1. Generate Access Token
	accessToken, err := tm.GenerateAccessToken(userID, customClaims)
	require.NoError(t, err, "Should not fail to generate access token")
	require.NotEmpty(t, accessToken, "Access token should not be empty")

	// 2. Validate Access Token
	claims, err := tm.ValidateAccessToken(accessToken)
	require.NoError(t, err, "Should not fail to validate access token")
	assert.Equal(t, userID, claims["sub"], "Subject should match user ID")
	assert.Equal(t, "admin", claims["role"], "Custom claim 'role' should be present")
	assert.Equal(t, "read:all", claims["scope"], "Custom claim 'scope' should be present")

	// 3. Generate Refresh Token
	refreshToken, err := tm.GenerateRefreshToken(userID)
	require.NoError(t, err, "Should not fail to generate refresh token")
	require.NotEmpty(t, refreshToken, "Refresh token should not be empty")

	// 4. Validate Refresh Token
	refreshClaims, err := tm.ValidateRefreshToken(refreshToken)
	require.NoError(t, err, "Should not fail to validate refresh token")
	assert.Equal(t, userID, refreshClaims["sub"], "Subject should match user ID")
	assert.NotContains(t, refreshClaims, "role", "Refresh token should not contain custom claims")
}

// TestTokenRefreshFlow tests the complete token refresh cycle.
func TestTokenRefreshFlow(t *testing.T) {
	tm := setupTestManager(t)
	userID := "user-456"

	// 1. Generate a refresh token
	refreshToken, err := tm.GenerateRefreshToken(userID)
	require.NoError(t, err)

	// 2. Use it to get a new access token
	newAccessToken, err := tm.RefreshAccessToken(refreshToken)
	require.NoError(t, err, "Should successfully refresh the access token")
	require.NotEmpty(t, newAccessToken)

	// 3. Validate the new access token
	claims, err := tm.ValidateAccessToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims["sub"], "New access token should have correct user ID")
	assert.NotContains(t, claims, "role", "Refreshed token should be clean and not have old custom claims")
}

// TestExpiredTokenValidation checks that expired tokens fail validation.
func TestExpiredTokenValidation(t *testing.T) {
	// Create a manager with a very short-lived access token
	cfg := JWTConfig{
		AccessSecret:   []byte("short-lived-secret"),
		Issuer:         "test-expiry",
		AccessTokenTTL: 1 * time.Millisecond, // Practically instant expiry
		SigningMethod:  HS256,
	}
	tm := NewJWTManager(cfg)

	accessToken, err := tm.GenerateAccessToken("user-789", nil)
	require.NoError(t, err)

	// Wait for the token to expire
	time.Sleep(5 * time.Millisecond)

	// Validation should now fail
	_, err = tm.ValidateAccessToken(accessToken)
	assert.Error(t, err, "Validation should fail for an expired token")
}

// TestInvalidTokenErrors covers various failure scenarios.
func TestInvalidTokenErrors(t *testing.T) {
	tm := setupTestManager(t)

	// 1. Try to validate a malformed token
	_, err := tm.ValidateAccessToken("this.is.not.a.valid.token")
	assert.Error(t, err, "Should fail for a malformed token")

	// 2. Try to refresh using an access token instead of a refresh token
	accessToken, _ := tm.GenerateAccessToken("user-111", nil)
	_, err = tm.RefreshAccessToken(accessToken)
	assert.Error(t, err, "Should fail when using an access token for refresh")

	// 3. Try to validate a token with the wrong secret
	otherManager := NewJWTManager(JWTConfig{
		AccessSecret:  []byte("a-different-secret"),
		SigningMethod: HS256,
	})
	tokenFromOther, _ := otherManager.GenerateAccessToken("user-222", nil)
	_, err = tm.ValidateAccessToken(tokenFromOther)
	assert.Error(t, err, "Should fail validation for a token signed with a different secret")
}

// TestSecurityChecks ensures secrets are not interchangeable.
func TestSecurityChecks(t *testing.T) {
	tm := setupTestManager(t)
	userID := "user-sec"

	accessToken, _ := tm.GenerateAccessToken(userID, nil)
	refreshToken, _ := tm.GenerateRefreshToken(userID)

	// 1. Access token should NOT be valid as a refresh token
	_, err := tm.ValidateRefreshToken(accessToken)
	assert.Error(t, err, "Access token should not be valid when validated as a refresh token")

	// 2. Refresh token should NOT be valid as an access token
	_, err = tm.ValidateAccessToken(refreshToken)
	assert.Error(t, err, "Refresh token should not be valid when validated as an access token")
}
