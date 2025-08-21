package models

import (
	"testing"
	"time"
)

func TestEnhancedDataModelsIntegration(t *testing.T) {
	// Test User model with all enhanced fields
	user := &User{
		ID:           "user-123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastLoginAt:  &time.Time{},
		IsActive:     true,
		Metadata: map[string]interface{}{
			"role":        "admin",
			"preferences": map[string]interface{}{"theme": "dark"},
		},
	}

	// Test UserProfile conversion
	profile := user.ToUserProfile()
	
	// Verify UserProfile has all expected fields
	if profile.ID != user.ID {
		t.Errorf("Expected profile ID %s, got %s", user.ID, profile.ID)
	}
	if profile.Username != user.Username {
		t.Errorf("Expected profile Username %s, got %s", user.Username, profile.Username)
	}
	if profile.Email != user.Email {
		t.Errorf("Expected profile Email %s, got %s", user.Email, profile.Email)
	}
	if !profile.CreatedAt.Equal(user.CreatedAt) {
		t.Errorf("Expected profile CreatedAt %v, got %v", user.CreatedAt, profile.CreatedAt)
	}
	if !profile.UpdatedAt.Equal(user.UpdatedAt) {
		t.Errorf("Expected profile UpdatedAt %v, got %v", user.UpdatedAt, profile.UpdatedAt)
	}
	if profile.LastLoginAt != user.LastLoginAt {
		t.Errorf("Expected profile LastLoginAt %v, got %v", user.LastLoginAt, profile.LastLoginAt)
	}
	if profile.IsActive != user.IsActive {
		t.Errorf("Expected profile IsActive %v, got %v", user.IsActive, profile.IsActive)
	}
	if len(profile.Metadata) != len(user.Metadata) {
		t.Errorf("Expected profile Metadata length %d, got %d", len(user.Metadata), len(profile.Metadata))
	}

	// Test BlacklistedToken model
	blacklistedToken := &BlacklistedToken{
		TokenID:   "token-456",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	if blacklistedToken.TokenID == "" {
		t.Error("BlacklistedToken TokenID should not be empty")
	}
	if blacklistedToken.ExpiresAt.IsZero() {
		t.Error("BlacklistedToken ExpiresAt should not be zero")
	}
	if blacklistedToken.CreatedAt.IsZero() {
		t.Error("BlacklistedToken CreatedAt should not be zero")
	}

	// Test Migration model
	migration := &Migration{
		Version:     1,
		Description: "Add enhanced user fields",
		AppliedAt:   time.Now(),
	}

	if migration.Version == 0 {
		t.Error("Migration Version should not be zero")
	}
	if migration.Description == "" {
		t.Error("Migration Description should not be empty")
	}
	if migration.AppliedAt.IsZero() {
		t.Error("Migration AppliedAt should not be zero")
	}
}

func TestUserProfileSafety(t *testing.T) {
	// Create a user with sensitive data
	user := &User{
		ID:           "user-123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "super-secret-hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Convert to UserProfile
	profile := user.ToUserProfile()

	// Verify that UserProfile doesn't expose sensitive data
	// This is a compile-time check - accessing profile.PasswordHash would fail
	_ = profile.ID
	_ = profile.Username
	_ = profile.Email
	_ = profile.CreatedAt
	_ = profile.UpdatedAt
	_ = profile.LastLoginAt
	_ = profile.IsActive
	_ = profile.Metadata

	// The following line would cause a compile error if uncommented:
	// _ = profile.PasswordHash // This field doesn't exist in UserProfile

	t.Log("UserProfile successfully excludes sensitive PasswordHash field")
}