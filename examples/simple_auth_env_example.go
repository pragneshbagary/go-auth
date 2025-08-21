package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("=== SimpleAuth Environment Configuration Example ===")

	// Set up environment variables for demonstration
	// In a real application, these would be set in your environment or .env file
	os.Setenv("AUTH_JWT_ACCESS_SECRET", "my-super-secret-access-key")
	os.Setenv("AUTH_JWT_REFRESH_SECRET", "my-super-secret-refresh-key")
	os.Setenv("AUTH_DB_TYPE", "memory") // Use in-memory for this example
	os.Setenv("AUTH_JWT_ISSUER", "my-awesome-app")
	os.Setenv("AUTH_ACCESS_TOKEN_TTL", "30m")
	os.Setenv("AUTH_REFRESH_TOKEN_TTL", "24h")
	os.Setenv("AUTH_APP_NAME", "My Awesome App")

	fmt.Println("\n1. Environment variables set:")
	fmt.Println("  AUTH_JWT_ACCESS_SECRET: [HIDDEN]")
	fmt.Println("  AUTH_JWT_REFRESH_SECRET: [HIDDEN]")
	fmt.Println("  AUTH_DB_TYPE: memory")
	fmt.Println("  AUTH_JWT_ISSUER: my-awesome-app")
	fmt.Println("  AUTH_ACCESS_TOKEN_TTL: 30m")
	fmt.Println("  AUTH_REFRESH_TOKEN_TTL: 24h")
	fmt.Println("  AUTH_APP_NAME: My Awesome App")

	// Create SimpleAuth from environment variables
	fmt.Println("\n2. Creating SimpleAuth from environment:")
	simpleAuth, err := auth.QuickFromEnv()
	if err != nil {
		log.Fatalf("Failed to create SimpleAuth from environment: %v", err)
	}
	fmt.Println("✓ SimpleAuth created successfully from environment variables")

	// Test the functionality
	fmt.Println("\n3. Testing functionality:")
	
	// Register a user
	user, err := simpleAuth.Register("env_user", "env@example.com", "env_password123")
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("✓ User registered: %s\n", user.Username)

	// Login
	loginResult, err := simpleAuth.Login("env_user", "env_password123")
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Println("✓ Login successful")

	// Validate token
	claims, err := simpleAuth.ValidateToken(loginResult.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}
	fmt.Printf("✓ Token validated - User: %v, Issuer: %v\n", claims["username"], claims["iss"])

	// Health check
	if err := simpleAuth.Health(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("✓ Health check passed")

	fmt.Println("\n=== Environment configuration example completed! ===")
	fmt.Println("\nTip: In production, set these environment variables in your deployment environment")
	fmt.Println("or use a .env file with a library like godotenv.")
}