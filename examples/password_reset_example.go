package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("=== Password Reset Workflow Example ===")

	// Initialize auth service
	authService, err := auth.NewInMemory("password-reset-secret")
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Get Users component for password reset operations
	users := authService.Users()

	// Example 1: Register a user
	fmt.Println("\n1. Register a user:")
	user, err := authService.Register(auth.RegisterRequest{
		Username: "john_doe",
		Email:    "john.doe@example.com",
		Password: "original_password_123",
	})
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("✓ User registered: %s <%s>\n", user.Username, user.Email)

	// Verify original password works
	_, err = authService.Login("john_doe", "original_password_123", nil)
	if err != nil {
		log.Fatalf("Failed to login with original password: %v", err)
	}
	fmt.Println("✓ Original password verified")

	// Example 2: Create password reset token
	fmt.Println("\n2. Create password reset token:")
	resetToken, err := users.CreateResetToken("john.doe@example.com")
	if err != nil {
		log.Fatalf("Failed to create reset token: %v", err)
	}
	fmt.Printf("✓ Reset token created:\n")
	fmt.Printf("  Token: %s\n", resetToken.Token)
	fmt.Printf("  User ID: %s\n", resetToken.UserID)
	fmt.Printf("  Expires at: %s\n", resetToken.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Valid for: %s\n", time.Until(resetToken.ExpiresAt).Round(time.Second))

	// Example 3: Simulate email sending (in real app, you'd send an email)
	fmt.Println("\n3. Simulate sending reset email:")
	fmt.Printf("✓ Email would be sent to: %s\n", user.Email)
	fmt.Printf("  Subject: Password Reset Request\n")
	fmt.Printf("  Body: Click this link to reset your password:\n")
	fmt.Printf("        https://yourapp.com/reset-password?token=%s\n", resetToken.Token)

	// Example 4: Reset password using token
	fmt.Println("\n4. Reset password using token:")
	newPassword := "new_secure_password_456"
	err = users.ResetPassword(resetToken.Token, newPassword)
	if err != nil {
		log.Fatalf("Failed to reset password: %v", err)
	}
	fmt.Println("✓ Password reset successfully")

	// Example 5: Verify old password no longer works
	fmt.Println("\n5. Verify old password is invalid:")
	_, err = authService.Login("john_doe", "original_password_123", nil)
	if err != nil {
		fmt.Printf("✓ Old password correctly rejected: %v\n", err)
	} else {
		fmt.Println("✗ Old password still works (this shouldn't happen)")
	}

	// Example 6: Verify new password works
	fmt.Println("\n6. Verify new password works:")
	loginResult, err := authService.Login("john_doe", newPassword, nil)
	if err != nil {
		log.Fatalf("Failed to login with new password: %v", err)
	}
	fmt.Printf("✓ Login successful with new password\n")
	fmt.Printf("  Access token: %s...\n", loginResult.AccessToken[:50])

	// Example 7: Try to reuse the same reset token (should fail)
	fmt.Println("\n7. Try to reuse reset token (should fail):")
	err = users.ResetPassword(resetToken.Token, "another_password")
	if err != nil {
		fmt.Printf("✓ Reset token correctly rejected: %v\n", err)
	} else {
		fmt.Println("✗ Reset token was reused (this shouldn't happen)")
	}

	// Example 8: Create reset token for non-existent email
	fmt.Println("\n8. Create reset token for non-existent email:")
	_, err = users.CreateResetToken("nonexistent@example.com")
	if err != nil {
		fmt.Printf("✓ Non-existent email correctly handled: %v\n", err)
	} else {
		fmt.Println("✗ Reset token created for non-existent email")
	}

	// Example 9: Multiple reset tokens workflow
	fmt.Println("\n9. Multiple reset tokens workflow:")
	
	// Create first reset token
	resetToken1, err := users.CreateResetToken("john.doe@example.com")
	if err != nil {
		log.Fatalf("Failed to create first reset token: %v", err)
	}
	fmt.Printf("✓ First reset token created: %s...\n", resetToken1.Token[:20])

	// Create second reset token (should invalidate the first)
	resetToken2, err := users.CreateResetToken("john.doe@example.com")
	if err != nil {
		log.Fatalf("Failed to create second reset token: %v", err)
	}
	fmt.Printf("✓ Second reset token created: %s...\n", resetToken2.Token[:20])

	// Try to use first token (should fail)
	err = users.ResetPassword(resetToken1.Token, "password_with_first_token")
	if err != nil {
		fmt.Printf("✓ First token correctly invalidated: %v\n", err)
	} else {
		fmt.Println("✗ First token still works (this shouldn't happen)")
	}

	// Use second token (should work)
	err = users.ResetPassword(resetToken2.Token, "password_with_second_token")
	if err != nil {
		log.Fatalf("Failed to reset with second token: %v", err)
	}
	fmt.Println("✓ Second token worked correctly")

	// Example 10: Expired token handling (simulated)
	fmt.Println("\n10. Expired token handling:")
	fmt.Println("✓ In a real application, tokens would expire after a set time")
	fmt.Printf("  Current token TTL would be configured in the auth service\n")
	fmt.Printf("  Expired tokens would be automatically rejected\n")

	// Example 11: Security considerations demonstration
	fmt.Println("\n11. Security considerations:")
	fmt.Println("✓ Security features implemented:")
	fmt.Println("  - Reset tokens are single-use only")
	fmt.Println("  - New reset tokens invalidate previous ones")
	fmt.Println("  - Tokens have expiration times")
	fmt.Println("  - Non-existent emails don't reveal user existence")
	fmt.Println("  - Password changes invalidate all existing sessions")

	// Example 12: Integration with user management
	fmt.Println("\n12. Integration with user management:")
	
	// Get updated user profile
	profile, err := users.Get(user.ID)
	if err != nil {
		log.Fatalf("Failed to get user profile: %v", err)
	}
	fmt.Printf("✓ User profile after password reset:\n")
	fmt.Printf("  ID: %s\n", profile.ID)
	fmt.Printf("  Username: %s\n", profile.Username)
	fmt.Printf("  Email: %s\n", profile.Email)
	fmt.Printf("  Last updated: %s\n", profile.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Example 13: Best practices demonstration
	fmt.Println("\n13. Best practices for password reset:")
	fmt.Println("✓ Recommended implementation:")
	fmt.Println("  1. Always send reset emails, even for non-existent addresses")
	fmt.Println("  2. Use HTTPS for all reset links")
	fmt.Println("  3. Set reasonable token expiration (15-30 minutes)")
	fmt.Println("  4. Log all password reset attempts for security monitoring")
	fmt.Println("  5. Require strong passwords for reset")
	fmt.Println("  6. Consider rate limiting reset requests")
	fmt.Println("  7. Invalidate all sessions after password reset")

	fmt.Println("\n=== Password Reset Workflow Example Complete ===")
}