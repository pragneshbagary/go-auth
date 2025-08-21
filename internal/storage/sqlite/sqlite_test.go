package sqlite

import (
	"os"
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

func TestSQLiteStorage_EnhancedInterface(t *testing.T) {
	// Create a temporary database file
	dbFile := "test.db"
	defer os.Remove(dbFile)

	s, err := NewSQLiteStorage(dbFile)
	if err != nil {
		t.Fatalf("NewSQLiteStorage failed: %v", err)
	}

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
	err = s.CreateUser(user)
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
	if version < 0 {
		t.Errorf("Expected non-negative schema version, got %d", version)
	}
}