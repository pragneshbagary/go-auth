package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pragneshbagary/go-auth/internal/storage/memory"
	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("--- High-Level Auth Service Example ---")

	// 1. Setup: Initialize all the components using the unified Config.
	cfg := auth.Config{
		Storage: memory.NewInMemoryStorage(), // Use your own DB implementation here
		JWT: auth.JWTConfig{
			AccessSecret:    []byte("your-super-secret-access-key"),
			RefreshSecret:   []byte("your-super-secret-refresh-key"),
			Issuer:          "my-awesome-app",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			SigningMethod:   auth.HS256,
		},
	}
	authService, err := auth.NewAuthService(cfg)
	if err != nil {
		log.Fatalf("Failed to create AuthService: %v", err)
	}

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
	claims, err := authService.ValidateAccessToken(loginResponse.AccessToken)
	if err != nil {
		log.Fatalf("Token validation failed: %v", err)
	}
	fmt.Println("Token is valid. Inspecting claims:")
	fmt.Printf("  User ID (sub): %s\n", claims["sub"])
	fmt.Printf("  Username: %s\n", claims["username"])
	fmt.Printf("  Role: %s\n", claims["role"])
	fmt.Printf("  Scopes: %v\n", claims["scopes"])
}
