package auth

import (
	"errors"
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/internal/jwtutils"
	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"

	"github.com/google/uuid"
)

// AuthService provides high-level authentication operations.
// It orchestrates user storage, password hashing, and token generation.
type AuthService struct {
	storage    storage.Storage
	jwtManager jwtutils.TokenManager
}

// NewAuthService creates a new AuthService.
// It requires a storage implementation and a JWT manager.
func NewAuthService(storage storage.Storage, jwtManager jwtutils.TokenManager) *AuthService {
	return &AuthService{
		storage:    storage,
		jwtManager: jwtManager,
	}
}

// RegisterPayload defines the data required to register a new user.
// Using a struct makes the API cleaner and more extensible.
type RegisterPayload struct {
	Username string
	Email    string
	Password string
}

// Register creates a new user, hashes their password, and saves them to storage.
// It returns the newly created user.
func (s *AuthService) Register(payload RegisterPayload) (*models.User, error) {
	// Basic validation (in a real app, you'd add more, e.g., password strength)
	if payload.Username == "" || payload.Password == "" {
		return nil, errors.New("username and password cannot be empty")
	}

	// Check if user already exists
	if _, err := s.storage.GetUserByUsername(payload.Username); err == nil {
		return nil, fmt.Errorf("user with username '%s' already exists", payload.Username)
	}

	passwordHash, err := HashPassword(payload.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := models.User{
		ID:           uuid.New().String(), // Generate a new unique ID for the user.
		Username:     payload.Username,
		Email:        payload.Email,
		PasswordHash: passwordHash,
	}

	if err := s.storage.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &newUser, nil
}

// LoginResponse is the structure returned upon a successful login.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login authenticates a user and returns an access and refresh token pair.
// It accepts customClaims to be embedded in the access token for authorization purposes.
func (s *AuthService) Login(username, password string, customClaims map[string]interface{}) (*LoginResponse, error) {
	user, err := s.storage.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password") // Generic error for security
	}

	match, err := CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		log.Printf("error checking password hash for user %s: %v", username, err)
		return nil, errors.New("invalid username or password")
	}
	if !match {
		return nil, errors.New("invalid username or password")
	}

	// Add standard claims like username to the token by default.
	claims := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	}
	// Merge with any custom claims provided by the caller.
	for k, v := range customClaims {
		claims[k] = v
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
