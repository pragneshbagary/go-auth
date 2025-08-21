package auth

import (
	"os"
	"testing"
	"time"
)

func TestNewInMemory(t *testing.T) {
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create in-memory auth: %v", err)
	}

	if auth == nil {
		t.Fatal("Auth instance should not be nil")
	}

	// Test health check
	if err := auth.Health(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

func TestNewSQLite(t *testing.T) {
	// Create a temporary database file
	dbPath := "test_auth.db"
	defer os.Remove(dbPath)

	auth, err := NewSQLite(dbPath, "test-secret")
	if err != nil {
		t.Fatalf("Failed to create SQLite auth: %v", err)
	}

	if auth == nil {
		t.Fatal("Auth instance should not be nil")
	}

	// Test health check
	if err := auth.Health(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

func TestNew(t *testing.T) {
	// Test that New() is an alias for NewSQLite()
	dbPath := "test_auth_new.db"
	defer os.Remove(dbPath)

	auth, err := New(dbPath, "test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth with New(): %v", err)
	}

	if auth == nil {
		t.Fatal("Auth instance should not be nil")
	}

	// Test health check
	if err := auth.Health(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

func TestNewWithConfig(t *testing.T) {
	config := &AuthConfig{
		JWTSecret:       "test-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTIssuer:       "test-issuer",
		AccessTokenTTL:  30 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		AppName:         "test-app",
	}

	auth, err := NewWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create auth with config: %v", err)
	}

	if auth == nil {
		t.Fatal("Auth instance should not be nil")
	}

	// Verify config was applied
	if auth.config.JWTIssuer != "test-issuer" {
		t.Errorf("Expected issuer 'test-issuer', got '%s'", auth.config.JWTIssuer)
	}

	if auth.config.AppName != "test-app" {
		t.Errorf("Expected app name 'test-app', got '%s'", auth.config.AppName)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth: %v", err)
	}

	// Test user registration
	payload := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword123",
	}

	user, err := auth.Register(payload)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	// Test user login
	loginResp, err := auth.Login("testuser", "testpassword123", nil)
	if err != nil {
		t.Fatalf("Failed to login user: %v", err)
	}

	if loginResp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}

	if loginResp.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}

	// Test token validation
	claims, err := auth.ValidateAccessToken(loginResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims["username"] != "testuser" {
		t.Errorf("Expected username 'testuser' in claims, got '%v'", claims["username"])
	}
}

func TestGetUserMethods(t *testing.T) {
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth: %v", err)
	}

	// Register a user first
	payload := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword123",
	}

	user, err := auth.Register(payload)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test GetUser
	profile, err := auth.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", profile.Username)
	}

	// Test GetUserByUsername
	profile, err = auth.GetUserByUsername("testuser")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}

	if profile.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", profile.Email)
	}

	// Test GetUserByEmail
	profile, err = auth.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", profile.Username)
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &AuthConfig{
		JWTSecret: "test-secret",
	}

	auth, err := NewWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create auth with minimal config: %v", err)
	}

	// Check that defaults were applied
	if auth.config.JWTRefreshSecret == "" {
		t.Error("JWT refresh secret should have a default value")
	}

	if auth.config.JWTIssuer != "go-auth" {
		t.Errorf("Expected default issuer 'go-auth', got '%s'", auth.config.JWTIssuer)
	}

	if auth.config.AccessTokenTTL != 15*time.Minute {
		t.Errorf("Expected default access token TTL 15m, got %v", auth.config.AccessTokenTTL)
	}

	if auth.config.RefreshTokenTTL != 7*24*time.Hour {
		t.Errorf("Expected default refresh token TTL 7 days, got %v", auth.config.RefreshTokenTTL)
	}

	if auth.config.AppName != "go-auth-app" {
		t.Errorf("Expected default app name 'go-auth-app', got '%s'", auth.config.AppName)
	}
}