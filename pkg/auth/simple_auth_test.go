package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestQuick(t *testing.T) {
	// Test with valid JWT secret
	simpleAuth, err := Quick("test-secret-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if simpleAuth == nil {
		t.Fatal("Expected SimpleAuth instance, got nil")
	}
	if simpleAuth.auth == nil {
		t.Fatal("Expected underlying Auth instance, got nil")
	}

	// Test health check
	if err := simpleAuth.Health(); err != nil {
		t.Fatalf("Expected health check to pass, got %v", err)
	}

	// Test with empty JWT secret
	_, err = Quick("")
	if err == nil {
		t.Fatal("Expected error for empty JWT secret, got nil")
	}
}

func TestQuickSQLite(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Test with valid parameters
	simpleAuth, err := QuickSQLite(dbPath, "test-secret-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if simpleAuth == nil {
		t.Fatal("Expected SimpleAuth instance, got nil")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Expected database file to be created")
	}

	// Test with empty JWT secret
	_, err = QuickSQLite(dbPath, "")
	if err == nil {
		t.Fatal("Expected error for empty JWT secret, got nil")
	}

	// Test with empty database path
	_, err = QuickSQLite("", "test-secret-key")
	if err == nil {
		t.Fatal("Expected error for empty database path, got nil")
	}
}

func TestQuickInMemory(t *testing.T) {
	// Test with valid JWT secret
	simpleAuth, err := QuickInMemory("test-secret-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if simpleAuth == nil {
		t.Fatal("Expected SimpleAuth instance, got nil")
	}

	// Test health check
	if err := simpleAuth.Health(); err != nil {
		t.Fatalf("Expected health check to pass, got %v", err)
	}

	// Test with empty JWT secret
	_, err = QuickInMemory("")
	if err == nil {
		t.Fatal("Expected error for empty JWT secret, got nil")
	}
}

func TestQuickFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"AUTH_JWT_ACCESS_SECRET",
		"AUTH_JWT_REFRESH_SECRET",
		"AUTH_DB_TYPE",
		"AUTH_DB_URL",
		"AUTH_JWT_ISSUER",
		"AUTH_ACCESS_TOKEN_TTL",
		"AUTH_REFRESH_TOKEN_TTL",
		"AUTH_APP_NAME",
	}
	
	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
	}
	
	// Clean up environment after test
	defer func() {
		for _, env := range envVars {
			if val, exists := originalEnv[env]; exists && val != "" {
				os.Setenv(env, val)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	// Test with minimal required environment variables
	os.Setenv("AUTH_JWT_ACCESS_SECRET", "test-access-secret")
	os.Setenv("AUTH_JWT_REFRESH_SECRET", "test-refresh-secret")
	
	simpleAuth, err := QuickFromEnv()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if simpleAuth == nil {
		t.Fatal("Expected SimpleAuth instance, got nil")
	}

	// Test with additional environment variables
	os.Setenv("AUTH_DB_TYPE", "memory")
	os.Setenv("AUTH_JWT_ISSUER", "test-issuer")
	os.Setenv("AUTH_ACCESS_TOKEN_TTL", "30m")
	os.Setenv("AUTH_REFRESH_TOKEN_TTL", "24h")
	os.Setenv("AUTH_APP_NAME", "test-app")
	
	simpleAuth2, err := QuickFromEnv()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if simpleAuth2 == nil {
		t.Fatal("Expected SimpleAuth instance, got nil")
	}

	// Test without required environment variables
	os.Unsetenv("AUTH_JWT_ACCESS_SECRET")
	os.Unsetenv("AUTH_JWT_REFRESH_SECRET")
	
	_, err = QuickFromEnv()
	if err == nil {
		t.Fatal("Expected error for missing required environment variables, got nil")
	}
}

func TestSimpleAuthRegisterAndLogin(t *testing.T) {
	simpleAuth, err := QuickInMemory("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create SimpleAuth: %v", err)
	}

	// Test user registration
	user, err := simpleAuth.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Expected no error during registration, got %v", err)
	}
	if user == nil {
		t.Fatal("Expected user, got nil")
	}
	if user.Username != "testuser" {
		t.Fatalf("Expected username 'testuser', got %s", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("Expected email 'test@example.com', got %s", user.Email)
	}

	// Test user login
	loginResult, err := simpleAuth.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Expected no error during login, got %v", err)
	}
	if loginResult == nil {
		t.Fatal("Expected login result, got nil")
	}
	if loginResult.AccessToken == "" {
		t.Fatal("Expected access token, got empty string")
	}
	if loginResult.RefreshToken == "" {
		t.Fatal("Expected refresh token, got empty string")
	}

	// Test login with wrong password
	_, err = simpleAuth.Login("testuser", "wrongpassword")
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}

	// Test login with non-existent user
	_, err = simpleAuth.Login("nonexistent", "password123")
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

func TestSimpleAuthLoginWithClaims(t *testing.T) {
	simpleAuth, err := QuickInMemory("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create SimpleAuth: %v", err)
	}

	// Register a user
	_, err = simpleAuth.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test login with custom claims
	customClaims := map[string]interface{}{
		"role":        "admin",
		"permissions": []string{"read", "write", "delete"},
		"department":  "engineering",
	}

	loginResult, err := simpleAuth.LoginWithClaims("testuser", "password123", customClaims)
	if err != nil {
		t.Fatalf("Expected no error during login with claims, got %v", err)
	}
	if loginResult == nil {
		t.Fatal("Expected login result, got nil")
	}

	// Validate the token and check if custom claims are present
	claims, err := simpleAuth.ValidateToken(loginResult.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error during token validation, got %v", err)
	}

	// Check standard claims
	if claims["username"] != "testuser" {
		t.Fatalf("Expected username 'testuser', got %v", claims["username"])
	}
	if claims["email"] != "test@example.com" {
		t.Fatalf("Expected email 'test@example.com', got %v", claims["email"])
	}

	// Check custom claims
	if claims["role"] != "admin" {
		t.Fatalf("Expected role 'admin', got %v", claims["role"])
	}
	if claims["department"] != "engineering" {
		t.Fatalf("Expected department 'engineering', got %v", claims["department"])
	}
}

func TestSimpleAuthValidateToken(t *testing.T) {
	simpleAuth, err := QuickInMemory("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create SimpleAuth: %v", err)
	}

	// Register and login a user
	_, err = simpleAuth.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginResult, err := simpleAuth.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to login user: %v", err)
	}

	// Test token validation
	claims, err := simpleAuth.ValidateToken(loginResult.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error during token validation, got %v", err)
	}
	if claims == nil {
		t.Fatal("Expected claims, got nil")
	}

	// Check standard claims
	if claims["username"] != "testuser" {
		t.Fatalf("Expected username 'testuser', got %v", claims["username"])
	}
	if claims["email"] != "test@example.com" {
		t.Fatalf("Expected email 'test@example.com', got %v", claims["email"])
	}

	// Test with invalid token
	_, err = simpleAuth.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}

	// Test with empty token
	_, err = simpleAuth.ValidateToken("")
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
}

func TestSimpleAuthRefreshToken(t *testing.T) {
	simpleAuth, err := QuickInMemory("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create SimpleAuth: %v", err)
	}

	// Register and login a user
	_, err = simpleAuth.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginResult, err := simpleAuth.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to login user: %v", err)
	}

	// Test token refresh
	refreshResult, err := simpleAuth.RefreshToken(loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("Expected no error during token refresh, got %v", err)
	}
	if refreshResult == nil {
		t.Fatal("Expected refresh result, got nil")
	}
	if refreshResult.AccessToken == "" {
		t.Fatal("Expected new access token, got empty string")
	}
	if refreshResult.RefreshToken == "" {
		t.Fatal("Expected new refresh token, got empty string")
	}

	// Verify the new access token is valid
	claims, err := simpleAuth.ValidateToken(refreshResult.AccessToken)
	if err != nil {
		t.Fatalf("Expected new access token to be valid, got %v", err)
	}
	if claims["username"] != "testuser" {
		t.Fatalf("Expected username 'testuser' in new token, got %v", claims["username"])
	}

	// Test with invalid refresh token
	_, err = simpleAuth.RefreshToken("invalid-refresh-token")
	if err == nil {
		t.Fatal("Expected error for invalid refresh token, got nil")
	}
}

func TestSimpleAuthGetUser(t *testing.T) {
	simpleAuth, err := QuickInMemory("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create SimpleAuth: %v", err)
	}

	// Register a user
	user, err := simpleAuth.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test GetUser by ID
	userProfile, err := simpleAuth.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Expected no error getting user by ID, got %v", err)
	}
	if userProfile == nil {
		t.Fatal("Expected user profile, got nil")
	}
	if userProfile.Username != "testuser" {
		t.Fatalf("Expected username 'testuser', got %s", userProfile.Username)
	}

	// Test GetUserByUsername
	userProfile2, err := simpleAuth.GetUserByUsername("testuser")
	if err != nil {
		t.Fatalf("Expected no error getting user by username, got %v", err)
	}
	if userProfile2 == nil {
		t.Fatal("Expected user profile, got nil")
	}
	if userProfile2.ID != user.ID {
		t.Fatalf("Expected user ID %s, got %s", user.ID, userProfile2.ID)
	}

	// Test GetUserByEmail
	userProfile3, err := simpleAuth.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Expected no error getting user by email, got %v", err)
	}
	if userProfile3 == nil {
		t.Fatal("Expected user profile, got nil")
	}
	if userProfile3.ID != user.ID {
		t.Fatalf("Expected user ID %s, got %s", user.ID, userProfile3.ID)
	}

	// Test with non-existent user ID
	_, err = simpleAuth.GetUser("non-existent-id")
	if err == nil {
		t.Fatal("Expected error for non-existent user ID, got nil")
	}

	// Test with non-existent username
	_, err = simpleAuth.GetUserByUsername("non-existent-user")
	if err == nil {
		t.Fatal("Expected error for non-existent username, got nil")
	}

	// Test with non-existent email
	_, err = simpleAuth.GetUserByEmail("non-existent@example.com")
	if err == nil {
		t.Fatal("Expected error for non-existent email, got nil")
	}
}

func TestSimpleAuthGetAuth(t *testing.T) {
	simpleAuth, err := QuickInMemory("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create SimpleAuth: %v", err)
	}

	// Test GetAuth method
	auth := simpleAuth.GetAuth()
	if auth == nil {
		t.Fatal("Expected Auth instance, got nil")
	}
	if auth != simpleAuth.auth {
		t.Fatal("Expected same Auth instance as internal auth")
	}
}

func TestSimpleAuthDirectoryCreation(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Test with nested directory path
	dbPath := filepath.Join(tempDir, "nested", "dir", "test.db")
	
	simpleAuth, err := QuickSQLite(dbPath, "test-secret-key")
	if err != nil {
		t.Fatalf("Expected no error creating nested directory, got %v", err)
	}
	if simpleAuth == nil {
		t.Fatal("Expected SimpleAuth instance, got nil")
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Fatal("Expected nested directory to be created")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Expected database file to be created")
	}
}

func TestSimpleAuthEnvironmentVariableTypes(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"AUTH_JWT_ACCESS_SECRET",
		"AUTH_JWT_REFRESH_SECRET",
		"AUTH_DB_TYPE",
		"AUTH_ACCESS_TOKEN_TTL",
		"AUTH_REFRESH_TOKEN_TTL",
	}
	
	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
	}
	
	// Clean up environment after test
	defer func() {
		for _, env := range envVars {
			if val, exists := originalEnv[env]; exists && val != "" {
				os.Setenv(env, val)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	// Test with different database types
	testCases := []struct {
		dbType   string
		dbURL    string
		shouldSucceed bool
	}{
		{"sqlite", "test.db", true},
		{"memory", "", true},
		{"postgres", "postgres://user:pass@localhost/db", false}, // Will fail without real postgres
		{"invalid", "test.db", false},
	}

	for _, tc := range testCases {
		t.Run("db_type_"+tc.dbType, func(t *testing.T) {
			// Set required environment variables
			os.Setenv("AUTH_JWT_ACCESS_SECRET", "test-access-secret")
			os.Setenv("AUTH_JWT_REFRESH_SECRET", "test-refresh-secret")
			os.Setenv("AUTH_DB_TYPE", tc.dbType)
			if tc.dbURL != "" {
				os.Setenv("AUTH_DB_URL", tc.dbURL)
			}
			os.Setenv("AUTH_ACCESS_TOKEN_TTL", "30m")
			os.Setenv("AUTH_REFRESH_TOKEN_TTL", "24h")

			_, err := QuickFromEnv()
			if tc.shouldSucceed && err != nil {
				t.Fatalf("Expected success for db type %s, got error: %v", tc.dbType, err)
			}
			if !tc.shouldSucceed && err == nil {
				t.Fatalf("Expected error for db type %s, got success", tc.dbType)
			}
		})
	}
}