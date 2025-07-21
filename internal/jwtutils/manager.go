// manager.go
package jwtutils

import "github.com/golang-jwt/jwt/v5"

type TokenManager interface {
	GenerateAccessToken(userID string, customClaims map[string]any) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	RefreshAccessToken(refreshToken string) (string, error)
	ValidateAccessToken(accessToken string) (jwt.MapClaims, error)
	ValidateRefreshToken(refreshToken string) (jwt.MapClaims, error)
}
