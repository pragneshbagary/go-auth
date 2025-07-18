package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SecretKey       []byte            // used for HS256 signing
	Issuer          string            // e.g. "neighborhood-app"
	AccessTokenTTL  time.Duration     // e.g. 15 * time.Minute
	RefreshTokenTTL time.Duration     // e.g. 7 * 24 * time.Hour
	SigningMethod   jwt.SigningMethod // jwt.SigningMethodHS256 or RS256
	TokenType       string            // optional: "access", "refresh"
}

type Claims struct {
	jwt.RegisteredClaims
	UserID    string   `json:"user_id"`
	Username  string   `json:"username,omitempty"`
	Scopes    []string `json:"scopes,omitempty"`
	Role      string   `json:"role,omitempty"`
	AuthLevel string   `json:"auth_level,omitempty"`
	DeviceId  string   `json:"device_id,omitempty"`
	TenantId  string   `json:"tenant_id,omitempty"`
}

// GenerateToken creates a new JWT for a given user ID and configuration.
// It includes detailed error handling for common configuration issues.
func GenerateToken(userId string, cfg JWTConfig) (string, error) {
	// 1. Validate the configuration
	if cfg.SigningMethod == nil {
		return "", errors.New("JWT signing method cannot be nil")
	}
	if len(cfg.SecretKey) == 0 {
		// This check is crucial for symmetric algorithms like HS256
		// For asymmetric algorithms (like RS256), the key would be an rsa.PrivateKey,
		// so this check might need to be more sophisticated if you support both key types.
		return "", errors.New("JWT secret key cannot be empty")
	}

	// 2. Create the claims
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userId,
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		UserID: userId,
	}

	// 3. Create the token object
	token := jwt.NewWithClaims(cfg.SigningMethod, claims)

	// 4. Sign the token
	signedToken, err := token.SignedString(cfg.SecretKey)
	if err != nil {
		// Wrap the original error with more context
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}