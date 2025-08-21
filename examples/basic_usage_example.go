package main

import (
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("=== Basic Usage Example - New API ===")

	// Example 1: Simple SQLite setup (most common use case)
	fmt.Println("\n1. Simple SQLite Setup:")
	authService, err := auth.New("basic_example.db", "my-super-secret-jwt-key")
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}
	fmt.Println("✓ Auth service created with SQLite database")

	// Example 2: Register a new user
	fmt.Println("\n2. Register a new user:")
	user, err := authService.Register(auth.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secure_password123",
	})
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("✓ User registered: %s (ID: %s)\n", user.Username, user.ID)

	// Example 3: Login and get tokens
	fmt.Println("\n3. Login and get tokens:")
	loginResult, err := authService.Login("alice", "secure_password123", nil)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Printf("✓ Login successful!\n")
	fmt.Printf("  Access Token: %s...\n", loginResult.AccessToken[:50])
	fmt.Printf("  Refresh Token: %s...\n", loginResult.RefreshToken[:50])

	// Example 4: Validate access token
	fmt.Println("\n4. Validate access token:")
	claims, err := authService.ValidateAccessToken(loginResult.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}
	fmt.Printf("✓ Token is valid!\n")
	fmt.Printf("  Username: %v\n", claims["username"])
	fmt.Printf("  Email: %v\n", claims["email"])
	fmt.Printf("  User ID: %v\n", claims["user_id"])

	// Example 5: Refresh tokens
	fmt.Println("\n5. Refresh tokens:")
	refreshResult, err := authService.RefreshToken(loginResult.RefreshToken)
	if err != nil {
		log.Fatalf("Failed to refresh token: %v", err)
	}
	fmt.Printf("✓ Tokens refreshed successfully!\n")
	fmt.Printf("  New Access Token: %s...\n", refreshResult.AccessToken[:50])
	fmt.Printf("  New Refresh Token: %s...\n", refreshResult.RefreshToken[:50])

	// Example 6: Get user profile
	fmt.Println("\n6. Get user profile:")
	profile, err := authService.GetUser(user.ID)
	if err != nil {
		log.Fatalf("Failed to get user profile: %v", err)
	}
	fmt.Printf("✓ User profile retrieved:\n")
	fmt.Printf("  ID: %s\n", profile.ID)
	fmt.Printf("  Username: %s\n", profile.Username)
	fmt.Printf("  Email: %s\n", profile.Email)
	fmt.Printf("  Created: %s\n", profile.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Active: %t\n", profile.IsActive)

	// Example 7: Login with custom claims
	fmt.Println("\n7. Login with custom claims:")
	customClaims := map[string]interface{}{
		"role":        "admin",
		"permissions": []string{"read", "write", "delete"},
		"department":  "engineering",
	}
	
	loginWithClaims, err := authService.Login("alice", "secure_password123", customClaims)
	if err != nil {
		log.Fatalf("Failed to login with claims: %v", err)
	}
	
	// Validate the token with custom claims
	claimsWithCustom, err := authService.ValidateAccessToken(loginWithClaims.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate token with custom claims: %v", err)
	}
	fmt.Printf("✓ Login with custom claims successful!\n")
	fmt.Printf("  Role: %v\n", claimsWithCustom["role"])
	fmt.Printf("  Department: %v\n", claimsWithCustom["department"])

	// Example 8: Health check
	fmt.Println("\n8. Health check:")
	if err := authService.Health(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("✓ Service is healthy!")

	fmt.Println("\n=== Basic Usage Example Complete ===")
}