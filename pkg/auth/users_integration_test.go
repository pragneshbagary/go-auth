package auth

import (
	"fmt"
	"testing"
	"time"
)

func TestAuth_Users_Integration(t *testing.T) {
	// Create an in-memory auth instance for testing
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Get the Users component
	users := auth.Users()

	// Register a test user first
	registerReq := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword123",
	}
	
	user, err := auth.Register(registerReq)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	t.Run("get user by ID", func(t *testing.T) {
		profile, err := users.Get(user.ID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if profile.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got %s", profile.Username)
		}
		if profile.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %s", profile.Email)
		}
	})

	t.Run("get user by email", func(t *testing.T) {
		profile, err := users.GetByEmail("test@example.com")
		if err != nil {
			t.Fatalf("Failed to get user by email: %v", err)
		}

		if profile.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, profile.ID)
		}
	})

	t.Run("get user by username", func(t *testing.T) {
		profile, err := users.GetByUsername("testuser")
		if err != nil {
			t.Fatalf("Failed to get user by username: %v", err)
		}

		if profile.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, profile.ID)
		}
	})

	t.Run("update user", func(t *testing.T) {
		newEmail := "newemail@example.com"
		metadata := map[string]interface{}{"role": "admin"}

		updates := UserUpdate{
			Email:    &newEmail,
			Metadata: metadata,
		}

		err := users.Update(user.ID, updates)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// Verify the update
		profile, err := users.Get(user.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}

		if profile.Email != newEmail {
			t.Errorf("Expected email %s, got %s", newEmail, profile.Email)
		}
		if profile.Metadata["role"] != "admin" {
			t.Errorf("Expected role 'admin', got %v", profile.Metadata["role"])
		}
	})

	t.Run("change password", func(t *testing.T) {
		err := users.ChangePassword(user.ID, "testpassword123", "newpassword123")
		if err != nil {
			t.Fatalf("Failed to change password: %v", err)
		}

		// Verify the password change by attempting to login with new password
		_, err = auth.Login("testuser", "newpassword123", nil)
		if err != nil {
			t.Fatalf("Failed to login with new password: %v", err)
		}

		// Verify old password no longer works
		_, err = auth.Login("testuser", "testpassword123", nil)
		if err == nil {
			t.Error("Expected login to fail with old password")
		}
	})

	t.Run("password reset flow", func(t *testing.T) {
		// Create reset token
		resetToken, err := users.CreateResetToken("newemail@example.com")
		if err != nil {
			t.Fatalf("Failed to create reset token: %v", err)
		}

		if resetToken.Token == "" {
			t.Error("Expected non-empty reset token")
		}
		if resetToken.UserID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, resetToken.UserID)
		}
		if resetToken.ExpiresAt.Before(time.Now()) {
			t.Error("Expected token to expire in the future")
		}

		// Reset password using token
		err = users.ResetPassword(resetToken.Token, "resetpassword123")
		if err != nil {
			t.Fatalf("Failed to reset password: %v", err)
		}

		// Verify the password reset by attempting to login with new password
		_, err = auth.Login("testuser", "resetpassword123", nil)
		if err != nil {
			t.Fatalf("Failed to login with reset password: %v", err)
		}
	})

	t.Run("list users", func(t *testing.T) {
		// Register a few more users for listing
		for i := 1; i <= 3; i++ {
			registerReq := RegisterRequest{
				Username: fmt.Sprintf("user%d", i),
				Email:    fmt.Sprintf("user%d@example.com", i),
				Password: "password123",
			}
			_, err := auth.Register(registerReq)
			if err != nil {
				t.Fatalf("Failed to register user %d: %v", i, err)
			}
		}

		// List users
		profiles, err := users.List(10, 0)
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}

		if len(profiles) != 4 { // Original user + 3 new users
			t.Errorf("Expected 4 users, got %d", len(profiles))
		}
	})

	t.Run("delete user", func(t *testing.T) {
		// Register a user to delete
		registerReq := RegisterRequest{
			Username: "deleteuser",
			Email:    "delete@example.com",
			Password: "password123",
		}
		
		userToDelete, err := auth.Register(registerReq)
		if err != nil {
			t.Fatalf("Failed to register user to delete: %v", err)
		}

		// Delete the user
		err = users.Delete(userToDelete.ID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify user was deleted
		_, err = users.Get(userToDelete.ID)
		if err == nil {
			t.Error("Expected error when getting deleted user")
		}
	})
}