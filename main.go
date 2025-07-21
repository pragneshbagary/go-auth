package main

import (
	"fmt"
	"go-Auth/internal/jwtutils"
	"log"
	"time"
)

func main() {
	fmt.Println("JWT Token Generation Example")

	// Example JWT configuration
	jwtConfig := jwtutils.JWTConfig{
		AccessSecret:    []byte("your-secret-key"),         // Use a secure key in production
		RefreshSecret:   []byte("your-refresh-secret-key"), // Use a secure key in production
		Issuer:          "go-Auth example",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 1 week
		SigningMethod:   jwtutils.HS256,     // Choose from HS256, HS384, HS512, RS256
	}

	tm := jwtutils.NewJWTManager(jwtConfig)
	userID := "12345"

	// Add any custom claims you want
	customClaims := map[string]interface{}{
		"username":   "johndoe",
		"role":       "admin",
		"device_id":  "abc-123",
		"tenant_id":  "tenant-456",
		"auth_level": "full",
		"scopes":     []string{"read:data", "write:data"},
	}

	accessToken, _ := tm.GenerateAccessToken(userID, customClaims)
	refreshToken, _ := tm.GenerateRefreshToken(userID)

	fmt.Println("Token generated successfully for user:", userID)
	fmt.Println("Token:", accessToken)

	parsedClaims, err := tm.ValidateAccessToken(accessToken, jwtConfig.AccessSecret)
	if err != nil {
		log.Fatalf("Parse/verify failed: %v", err)
	}
	refreshValid, error := tm.ValidateRefreshToken(refreshToken, jwtConfig.RefreshSecret)
	if error != nil {
		log.Fatalf("Parse/verify failed: %v", err)
	}

	fmt.Println("Parsed claims:")
	for key, value := range parsedClaims {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("refresh")
	for key, value := range refreshValid {
		fmt.Printf("  %s: %v\n", key, value)
	}

}
