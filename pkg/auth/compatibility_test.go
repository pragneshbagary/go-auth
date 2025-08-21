package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/internal/storage/memory"
)

func TestCompatibilityLayer(t *testing.T) {
	t.Run("NewAuthFromLegacyConfig", func(t *testing.T) {
		// Create enhanced storage for legacy config
		memStorage := memory.NewInMemoryStorage()

		// Test legacy config structure
		legacyConfig := Config{
			Storage: memStorage,
			JWT: JWTConfig{
				AccessSecret:    []byte("legacy-access-secret"),
				RefreshSecret:   []byte("legacy-refresh-secret"),
				Issuer:          "legacy-app",
				AccessTokenTTL:  time.Hour,
				RefreshTokenTTL: 24 * time.Hour,
				SigningMethod:   HS256,
			},
		}

		auth, err := NewAuthFromLegacyConfig(legacyConfig)
		if err != nil {
			t.Fatalf("Failed to create auth from legacy config: %v", err)
		}

		if auth == nil {
			t.Fatal("Auth instance should not be nil")
		}

		// Test basic functionality
		req := RegisterRequest{
			Username: "legacyuser",
			Email:    "legacy@example.com",
			Password: "password123",
		}

		user, err := auth.Register(req)
		if err != nil {
			t.Fatalf("Failed to register user with legacy config: %v", err)
		}

		if user.ID == "" {
			t.Error("User ID should not be empty")
		}
	})

	t.Run("CreateJWTManagerFromOldConfig", func(t *testing.T) {
		jwtConfig := JWTConfig{
			AccessSecret:    []byte("test-access-secret"),
			RefreshSecret:   []byte("test-refresh-secret"),
			Issuer:          "test-app",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
			SigningMethod:   HS256,
		}

		jwtManager := createJWTManagerFromOldConfig(jwtConfig)
		if jwtManager == nil {
			t.Fatal("JWT manager should not be nil")
		}

		// Test token generation
		claims := map[string]interface{}{
			"user_id":  "test-user",
			"username": "testuser",
		}

		accessToken, err := jwtManager.GenerateAccessToken("test-user", claims)
		if err != nil {
			t.Fatalf("Failed to generate access token: %v", err)
		}

		refreshToken, err := jwtManager.GenerateRefreshToken("test-user")
		if err != nil {
			t.Fatalf("Failed to generate refresh token: %v", err)
		}

		if accessToken == "" {
			t.Error("Access token should not be empty")
		}
		if refreshToken == "" {
			t.Error("Refresh token should not be empty")
		}
	})

	t.Run("MigrateFromAuthService", func(t *testing.T) {
		// Create a legacy AuthService-style configuration
		legacyStorage := memory.NewInMemoryStorage()

		legacyConfig := Config{
			Storage: legacyStorage,
			JWT: JWTConfig{
				AccessSecret:    []byte("access-secret"),
				RefreshSecret:   []byte("refresh-secret"),
				AccessTokenTTL:  time.Hour,
				RefreshTokenTTL: 24 * time.Hour,
				Issuer:          "legacy-issuer",
				SigningMethod:   HS256,
			},
		}

		// Test migration
		auth, err := MigrateFromAuthService(legacyConfig)
		if err != nil {
			t.Fatalf("Failed to migrate from AuthService: %v", err)
		}

		if auth == nil {
			t.Fatal("Migrated auth should not be nil")
		}

		// Test functionality
		req := RegisterRequest{
			Username: "migrateduser",
			Email:    "migrated@example.com",
			Password: "password123",
		}

		user, err := auth.Register(req)
		if err != nil {
			t.Fatalf("Failed to register user with migrated auth: %v", err)
		}

		if user.ID == "" {
			t.Error("User ID should not be empty")
		}

		// Verify user can login
		loginTokens, err := auth.Login("migrateduser", "password123", nil)
		if err != nil {
			t.Fatalf("Failed to login with migrated auth: %v", err)
		}

		if loginTokens.AccessToken == "" {
			t.Error("Login access token should not be empty")
		}
	})

	t.Run("BackwardCompatibilityMethods", func(t *testing.T) {
		auth, err := NewInMemory("test-secret")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Test that new Auth instance provides backward compatibility
		// These methods should exist and work
		
		// Test Logger access
		logger := auth.Logger()
		if logger == nil {
			t.Error("Logger should not be nil")
		}

		// Test EventLogger access
		eventLogger := auth.EventLogger()
		if eventLogger == nil {
			t.Error("EventLogger should not be nil")
		}

		// Test MetricsCollector access
		metrics := auth.MetricsCollector()
		if metrics == nil {
			t.Error("MetricsCollector should not be nil")
		}

		// Test Monitor access
		monitor := auth.Monitor()
		if monitor == nil {
			t.Error("Monitor should not be nil")
		}

		// Test health methods
		err = auth.Health()
		if err != nil {
			t.Errorf("Health check should pass: %v", err)
		}

		systemHealth := auth.GetSystemHealth()
		if systemHealth.Status == "" {
			t.Error("SystemHealth status should not be empty")
		}

		systemInfo := auth.GetSystemInfo()
		if systemInfo.Version == "" {
			t.Error("SystemInfo version should not be empty")
		}

		metricsData := auth.GetMetrics()
		if metricsData.StartTime.IsZero() {
			t.Error("Metrics start time should not be zero")
		}
	})

	t.Run("LegacyConfigDefaults", func(t *testing.T) {
		// Create minimal storage
		memStorage := memory.NewInMemoryStorage()

		// Test that legacy config gets proper defaults
		legacyConfig := Config{
			Storage: memStorage,
			JWT: JWTConfig{
				AccessSecret:  []byte("test-secret"),
				RefreshSecret: []byte("test-refresh-secret"),
				// Other fields left empty to test defaults
			},
		}

		auth, err := NewAuthFromLegacyConfig(legacyConfig)
		if err != nil {
			t.Fatalf("Failed to create auth from minimal legacy config: %v", err)
		}

		// Test that defaults were applied by trying to use the auth
		req := RegisterRequest{
			Username: "defaultuser",
			Email:    "default@example.com",
			Password: "password123",
		}

		user, err := auth.Register(req)
		if err != nil {
			t.Fatalf("Failed to register user with default config: %v", err)
		}

		if user.ID == "" {
			t.Error("User ID should not be empty with default config")
		}
	})

	t.Run("AuthServiceCompatLayer", func(t *testing.T) {
		// Test the original AuthService API for backward compatibility
		memStorage := memory.NewInMemoryStorage()
		
		legacyConfig := Config{
			Storage: memStorage,
			JWT: JWTConfig{
				AccessSecret:    []byte("compat-secret"),
				RefreshSecret:   []byte("compat-refresh-secret"),
				AccessTokenTTL:  time.Hour,
				RefreshTokenTTL: 24 * time.Hour,
				Issuer:          "compat-test",
				SigningMethod:   HS256,
			},
		}

		// Test NewAuthService compatibility function (original API)
		authService, err := NewAuthService(legacyConfig)
		if err != nil {
			t.Fatalf("Failed to create AuthService: %v", err)
		}

		if authService == nil {
			t.Fatal("AuthService should not be nil")
		}

		// Test Register with old payload format (original API)
		user, err := authService.Register(RegisterPayload{
			Username: "compatuser",
			Email:    "compat@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Failed to register user with compat layer: %v", err)
		}

		if user.ID == "" {
			t.Error("User ID should not be empty")
		}

		// Test Login with old response format (original API)
		loginResp, err := authService.Login("compatuser", "password123", map[string]interface{}{
			"role": "test",
		})
		if err != nil {
			t.Fatalf("Failed to login with compat layer: %v", err)
		}

		if loginResp.AccessToken == "" {
			t.Error("Access token should not be empty")
		}
		if loginResp.RefreshToken == "" {
			t.Error("Refresh token should not be empty")
		}

		// Test token validation (original API)
		claims, err := authService.ValidateAccessToken(loginResp.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate access token: %v", err)
		}

		if claims["username"] != "compatuser" {
			t.Error("Username claim should match")
		}

		// Test refresh token validation (original API)
		refreshClaims, err := authService.ValidateRefreshToken(loginResp.RefreshToken)
		if err != nil {
			t.Fatalf("Failed to validate refresh token: %v", err)
		}

		// Note: The original AuthService may not include user_id in refresh token claims
		// Just verify that we got valid claims back
		if refreshClaims == nil {
			t.Error("Refresh token claims should not be nil")
		}
	})

	t.Run("MigrationHelper", func(t *testing.T) {
		helper := NewMigrationHelper()

		// Test config conversion
		oldConfig := Config{
			Storage: memory.NewInMemoryStorage(),
			JWT: JWTConfig{
				AccessSecret:    []byte("test-secret"),
				RefreshSecret:   []byte("test-refresh-secret"),
				Issuer:          "test-app",
				AccessTokenTTL:  time.Hour,
				RefreshTokenTTL: 24 * time.Hour,
				SigningMethod:   HS256,
			},
		}

		newConfig := helper.ConvertConfigToAuthConfig(oldConfig)
		if newConfig.JWTSecret != "test-secret" {
			t.Error("JWT secret should be converted correctly")
		}
		if newConfig.JWTRefreshSecret != "test-refresh-secret" {
			t.Error("JWT refresh secret should be converted correctly")
		}
		if newConfig.JWTIssuer != "test-app" {
			t.Error("JWT issuer should be converted correctly")
		}

		// Test payload conversion
		oldPayload := RegisterPayload{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		newRequest := helper.ConvertRegisterPayloadToRequest(oldPayload)
		if newRequest.Username != oldPayload.Username {
			t.Error("Username should be converted correctly")
		}
		if newRequest.Email != oldPayload.Email {
			t.Error("Email should be converted correctly")
		}
		if newRequest.Password != oldPayload.Password {
			t.Error("Password should be converted correctly")
		}

		// Test response conversion
		newResult := &LoginResult{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}

		oldResponse := helper.ConvertLoginResultToResponse(newResult)
		if oldResponse.AccessToken != newResult.AccessToken {
			t.Error("Access token should be converted correctly")
		}
		if oldResponse.RefreshToken != newResult.RefreshToken {
			t.Error("Refresh token should be converted correctly")
		}

		// Test migration report generation
		report := helper.GenerateMigrationReport()
		if !strings.Contains(report, "Migration Guide") {
			t.Error("Migration report should contain migration guide")
		}
		if !strings.Contains(report, "NewAuthService") {
			t.Error("Migration report should mention old API patterns")
		}

		// Test storage compatibility validation
		err := helper.ValidateStorageCompatibility(memory.NewInMemoryStorage())
		if err != nil {
			t.Errorf("Memory storage should be compatible: %v", err)
		}
	})

	t.Run("AutoMigrationTool", func(t *testing.T) {
		tool := NewAutoMigrationTool()

		// Test migration suggestions
		suggestion := tool.SuggestMigration("NewAuthService")
		if !strings.Contains(suggestion, "auth.New") {
			t.Error("Should suggest using auth.New")
		}

		suggestion = tool.SuggestMigration("RegisterPayload")
		if !strings.Contains(suggestion, "RegisterRequest") {
			t.Error("Should suggest using RegisterRequest")
		}

		suggestion = tool.SuggestMigration("UnknownPattern")
		if !strings.Contains(suggestion, "No specific migration suggestion") {
			t.Error("Should provide generic message for unknown patterns")
		}

		// Test code migration example generation
		example := tool.GenerateCodeMigrationExample()
		if !strings.Contains(example, "BEFORE") || !strings.Contains(example, "AFTER") {
			t.Error("Code migration example should contain before/after sections")
		}
	})

	t.Run("DeprecationWarnings", func(t *testing.T) {
		// Temporarily disable warnings for this test
		originalWarnings := ShowDeprecationWarnings
		ShowDeprecationWarnings = false
		defer func() {
			ShowDeprecationWarnings = originalWarnings
		}()

		// Test that functions work without warnings
		memStorage := memory.NewInMemoryStorage()
		legacyConfig := Config{
			Storage: memStorage,
			JWT: JWTConfig{
				AccessSecret:  []byte("test-secret"),
				RefreshSecret: []byte("test-refresh-secret"),
			},
		}

		// These should work without printing warnings
		auth, err := NewAuthFromLegacyConfig(legacyConfig)
		if err != nil {
			t.Fatalf("Failed to create auth without warnings: %v", err)
		}

		_, err = MigrateFromAuthService(legacyConfig)
		if err != nil {
			t.Fatalf("Failed to migrate without warnings: %v", err)
		}

		_, err = NewAuthService(legacyConfig)
		if err != nil {
			t.Fatalf("Failed to create AuthService without warnings: %v", err)
		}

		// Verify auth instance works
		if auth == nil {
			t.Error("Auth instance should not be nil")
		}
	})
}