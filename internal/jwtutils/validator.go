package jwtutils

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateAccessToken validates an access token string using the secret from the JWTManager config.
func (m *JWTManager) ValidateAccessToken(accessToken string) (jwt.MapClaims, error) {
	return m.parseToken(accessToken, m.cfg.AccessSecret)
}

// ValidateRefreshToken validates a refresh token string using the secret from the JWTManager config.
func (m *JWTManager) ValidateRefreshToken(refreshToken string) (jwt.MapClaims, error) {
	return m.parseToken(refreshToken, m.cfg.RefreshSecret)
}

// parseToken is an internal helper that parses a token string with a given secret.
func (m *JWTManager) parseToken(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Check that the signing method is the one specified in the config.
		if _, ok := signingMethods[m.cfg.SigningMethod]; !ok {
			return nil, fmt.Errorf("unsupported signing method: %s", m.cfg.SigningMethod)
		}
		if token.Method.Alg() != signingMethods[m.cfg.SigningMethod].Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		// The library returns a detailed error, e.g., if the token is expired.
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		// This case handles other issues, like if the claims aren't a map or the token is invalid for other reasons.
		return nil, errors.New("invalid token or claims")
	}

	return claims, nil
}