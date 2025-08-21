package main

import (
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("=== SimpleAuth Example ===")

	// Example 1: Ultra-simple setup with default SQLite database
	fmt.Println("\n1. Quick setup with default SQLite:")
	simpleAuth, err := auth.Quick("my-super-secret-jwt-key")
	if err != nil {
		log.Fatalf("Failed to create SimpleAuth: %v", err)
	}
	fmt.Println("✓ SimpleAuth created successfully with default SQLite database")

	// Example 2: Register a new user
	fmt.Println("\n2. Register a new user:")
	user, err := simpleAuth.Register("john_doe", "john@example.com", "secure_password123")
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("✓ User registered: %s (ID: %s)\n", user.Username, user.ID)

	// Example 3: Login and get tokens
	fmt.Println("\n3. Login and get tokens:")
	loginResult, err := simpleAuth.Login("john_doe", "secure_password123")
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Printf("✓ Login successful!\n")
	fmt.Printf("  Access Token: %s...\n", loginResult.AccessToken[:50])
	fmt.Printf("  Refresh Token: %s...\n", loginResult.RefreshToken[:50])

	// Example 4: Validate access token
	fmt.Println("\n4. Validate access token:")
	claims, err := simpleAuth.ValidateToken(loginResult.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}
	fmt.Printf("✓ Token is valid!\n")
	fmt.Printf("  Username: %v\n", claims["username"])
	fmt.Printf("  Email: %v\n", claims["email"])
	fmt.Printf("  User ID: %v\n", claims["user_id"])

	// Example 5: Refresh token
	fmt.Println("\n5. Refresh token:")
	refreshResult, err := simpleAuth.RefreshToken(loginResult.RefreshToken)
	if err != nil {
		log.Fatalf("Failed to refresh token: %v", err)
	}
	fmt.Printf("✓ Token refreshed successfully!\n")
	fmt.Printf("  New Access Token: %s...\n", refreshResult.AccessToken[:50])

	// Example 6: Get user information
	fmt.Println("\n6. Get user information:")
	userProfile, err := simpleAuth.GetUser(user.ID)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	fmt.Printf("✓ User profile retrieved:\n")
	fmt.Printf("  ID: %s\n", userProfile.ID)
	fmt.Printf("  Username: %s\n", userProfile.Username)
	fmt.Printf("  Email: %s\n", userProfile.Email)
	fmt.Printf("  Created: %s\n", userProfile.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Active: %t\n", userProfile.IsActive)

	// Example 7: Login with custom claims
	fmt.Println("\n7. Login with custom claims:")
	customClaims := map[string]interface{}{
		"role":        "admin",
		"permissions": []string{"read", "write", "delete"},
		"department":  "engineering",
	}
	
	loginWithClaims, err := simpleAuth.LoginWithClaims("john_doe", "secure_password123", customClaims)
	if err != nil {
		log.Fatalf("Failed to login with claims: %v", err)
	}
	
	// Validate the token with custom claims
	claimsWithCustom, err := simpleAuth.ValidateToken(loginWithClaims.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate token with custom claims: %v", err)
	}
	fmt.Printf("✓ Login with custom claims successful!\n")
	fmt.Printf("  Role: %v\n", claimsWithCustom["role"])
	fmt.Printf("  Department: %v\n", claimsWithCustom["department"])

	// Example 8: Health check
	fmt.Println("\n8. Health check:")
	if err := simpleAuth.Health(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("✓ Service is healthy!")

	fmt.Println("\n=== All examples completed successfully! ===")
}