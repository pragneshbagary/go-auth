package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("=== Advanced Token Management Example ===")

	// Initialize auth service
	authService, err := auth.NewInMemory("token-management-secret")
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Get Tokens component for advanced token operations
	tokens := authService.Tokens()

	// Example 1: Register users and login
	fmt.Println("\n1. Setup: Register users and login:")
	
	// Register first user
	user1, err := authService.Register(auth.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "alice_password_123",
	})
	if err != nil {
		log.Fatalf("Failed to register alice: %v", err)
	}

	// Register second user
	user2, err := authService.Register(auth.RegisterRequest{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "bob_password_123",
	})
	if err != nil {
		log.Fatalf("Failed to register bob: %v", err)
	}

	// Login alice with custom claims
	aliceLogin, err := authService.Login("alice", "alice_password_123", map[string]interface{}{
		"role":       "admin",
		"department": "engineering",
		"session_id": "alice_session_1",
	})
	if err != nil {
		log.Fatalf("Failed to login alice: %v", err)
	}
	fmt.Printf("✓ Alice logged in: %s...\n", aliceLogin.AccessToken[:30])

	// Login bob with different claims
	bobLogin, err := authService.Login("bob", "bob_password_123", map[string]interface{}{
		"role":       "user",
		"department": "sales",
		"session_id": "bob_session_1",
	})
	if err != nil {
		log.Fatalf("Failed to login bob: %v", err)
	}
	fmt.Printf("✓ Bob logged in: %s...\n", bobLogin.AccessToken[:30])

	// Example 2: Token validation and session info
	fmt.Println("\n2. Token validation and session info:")
	
	// Validate alice's token
	aliceUser, err := tokens.Validate(aliceLogin.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate alice's token: %v", err)
	}
	fmt.Printf("✓ Alice's token validated for user: %s\n", aliceUser.Username)

	// Get session info
	aliceSession, err := tokens.GetSessionInfo(aliceLogin.AccessToken)
	if err != nil {
		log.Fatalf("Failed to get alice's session info: %v", err)
	}
	fmt.Printf("✓ Alice's session info:\n")
	fmt.Printf("  Token ID: %s\n", aliceSession.TokenID)
	fmt.Printf("  User ID: %s\n", aliceSession.UserID)
	fmt.Printf("  Token Type: %s\n", aliceSession.TokenType)

	// Quick validation check
	isValid := tokens.IsValid(aliceLogin.AccessToken)
	fmt.Printf("✓ Alice's token is valid: %t\n", isValid)

	// Example 3: Batch token validation
	fmt.Println("\n3. Batch token validation:")
	
	testTokens := []string{
		aliceLogin.AccessToken,
		bobLogin.AccessToken,
		"invalid-token-example",
		aliceLogin.RefreshToken,
	}

	results := tokens.ValidateBatch(testTokens)
	fmt.Printf("✓ Batch validation results:\n")
	for i, result := range results {
		if result.Valid {
			fmt.Printf("  Token %d: ✓ Valid (User: %s)\n", i+1, result.User.Username)
		} else {
			fmt.Printf("  Token %d: ✗ Invalid (%s)\n", i+1, result.Error)
		}
	}

	// Example 4: Token refresh with automatic rotation
	fmt.Println("\n4. Token refresh with automatic rotation:")
	
	// Refresh alice's tokens
	aliceRefresh, err := tokens.Refresh(aliceLogin.RefreshToken)
	if err != nil {
		log.Fatalf("Failed to refresh alice's tokens: %v", err)
	}
	fmt.Printf("✓ Alice's tokens refreshed:\n")
	fmt.Printf("  New Access Token: %s...\n", aliceRefresh.AccessToken[:30])
	fmt.Printf("  New Refresh Token: %s...\n", aliceRefresh.RefreshToken[:30])

	// Verify old refresh token is now invalid
	_, err = tokens.Refresh(aliceLogin.RefreshToken)
	if err != nil {
		fmt.Printf("✓ Old refresh token correctly invalidated: %v\n", err)
	}

	// Verify new access token works
	_, err = tokens.Validate(aliceRefresh.AccessToken)
	if err != nil {
		log.Fatalf("New access token should be valid: %v", err)
	}
	fmt.Println("✓ New access token is valid")

	// Example 5: Token revocation
	fmt.Println("\n5. Token revocation:")
	
	// Revoke alice's current access token
	err = tokens.Revoke(aliceRefresh.AccessToken)
	if err != nil {
		log.Fatalf("Failed to revoke alice's token: %v", err)
	}
	fmt.Println("✓ Alice's access token revoked")

	// Verify token is now invalid
	isValid = tokens.IsValid(aliceRefresh.AccessToken)
	fmt.Printf("✓ Revoked token validity: %t\n", isValid)

	// Try to validate revoked token
	_, err = tokens.Validate(aliceRefresh.AccessToken)
	if err != nil {
		fmt.Printf("✓ Revoked token correctly rejected: %v\n", err)
	}

	// Example 6: Multiple sessions for same user
	fmt.Println("\n6. Multiple sessions for same user:")
	
	// Create multiple sessions for alice
	aliceSession2, err := authService.Login("alice", "alice_password_123", map[string]interface{}{
		"role":       "admin",
		"session_id": "alice_session_2",
		"device":     "mobile",
	})
	if err != nil {
		log.Fatalf("Failed to create alice's second session: %v", err)
	}

	aliceSession3, err := authService.Login("alice", "alice_password_123", map[string]interface{}{
		"role":       "admin",
		"session_id": "alice_session_3",
		"device":     "tablet",
	})
	if err != nil {
		log.Fatalf("Failed to create alice's third session: %v", err)
	}

	fmt.Printf("✓ Created multiple sessions for alice:\n")
	fmt.Printf("  Session 2 (mobile): %s...\n", aliceSession2.AccessToken[:30])
	fmt.Printf("  Session 3 (tablet): %s...\n", aliceSession3.AccessToken[:30])

	// List active sessions
	activeSessions, err := tokens.ListActiveSessions(user1.ID)
	if err != nil {
		log.Fatalf("Failed to list active sessions: %v", err)
	}
	fmt.Printf("✓ Alice has %d active sessions\n", len(activeSessions))

	// Example 7: Revoke all tokens for a user
	fmt.Println("\n7. Revoke all tokens for a user:")
	
	// Revoke all of alice's tokens
	err = tokens.RevokeAll(user1.ID)
	if err != nil {
		log.Fatalf("Failed to revoke all tokens for alice: %v", err)
	}
	fmt.Println("✓ All tokens revoked for alice")

	// Verify all tokens are now invalid
	fmt.Printf("✓ Verifying all alice's tokens are invalid:\n")
	
	tokens_to_check := []string{
		aliceRefresh.RefreshToken,
		aliceSession2.AccessToken,
		aliceSession3.AccessToken,
	}

	for i, token := range tokens_to_check {
		isValid := tokens.IsValid(token)
		fmt.Printf("  Token %d valid: %t\n", i+1, isValid)
	}

	// Example 8: Token cleanup operations
	fmt.Println("\n8. Token cleanup operations:")
	
	// Clean up expired tokens
	err = tokens.CleanupExpired()
	if err != nil {
		log.Fatalf("Failed to cleanup expired tokens: %v", err)
	}
	fmt.Println("✓ Expired tokens cleaned up")

	// Example 9: Session management
	fmt.Println("\n9. Session management:")
	
	// Create new session for alice after revocation
	aliceNewSession, err := authService.Login("alice", "alice_password_123", map[string]interface{}{
		"role":       "admin",
		"session_id": "alice_new_session",
		"login_time": time.Now().Unix(),
	})
	if err != nil {
		log.Fatalf("Failed to create new session for alice: %v", err)
	}
	fmt.Printf("✓ New session created for alice: %s...\n", aliceNewSession.AccessToken[:30])

	// Get session details
	sessionInfo, err := tokens.GetSessionInfo(aliceNewSession.AccessToken)
	if err != nil {
		log.Fatalf("Failed to get session info: %v", err)
	}
	fmt.Printf("✓ Session details:\n")
	fmt.Printf("  Token ID: %s\n", sessionInfo.TokenID)
	fmt.Printf("  User ID: %s\n", sessionInfo.UserID)
	fmt.Printf("  Token Type: %s\n", sessionInfo.TokenType)

	// Example 10: Token security features
	fmt.Println("\n10. Token security features:")
	
	// Demonstrate token blacklisting
	fmt.Println("✓ Security features demonstrated:")
	fmt.Println("  - Automatic token rotation on refresh")
	fmt.Println("  - Token blacklisting on revocation")
	fmt.Println("  - Single-use refresh tokens")
	fmt.Println("  - Batch validation for performance")
	fmt.Println("  - Session tracking and management")
	fmt.Println("  - Expired token cleanup")

	// Example 11: Performance considerations
	fmt.Println("\n11. Performance considerations:")
	
	// Batch validate multiple tokens for performance
	performanceTokens := []string{
		aliceNewSession.AccessToken,
		bobLogin.AccessToken,
	}

	start := time.Now()
	batchResults := tokens.ValidateBatch(performanceTokens)
	batchDuration := time.Since(start)

	fmt.Printf("✓ Batch validation performance:\n")
	fmt.Printf("  Validated %d tokens in %v\n", len(performanceTokens), batchDuration)
	fmt.Printf("  Results: %d valid, %d invalid\n", 
		countValidTokens(batchResults), 
		len(batchResults)-countValidTokens(batchResults))

	// Example 12: Monitoring and metrics
	fmt.Println("\n12. Token metrics:")
	
	// Get current metrics
	metrics := authService.GetMetrics()
	fmt.Printf("✓ Token-related metrics:\n")
	fmt.Printf("  Tokens generated: %d\n", metrics.TokensGenerated)
	fmt.Printf("  Token refreshes: %d\n", metrics.TokenRefreshes)
	fmt.Printf("  Token validations: %d\n", metrics.TokenValidations)
	fmt.Printf("  Token revocations: %d\n", metrics.TokenRevocations)
	fmt.Printf("  Failed validations: %d\n", metrics.TokenValidationFail)

	// Get success rates
	collector := authService.MetricsCollector()
	tokenSuccessRate := collector.GetTokenValidationSuccessRate()
	fmt.Printf("  Token validation success rate: %.1f%%\n", tokenSuccessRate)

	fmt.Println("\n=== Advanced Token Management Example Complete ===")
}

// Helper function to count valid tokens in batch results
func countValidTokens(results []auth.BatchValidationResult) int {
	count := 0
	for _, result := range results {
		if result.Valid {
			count++
		}
	}
	return count
}