// manager.go
package jwtutils

type TokenManager interface {
	GenerateAccessToken(claims Claims) (string, error)
	GenerateRefreshToken(claims Claims) (string, error)
	ParseToken(token string) (*Claims, error)
}
