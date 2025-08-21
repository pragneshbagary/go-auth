package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Create an auth instance
	authService, err := auth.NewInMemory("your-secret-key")
	if err != nil {
		log.Fatal("Failed to create auth service:", err)
	}

	// Example 1: Demonstrate structured error handling in registration
	fmt.Println("=== Example 1: Registration Error Handling ===")
	
	// Try to register with invalid data
	_, err = authService.Register(auth.RegisterRequest{
		Username: "", // Empty username should trigger validation error
		Password: "password123",
		Email:    "user@example.com",
	})
	
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("Error Code: %s\n", authErr.Code)
			fmt.Printf("Error Message: %s\n", authErr.Message)
			if authErr.Details != "" {
				fmt.Printf("Error Details: %s\n", authErr.Details)
			}
		} else {
			fmt.Printf("Generic error: %s\n", err.Error())
		}
	}

	// Example 2: Demonstrate structured error handling in login
	fmt.Println("\n=== Example 2: Login Error Handling ===")
	
	// Try to login with invalid credentials
	_, err = authService.Login("nonexistent", "wrongpassword", nil)
	
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("Error Code: %s\n", authErr.Code)
			fmt.Printf("Error Message: %s\n", authErr.Message)
		}
	}

	// Example 3: HTTP Error Response Handling
	fmt.Println("\n=== Example 3: HTTP Error Response Handling ===")
	
	// Create a simple HTTP handler that demonstrates error responses
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			auth.WriteJSONError(w, auth.NewAuthError(auth.ErrCodeValidationError, "Method not allowed"))
			return
		}

		// Simulate login attempt with invalid credentials
		_, err := authService.Login("invalid", "credentials", nil)
		if err != nil {
			auth.WriteJSONError(w, err)
			return
		}

		// This won't be reached in this example
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// Example 4: Demonstrate different error types
	fmt.Println("\n=== Example 4: Different Error Types ===")
	
	// Register a user first
	user, err := authService.Register(auth.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	})
	if err != nil {
		log.Fatal("Failed to register user:", err)
	}

	// Try to register the same user again (should get user exists error)
	_, err = authService.Register(auth.RegisterRequest{
		Username: "testuser",
		Password: "password456",
		Email:    "test2@example.com",
	})
	
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("Duplicate User Error - Code: %s, Message: %s\n", authErr.Code, authErr.Message)
		}
	}

	// Example 5: User management errors
	fmt.Println("\n=== Example 5: User Management Error Handling ===")
	
	users := authService.Users()
	
	// Try to get a non-existent user
	_, err = users.Get("non-existent-id")
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("User Not Found - Code: %s, Message: %s\n", authErr.Code, authErr.Message)
		}
	}

	// Try to change password with wrong old password
	err = users.ChangePassword(user.ID, "wrongoldpassword", "newpassword123")
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("Password Change Error - Code: %s, Message: %s\n", authErr.Code, authErr.Message)
		}
	}

	// Try to use a weak password
	err = users.ChangePassword(user.ID, "password123", "weak")
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("Weak Password Error - Code: %s, Message: %s, Details: %s\n", 
				authErr.Code, authErr.Message, authErr.Details)
		}
	}

	// Example 6: Token management errors
	fmt.Println("\n=== Example 6: Token Management Error Handling ===")
	
	tokens := authService.Tokens()
	
	// Try to refresh an invalid token
	_, err = tokens.Refresh("invalid-refresh-token")
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("Token Refresh Error - Code: %s, Message: %s\n", authErr.Code, authErr.Message)
		}
	}

	// Try to validate an invalid token
	_, err = tokens.Validate("invalid-access-token")
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			fmt.Printf("Token Validation Error - Code: %s, Message: %s\n", authErr.Code, authErr.Message)
		}
	}

	fmt.Println("\n=== Error Handling Examples Complete ===")
	fmt.Println("The new structured error handling system provides:")
	fmt.Println("1. Consistent error codes for programmatic handling")
	fmt.Println("2. Human-readable error messages")
	fmt.Println("3. Optional detailed context without exposing sensitive information")
	fmt.Println("4. Automatic HTTP status code mapping")
	fmt.Println("5. JSON-formatted error responses for APIs")
}