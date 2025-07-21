package jwtutils

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager is the concrete implementation of the TokenManager interface.
type JWTManager struct {
	cfg JWTConfig
}

// NewJWTManager creates a new TokenManager with the given configuration.
// It returns an interface, promoting loose coupling.
func NewJWTManager(cfg JWTConfig) TokenManager {
	return &JWTManager{
		cfg: cfg,
	}
}

// GenerateAccessToken creates a new access token with the specified custom claims.
func (m *JWTManager) GenerateAccessToken(userID string, customClaims map[string]any) (string, error) {
	if m.cfg.SigningMethod == "" {
		return "", errors.New("JWT signing method cannot be empty in config")
	}
	if len(m.cfg.AccessSecret) == 0 {
		return "", errors.New("JWT access secret key cannot be empty in config")
	}

	claims := jwt.MapClaims{
		"iss":        m.cfg.Issuer,
		"exp":        time.Now().Add(m.cfg.AccessTokenTTL).Unix(),
		"iat":        time.Now().Unix(),
		"sub":        userID,
		"nbf":        time.Now().Unix(),
		"jti":        uuid.New().String(),
		"token_type": "access",
	}

	// Securely copy custom claims, ensuring they don't overwrite standard claims.
	if customClaims != nil {
		maps.Copy(claims, customClaims)
	}

	method, ok := signingMethods[m.cfg.SigningMethod]
	if !ok {
		return "", fmt.Errorf("unsupported signing method in config: %s", m.cfg.SigningMethod)
	}

	token := jwt.NewWithClaims(method, claims)
	signedToken, err := token.SignedString(m.cfg.AccessSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken creates a simple, long-lived refresh token.
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	if m.cfg.SigningMethod == "" {
		return "", errors.New("JWT signing method cannot be empty in config")
	}
	if len(m.cfg.RefreshSecret) == 0 {
		return "", errors.New("JWT refresh secret key cannot be empty in config")
	}

	claims := jwt.MapClaims{
		"iss":        m.cfg.Issuer,
		"exp":        time.Now().Add(m.cfg.RefreshTokenTTL).Unix(),
		"iat":        time.Now().Unix(),
		"sub":        userID,
		"jti":        uuid.New().String(),
		"token_type": "refresh",
	}

	method, ok := signingMethods[m.cfg.SigningMethod]
	if !ok {
		return "", fmt.Errorf("unsupported signing method in config: %s", m.cfg.SigningMethod)
	}

	token := jwt.NewWithClaims(method, claims)
	signedToken, err := token.SignedString(m.cfg.RefreshSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

// RefreshAccessToken validates a refresh token and issues a new, clean access token.
func (m *JWTManager) RefreshAccessToken(refreshTokenStr string) (string, error) {
	// 1. Validate the refresh token using the simplified validation method.
	claims, err := m.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return "", fmt.Errorf("could not validate refresh token: %w", err)
	}

	// 2. Double-check that it's a refresh token.
	if tokenType, ok := claims["token_type"].(string); !ok || tokenType != "refresh" {
		return "", errors.New("token is not a valid refresh token")
	}

	// 3. Get user ID from the 'subject' claim.
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid user ID in refresh token claims")
	}

	// 4. Generate a new access token.
	// CRITICAL SECURITY FIX: Pass `nil` for custom claims to generate a clean token,
	// preventing old claims from being carried over.
	newAccessToken, err := m.GenerateAccessToken(userID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessToken, nil
}