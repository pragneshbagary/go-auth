package main

import (
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Example 1: Simple SQLite setup (most common use case)
	fmt.Println("=== Example 1: Simple SQLite Setup ===")
	authService, err := auth.New("example.db", "my-secret-key")
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Register a new user
	user, err := authService.Register(auth.RegisterRequest{
		Username: "john_doe",
		Email:    "john@example.com",
		Password: "secure_password123",
	})
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("Registered user: %s (ID: %s)\n", user.Username, user.ID)

	// Login the user
	loginResult, err := authService.Login("john_doe", "secure_password123", nil)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Printf("Login successful! Access token: %s...\n", loginResult.AccessToken[:20])

	// Validate the token
	claims, err := authService.ValidateAccessToken(loginResult.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}
	fmt.Printf("Token valid! Username from claims: %s\n", claims["username"])

	// Get user profile
	profile, err := authService.GetUser(user.ID)
	if err != nil {
		log.Fatalf("Failed to get user profile: %v", err)
	}
	fmt.Printf("User profile: %s (%s)\n", profile.Username, profile.Email)

	fmt.Println()

	// Example 2: In-memory setup (for testing)
	fmt.Println("=== Example 2: In-Memory Setup ===")
	memoryAuth, err := auth.NewInMemory("test-secret")
	if err != nil {
		log.Fatalf("Failed to create in-memory auth: %v", err)
	}

	// Register and login with in-memory storage
	testUser, err := memoryAuth.Register(auth.RegisterRequest{
		Username: "test_user",
		Email:    "test@example.com",
		Password: "test123",
	})
	if err != nil {
		log.Fatalf("Failed to register test user: %v", err)
	}
	fmt.Printf("Test user registered: %s\n", testUser.Username)

	// Example 3: Custom configuration
	fmt.Println("=== Example 3: Custom Configuration ===")
	customAuth, err := auth.NewWithConfig(&auth.AuthConfig{
		JWTSecret:       "custom-secret",
		JWTRefreshSecret: "custom-refresh-secret",
		JWTIssuer:       "my-app",
		AppName:         "My Awesome App",
		// Uses in-memory storage by default when no database is specified
	})
	if err != nil {
		log.Fatalf("Failed to create custom auth: %v", err)
	}

	// Health check
	if err := customAuth.Health(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("Custom auth service is healthy!")

	fmt.Println("\n=== All examples completed successfully! ===")
}