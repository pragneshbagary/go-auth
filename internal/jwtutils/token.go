package jwtutils

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	cfg JWTConfig
}

func NewJWTManager(cfg JWTConfig) *JWTManager {
	return &JWTManager{
		cfg: cfg,
	}
}

func (m *JWTManager) GenerateAccessToken(userID string, customClaims map[string]any) (string, error) {
	if m.cfg.SigningMethod == "" {
		return "", errors.New("JWT signing method cannot be nil")
	}
	if len(m.cfg.AccessSecret) == 0 {
		return "", errors.New("JWT secret key cannot be empty")
	}

	// Create the claims
	claims := jwt.MapClaims{
		"iss":        m.cfg.Issuer,
		"exp":        time.Now().Add(m.cfg.AccessTokenTTL).Unix(),
		"iat":        time.Now().Unix(),
		"sub":        userID,
		"nbf":        time.Now().Unix(),
		"jti":        uuid.New().String(),
		"token_type": "access",
	}

	// Add custom claims
	maps.Copy(claims, customClaims)
	method := signingMethods[m.cfg.SigningMethod]
	fmt.Println(method)

	token := jwt.NewWithClaims(method, claims)

	// Sign the token
	signedToken, err := token.SignedString(m.cfg.AccessSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	if m.cfg.SigningMethod == "" {
		return "", errors.New("JWT signing method cannot be nil")
	}
	if len(m.cfg.RefreshSecret) == 0 {
		return "", errors.New("JWT secret key cannot be empty")
	}

	claims := jwt.MapClaims{
		"iss":        m.cfg.Issuer,
		"exp":        time.Now().Add(m.cfg.RefreshTokenTTL).Unix(),
		"iat":        time.Now().Unix(),
		"sub":        userID,
		"nbf":        time.Now().Unix(),
		"jti":        uuid.New().String(),
		"token_type": "refresh",
	}

	token := jwt.NewWithClaims(signingMethods[m.cfg.SigningMethod], claims)

	signedToken, err := token.SignedString(m.cfg.RefreshSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

// RefreshAccessToken validates a refresh token and issues a new access token.
func (m *JWTManager) RefreshAccessToken(refreshTokenStr string) (string, error) {
	// 1. Validate the refresh token
	claims, err := m.ValidateRefreshToken(refreshTokenStr, m.cfg.RefreshSecret)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Check token type
	if claims["token_type"] != "refresh" {
		return "", errors.New("token is not a refresh token")
	}

	// 3. Get user ID from subject
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid user ID in refresh token")
	}

	// 4. Generate a new access token (without custom claims for security)
	newAccessToken, err := m.GenerateAccessToken(userID, claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessToken, nil
}
