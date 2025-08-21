package main

import (
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Create an in-memory auth instance for this example
	authService, err := auth.NewInMemory("your-secret-key")
	if err != nil {
		log.Fatal("Failed to create auth service:", err)
	}

	// Get the Users component
	users := authService.Users()

	fmt.Println("=== User Management Example ===")

	// 1. Register a user first
	fmt.Println("\n1. Registering a user...")
	user, err := authService.Register(auth.RegisterRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "securepassword123",
	})
	if err != nil {
		log.Fatal("Failed to register user:", err)
	}
	fmt.Printf("✓ User registered: %s (ID: %s)\n", user.Username, user.ID)

	// 2. Get user by ID
	fmt.Println("\n2. Getting user by ID...")
	profile, err := users.Get(user.ID)
	if err != nil {
		log.Fatal("Failed to get user:", err)
	}
	fmt.Printf("✓ Found user: %s <%s>\n", profile.Username, profile.Email)

	// 3. Get user by email
	fmt.Println("\n3. Getting user by email...")
	profile, err = users.GetByEmail("john@example.com")
	if err != nil {
		log.Fatal("Failed to get user by email:", err)
	}
	fmt.Printf("✓ Found user by email: %s\n", profile.Username)

	// 4. Get user by username
	fmt.Println("\n4. Getting user by username...")
	profile, err = users.GetByUsername("johndoe")
	if err != nil {
		log.Fatal("Failed to get user by username:", err)
	}
	fmt.Printf("✓ Found user by username: %s\n", profile.Email)

	// 5. Update user profile
	fmt.Println("\n5. Updating user profile...")
	newEmail := "john.doe@example.com"
	metadata := map[string]interface{}{
		"role":        "admin",
		"department":  "engineering",
		"last_active": "2024-01-15",
	}
	
	err = users.Update(user.ID, auth.UserUpdate{
		Email:    &newEmail,
		Metadata: metadata,
	})
	if err != nil {
		log.Fatal("Failed to update user:", err)
	}
	fmt.Printf("✓ User updated successfully\n")

	// Verify the update
	profile, err = users.Get(user.ID)
	if err != nil {
		log.Fatal("Failed to get updated user:", err)
	}
	fmt.Printf("  New email: %s\n", profile.Email)
	fmt.Printf("  Role: %s\n", profile.Metadata["role"])

	// 6. Change password
	fmt.Println("\n6. Changing user password...")
	err = users.ChangePassword(user.ID, "securepassword123", "newsecurepassword456")
	if err != nil {
		log.Fatal("Failed to change password:", err)
	}
	fmt.Printf("✓ Password changed successfully\n")

	// Verify password change by attempting login
	_, err = authService.Login("johndoe", "newsecurepassword456", nil)
	if err != nil {
		log.Fatal("Failed to login with new password:", err)
	}
	fmt.Printf("✓ Login with new password successful\n")

	// 7. Password reset flow
	fmt.Println("\n7. Demonstrating password reset flow...")
	
	// Create reset token
	resetToken, err := users.CreateResetToken("john.doe@example.com")
	if err != nil {
		log.Fatal("Failed to create reset token:", err)
	}
	fmt.Printf("✓ Reset token created (expires at: %s)\n", resetToken.ExpiresAt.Format("2006-01-02 15:04:05"))
	
	// Reset password using token
	err = users.ResetPassword(resetToken.Token, "resetpassword789")
	if err != nil {
		log.Fatal("Failed to reset password:", err)
	}
	fmt.Printf("✓ Password reset successfully\n")

	// Verify password reset by attempting login
	_, err = authService.Login("johndoe", "resetpassword789", nil)
	if err != nil {
		log.Fatal("Failed to login with reset password:", err)
	}
	fmt.Printf("✓ Login with reset password successful\n")

	// 8. Register more users for listing demonstration
	fmt.Println("\n8. Registering additional users for listing...")
	for i := 1; i <= 3; i++ {
		_, err := authService.Register(auth.RegisterRequest{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "password123",
		})
		if err != nil {
			log.Printf("Failed to register user%d: %v", i, err)
			continue
		}
		fmt.Printf("✓ Registered user%d\n", i)
	}

	// 9. List users
	fmt.Println("\n9. Listing users...")
	profiles, err := users.List(10, 0)
	if err != nil {
		log.Fatal("Failed to list users:", err)
	}
	fmt.Printf("✓ Found %d users:\n", len(profiles))
	for i, p := range profiles {
		fmt.Printf("  %d. %s <%s> (ID: %s)\n", i+1, p.Username, p.Email, p.ID)
	}

	// 10. Delete a user
	fmt.Println("\n10. Deleting a user...")
	
	// Register a user to delete
	userToDelete, err := authService.Register(auth.RegisterRequest{
		Username: "tempuser",
		Email:    "temp@example.com",
		Password: "temppassword123",
	})
	if err != nil {
		log.Fatal("Failed to register temp user:", err)
	}
	fmt.Printf("✓ Created temp user: %s\n", userToDelete.Username)

	// Delete the user
	err = users.Delete(userToDelete.ID)
	if err != nil {
		log.Fatal("Failed to delete user:", err)
	}
	fmt.Printf("✓ User deleted successfully\n")

	// Verify deletion
	_, err = users.Get(userToDelete.ID)
	if err != nil {
		fmt.Printf("✓ Confirmed user deletion (user not found)\n")
	} else {
		fmt.Printf("✗ User still exists after deletion\n")
	}

	fmt.Println("\n=== User Management Example Complete ===")
}