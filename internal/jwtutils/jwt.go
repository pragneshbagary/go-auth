package jwtutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	AccessSecret    []byte
	RefreshSecret   []byte        // used for HS256 signing
	Issuer          string        // e.g. "neighborhood-app"
	AccessTokenTTL  time.Duration // e.g. 15 * time.Minute
	RefreshTokenTTL time.Duration // e.g. 7 * 24 * time.Hour
	SigningMethod   string        // jwt.SigningMethodHS256 or RS256
}

var signingMethods = map[string]jwt.SigningMethod{
	HS256: jwt.SigningMethodHS256,
	HS384: jwt.SigningMethodHS384,
	HS512: jwt.SigningMethodHS512,
	RS256: jwt.SigningMethodRS256,
}
