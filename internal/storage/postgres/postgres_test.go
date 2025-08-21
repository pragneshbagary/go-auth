package postgres

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
	_ "github.com/lib/pq"
)

func getTestPostgresURL() string {
	url := os.Getenv("POSTGRES_TEST_URL")
	if url == "" {
		// Default test database URL - skip tests if not available
		url = "postgres://postgres:password@localhost:5432/auth_test?sslmode=disable"
	}
	return url
}

func setupTestDB(t *testing.T) *PostgresStorage {
	url := getTestPostgresURL()
	
	// Try to connect to test database
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Skipf("Skipping PostgreSQL tests: %v", err)
	}
	
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping PostgreSQL tests - database not available: %v", err)
	}
	db.Close()

	storage, err := NewPostgresStorage(url)
	if err != nil {
		t.Skipf("Skipping PostgreSQL tests: %v", err)
	}

	// Clean up any existing test data
	cleanupTestData(t, storage)

	return storage
}

func cleanupTestData(t *testing.T, storage *PostgresStorage) {
	// Clean up test data
	_, err := storage.db.Exec("DELETE FROM users WHERE username LIKE 'test%' OR email LIKE 'test%'")
	if err != nil {
		t.Logf("Warning: failed to clean up test data: %v", err)
	}
	
	_, err = storage.db.Exec("DELETE FROM blacklisted_tokens WHERE token_id LIKE 'test%'")
	if err != nil {
		t.Logf("Warning: failed to clean up blacklisted tokens: %v", err)
	}
}

func TestPostgresStorage_BasicOperations(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()
	defer cleanupTestData(t, storage)

	user := &models.User{
		ID:           "test-user-1",
		Username:     "testuser1",
		Email:        "test1@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
		Metadata:     map[string]interface{}{"role": "user"},
	}

	t.Run("CreateUser", func(t *testing.T) {
		err := storage.CreateUser(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	})

	t.Run("GetUserByID", func(t *testing.T) {
		retrievedUser, err := storage.GetUserByID(user.ID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}
		if retrievedUser.Username != user.Username {
			t.Errorf("Expected username %s, got %s", user.Username, retrievedUser.Username)
		}
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		retrievedUser, err := storage.GetUserByUsername(user.Username)
		if err != nil {
			t.Fatalf("Failed to get user by username: %v", err)
		}
		if retrievedUser.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, retrievedUser.ID)
		}
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		retrievedUser, err := storage.GetUserByEmail(user.Email)
		if err != nil {
			t.Fatalf("Failed to get user by email: %v", err)
		}
		if retrievedUser.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, retrievedUser.ID)
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		updates := models.UserUpdates{
			Email:    &[]string{"updated@example.com"}[0],
			Metadata: map[string]interface{}{"role": "admin", "updated": true},
		}
		
		err := storage.UpdateUser(user.ID, updates)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// Verify update
		updatedUser, err := storage.GetUserByID(user.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}
		if updatedUser.Email != "updated@example.com" {
			t.Errorf("Expected email updated@example.com, got %s", updatedUser.Email)
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		users, err := storage.ListUsers(10, 0)
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}
		if len(users) == 0 {
			t.Error("Expected at least one user")
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		err := storage.DeleteUser(user.ID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify deletion
		_, err = storage.GetUserByID(user.ID)
		if err == nil {
			t.Error("Expected error when getting deleted user")
		}
	})
}

func TestPostgresStorage_TokenBlacklist(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()
	defer cleanupTestData(t, storage)

	tokenID := "test-token-1"
	expiresAt := time.Now().Add(time.Hour)

	t.Run("BlacklistToken", func(t *testing.T) {
		err := storage.BlacklistToken(tokenID, expiresAt)
		if err != nil {
			t.Fatalf("Failed to blacklist token: %v", err)
		}
	})

	t.Run("IsTokenBlacklisted", func(t *testing.T) {
		blacklisted, err := storage.IsTokenBlacklisted(tokenID)
		if err != nil {
			t.Fatalf("Failed to check if token is blacklisted: %v", err)
		}
		if !blacklisted {
			t.Error("Expected token to be blacklisted")
		}
	})

	t.Run("CleanupExpiredTokens", func(t *testing.T) {
		// Add an expired token
		expiredTokenID := "test-expired-token"
		pastTime := time.Now().Add(-time.Hour)
		err := storage.BlacklistToken(expiredTokenID, pastTime)
		if err != nil {
			t.Fatalf("Failed to blacklist expired token: %v", err)
		}

		// Clean up expired tokens
		err = storage.CleanupExpiredTokens()
		if err != nil {
			t.Fatalf("Failed to cleanup expired tokens: %v", err)
		}

		// Verify expired token is removed
		blacklisted, err := storage.IsTokenBlacklisted(expiredTokenID)
		if err != nil {
			t.Fatalf("Failed to check expired token: %v", err)
		}
		if blacklisted {
			t.Error("Expected expired token to be cleaned up")
		}

		// Verify non-expired token is still there
		blacklisted, err = storage.IsTokenBlacklisted(tokenID)
		if err != nil {
			t.Fatalf("Failed to check non-expired token: %v", err)
		}
		if !blacklisted {
			t.Error("Expected non-expired token to still be blacklisted")
		}
	})
}

func TestPostgresStorage_PasswordOperations(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()
	defer cleanupTestData(t, storage)

	user := &models.User{
		ID:           "test-user-pwd",
		Username:     "testuserpwd",
		Email:        "testpwd@example.com",
		PasswordHash: "original-hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Create user first
	err := storage.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("UpdatePassword", func(t *testing.T) {
		newPasswordHash := "new-password-hash"
		err := storage.UpdatePassword(user.ID, newPasswordHash)
		if err != nil {
			t.Fatalf("Failed to update password: %v", err)
		}

		// Verify password was updated
		updatedUser, err := storage.GetUserByID(user.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}
		if updatedUser.PasswordHash != newPasswordHash {
			t.Errorf("Expected password hash %s, got %s", newPasswordHash, updatedUser.PasswordHash)
		}
	})
}

func TestPostgresStorage_HealthCheck(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()

	t.Run("Ping", func(t *testing.T) {
		err := storage.Ping()
		if err != nil {
			t.Fatalf("Failed to ping database: %v", err)
		}
	})
}

func TestPostgresStorage_Migration(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()

	t.Run("Migrate", func(t *testing.T) {
		err := storage.Migrate()
		if err != nil {
			t.Fatalf("Failed to migrate: %v", err)
		}
	})

	t.Run("GetSchemaVersion", func(t *testing.T) {
		version, err := storage.GetSchemaVersion()
		if err != nil {
			t.Fatalf("Failed to get schema version: %v", err)
		}
		if version < 0 {
			t.Errorf("Expected non-negative schema version, got %d", version)
		}
	})
}

func TestPostgresStorage_ErrorCases(t *testing.T) {
	storage := setupTestDB(t)
	defer storage.Close()
	defer cleanupTestData(t, storage)

	t.Run("GetNonexistentUser", func(t *testing.T) {
		_, err := storage.GetUserByID("nonexistent")
		if err == nil {
			t.Error("Expected error when getting nonexistent user")
		}
	})

	t.Run("UpdateNonexistentUser", func(t *testing.T) {
		updates := models.UserUpdates{
			Email: &[]string{"test@example.com"}[0],
		}
		err := storage.UpdateUser("nonexistent", updates)
		if err == nil {
			t.Error("Expected error when updating nonexistent user")
		}
	})

	t.Run("DeleteNonexistentUser", func(t *testing.T) {
		err := storage.DeleteUser("nonexistent")
		if err == nil {
			t.Error("Expected error when deleting nonexistent user")
		}
	})

	t.Run("DuplicateUser", func(t *testing.T) {
		user := &models.User{
			ID:           "duplicate-test",
			Username:     "duplicateuser",
			Email:        "duplicate@example.com",
			PasswordHash: "hash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			IsActive:     true,
		}

		// Create user first time
		err := storage.CreateUser(user)
		if err != nil {
			t.Fatalf("Failed to create user first time: %v", err)
		}

		// Try to create same user again
		err = storage.CreateUser(user)
		if err == nil {
			t.Error("Expected error when creating duplicate user")
		}
	})
}