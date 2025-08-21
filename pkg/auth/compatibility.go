package auth

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pragneshbagary/go-auth/internal/jwtutils"
	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// ErrStorageNotEnhanced is returned when the provided storage doesn't implement EnhancedStorage
var ErrStorageNotEnhanced = errors.New("storage implementation must implement EnhancedStorage interface")

// Deprecation warning flag - can be disabled by setting this to false
var ShowDeprecationWarnings = true

// logDeprecationWarning logs a deprecation warning if warnings are enabled
func logDeprecationWarning(oldMethod, newMethod string) {
	if ShowDeprecationWarnings {
		log.Printf("DEPRECATION WARNING: %s is deprecated. Use %s instead. See migration guide: https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md", oldMethod, newMethod)
	}
}

// NewAuthFromLegacyConfig creates a new Auth instance from the old Config structure.
// This provides a migration path from the old API to the new API.
// Deprecated: Use New, NewSQLite, NewPostgres, or NewInMemory instead.
func NewAuthFromLegacyConfig(cfg Config) (*Auth, error) {
	logDeprecationWarning("NewAuthFromLegacyConfig", "auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory")
	
	// Convert old Config to new AuthConfig
	authConfig := &AuthConfig{
		JWTSecret:       string(cfg.JWT.AccessSecret),
		JWTRefreshSecret: string(cfg.JWT.RefreshSecret),
		JWTIssuer:       cfg.JWT.Issuer,
		AccessTokenTTL:  cfg.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.JWT.RefreshTokenTTL,
		AppName:         "go-auth-app",
		LogLevel:        "info",
	}

	// Set up storage - we need to handle the case where the old storage interface
	// might not implement EnhancedStorage
	var enhancedStorage storage.EnhancedStorage
	if es, ok := cfg.Storage.(storage.EnhancedStorage); ok {
		enhancedStorage = es
	} else {
		// For basic storage, we can't provide full functionality
		return nil, ErrStorageNotEnhanced
	}

	return newAuthWithStorage(enhancedStorage, authConfig)
}

// createJWTManagerFromOldConfig creates a JWT manager from the old config structure
func createJWTManagerFromOldConfig(cfg JWTConfig) jwtutils.TokenManager {
	return jwtutils.NewJWTManager(jwtutils.JWTConfig{
		AccessSecret:    cfg.AccessSecret,
		RefreshSecret:   cfg.RefreshSecret,
		Issuer:          cfg.Issuer,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
		SigningMethod:   cfg.SigningMethod,
	})
}

// Migration helper functions

// MigrateFromAuthService helps users migrate from the old AuthService to the new Auth.
// It provides the same interface but with the new implementation.
// Deprecated: Use auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory instead.
func MigrateFromAuthService(oldConfig Config) (*Auth, error) {
	logDeprecationWarning("MigrateFromAuthService", "auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory")
	return NewAuthFromLegacyConfig(oldConfig)
}

// ============================================================================
// AuthService Backward Compatibility Layer
// ============================================================================

// AuthServiceCompat provides backward compatibility for the old AuthService API.
// It wraps the new Auth struct and provides the old method signatures.
// Deprecated: Use Auth directly instead.
type AuthServiceCompat struct {
	auth *Auth
}

// NewAuthServiceCompat creates a new AuthServiceCompat from a configuration.
// This is provided for backward compatibility with the old API.
// Deprecated: Use auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory instead.
func NewAuthServiceCompat(cfg Config) (*AuthServiceCompat, error) {
	logDeprecationWarning("NewAuthService", "auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory")
	
	auth, err := NewAuthFromLegacyConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &AuthServiceCompat{auth: auth}, nil
}

// RegisterPayloadCompat defines the data required to register a new user (old API).
// Deprecated: Use RegisterRequest instead.
type RegisterPayloadCompat struct {
	Username string
	Email    string
	Password string
}

// Register creates a new user using the old API signature.
// Deprecated: Use Auth.Register with RegisterRequest instead.
func (s *AuthServiceCompat) Register(payload RegisterPayloadCompat) (*models.User, error) {
	logDeprecationWarning("AuthService.Register(RegisterPayload)", "Auth.Register(RegisterRequest)")
	
	// Convert old payload to new request format
	req := RegisterRequest{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	}
	
	return s.auth.Register(req)
}

// LoginResponseCompat is the old structure returned upon a successful login.
// Deprecated: Use LoginResult instead.
type LoginResponseCompat struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login authenticates a user using the old API signature.
// Deprecated: Use Auth.Login instead.
func (s *AuthServiceCompat) Login(username, password string, customClaims map[string]interface{}) (*LoginResponseCompat, error) {
	logDeprecationWarning("AuthService.Login", "Auth.Login")
	
	result, err := s.auth.Login(username, password, customClaims)
	if err != nil {
		return nil, err
	}
	
	// Convert new result to old response format
	return &LoginResponseCompat{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

// ValidateAccessToken validates an access token using the old API signature.
// Deprecated: Use Auth.ValidateAccessToken instead.
func (s *AuthServiceCompat) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	logDeprecationWarning("AuthService.ValidateAccessToken", "Auth.ValidateAccessToken")
	return s.auth.ValidateAccessToken(tokenString)
}

// ValidateRefreshToken validates a refresh token using the old API signature.
// Deprecated: Use Auth.ValidateRefreshToken instead.
func (s *AuthServiceCompat) ValidateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	logDeprecationWarning("AuthService.ValidateRefreshToken", "Auth.ValidateRefreshToken")
	return s.auth.ValidateRefreshToken(tokenString)
}

// GetAuth returns the underlying Auth instance for migration purposes.
// This allows users to gradually migrate to the new API.
func (s *AuthServiceCompat) GetAuth() *Auth {
	return s.auth
}

// ============================================================================
// Direct Alias Functions for Exact Backward Compatibility
// ============================================================================

// These functions provide exact aliases for the original service.go API
// to ensure existing code continues to work without any changes.

// NewAuthService provides exact backward compatibility with the original API.
// It returns the original AuthService type from service.go.
// Deprecated: Use auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory instead.
func NewAuthService(cfg Config) (*AuthService, error) {
	logDeprecationWarning("NewAuthService", "auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory")
	
	// Use the original NewAuthService from service.go
	// This ensures 100% backward compatibility
	return newAuthServiceFromConfig(cfg)
}

// CreateAuthService is an alias for NewAuthService for even more backward compatibility.
// Deprecated: Use auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory instead.
func CreateAuthService(cfg Config) (*AuthService, error) {
	logDeprecationWarning("CreateAuthService", "auth.New, auth.NewSQLite, auth.NewPostgres, or auth.NewInMemory")
	return NewAuthService(cfg)
}

// newAuthServiceFromConfig is the internal implementation that creates the original AuthService.
// This delegates to the renamed legacy function.
func newAuthServiceFromConfig(cfg Config) (*AuthService, error) {
	return NewAuthServiceLegacy(cfg)
}

// ============================================================================
// Migration Utilities
// ============================================================================

// MigrationHelper provides utilities to help users migrate from v1 to v2 API.
type MigrationHelper struct{}

// NewMigrationHelper creates a new migration helper.
func NewMigrationHelper() *MigrationHelper {
	return &MigrationHelper{}
}

// ConvertConfigToAuthConfig converts old Config to new AuthConfig.
func (m *MigrationHelper) ConvertConfigToAuthConfig(oldConfig Config) *AuthConfig {
	return &AuthConfig{
		JWTSecret:       string(oldConfig.JWT.AccessSecret),
		JWTRefreshSecret: string(oldConfig.JWT.RefreshSecret),
		JWTIssuer:       oldConfig.JWT.Issuer,
		AccessTokenTTL:  oldConfig.JWT.AccessTokenTTL,
		RefreshTokenTTL: oldConfig.JWT.RefreshTokenTTL,
		AppName:         "go-auth-app",
		Version:         "2.0.0",
		LogLevel:        "info",
	}
}

// ConvertRegisterPayloadToRequest converts old RegisterPayload to new RegisterRequest.
func (m *MigrationHelper) ConvertRegisterPayloadToRequest(payload RegisterPayload) RegisterRequest {
	return RegisterRequest{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	}
}

// ConvertLoginResultToResponse converts new LoginResult to old LoginResponse.
func (m *MigrationHelper) ConvertLoginResultToResponse(result *LoginResult) *LoginResponse {
	if result == nil {
		return nil
	}
	return &LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}
}

// GenerateMigrationReport analyzes old code patterns and suggests new equivalents.
func (m *MigrationHelper) GenerateMigrationReport() string {
	return `
Go-Auth v2 Migration Guide
==========================

Old API -> New API Mappings:

1. Service Creation:
   OLD: auth.NewAuthService(config)
   NEW: auth.New(dbPath, jwtSecret) or auth.NewSQLite(dbPath, jwtSecret)

2. Registration:
   OLD: service.Register(auth.RegisterPayload{...})
   NEW: service.Register(auth.RegisterRequest{...})

3. Login:
   OLD: service.Login(username, password, claims) -> *LoginResponse
   NEW: service.Login(username, password, claims) -> *LoginResult

4. User Management:
   OLD: Direct storage access
   NEW: service.Users().Get(), service.Users().Update(), etc.

5. Token Management:
   OLD: service.ValidateAccessToken(), service.ValidateRefreshToken()
   NEW: service.Tokens().Validate(), service.Tokens().Refresh(), etc.

6. Middleware:
   OLD: Custom middleware implementation
   NEW: service.Middleware().Gin(), service.Middleware().Echo(), etc.

For detailed migration instructions, see: https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md
`
}

// ValidateStorageCompatibility checks if storage implements required interfaces.
func (m *MigrationHelper) ValidateStorageCompatibility(stor storage.Storage) error {
	if _, ok := stor.(storage.EnhancedStorage); !ok {
		return fmt.Errorf("storage must implement EnhancedStorage interface for v2 compatibility. Current storage type: %T", stor)
	}
	return nil
}

// ============================================================================
// Automated Migration Tool Functions
// ============================================================================

// AutoMigrationTool provides automated code migration suggestions.
type AutoMigrationTool struct {
	helper *MigrationHelper
}

// NewAutoMigrationTool creates a new automated migration tool.
func NewAutoMigrationTool() *AutoMigrationTool {
	return &AutoMigrationTool{
		helper: NewMigrationHelper(),
	}
}

// SuggestMigration provides migration suggestions for common patterns.
func (t *AutoMigrationTool) SuggestMigration(oldPattern string) string {
	suggestions := map[string]string{
		"NewAuthService":           "Use auth.New(dbPath, jwtSecret) for SQLite or auth.NewPostgres(connStr, jwtSecret) for PostgreSQL",
		"RegisterPayload":          "Use RegisterRequest instead of RegisterPayload",
		"LoginResponse":            "Use LoginResult instead of LoginResponse",
		"ValidateAccessToken":      "Use auth.ValidateAccessToken() or auth.Tokens().Validate()",
		"ValidateRefreshToken":     "Use auth.Tokens().Refresh() for token refresh operations",
		"storage.GetUserByID":      "Use auth.Users().Get() for safe user data access",
		"storage.CreateUser":       "Use auth.Register() for user creation",
		"storage.GetUserByUsername": "Use auth.Users().GetByUsername() for safe user data access",
	}

	if suggestion, exists := suggestions[oldPattern]; exists {
		return suggestion
	}

	return fmt.Sprintf("No specific migration suggestion available for '%s'. Please refer to the migration guide.", oldPattern)
}

// GenerateCodeMigrationExample generates example code showing before/after migration.
func (t *AutoMigrationTool) GenerateCodeMigrationExample() string {
	return `
Code Migration Examples
=======================

BEFORE (v1 API):
----------------
config := auth.Config{
    Storage: sqliteStorage,
    JWT: auth.JWTConfig{
        AccessSecret:    []byte("secret"),
        RefreshSecret:   []byte("refresh-secret"),
        AccessTokenTTL:  time.Hour,
        RefreshTokenTTL: 24 * time.Hour,
        Issuer:          "my-app",
        SigningMethod:   auth.HS256,
    },
}

authService, err := auth.NewAuthService(config)
if err != nil {
    log.Fatal(err)
}

user, err := authService.Register(auth.RegisterPayload{
    Username: "alice",
    Email:    "alice@example.com",
    Password: "password123",
})

loginResp, err := authService.Login("alice", "password123", nil)

AFTER (v2 API):
---------------
authService, err := auth.NewSQLite("auth.db", "secret")
if err != nil {
    log.Fatal(err)
}

user, err := authService.Register(auth.RegisterRequest{
    Username: "alice",
    Email:    "alice@example.com",
    Password: "password123",
})

loginResult, err := authService.Login("alice", "password123", nil)

// Enhanced features available:
users := authService.Users()
profile, err := users.Get(user.ID)

tokens := authService.Tokens()
newTokens, err := tokens.Refresh(loginResult.RefreshToken)

middleware := authService.Middleware()
router.Use(middleware.Gin())
`
}