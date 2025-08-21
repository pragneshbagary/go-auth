package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pragneshbagary/go-auth/internal/jwtutils"
	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// Tokens provides enhanced token management operations with clear method names.
type Tokens struct {
	jwtManager       jwtutils.TokenManager
	storage          storage.EnhancedStorage
	eventLogger      *AuthEventLogger
	metricsCollector *MetricsCollector
}

// RefreshResult represents the result of a token refresh operation.
type RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// SessionInfo represents information about an active session.
type SessionInfo struct {
	TokenID   string    `json:"token_id"`
	UserID    string    `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
}

// ValidationResult represents the result of token validation.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Claims jwt.MapClaims     `json:"claims,omitempty"`
	User   *models.User      `json:"user,omitempty"`
	Error  string            `json:"error,omitempty"`
}

// Refresh validates a refresh token and generates new access and refresh tokens.
// This implements automatic token rotation for enhanced security.
func (t *Tokens) Refresh(refreshToken string) (*RefreshResult, error) {
	start := time.Now()
	var userID string
	var success bool
	var err error

	defer func() {
		duration := time.Since(start)
		if t.eventLogger != nil {
			t.eventLogger.LogTokenRefresh(userID, "", "", success, duration, err)
		}
		if t.metricsCollector != nil {
			t.metricsCollector.RecordTokenRefresh(success, duration)
		}
	}()

	// First, validate the refresh token
	claims, validateErr := t.jwtManager.ValidateRefreshToken(refreshToken)
	if validateErr != nil {
		err = ErrInvalidToken()
		return nil, err
	}

	// Extract token ID and check if it's blacklisted
	tokenID, ok := claims["jti"].(string)
	if !ok {
		err = NewAuthErrorWithDetails(ErrCodeInvalidToken, 
			"Refresh token missing token ID", "Token must contain a valid 'jti' claim")
		return nil, err
	}

	// Check if the token is blacklisted
	blacklisted, blacklistErr := t.storage.IsTokenBlacklisted(tokenID)
	if blacklistErr != nil {
		err = WrapDatabaseError(blacklistErr)
		return nil, err
	}
	if blacklisted {
		err = ErrTokenRevoked()
		return nil, err
	}

	// Extract user ID
	userID, ok = claims["sub"].(string)
	if !ok {
		err = NewAuthErrorWithDetails(ErrCodeInvalidToken, 
			"Refresh token missing user ID", "Token must contain a valid 'sub' claim")
		return nil, err
	}

	// Verify user still exists and is active
	user, getUserErr := t.storage.GetUserByID(userID)
	if getUserErr != nil {
		err = ErrUserNotFound()
		return nil, err
	}
	if !user.IsActive {
		err = ErrUserInactive()
		return nil, err
	}

	// Generate new access token with user claims
	userClaims := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"user_id":  user.ID,
	}

	newAccessToken, accessErr := t.jwtManager.GenerateAccessToken(userID, userClaims)
	if accessErr != nil {
		err = WrapError(accessErr, ErrCodeInternalError, "Failed to generate new access token")
		return nil, err
	}

	// Generate new refresh token (token rotation)
	newRefreshToken, refreshErr := t.jwtManager.GenerateRefreshToken(userID)
	if refreshErr != nil {
		err = WrapError(refreshErr, ErrCodeInternalError, "Failed to generate new refresh token")
		return nil, err
	}

	// Blacklist the old refresh token to prevent reuse
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt := time.Unix(int64(exp), 0)
		if blacklistErr := t.storage.BlacklistToken(tokenID, expiresAt); blacklistErr != nil {
			// Log the error but don't fail the refresh operation
			// The new tokens are still valid
			fmt.Printf("Warning: failed to blacklist old refresh token: %v\n", blacklistErr)
		}
	}

	success = true
	return &RefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Revoke blacklists a specific token, preventing its future use.
// This works for both access and refresh tokens.
func (t *Tokens) Revoke(tokenString string) error {
	// Try to parse as access token first
	claims, err := t.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		// If access token validation fails, try refresh token
		claims, err = t.jwtManager.ValidateRefreshToken(tokenString)
		if err != nil {
			return ErrInvalidToken()
		}
	}

	// Extract token ID
	tokenID, ok := claims["jti"].(string)
	if !ok {
		return NewAuthErrorWithDetails(ErrCodeInvalidToken, 
			"Token missing token ID", "Token must contain a valid 'jti' claim")
	}

	// Extract expiration time
	exp, ok := claims["exp"].(float64)
	if !ok {
		return NewAuthErrorWithDetails(ErrCodeInvalidToken, 
			"Token missing expiration time", "Token must contain a valid 'exp' claim")
	}
	expiresAt := time.Unix(int64(exp), 0)

	// Blacklist the token
	if err := t.storage.BlacklistToken(tokenID, expiresAt); err != nil {
		return WrapDatabaseError(err)
	}

	return nil
}

// RevokeAll blacklists all tokens for a specific user.
// This is useful for scenarios like password changes or account compromise.
func (t *Tokens) RevokeAll(userID string) error {
	// Note: This is a simplified implementation. In a production system,
	// you might want to track active tokens per user in the database.
	// For now, we'll implement this by updating the user's UpdatedAt timestamp
	// and checking it during token validation.
	
	// Update user's UpdatedAt timestamp to invalidate all existing tokens
	updates := storage.UserUpdates{}
	if err := t.storage.UpdateUser(userID, updates); err != nil {
		return WrapDatabaseError(err)
	}

	return nil
}

// Validate checks if a token is valid and returns the associated user.
// This method checks both token validity and blacklist status.
func (t *Tokens) Validate(tokenString string) (*models.User, error) {
	// Try to parse as access token first
	claims, err := t.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken()
	}

	// Extract token ID and check blacklist
	tokenID, ok := claims["jti"].(string)
	if !ok {
		return nil, NewAuthErrorWithDetails(ErrCodeInvalidToken, 
			"Token missing token ID", "Token must contain a valid 'jti' claim")
	}

	blacklisted, err := t.storage.IsTokenBlacklisted(tokenID)
	if err != nil {
		return nil, WrapDatabaseError(err)
	}
	if blacklisted {
		return nil, ErrTokenRevoked()
	}

	// Extract user ID and fetch user
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, NewAuthErrorWithDetails(ErrCodeInvalidToken, 
			"Token missing user ID", "Token must contain a valid 'sub' claim")
	}

	user, err := t.storage.GetUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound()
	}

	if !user.IsActive {
		return nil, ErrUserInactive()
	}

	return user, nil
}

// IsValid performs a quick validation check without returning user data.
// This is useful for middleware that only needs to know if a token is valid.
func (t *Tokens) IsValid(tokenString string) bool {
	_, err := t.Validate(tokenString)
	return err == nil
}

// ValidateBatch validates multiple tokens in a single operation.
// This is useful for scenarios where you need to validate many tokens at once.
func (t *Tokens) ValidateBatch(tokens []string) []ValidationResult {
	results := make([]ValidationResult, len(tokens))
	
	for i, token := range tokens {
		user, err := t.Validate(token)
		if err != nil {
			results[i] = ValidationResult{
				Valid: false,
				Error: err.Error(),
			}
		} else {
			// Get claims for the result
			claims, _ := t.jwtManager.ValidateAccessToken(token)
			results[i] = ValidationResult{
				Valid:  true,
				Claims: claims,
				User:   user,
			}
		}
	}
	
	return results
}

// GetSessionInfo extracts session information from a token without full validation.
// This is useful for logging and monitoring purposes.
func (t *Tokens) GetSessionInfo(tokenString string) (*SessionInfo, error) {
	// Try to parse as access token first
	claims, err := t.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		// Try refresh token
		claims, err = t.jwtManager.ValidateRefreshToken(tokenString)
		if err != nil {
			return nil, ErrInvalidToken()
		}
	}

	// Extract session information
	tokenID, _ := claims["jti"].(string)
	userID, _ := claims["sub"].(string)
	tokenType, _ := claims["token_type"].(string)
	
	var issuedAt, expiresAt time.Time
	if iat, ok := claims["iat"].(float64); ok {
		issuedAt = time.Unix(int64(iat), 0)
	}
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(exp), 0)
	}

	return &SessionInfo{
		TokenID:   tokenID,
		UserID:    userID,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		TokenType: tokenType,
	}, nil
}

// CleanupExpired removes expired tokens from the blacklist.
// This should be called periodically to maintain database performance.
func (t *Tokens) CleanupExpired() error {
	return t.storage.CleanupExpiredTokens()
}

// ListActiveSessions returns information about active sessions for a user.
// Note: This is a simplified implementation. In a production system,
// you might want to track active sessions more explicitly.
func (t *Tokens) ListActiveSessions(userID string) ([]*SessionInfo, error) {
	// This is a placeholder implementation since we don't currently track
	// active sessions in the database. In a full implementation, you would
	// store session information and query it here.
	return []*SessionInfo{}, nil
}