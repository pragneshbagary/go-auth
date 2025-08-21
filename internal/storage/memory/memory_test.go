package memory

import (
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

func TestInMemoryStorage_EnhancedInterface(t *testing.T) {
	s := NewInMemoryStorage()

	// Test that it implements the EnhancedStorage interface
	var _ storage.EnhancedStorage = s

	// Create a test user
	user := models.User{
		ID:           "test-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		IsActive:     true,
		Metadata:     map[string]interface{}{"role": "user"},
	}

	// Test CreateUser
	err := s.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Test GetUserByID
	retrievedUser, err := s.GetUserByID("test-id")
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if retrievedUser.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrievedUser.Username)
	}

	// Test GetUserByEmail
	retrievedUser, err = s.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if retrievedUser.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", retrievedUser.ID)
	}

	// Test UpdateUser
	newEmail := "newemail@example.com"
	updates := storage.UserUpdates{
		Email: &newEmail,
	}
	err = s.UpdateUser("test-id", updates)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// Verify update
	retrievedUser, err = s.GetUserByID("test-id")
	if err != nil {
		t.Fatalf("GetUserByID after update failed: %v", err)
	}
	if retrievedUser.Email != "newemail@example.com" {
		t.Errorf("Expected email 'newemail@example.com', got '%s'", retrievedUser.Email)
	}

	// Test UpdatePassword
	err = s.UpdatePassword("test-id", "newhashedpassword")
	if err != nil {
		t.Fatalf("UpdatePassword failed: %v", err)
	}

	// Test BlacklistToken
	tokenID := "test-token-id"
	expiresAt := time.Now().Add(time.Hour)
	err = s.BlacklistToken(tokenID, expiresAt)
	if err != nil {
		t.Fatalf("BlacklistToken failed: %v", err)
	}

	// Test IsTokenBlacklisted
	isBlacklisted, err := s.IsTokenBlacklisted(tokenID)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted failed: %v", err)
	}
	if !isBlacklisted {
		t.Error("Expected token to be blacklisted")
	}

	// Test ListUsers
	users, err := s.ListUsers(10, 0)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	// Test Ping
	err = s.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Test GetSchemaVersion
	version, err := s.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion failed: %v", err)
	}
	if version != 1 {
		t.Errorf("Expected schema version 1, got %d", version)
	}

	// Test DeleteUser
	err = s.DeleteUser("test-id")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify deletion
	_, err = s.GetUserByID("test-id")
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestInMemoryStorage_CleanupExpiredTokens(t *testing.T) {
	s := NewInMemoryStorage()

	// Add an expired token
	expiredTokenID := "expired-token"
	expiredTime := time.Now().Add(-time.Hour)
	err := s.BlacklistToken(expiredTokenID, expiredTime)
	if err != nil {
		t.Fatalf("BlacklistToken failed: %v", err)
	}

	// Add a valid token
	validTokenID := "valid-token"
	validTime := time.Now().Add(time.Hour)
	err = s.BlacklistToken(validTokenID, validTime)
	if err != nil {
		t.Fatalf("BlacklistToken failed: %v", err)
	}

	// Cleanup expired tokens
	err = s.CleanupExpiredTokens()
	if err != nil {
		t.Fatalf("CleanupExpiredTokens failed: %v", err)
	}

	// Check that expired token is no longer blacklisted
	isBlacklisted, err := s.IsTokenBlacklisted(expiredTokenID)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted failed: %v", err)
	}
	if isBlacklisted {
		t.Error("Expected expired token to not be blacklisted after cleanup")
	}

	// Check that valid token is still blacklisted
	isBlacklisted, err = s.IsTokenBlacklisted(validTokenID)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted failed: %v", err)
	}
	if !isBlacklisted {
		t.Error("Expected valid token to still be blacklisted after cleanup")
	}
}