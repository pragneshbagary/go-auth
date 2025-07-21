// manager.go
package jwtutils

type TokenManager interface {
	GenerateAccessToken(userID string, customClaims map[string]any) (string, error)
	GenerateRefreshToken(userID string, customClaims map[string]any) (string, error)
	RefreshAccessToken(refreshToken string) (string, error)
	// ParseToken(token string) (*jwt.MapClaims, error)
}
