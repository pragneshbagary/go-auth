package auth

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// mockEnhancedStorage is a mock implementation of EnhancedStorage for testing
type mockEnhancedStorage struct {
	users         map[string]*models.User
	usersByEmail  map[string]*models.User
	usersByName   map[string]*models.User
	shouldFailGet bool
}

func newMockEnhancedStorage() *mockEnhancedStorage {
	return &mockEnhancedStorage{
		users:        make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		usersByName:  make(map[string]*models.User),
	}
}

func (m *mockEnhancedStorage) CreateUser(user models.User) error {
	m.users[user.ID] = &user
	m.usersByEmail[user.Email] = &user
	m.usersByName[user.Username] = &user
	return nil
}

func (m *mockEnhancedStorage) GetUserByUsername(username string) (*models.User, error) {
	if m.shouldFailGet {
		return nil, storage.ErrUserNotFound
	}
	user, exists := m.usersByName[username]
	if !exists {
		return nil, storage.ErrUserNotFound
	}
	return user, nil
}

func (m *mockEnhancedStorage) GetUserByID(userID string) (*models.User, error) {
	if m.shouldFailGet {
		return nil, storage.ErrUserNotFound
	}
	user, exists := m.users[userID]
	if !exists {
		return nil, storage.ErrUserNotFound
	}
	return user, nil
}

func (m *mockEnhancedStorage) GetUserByEmail(email string) (*models.User, error) {
	if m.shouldFailGet {
		return nil, storage.ErrUserNotFound
	}
	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, storage.ErrUserNotFound
	}
	return user, nil
}

func (m *mockEnhancedStorage) UpdateUser(userID string, updates storage.UserUpdates) error {
	user, exists := m.users[userID]
	if !exists {
		return storage.ErrUserNotFound
	}

	// Remove old entries if username or email is changing
	if updates.Username != nil {
		delete(m.usersByName, user.Username)
		user.Username = *updates.Username
		m.usersByName[user.Username] = user
	}
	if updates.Email != nil {
		delete(m.usersByEmail, user.Email)
		user.Email = *updates.Email
		m.usersByEmail[user.Email] = user
	}
	if updates.Metadata != nil {
		user.Metadata = updates.Metadata
	}

	user.UpdatedAt = time.Now()
	return nil
}

func (m *mockEnhancedStorage) UpdatePassword(userID string, passwordHash string) error {
	user, exists := m.users[userID]
	if !exists {
		return storage.ErrUserNotFound
	}
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()
	return nil
}

func (m *mockEnhancedStorage) DeleteUser(userID string) error {
	user, exists := m.users[userID]
	if !exists {
		return storage.ErrUserNotFound
	}
	delete(m.users, userID)
	delete(m.usersByEmail, user.Email)
	delete(m.usersByName, user.Username)
	return nil
}

func (m *mockEnhancedStorage) ListUsers(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	count := 0
	for _, user := range m.users {
		if count >= offset && len(users) < limit {
			users = append(users, user)
		}
		count++
	}
	return users, nil
}

// Unused methods for interface compliance
func (m *mockEnhancedStorage) BlacklistToken(tokenID string, expiresAt time.Time) error { return nil }
func (m *mockEnhancedStorage) IsTokenBlacklisted(tokenID string) (bool, error)          { return false, nil }
func (m *mockEnhancedStorage) CleanupExpiredTokens() error                              { return nil }
func (m *mockEnhancedStorage) Ping() error                                              { return nil }
func (m *mockEnhancedStorage) Migrate() error                                           { return nil }
func (m *mockEnhancedStorage) GetSchemaVersion() (int, error)                           { return 1, nil }
func (m *mockEnhancedStorage) RecordMigration(version int, description string) error    { return nil }
func (m *mockEnhancedStorage) RemoveMigrationRecord(version int) error                  { return nil }
func (m *mockEnhancedStorage) GetAppliedMigrations() ([]models.Migration, error) {
	return []models.Migration{{Version: 1, Description: "Initial", AppliedAt: time.Now()}}, nil
}

func TestUsers_Update(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	storage.CreateUser(testUser)

	t.Run("successful update", func(t *testing.T) {
		newEmail := "newemail@example.com"
		newUsername := "newusername"
		metadata := map[string]interface{}{"role": "admin"}

		updates := UserUpdate{
			Email:    &newEmail,
			Username: &newUsername,
			Metadata: metadata,
		}

		err := users.Update("user1", updates)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify the update
		updatedUser, err := storage.GetUserByID("user1")
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}

		if updatedUser.Email != newEmail {
			t.Errorf("Expected email %s, got %s", newEmail, updatedUser.Email)
		}
		if updatedUser.Username != newUsername {
			t.Errorf("Expected username %s, got %s", newUsername, updatedUser.Username)
		}
		if updatedUser.Metadata["role"] != "admin" {
			t.Errorf("Expected metadata role to be admin, got %v", updatedUser.Metadata["role"])
		}
	})

	t.Run("empty user ID", func(t *testing.T) {
		err := users.Update("", UserUpdate{})
		if err == nil || !strings.Contains(err.Error(), "user ID") {
			t.Errorf("Expected user ID validation error, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := users.Update("nonexistent", UserUpdate{})
		if err == nil {
			t.Error("Expected error for nonexistent user")
		}
	})
}

func TestUsers_ChangePassword(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user with a known password
	oldPassword := "oldpassword123"
	hashedPassword, _ := HashPassword(oldPassword)
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	storage.CreateUser(testUser)

	t.Run("successful password change", func(t *testing.T) {
		newPassword := "newpassword123"
		err := users.ChangePassword("user1", oldPassword, newPassword)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify the password was changed
		updatedUser, _ := storage.GetUserByID("user1")
		match, err := CheckPasswordHash(newPassword, updatedUser.PasswordHash)
		if err != nil {
			t.Fatalf("Failed to check new password: %v", err)
		}
		if !match {
			t.Error("New password does not match stored hash")
		}
	})

	t.Run("empty user ID", func(t *testing.T) {
		err := users.ChangePassword("", "old", "new")
		if err == nil || !strings.Contains(err.Error(), "user ID") {
			t.Errorf("Expected user ID validation error, got %v", err)
		}
	})

	t.Run("empty old password", func(t *testing.T) {
		err := users.ChangePassword("user1", "", "newpassword123")
		if err == nil || !strings.Contains(err.Error(), "old password") {
			t.Errorf("Expected old password validation error, got %v", err)
		}
	})

	t.Run("empty new password", func(t *testing.T) {
		err := users.ChangePassword("user1", "oldpassword", "")
		if err == nil || !strings.Contains(err.Error(), "new password") {
			t.Errorf("Expected new password validation error, got %v", err)
		}
	})

	t.Run("weak new password", func(t *testing.T) {
		err := users.ChangePassword("user1", oldPassword, "weak")
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "password") {
			t.Errorf("Expected weak password error, got %v", err)
		}
	})

	t.Run("incorrect old password", func(t *testing.T) {
		err := users.ChangePassword("user1", "wrongpassword", "newpassword123")
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "password") {
			t.Errorf("Expected incorrect old password error, got %v", err)
		}
	})
}

func TestUsers_CreateResetToken(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	storage.CreateUser(testUser)

	t.Run("successful token creation", func(t *testing.T) {
		token, err := users.CreateResetToken("test@example.com")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if token.Token == "" {
			t.Error("Expected non-empty token")
		}
		if token.UserID != "user1" {
			t.Errorf("Expected user ID 'user1', got %s", token.UserID)
		}
		if token.ExpiresAt.Before(time.Now()) {
			t.Error("Expected token to expire in the future")
		}
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := users.CreateResetToken("")
		if err == nil || !strings.Contains(err.Error(), "email") {
			t.Errorf("Expected email validation error, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := users.CreateResetToken("nonexistent@example.com")
		if err == nil {
			t.Error("Expected error for nonexistent user")
		}
	})
}

func TestUsers_ResetPassword(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	storage.CreateUser(testUser)

	// Create a reset token
	token, _ := users.CreateResetToken("test@example.com")

	t.Run("successful password reset", func(t *testing.T) {
		newPassword := "newpassword123"
		err := users.ResetPassword(token.Token, newPassword)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify the password was changed
		updatedUser, _ := storage.GetUserByID("user1")
		match, err := CheckPasswordHash(newPassword, updatedUser.PasswordHash)
		if err != nil {
			t.Fatalf("Failed to check new password: %v", err)
		}
		if !match {
			t.Error("New password does not match stored hash")
		}

		// Verify token was consumed
		err = users.ResetPassword(token.Token, "anotherpassword123")
		if err == nil {
			t.Error("Expected error when reusing consumed token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		err := users.ResetPassword("", "newpassword123")
		if err == nil || !strings.Contains(err.Error(), "token") {
			t.Errorf("Expected reset token validation error, got %v", err)
		}
	})

	t.Run("empty password", func(t *testing.T) {
		token, _ := users.CreateResetToken("test@example.com")
		err := users.ResetPassword(token.Token, "")
		if err == nil || !strings.Contains(err.Error(), "password") {
			t.Errorf("Expected new password validation error, got %v", err)
		}
	})

	t.Run("weak password", func(t *testing.T) {
		token, _ := users.CreateResetToken("test@example.com")
		err := users.ResetPassword(token.Token, "weak")
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "password") {
			t.Errorf("Expected weak password error, got %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		err := users.ResetPassword("invalidtoken", "newpassword123")
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "token") {
			t.Errorf("Expected invalid token error, got %v", err)
		}
	})
}

func TestUsers_Get(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
		Metadata:     map[string]interface{}{"role": "user"},
	}
	storage.CreateUser(testUser)

	t.Run("successful get", func(t *testing.T) {
		profile, err := users.Get("user1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if profile.ID != "user1" {
			t.Errorf("Expected ID 'user1', got %s", profile.ID)
		}
		if profile.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got %s", profile.Username)
		}
		if profile.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %s", profile.Email)
		}
		if profile.Metadata["role"] != "user" {
			t.Errorf("Expected role 'user', got %v", profile.Metadata["role"])
		}
	})

	t.Run("empty user ID", func(t *testing.T) {
		_, err := users.Get("")
		if err == nil || !strings.Contains(err.Error(), "user ID") {
			t.Errorf("Expected user ID validation error, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := users.Get("nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent user")
		}
	})
}

func TestUsers_GetByEmail(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	storage.CreateUser(testUser)

	t.Run("successful get by email", func(t *testing.T) {
		profile, err := users.GetByEmail("test@example.com")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if profile.ID != "user1" {
			t.Errorf("Expected ID 'user1', got %s", profile.ID)
		}
		if profile.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %s", profile.Email)
		}
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := users.GetByEmail("")
		if err == nil || !strings.Contains(err.Error(), "email") {
			t.Errorf("Expected email validation error, got %v", err)
		}
	})
}

func TestUsers_GetByUsername(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	storage.CreateUser(testUser)

	t.Run("successful get by username", func(t *testing.T) {
		profile, err := users.GetByUsername("testuser")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if profile.ID != "user1" {
			t.Errorf("Expected ID 'user1', got %s", profile.ID)
		}
		if profile.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got %s", profile.Username)
		}
	})

	t.Run("empty username", func(t *testing.T) {
		_, err := users.GetByUsername("")
		if err == nil || !strings.Contains(err.Error(), "username") {
			t.Errorf("Expected username validation error, got %v", err)
		}
	})
}

func TestUsers_List(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create test users
	for i := 0; i < 15; i++ {
		testUser := models.User{
			ID:           fmt.Sprintf("user%d", i),
			Username:     fmt.Sprintf("testuser%d", i),
			Email:        fmt.Sprintf("test%d@example.com", i),
			PasswordHash: "hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			IsActive:     true,
		}
		storage.CreateUser(testUser)
	}

	t.Run("successful list with default limit", func(t *testing.T) {
		profiles, err := users.List(0, 0) // Should use default limit of 10
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(profiles) != 10 {
			t.Errorf("Expected 10 profiles, got %d", len(profiles))
		}
	})

	t.Run("successful list with custom limit", func(t *testing.T) {
		profiles, err := users.List(5, 0)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(profiles) != 5 {
			t.Errorf("Expected 5 profiles, got %d", len(profiles))
		}
	})

	t.Run("list with offset", func(t *testing.T) {
		profiles, err := users.List(5, 10)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(profiles) != 5 {
			t.Errorf("Expected 5 profiles, got %d", len(profiles))
		}
	})

	t.Run("list with excessive limit", func(t *testing.T) {
		profiles, err := users.List(200, 0) // Should be capped at 100
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should return all 15 users since that's less than the 100 cap
		if len(profiles) != 15 {
			t.Errorf("Expected 15 profiles, got %d", len(profiles))
		}
	})
}

func TestUsers_Delete(t *testing.T) {
	storage := newMockEnhancedStorage()
	users := &Users{storage: storage}

	// Create a test user
	testUser := models.User{
		ID:           "user1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	storage.CreateUser(testUser)

	t.Run("successful delete", func(t *testing.T) {
		err := users.Delete("user1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify user was deleted
		_, err = storage.GetUserByID("user1")
		if err == nil {
			t.Error("Expected user to be deleted")
		}
	})

	t.Run("empty user ID", func(t *testing.T) {
		err := users.Delete("")
		if err == nil || !strings.Contains(err.Error(), "user ID") {
			t.Errorf("Expected user ID validation error, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := users.Delete("nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent user")
		}
	})
}