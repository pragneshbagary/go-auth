package jwtutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SecretKey       []byte            // used for HS256 signing
	Issuer          string            // e.g. "neighborhood-app"
	AccessTokenTTL  time.Duration     // e.g. 15 * time.Minute
	RefreshTokenTTL time.Duration     // e.g. 7 * 24 * time.Hour
	SigningMethod   jwt.SigningMethod // jwt.SigningMethodHS256 or RS256
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
	TokenType string   `json:"token_type,omitempty"`
}
