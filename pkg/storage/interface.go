package storage

import "github.com/pragneshbagary/go-auth/pkg/models"

// Storage defines the interface for data storage operations.
// This allows the auth service to be independent of the database implementation,
// enabling users to plug in their own backend (e.g., PostgreSQL, MongoDB).
type Storage interface {
	// CreateUser saves a new user to the database.
	// It should return an error if a user with the same username already exists.
	CreateUser(user models.User) error

	// GetUserByUsername retrieves a user by their username.
	// It should return an error if the user is not found.
	GetUserByUsername(username string) (*models.User, error)
}
