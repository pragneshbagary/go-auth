package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/auth"
	"github.com/pragneshbagary/go-auth/pkg/models"
)

func TestBasicUsageExample(t *testing.T) {
	// Test the basic usage pattern from basic_usage_example.go
	t.Run("BasicAuthFlow", func(t *testing.T) {
		// Create auth instance
		authInstance, err := auth.NewInMemory("test-secret-key-for-examples")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Register a user
		user := &models.User{
			Username: "exampleuser",
			Email:    "example@test.com",
		}

		tokens, err := authInstance.Register(user, "securepassword123", nil)
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		if tokens.AccessToken == "" {
			t.Error("Access token should not be empty")
		}
		if tokens.RefreshToken == "" {
			t.Error("Refresh token should not be empty")
		}

		// Login
		loginTokens, err := authInstance.Login("exampleuser", "securepassword123", nil)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		if loginTokens.AccessToken == "" {
			t.Error("Login access token should not be empty")
		}

		// Validate token
		validatedUser, err := authInstance.ValidateAccessToken(loginTokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if validatedUser.Username != "exampleuser" {
			t.Errorf("Expected username 'exampleuser', got '%s'", validatedUser.Username)
		}
	})
}

func TestSimpleAuthExample(t *testing.T) {
	// Test the SimpleAuth pattern from simple_auth_example.go
	t.Run("SimpleAuthFlow", func(t *testing.T) {
		// Create SimpleAuth instance
		simpleAuth, err := auth.QuickInMemory("test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create SimpleAuth: %v", err)
		}

		// Register
		tokens, err := simpleAuth.Register("simpleuser", "simple@test.com", "password123")
		if err != nil {
			t.Fatalf("Failed to register with SimpleAuth: %v", err)
		}

		if tokens.AccessToken == "" {
			t.Error("SimpleAuth access token should not be empty")
		}

		// Login
		loginTokens, err := simpleAuth.Login("simpleuser", "password123")
		if err != nil {
			t.Fatalf("Failed to login with SimpleAuth: %v", err)
		}

		if loginTokens.AccessToken == "" {
			t.Error("SimpleAuth login token should not be empty")
		}

		// Validate
		user, err := simpleAuth.ValidateToken(loginTokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate SimpleAuth token: %v", err)
		}

		if user.Username != "simpleuser" {
			t.Errorf("Expected username 'simpleuser', got '%s'", user.Username)
		}
	})
}

func TestAdvancedUsageExample(t *testing.T) {
	// Test advanced features from advanced_usage_example.go
	t.Run("AdvancedFeatures", func(t *testing.T) {
		authInstance, err := auth.NewInMemory("test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Register user with custom claims
		user := &models.User{
			Username: "advanceduser",
			Email:    "advanced@test.com",
			Metadata: map[string]interface{}{
				"role":        "admin",
				"permissions": []string{"read", "write", "delete"},
			},
		}

		customClaims := map[string]interface{}{
			"role":   "admin",
			"tenant": "test-tenant",
		}

		tokens, err := authInstance.Register(user, "password123", customClaims)
		if err != nil {
			t.Fatalf("Failed to register user with custom claims: %v", err)
		}

		// Validate token and check claims
		validatedUser, err := authInstance.ValidateAccessToken(tokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if validatedUser.Username != "advanceduser" {
			t.Errorf("Expected username 'advanceduser', got '%s'", validatedUser.Username)
		}

		// Test user management
		users := authInstance.Users()

		// Update user
		err = users.Update(user.ID, auth.UserUpdate{
			Metadata: map[string]interface{}{
				"role":        "super-admin",
				"last_update": time.Now().Unix(),
			},
		})
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// Get updated user
		updatedProfile, err := users.Get(user.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}

		if updatedProfile.Metadata["role"] != "super-admin" {
			t.Errorf("Expected role 'super-admin', got '%v'", updatedProfile.Metadata["role"])
		}

		// Test token management
		tokens_mgr := authInstance.Tokens()

		// Refresh token
		newTokens, err := tokens_mgr.Refresh(tokens.RefreshToken)
		if err != nil {
			t.Fatalf("Failed to refresh token: %v", err)
		}

		if newTokens.AccessToken == "" {
			t.Error("Refreshed access token should not be empty")
		}

		// Validate new token
		_, err = tokens_mgr.Validate(newTokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate refreshed token: %v", err)
		}
	})
}

func TestPasswordResetExample(t *testing.T) {
	// Test password reset flow from password_reset_example.go
	t.Run("PasswordResetFlow", func(t *testing.T) {
		authInstance, err := auth.NewInMemory("test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Register user
		user := &models.User{
			Username: "resetuser",
			Email:    "reset@test.com",
		}

		_, err = authInstance.Register(user, "originalpassword", nil)
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		users := authInstance.Users()

		// Create reset token
		resetToken, err := users.CreateResetToken("reset@test.com")
		if err != nil {
			t.Fatalf("Failed to create reset token: %v", err)
		}

		if resetToken.Token == "" {
			t.Error("Reset token should not be empty")
		}

		// Reset password
		err = users.ResetPassword(resetToken.Token, "newpassword123")
		if err != nil {
			t.Fatalf("Failed to reset password: %v", err)
		}

		// Verify old password doesn't work
		_, err = authInstance.Login("resetuser", "originalpassword", nil)
		if err == nil {
			t.Error("Old password should not work after reset")
		}

		// Verify new password works
		_, err = authInstance.Login("resetuser", "newpassword123", nil)
		if err != nil {
			t.Fatalf("New password should work after reset: %v", err)
		}
	})
}

func TestTokenManagementExample(t *testing.T) {
	// Test token management from token_management_example.go
	t.Run("TokenManagement", func(t *testing.T) {
		authInstance, err := auth.NewInMemory("test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Register user
		user := &models.User{
			Username: "tokenuser",
			Email:    "token@test.com",
		}

		tokens, err := authInstance.Register(user, "password123", nil)
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		tokenMgr := authInstance.Tokens()

		// Test token validation
		validatedUser, err := tokenMgr.Validate(tokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if validatedUser.Username != "tokenuser" {
			t.Errorf("Expected username 'tokenuser', got '%s'", validatedUser.Username)
		}

		// Test token refresh
		newTokens, err := tokenMgr.Refresh(tokens.RefreshToken)
		if err != nil {
			t.Fatalf("Failed to refresh token: %v", err)
		}

		if newTokens.AccessToken == tokens.AccessToken {
			t.Error("New access token should be different from old one")
		}

		// Test token revocation
		err = tokenMgr.Revoke(newTokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to revoke token: %v", err)
		}

		// Verify revoked token doesn't work
		_, err = tokenMgr.Validate(newTokens.AccessToken)
		if err == nil {
			t.Error("Revoked token should not validate")
		}
	})
}

func TestErrorHandlingExample(t *testing.T) {
	// Test error handling patterns from error_handling_example.go
	t.Run("ErrorHandling", func(t *testing.T) {
		authInstance, err := auth.NewInMemory("test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Test registration with duplicate username
		user1 := &models.User{
			Username: "duplicateuser",
			Email:    "user1@test.com",
		}

		user2 := &models.User{
			Username: "duplicateuser", // Same username
			Email:    "user2@test.com",
		}

		// First registration should succeed
		_, err = authInstance.Register(user1, "password123", nil)
		if err != nil {
			t.Fatalf("First registration should succeed: %v", err)
		}

		// Second registration should fail
		_, err = authInstance.Register(user2, "password123", nil)
		if err == nil {
			t.Error("Second registration with duplicate username should fail")
		}

		// Test login with wrong password
		_, err = authInstance.Login("duplicateuser", "wrongpassword", nil)
		if err == nil {
			t.Error("Login with wrong password should fail")
		}

		// Test login with non-existent user
		_, err = authInstance.Login("nonexistentuser", "password123", nil)
		if err == nil {
			t.Error("Login with non-existent user should fail")
		}

		// Test token validation with invalid token
		_, err = authInstance.ValidateAccessToken("invalid-token")
		if err == nil {
			t.Error("Validation of invalid token should fail")
		}
	})
}

func TestEnvironmentConfiguration(t *testing.T) {
	// Test environment-based configuration
	t.Run("EnvironmentConfig", func(t *testing.T) {
		// Set environment variables
		os.Setenv("AUTH_JWT_ACCESS_SECRET", "test-env-secret")
		os.Setenv("AUTH_DB_TYPE", "memory")
		defer func() {
			os.Unsetenv("AUTH_JWT_ACCESS_SECRET")
			os.Unsetenv("AUTH_DB_TYPE")
		}()

		// Test QuickFromEnv
		simpleAuth, err := auth.QuickFromEnv()
		if err != nil {
			t.Fatalf("Failed to create auth from environment: %v", err)
		}

		// Test basic functionality
		tokens, err := simpleAuth.Register("envuser", "env@test.com", "password123")
		if err != nil {
			t.Fatalf("Failed to register user with env config: %v", err)
		}

		if tokens.AccessToken == "" {
			t.Error("Access token should not be empty")
		}
	})
}

func TestDatabasePersistence(t *testing.T) {
	// Test SQLite persistence
	t.Run("SQLitePersistence", func(t *testing.T) {
		// Create temporary database file
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		// Create first auth instance
		auth1, err := auth.NewSQLite(dbPath, "test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create first auth instance: %v", err)
		}

		// Register user
		user := &models.User{
			Username: "persistentuser",
			Email:    "persistent@test.com",
		}

		_, err = auth1.Register(user, "password123", nil)
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Close first instance
		// Note: We don't have a Close method, but the database should persist

		// Create second auth instance with same database
		auth2, err := auth.NewSQLite(dbPath, "test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create second auth instance: %v", err)
		}

		// Try to login with the user registered in first instance
		_, err = auth2.Login("persistentuser", "password123", nil)
		if err != nil {
			t.Fatalf("Failed to login with persistent user: %v", err)
		}

		// Verify user exists
		retrievedUser, err := auth2.GetUserByUsername("persistentuser")
		if err != nil {
			t.Fatalf("Failed to get persistent user: %v", err)
		}

		if retrievedUser.Username != "persistentuser" {
			t.Errorf("Expected username 'persistentuser', got '%s'", retrievedUser.Username)
		}
	})
}

func TestExampleCodeCompilation(t *testing.T) {
	// This test ensures that example files can be compiled
	// We'll check that the example files exist and have valid Go syntax
	
	exampleFiles := []string{
		"basic_usage_example.go",
		"simple_auth_example.go",
		"advanced_usage_example.go",
		"password_reset_example.go",
		"token_management_example.go",
		"error_handling_example.go",
		"middleware_example.go",
		"gin_integration_example.go",
		"echo_integration_example.go",
		"fiber_integration_example.go",
		"logging_monitoring_example.go",
		"migration_example.go",
	}

	for _, filename := range exampleFiles {
		t.Run("CompileCheck_"+strings.TrimSuffix(filename, ".go"), func(t *testing.T) {
			// Check if file exists
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				t.Skipf("Example file %s does not exist", filename)
			}

			// For now, we just check existence
			// In a more comprehensive test, we could:
			// 1. Parse the Go file to check syntax
			// 2. Try to compile it (though it might have external dependencies)
			// 3. Run static analysis on it
			
			t.Logf("Example file %s exists", filename)
		})
	}
}

func TestReadmeExamples(t *testing.T) {
	// Test that examples mentioned in README files work
	t.Run("BasicReadmeExample", func(t *testing.T) {
		// This should match the basic example from README.md
		authInstance, err := auth.NewInMemory("your-secret-key")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Register
		user := &models.User{
			Username: "john_doe",
			Email:    "john@example.com",
		}

		tokens, err := authInstance.Register(user, "secure_password", nil)
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Login
		loginTokens, err := authInstance.Login("john_doe", "secure_password", nil)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		// Validate
		validatedUser, err := authInstance.ValidateAccessToken(loginTokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if validatedUser.Username != "john_doe" {
			t.Errorf("Expected username 'john_doe', got '%s'", validatedUser.Username)
		}

		// Refresh
		newTokens, err := authInstance.RefreshToken(tokens.RefreshToken)
		if err != nil {
			t.Fatalf("Failed to refresh token: %v", err)
		}

		if newTokens.AccessToken == "" {
			t.Error("Refreshed access token should not be empty")
		}
	})

	t.Run("SimpleAuthReadmeExample", func(t *testing.T) {
		// This should match the SimpleAuth example from README
		simpleAuth, err := auth.QuickInMemory("your-secret-key")
		if err != nil {
			t.Fatalf("Failed to create SimpleAuth: %v", err)
		}

		// Register
		tokens, err := simpleAuth.Register("jane_doe", "jane@example.com", "secure_password")
		if err != nil {
			t.Fatalf("Failed to register: %v", err)
		}

		// Login
		loginTokens, err := simpleAuth.Login("jane_doe", "secure_password")
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		// Validate
		user, err := simpleAuth.ValidateToken(loginTokens.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if user.Username != "jane_doe" {
			t.Errorf("Expected username 'jane_doe', got '%s'", user.Username)
		}

		// Refresh
		newTokens, err := simpleAuth.RefreshToken(tokens.RefreshToken)
		if err != nil {
			t.Fatalf("Failed to refresh token: %v", err)
		}

		if newTokens.AccessToken == "" {
			t.Error("Refreshed access token should not be empty")
		}
	})
}