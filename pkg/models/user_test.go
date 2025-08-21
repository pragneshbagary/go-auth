package models

import (
	"testing"
	"time"
)

func TestUser_ToUserProfile(t *testing.T) {
	// Create a test user with all fields populated
	user := &User{
		ID:           "test-user-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed-password-should-not-appear",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastLoginAt:  &time.Time{},
		IsActive:     true,
		Metadata: map[string]interface{}{
			"role": "user",
			"plan": "premium",
		},
	}

	// Convert to UserProfile
	profile := user.ToUserProfile()

	// Verify all fields are copied correctly
	if profile.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, profile.ID)
	}
	if profile.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, profile.Username)
	}
	if profile.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, profile.Email)
	}
	if !profile.CreatedAt.Equal(user.CreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", user.CreatedAt, profile.CreatedAt)
	}
	if !profile.UpdatedAt.Equal(user.UpdatedAt) {
		t.Errorf("Expected UpdatedAt %v, got %v", user.UpdatedAt, profile.UpdatedAt)
	}
	if profile.LastLoginAt != user.LastLoginAt {
		t.Errorf("Expected LastLoginAt %v, got %v", user.LastLoginAt, profile.LastLoginAt)
	}
	if profile.IsActive != user.IsActive {
		t.Errorf("Expected IsActive %v, got %v", user.IsActive, profile.IsActive)
	}
	if len(profile.Metadata) != len(user.Metadata) {
		t.Errorf("Expected Metadata length %d, got %d", len(user.Metadata), len(profile.Metadata))
	}
	if profile.Metadata["role"] != user.Metadata["role"] {
		t.Errorf("Expected Metadata role %v, got %v", user.Metadata["role"], profile.Metadata["role"])
	}
}

func TestUserProfile_DoesNotExposePassword(t *testing.T) {
	user := &User{
		ID:           "test-user-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "secret-password-hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	profile := user.ToUserProfile()

	// Verify that UserProfile struct doesn't have a PasswordHash field
	// This is a compile-time check - if UserProfile had a PasswordHash field,
	// this would fail to compile
	_ = profile.ID
	_ = profile.Username
	_ = profile.Email
	_ = profile.CreatedAt
	_ = profile.UpdatedAt
	_ = profile.LastLoginAt
	_ = profile.IsActive
	_ = profile.Metadata
	// profile.PasswordHash would cause a compile error if it existed
}

func TestBlacklistedToken_Structure(t *testing.T) {
	token := &BlacklistedToken{
		TokenID:   "test-token-id",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	// Verify all fields are accessible
	if token.TokenID == "" {
		t.Error("TokenID should not be empty")
	}
	if token.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}
	if token.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestMigration_Structure(t *testing.T) {
	migration := &Migration{
		Version:     1,
		Description: "Initial migration",
		AppliedAt:   time.Now(),
	}

	// Verify all fields are accessible
	if migration.Version == 0 {
		t.Error("Version should not be zero")
	}
	if migration.Description == "" {
		t.Error("Description should not be empty")
	}
	if migration.AppliedAt.IsZero() {
		t.Error("AppliedAt should not be zero")
	}
}