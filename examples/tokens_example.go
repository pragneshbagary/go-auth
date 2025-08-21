package main

import (
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Create an Auth instance with in-memory storage for this example
	authService, err := auth.NewInMemory("your-secret-key")
	if err != nil {
		log.Fatal("Failed to create auth service:", err)
	}

	// Register a user
	user, err := authService.Register(auth.RegisterRequest{
		Username: "john_doe",
		Email:    "john@example.com",
		Password: "secure_password123",
	})
	if err != nil {
		log.Fatal("Failed to register user:", err)
	}
	fmt.Printf("Registered user: %s (ID: %s)\n", user.Username, user.ID)

	// Login to get initial tokens
	loginResult, err := authService.Login("john_doe", "secure_password123", map[string]interface{}{
		"role": "user",
	})
	if err != nil {
		log.Fatal("Failed to login:", err)
	}
	fmt.Println("Login successful!")
	fmt.Printf("Access Token: %s...\n", loginResult.AccessToken[:50])
	fmt.Printf("Refresh Token: %s...\n", loginResult.RefreshToken[:50])

	// Get the Tokens component for enhanced token management
	tokens := authService.Tokens()

	// Validate the access token
	validatedUser, err := tokens.Validate(loginResult.AccessToken)
	if err != nil {
		log.Fatal("Failed to validate token:", err)
	}
	fmt.Printf("Token validated for user: %s\n", validatedUser.Username)

	// Check if token is valid (quick check)
	isValid := tokens.IsValid(loginResult.AccessToken)
	fmt.Printf("Token is valid: %t\n", isValid)

	// Get session information from the token
	sessionInfo, err := tokens.GetSessionInfo(loginResult.AccessToken)
	if err != nil {
		log.Fatal("Failed to get session info:", err)
	}
	fmt.Printf("Session Info - Token ID: %s, User ID: %s, Type: %s\n", 
		sessionInfo.TokenID, sessionInfo.UserID, sessionInfo.TokenType)

	// Refresh the tokens (automatic token rotation)
	refreshResult, err := tokens.Refresh(loginResult.RefreshToken)
	if err != nil {
		log.Fatal("Failed to refresh tokens:", err)
	}
	fmt.Println("Tokens refreshed successfully!")
	fmt.Printf("New Access Token: %s...\n", refreshResult.AccessToken[:50])
	fmt.Printf("New Refresh Token: %s...\n", refreshResult.RefreshToken[:50])

	// The old refresh token is now automatically blacklisted
	_, err = tokens.Refresh(loginResult.RefreshToken)
	if err != nil {
		fmt.Printf("Old refresh token is now invalid: %v\n", err)
	}

	// Validate multiple tokens at once (batch validation)
	testTokens := []string{
		refreshResult.AccessToken,
		"invalid-token",
		loginResult.AccessToken, // This might still be valid
	}
	
	results := tokens.ValidateBatch(testTokens)
	fmt.Println("\nBatch validation results:")
	for i, result := range results {
		if result.Valid {
			fmt.Printf("Token %d: Valid (User: %s)\n", i+1, result.User.Username)
		} else {
			fmt.Printf("Token %d: Invalid (%s)\n", i+1, result.Error)
		}
	}

	// Revoke a specific token
	err = tokens.Revoke(refreshResult.AccessToken)
	if err != nil {
		log.Fatal("Failed to revoke token:", err)
	}
	fmt.Println("Access token revoked successfully!")

	// Verify the token is now invalid
	isValid = tokens.IsValid(refreshResult.AccessToken)
	fmt.Printf("Revoked token is valid: %t\n", isValid)

	// Revoke all tokens for a user (useful for logout from all devices)
	err = tokens.RevokeAll(user.ID)
	if err != nil {
		log.Fatal("Failed to revoke all tokens:", err)
	}
	fmt.Println("All tokens for user revoked successfully!")

	// Clean up expired tokens (maintenance operation)
	err = tokens.CleanupExpired()
	if err != nil {
		log.Fatal("Failed to cleanup expired tokens:", err)
	}
	fmt.Println("Expired tokens cleaned up!")

	// List active sessions (placeholder implementation)
	sessions, err := tokens.ListActiveSessions(user.ID)
	if err != nil {
		log.Fatal("Failed to list active sessions:", err)
	}
	fmt.Printf("Active sessions: %d\n", len(sessions))

	fmt.Println("\nTokens component example completed successfully!")
}