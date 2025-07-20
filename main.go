package main

import (
	"fmt"
	"go-Auth/internal/jwtutils"
	"log"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func main() {
	fmt.Println("JWT Token Generation Example")

	// Example JWT configuration
	jwtConfig := jwtutils.JWTConfig{
		SecretKey:      []byte("your-secret-key"), // Use a secure key in production
		Issuer:         "go-Auth example",
		AccessTokenTTL: 15 * time.Minute,
		SigningMethod:  jwtv5.SigningMethodHS256, // or "RS256" for RSA signing
	}
	// Generate a token for a user
	// myClaims := jwtutils.Claims{
	// 	UserID: "12345",
	// }

	tm := jwtutils.NewJWTManager(jwtConfig)
	userID := "12345" // Example user ID, replace with actual user ID

	token, err := tm.GenerateAccessToken(userID)

	if err != nil {
		log.Fatalf("Error generating access token: %v", err)
	}
	fmt.Printf("Access Token: %s\n\n", token)

	fmt.Println("Token generated successfully for user:", userID)
	fmt.Println("Token:", token)

	parsedClaims, err := tm.ParseToken(token)
	if err != nil {
		log.Fatalf("Parse/verify failed: %v", err)
	}
	fmt.Printf("Parsed claims: %#v\n", parsedClaims)

}
