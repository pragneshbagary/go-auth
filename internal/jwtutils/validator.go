package jwtutils

import (
	"errors"

	"github.com/golang-jwt/jwt"
)

func (tm *JWTManager) ValidateAccessToken(accessToken string, secret []byte) (jwt.MapClaims, error) {
	return tm.parseToken(accessToken, tm.cfg.AccessSecret)
}

func (tm *JWTManager) ValidateRefreshToken(refreshToken string, secret []byte) (jwt.MapClaims, error) {
	return tm.parseToken(refreshToken, tm.cfg.RefreshSecret)
}

func (tm *JWTManager) parseToken(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != signingMethods[tm.cfg.SigningMethod].Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
