package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// Users provides enhanced user management operations with intuitive method names.
type Users struct {
	storage          storage.EnhancedStorage
	eventLogger      *AuthEventLogger
	metricsCollector *MetricsCollector
}

// UserUpdate represents the fields that can be updated for a user.
type UserUpdate struct {
	Email    *string
	Username *string
	Metadata map[string]interface{}
}

// ResetToken represents a password reset token with expiration.
type ResetToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

// passwordResetTokens is an in-memory store for reset tokens
// In a production system, this should be stored in the database
var passwordResetTokens = make(map[string]*ResetToken)

// Update modifies user profile information.
// It allows updating email, username, and metadata fields.
func (u *Users) Update(userID string, updates UserUpdate) error {
	if userID == "" {
		return ErrValidationError("user ID")
	}

	// Validate that the user exists
	_, err := u.storage.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound()
	}

	// Check for username conflicts if username is being updated
	if updates.Username != nil && *updates.Username != "" {
		existingUser, err := u.storage.GetUserByUsername(*updates.Username)
		if err == nil && existingUser.ID != userID {
			return ErrUserExists("username")
		}
	}

	// Check for email conflicts if email is being updated
	if updates.Email != nil && *updates.Email != "" {
		existingUser, err := u.storage.GetUserByEmail(*updates.Email)
		if err == nil && existingUser.ID != userID {
			return ErrUserExists("email")
		}
	}

	// Convert to storage format
	storageUpdates := storage.UserUpdates{
		Email:    updates.Email,
		Username: updates.Username,
		Metadata: updates.Metadata,
	}

	if err := u.storage.UpdateUser(userID, storageUpdates); err != nil {
		return WrapDatabaseError(err)
	}
	
	return nil
}

// ChangePassword securely changes a user's password after validating the old password.
func (u *Users) ChangePassword(userID, oldPassword, newPassword string) error {
	if userID == "" {
		return ErrValidationError("user ID")
	}
	if oldPassword == "" {
		return ErrValidationError("old password")
	}
	if newPassword == "" {
		return ErrValidationError("new password")
	}

	// Basic password strength validation
	if len(newPassword) < 8 {
		return ErrWeakPassword("Password must be at least 8 characters long")
	}

	// Get the user to verify the old password
	user, err := u.storage.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound()
	}

	// Verify the old password
	match, err := CheckPasswordHash(oldPassword, user.PasswordHash)
	if err != nil {
		return WrapError(err, ErrCodeInternalError, "Failed to verify old password")
	}
	if !match {
		return NewAuthError(ErrCodePasswordMismatch, "Old password is incorrect")
	}

	// Hash the new password
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return WrapError(err, ErrCodeInternalError, "Failed to hash new password")
	}

	// Update the password in storage
	if err := u.storage.UpdatePassword(userID, newPasswordHash); err != nil {
		return WrapDatabaseError(err)
	}
	
	return nil
}

// CreateResetToken generates a password reset token for the user with the given email.
// The token expires after 1 hour.
func (u *Users) CreateResetToken(email string) (*ResetToken, error) {
	if email == "" {
		return nil, ErrValidationError("email")
	}

	// Verify that a user with this email exists
	user, err := u.storage.GetUserByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound()
	}

	// Generate a cryptographically secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, WrapError(err, ErrCodeInternalError, "Failed to generate reset token")
	}
	token := hex.EncodeToString(tokenBytes)

	// Create reset token with 1 hour expiration
	resetToken := &ResetToken{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Store the token (in production, this should be in the database)
	passwordResetTokens[token] = resetToken

	return resetToken, nil
}

// ResetPassword resets a user's password using a valid reset token.
func (u *Users) ResetPassword(token, newPassword string) error {
	if token == "" {
		return ErrValidationError("reset token")
	}
	if newPassword == "" {
		return ErrValidationError("new password")
	}

	// Basic password strength validation
	if len(newPassword) < 8 {
		return ErrWeakPassword("Password must be at least 8 characters long")
	}

	// Retrieve and validate the reset token
	resetToken, exists := passwordResetTokens[token]
	if !exists {
		return NewAuthError(ErrCodeInvalidResetToken, "Invalid or expired reset token")
	}

	// Check if token has expired
	if time.Now().After(resetToken.ExpiresAt) {
		// Clean up expired token
		delete(passwordResetTokens, token)
		return NewAuthError(ErrCodeResetTokenExpired, "Reset token has expired")
	}

	// Hash the new password
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return WrapError(err, ErrCodeInternalError, "Failed to hash new password")
	}

	// Update the password in storage
	err = u.storage.UpdatePassword(resetToken.UserID, newPasswordHash)
	if err != nil {
		return WrapDatabaseError(err)
	}

	// Remove the used token
	delete(passwordResetTokens, token)

	return nil
}

// Get retrieves a user by their ID, returning a safe UserProfile without sensitive data.
func (u *Users) Get(userID string) (*models.UserProfile, error) {
	if userID == "" {
		return nil, ErrValidationError("user ID")
	}

	user, err := u.storage.GetUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound()
	}

	return user.ToUserProfile(), nil
}

// GetByEmail retrieves a user by their email, returning a safe UserProfile without sensitive data.
func (u *Users) GetByEmail(email string) (*models.UserProfile, error) {
	if email == "" {
		return nil, ErrValidationError("email")
	}

	user, err := u.storage.GetUserByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound()
	}

	return user.ToUserProfile(), nil
}

// GetByUsername retrieves a user by their username, returning a safe UserProfile without sensitive data.
func (u *Users) GetByUsername(username string) (*models.UserProfile, error) {
	if username == "" {
		return nil, ErrValidationError("username")
	}

	user, err := u.storage.GetUserByUsername(username)
	if err != nil {
		return nil, ErrUserNotFound()
	}

	return user.ToUserProfile(), nil
}

// List retrieves a paginated list of users, returning safe UserProfile objects without sensitive data.
func (u *Users) List(limit, offset int) ([]*models.UserProfile, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit to prevent abuse
	}
	if offset < 0 {
		offset = 0
	}

	users, err := u.storage.ListUsers(limit, offset)
	if err != nil {
		return nil, WrapDatabaseError(err)
	}

	profiles := make([]*models.UserProfile, len(users))
	for i, user := range users {
		profiles[i] = user.ToUserProfile()
	}

	return profiles, nil
}

// Delete removes a user from the system.
func (u *Users) Delete(userID string) error {
	if userID == "" {
		return ErrValidationError("user ID")
	}

	// Verify that the user exists before attempting deletion
	_, err := u.storage.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound()
	}

	if err := u.storage.DeleteUser(userID); err != nil {
		return WrapDatabaseError(err)
	}
	
	return nil
}