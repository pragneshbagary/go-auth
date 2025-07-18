package main

import (
	"fmt"
	"go-Auth/jwt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func main() {
	fmt.Println("JWT Token Generation Example")

	// Example JWT configuration
	jwtConfig := jwt.JWTConfig{
		SecretKey:       []byte("testing my secret key"),
		Issuer:          "neighborhood-app",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		SigningMethod:   jwtv5.SigningMethodHS256, // or "RS256" for RSA signing
		TokenType:       "access",
	}
	// Generate a token for a user
	myClaims := jwt.Claims{
		UserID: "12345",
	}

	token, err := jwt.GenerateToken(myClaims.UserID, jwtConfig)

	if err != nil {
		fmt.Println("Error generating token:", err)
		return
	}

	fmt.Println("Token generated successfully for user:", myClaims.UserID)
	fmt.Println("Token:", token)

}
