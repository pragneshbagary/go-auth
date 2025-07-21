// user.go
package models

// User defines the structure for a user in the system.
// The password field should always store a hashed password, never plaintext.
type User struct {
	// ID is the unique identifier for the user (e.g., a UUID).
	ID string `json:"id"`
	// Username is the unique, user-chosen name for login.
	Username string `json:"username"`
	// Email is the user's email address.
	Email string `json:"email"`
	// PasswordHash is the secure, hashed version of the user's password.
	// The struct tag `json:"-"` ensures it is never exposed in API responses.
	PasswordHash string `json:"-"`
}
