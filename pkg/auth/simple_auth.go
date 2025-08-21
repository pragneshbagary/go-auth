package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pragneshbagary/go-auth/pkg/models"
)

// SimpleAuth provides an ultra-simple wrapper around Auth with minimal configuration requirements.
// It's designed for developers who want to get started with authentication in just a few lines of code.
type SimpleAuth struct {
	auth *Auth
}

// Quick creates a new SimpleAuth instance with SQLite storage using sensible defaults.
// It only requires a JWT secret and uses a default SQLite database file.
// This is the simplest way to get started with go-auth.
func Quick(jwtSecret string) (*SimpleAuth, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}

	// Use default SQLite database path
	defaultDBPath := "auth.db"
	
	auth, err := NewSQLite(defaultDBPath, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	return &SimpleAuth{auth: auth}, nil
}

// QuickSQLite creates a new SimpleAuth instance with SQLite storage at the specified path.
// This provides a bit more control over the database location while keeping things simple.
func QuickSQLite(dbPath string, jwtSecret string) (*SimpleAuth, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	if dbPath == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	// Ensure directory exists
	if dir := filepath.Dir(dbPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	auth, err := NewSQLite(dbPath, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	return &SimpleAuth{auth: auth}, nil
}

// QuickPostgres creates a new SimpleAuth instance with PostgreSQL storage.
// This is useful when you need a more robust database backend but still want simple setup.
func QuickPostgres(databaseURL string, jwtSecret string) (*SimpleAuth, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL cannot be empty")
	}

	auth, err := NewPostgres(databaseURL, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	return &SimpleAuth{auth: auth}, nil
}

// QuickInMemory creates a new SimpleAuth instance with in-memory storage.
// This is perfect for testing, development, or temporary applications.
func QuickInMemory(jwtSecret string) (*SimpleAuth, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}

	auth, err := NewInMemory(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	return &SimpleAuth{auth: auth}, nil
}

// QuickFromEnv creates a new SimpleAuth instance using configuration from environment variables.
// This is the most flexible option and follows 12-factor app principles.
// Required environment variables:
//   - AUTH_JWT_ACCESS_SECRET: JWT access token secret
//   - AUTH_JWT_REFRESH_SECRET: JWT refresh token secret
// Optional environment variables:
//   - AUTH_DB_TYPE: Database type (sqlite, postgres, memory) - defaults to sqlite
//   - AUTH_DB_URL: Database URL/path - defaults to "auth.db"
//   - AUTH_JWT_ISSUER: JWT issuer - defaults to "go-auth"
//   - AUTH_ACCESS_TOKEN_TTL: Access token TTL - defaults to "15m"
//   - AUTH_REFRESH_TOKEN_TTL: Refresh token TTL - defaults to "168h"
//   - AUTH_APP_NAME: Application name - defaults to "go-auth-app"
func QuickFromEnv() (*SimpleAuth, error) {
	// Load enhanced configuration from environment
	enhancedConfig, err := LoadConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
	}

	// Convert to AuthConfig format
	authConfig := &AuthConfig{
		JWTSecret:       enhancedConfig.JWTAccessSecret,
		JWTRefreshSecret: enhancedConfig.JWTRefreshSecret,
		JWTIssuer:       enhancedConfig.JWTIssuer,
		AccessTokenTTL:  enhancedConfig.AccessTokenTTL,
		RefreshTokenTTL: enhancedConfig.RefreshTokenTTL,
		AppName:         enhancedConfig.AppName,
	}

	// Set database configuration based on type
	switch enhancedConfig.DatabaseType {
	case "postgres":
		authConfig.DatabaseURL = enhancedConfig.DatabaseURL
	case "sqlite":
		authConfig.DatabasePath = enhancedConfig.DatabaseURL
	case "memory":
		// No additional config needed for memory storage
	default:
		return nil, fmt.Errorf("unsupported database type: %s", enhancedConfig.DatabaseType)
	}

	auth, err := NewWithConfig(authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	return &SimpleAuth{auth: auth}, nil
}

// Register creates a new user account.
// This is a simplified version that only requires username, email, and password.
func (s *SimpleAuth) Register(username, email, password string) (*models.User, error) {
	return s.auth.Register(RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	})
}

// Login authenticates a user and returns access and refresh tokens.
// This simplified version doesn't require custom claims - it uses standard user claims.
func (s *SimpleAuth) Login(username, password string) (*LoginResult, error) {
	return s.auth.Login(username, password, nil)
}

// LoginWithClaims authenticates a user and returns tokens with custom claims.
// Use this when you need to embed additional information in the access token.
func (s *SimpleAuth) LoginWithClaims(username, password string, customClaims map[string]interface{}) (*LoginResult, error) {
	return s.auth.Login(username, password, customClaims)
}

// ValidateToken validates an access token and returns the user claims.
// This is the most commonly used validation method.
func (s *SimpleAuth) ValidateToken(tokenString string) (map[string]interface{}, error) {
	claims, err := s.auth.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}
	
	// Convert jwt.MapClaims to map[string]interface{} for simpler usage
	result := make(map[string]interface{})
	for k, v := range claims {
		result[k] = v
	}
	
	return result, nil
}

// RefreshToken refreshes an access token using a refresh token.
// This returns a new access token and refresh token pair.
func (s *SimpleAuth) RefreshToken(refreshToken string) (*LoginResult, error) {
	// Validate the refresh token first
	claims, err := s.auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Extract user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token: missing user ID")
	}

	// Get user information
	user, err := s.auth.storage.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Generate new tokens
	userClaims := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"user_id":  user.ID,
	}

	accessToken, err := s.auth.jwtManager.GenerateAccessToken(user.ID, userClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.auth.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetUser retrieves user information by user ID.
// Returns a safe UserProfile without sensitive information.
func (s *SimpleAuth) GetUser(userID string) (*models.UserProfile, error) {
	return s.auth.GetUser(userID)
}

// GetUserByUsername retrieves user information by username.
// Returns a safe UserProfile without sensitive information.
func (s *SimpleAuth) GetUserByUsername(username string) (*models.UserProfile, error) {
	return s.auth.GetUserByUsername(username)
}

// GetUserByEmail retrieves user information by email.
// Returns a safe UserProfile without sensitive information.
func (s *SimpleAuth) GetUserByEmail(email string) (*models.UserProfile, error) {
	return s.auth.GetUserByEmail(email)
}

// Health checks if the authentication service is healthy.
// This includes checking database connectivity.
func (s *SimpleAuth) Health() error {
	return s.auth.Health()
}

// GetAuth returns the underlying Auth instance for advanced usage.
// Use this when you need access to more advanced features not exposed by SimpleAuth.
func (s *SimpleAuth) GetAuth() *Auth {
	return s.auth
}