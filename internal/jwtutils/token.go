package jwtutils

import (
	"errors"
	"fmt"
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

func (m *JWTManager) GenerateAccessToken(userId string) (string, error) {
	if m.cfg.SigningMethod == nil {
		return "", errors.New("JWT signing method cannot be nil")
	}
	if len(m.cfg.SecretKey) == 0 {

		return "", errors.New("JWT secret key cannot be empty")
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userId,
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		UserID:    userId,
		TokenType: "access",
	}

	token := jwt.NewWithClaims(m.cfg.SigningMethod, claims)

	// 4. Sign the token
	signedToken, err := token.SignedString(m.cfg.SecretKey)
	if err != nil {
		// Wrap the original error with more context
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (m *JWTManager) GenerateRefreshToken(userId string) (string, error) {
	if m.cfg.SigningMethod == nil {
		return "", errors.New("JWT signing method cannot be nil")
	}
	if len(m.cfg.SecretKey) == 0 {
		return "", errors.New("JWT secret key cannot be empty")
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userId,
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		UserID:    userId,
		TokenType: "refresh",
	}

	token := jwt.NewWithClaims(m.cfg.SigningMethod, claims)

	signedToken, err := token.SignedString(m.cfg.SecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

func (m *JWTManager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != m.cfg.SigningMethod.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return m.cfg.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
