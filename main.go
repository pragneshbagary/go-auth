package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pragneshbagary/go-auth/internal/jwtutils"
	"github.com/pragneshbagary/go-auth/internal/storage/memory"
	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("--- High-Level Auth Service Example ---")

	// 1. Setup: Initialize all the components.

	// The in-memory storage is for this example.
	// In a real application, you would plug in a real database implementation here.
	storage := memory.NewInMemoryStorage()

	// The JWT manager from your internal package.
	jwtConfig := jwtutils.JWTConfig{
		AccessSecret:    []byte("example-access-secret"),
		RefreshSecret:   []byte("example-refresh-secret"),
		Issuer:          "go-auth-example",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		SigningMethod:   jwtutils.HS256,
	}
	jwtManager := jwtutils.NewJWTManager(jwtConfig)

	// The high-level authentication service.
	authService := auth.NewAuthService(storage, jwtManager)

	// 2. Register a new user using the new payload struct.
	fmt.Println("\n--- Registering a new user... ---")
	registerPayload := auth.RegisterPayload{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "StrongPassword123!",
	}
	user, err := authService.Register(registerPayload)
	if err != nil {
		log.Fatalf("Registration failed: %v", err)
	}
	fmt.Printf("User '%s' registered successfully with ID: %s\n", user.Username, user.ID)

	// 3. Log in and add custom claims.
	fmt.Println("\n--- Logging in... ---")
	// In a real app, you might fetch the user's roles from a database here.
	customClaims := map[string]interface{}{
		"role":   "admin",
		"scopes": []string{"read:all", "write:all"},
	}
	loginResponse, err := authService.Login(registerPayload.Username, registerPayload.Password, customClaims)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Println("Login successful!")
	fmt.Printf("Access Token: %s\n", loginResponse.AccessToken)
	fmt.Printf("Refresh Token: %s\n", loginResponse.RefreshToken)

	// 4. Validate the token and inspect the claims.
	fmt.Println("\n--- Validating the received access token... ---")
	claims, err := jwtManager.ValidateAccessToken(loginResponse.AccessToken)
	if err != nil {
		log.Fatalf("Token validation failed: %v", err)
	}
	fmt.Println("Token is valid. Inspecting claims:")
	fmt.Printf("  User ID (sub): %s\n", claims["sub"])
	fmt.Printf("  Username: %s\n", claims["username"])
	fmt.Printf("  Role: %s\n", claims["role"])
	fmt.Printf("  Scopes: %v\n", claims["scopes"])
}
