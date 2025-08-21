package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("=== Advanced Usage Example ===")

	// Example 1: Custom configuration with PostgreSQL
	fmt.Println("\n1. Custom Configuration Setup:")
	
	// For this example, we'll use SQLite but show how PostgreSQL would work
	config := &auth.AuthConfig{
		JWTSecret:        "advanced-secret-key",
		JWTRefreshSecret: "advanced-refresh-secret",
		JWTIssuer:        "advanced-app",
		AppName:          "Advanced Auth Example",
		DatabasePath:     "advanced_example.db", // For SQLite
		// DatabaseURL:   "postgres://user:pass@localhost/authdb", // For PostgreSQL
	}
	
	authService, err := auth.NewWithConfig(config)
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}
	fmt.Println("✓ Auth service created with custom configuration")

	// Example 2: Advanced user management
	fmt.Println("\n2. Advanced User Management:")
	
	// Register user with metadata
	user, err := authService.Register(auth.RegisterRequest{
		Username: "bob",
		Email:    "bob@company.com",
		Password: "complex_password_123!",
	})
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("✓ User registered: %s\n", user.Username)

	// Get Users component for advanced operations
	users := authService.Users()

	// Update user profile with metadata
	newEmail := "bob.smith@company.com"
	metadata := map[string]interface{}{
		"role":        "manager",
		"department":  "sales",
		"hire_date":   "2024-01-15",
		"permissions": []string{"read", "write", "approve"},
	}
	
	err = users.Update(user.ID, auth.UserUpdate{
		Email:    &newEmail,
		Metadata: metadata,
	})
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	fmt.Println("✓ User profile updated with metadata")

	// Example 3: Password reset workflow
	fmt.Println("\n3. Password Reset Workflow:")
	
	// Create password reset token
	resetToken, err := users.CreateResetToken("bob.smith@company.com")
	if err != nil {
		log.Fatalf("Failed to create reset token: %v", err)
	}
	fmt.Printf("✓ Reset token created (expires: %s)\n", resetToken.ExpiresAt.Format("2006-01-02 15:04:05"))
	
	// Simulate sending email with reset token (in real app, you'd send an email)
	fmt.Printf("  Reset token: %s\n", resetToken.Token)
	
	// Reset password using token
	err = users.ResetPassword(resetToken.Token, "new_secure_password_456!")
	if err != nil {
		log.Fatalf("Failed to reset password: %v", err)
	}
	fmt.Println("✓ Password reset successfully")

	// Example 4: Advanced token management
	fmt.Println("\n4. Advanced Token Management:")
	
	// Login with new password
	loginResult, err := authService.Login("bob", "new_secure_password_456!", map[string]interface{}{
		"role":       "manager",
		"session_id": "sess_" + fmt.Sprintf("%d", time.Now().Unix()),
	})
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Println("✓ Login successful with new password")

	// Get Tokens component for advanced operations
	tokens := authService.Tokens()

	// Validate token and get session info
	sessionInfo, err := tokens.GetSessionInfo(loginResult.AccessToken)
	if err != nil {
		log.Fatalf("Failed to get session info: %v", err)
	}
	fmt.Printf("✓ Session info retrieved:\n")
	fmt.Printf("  Token ID: %s\n", sessionInfo.TokenID)
	fmt.Printf("  User ID: %s\n", sessionInfo.UserID)
	fmt.Printf("  Token Type: %s\n", sessionInfo.TokenType)

	// Batch token validation
	testTokens := []string{
		loginResult.AccessToken,
		"invalid-token-example",
		loginResult.RefreshToken,
	}
	
	results := tokens.ValidateBatch(testTokens)
	fmt.Println("✓ Batch validation results:")
	for i, result := range results {
		if result.Valid {
			fmt.Printf("  Token %d: Valid (User: %s)\n", i+1, result.User.Username)
		} else {
			fmt.Printf("  Token %d: Invalid (%s)\n", i+1, result.Error)
		}
	}

	// Example 5: User listing and management
	fmt.Println("\n5. User Listing and Management:")
	
	// Register additional users
	for i := 1; i <= 3; i++ {
		_, err := authService.Register(auth.RegisterRequest{
			Username: fmt.Sprintf("employee%d", i),
			Email:    fmt.Sprintf("employee%d@company.com", i),
			Password: "employee_password_123",
		})
		if err != nil {
			log.Printf("Failed to register employee%d: %v", i, err)
		}
	}

	// List all users
	userList, err := users.List(10, 0)
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}
	fmt.Printf("✓ Found %d users:\n", len(userList))
	for i, u := range userList {
		fmt.Printf("  %d. %s <%s>\n", i+1, u.Username, u.Email)
	}

	// Example 6: Token revocation and cleanup
	fmt.Println("\n6. Token Revocation and Cleanup:")
	
	// Revoke specific token
	err = tokens.Revoke(loginResult.AccessToken)
	if err != nil {
		log.Fatalf("Failed to revoke token: %v", err)
	}
	fmt.Println("✓ Access token revoked")

	// Verify token is now invalid
	isValid := tokens.IsValid(loginResult.AccessToken)
	fmt.Printf("✓ Revoked token validity: %t\n", isValid)

	// Login again to get new tokens
	newLogin, err := authService.Login("bob", "new_secure_password_456!", nil)
	if err != nil {
		log.Fatalf("Failed to login again: %v", err)
	}

	// Revoke all tokens for user (logout from all devices)
	err = tokens.RevokeAll(user.ID)
	if err != nil {
		log.Fatalf("Failed to revoke all tokens: %v", err)
	}
	fmt.Println("✓ All tokens for user revoked")

	// Clean up expired tokens
	err = tokens.CleanupExpired()
	if err != nil {
		log.Fatalf("Failed to cleanup expired tokens: %v", err)
	}
	fmt.Println("✓ Expired tokens cleaned up")

	// Example 7: Monitoring and metrics
	fmt.Println("\n7. Monitoring and Metrics:")
	
	// Get current metrics
	metrics := authService.GetMetrics()
	fmt.Printf("✓ Current metrics:\n")
	fmt.Printf("  Registration attempts: %d\n", metrics.RegistrationAttempts)
	fmt.Printf("  Login attempts: %d\n", metrics.LoginAttempts)
	fmt.Printf("  Token validations: %d\n", metrics.TokenValidations)
	fmt.Printf("  Token revocations: %d\n", metrics.TokenRevocations)

	// Get success rates
	collector := authService.MetricsCollector()
	loginRate := collector.GetLoginSuccessRate()
	fmt.Printf("  Login success rate: %.1f%%\n", loginRate)

	// System information
	info := authService.GetSystemInfo()
	fmt.Printf("✓ System info:\n")
	fmt.Printf("  App name: %s\n", info.AppName)
	fmt.Printf("  Version: %s\n", info.Version)
	fmt.Printf("  Uptime: %s\n", info.Uptime)

	fmt.Println("\n=== Advanced Usage Example Complete ===")
}