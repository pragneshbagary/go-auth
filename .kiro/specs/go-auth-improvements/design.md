# Design Document

## Overview

This design outlines improvements to the go-Auth package to create a more user-friendly, feature-rich authentication library. The design focuses on simplifying the developer experience while maintaining security and flexibility. Key improvements include simplified initialization, built-in database management, middleware helpers, enhanced user management, and better configuration options.

## Architecture

### Current Architecture Analysis

The existing package has a solid foundation with:
- Clean separation between storage interface and implementations
- JWT token management with access/refresh token support
- Secure password hashing using Argon2id
- Multiple storage backend support (memory, SQLite, PostgreSQL)

### Proposed Architecture Enhancements

```
┌─────────────────────────────────────────────────────────────┐
│                    Public API Layer                         │
├─────────────────────────────────────────────────────────────┤
│  SimpleAuth  │  Middleware  │  UserManager  │  TokenManager │
├─────────────────────────────────────────────────────────────┤
│                   Enhanced AuthService                      │
├─────────────────────────────────────────────────────────────┤
│  ConfigManager  │  MigrationManager  │  ErrorHandler       │
├─────────────────────────────────────────────────────────────┤
│              Storage Layer (Enhanced)                       │
├─────────────────────────────────────────────────────────────┤
│  Memory  │  SQLite+  │  PostgreSQL+  │  TokenBlacklist     │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Auth Component (Renamed from AuthService)

The main authentication component with improved naming for better user experience.

```go
type Auth struct {
    service *authService // internal service
    config  *Config
}

type Config struct {
    DatabasePath    string // For SQLite
    DatabaseURL     string // For PostgreSQL
    JWTSecret      string
    AppName        string
    TokenTTL       time.Duration
}

// Intuitive constructor functions
func New(databasePath string, jwtSecret string) (*Auth, error)
func NewWithPostgres(databaseURL string, jwtSecret string) (*Auth, error)
func NewFromEnv() (*Auth, error)

// Quick setup functions for common scenarios
func NewSQLite(dbPath string, jwtSecret string) (*Auth, error)
func NewPostgres(connectionString string, jwtSecret string) (*Auth, error)
func NewInMemory(jwtSecret string) (*Auth, error)
```

### 2. SimpleAuth Component (Optional Wrapper)

An even simpler wrapper for developers who want minimal configuration:

```go
type SimpleAuth struct {
    auth *Auth
}

// Ultra-simple constructors
func Quick(jwtSecret string) (*SimpleAuth, error) // Uses SQLite with default path
func QuickPostgres(dbURL string, jwtSecret string) (*SimpleAuth, error)
func QuickFromEnv() (*SimpleAuth, error) // Reads all config from environment
```

### 2. Enhanced Storage Interface

Extend the current storage interface to support additional operations:

```go
type EnhancedStorage interface {
    storage.Storage // Embed existing interface
    
    // User management
    UpdateUser(userID string, updates UserUpdates) error
    DeleteUser(userID string) error
    GetUserByID(userID string) (*models.User, error)
    GetUserByEmail(email string) (*models.User, error)
    ListUsers(limit, offset int) ([]*models.User, error)
    
    // Password management
    UpdatePassword(userID string, passwordHash string) error
    
    // Token management
    BlacklistToken(tokenID string, expiresAt time.Time) error
    IsTokenBlacklisted(tokenID string) (bool, error)
    CleanupExpiredTokens() error
    
    // Health and migration
    Ping() error
    Migrate() error
    GetSchemaVersion() (int, error)
}
```

### 3. Middleware Component

HTTP middleware for popular Go web frameworks with intuitive method names:

```go
type Middleware struct {
    auth *Auth
}

// Generic HTTP middleware with clear names
func (m *Middleware) Protect(next http.Handler) http.Handler
func (m *Middleware) Optional(next http.Handler) http.Handler

// Framework-specific adapters with consistent naming
func (m *Middleware) Gin() gin.HandlerFunc
func (m *Middleware) GinOptional() gin.HandlerFunc
func (m *Middleware) Echo() echo.MiddlewareFunc
func (m *Middleware) EchoOptional() echo.MiddlewareFunc
func (m *Middleware) Fiber() fiber.Handler
func (m *Middleware) FiberOptional() fiber.Handler

// Convenience methods on Auth struct
func (a *Auth) Middleware() *Middleware
func (a *Auth) Protect(next http.Handler) http.Handler // Direct access
```

### 4. Users Component (Renamed from UserManager)

Enhanced user management operations with intuitive method names:

```go
type Users struct {
    storage EnhancedStorage
}

type UserUpdate struct {
    Email    *string
    Username *string
    Metadata map[string]interface{}
}

type ResetToken struct {
    Token     string
    UserID    string
    ExpiresAt time.Time
}

// Clear, action-oriented method names
func (u *Users) Update(userID string, updates UserUpdate) error
func (u *Users) ChangePassword(userID, oldPassword, newPassword string) error
func (u *Users) CreateResetToken(email string) (*ResetToken, error)
func (u *Users) ResetPassword(token, newPassword string) error
func (u *Users) Get(userID string) (*UserProfile, error)
func (u *Users) GetByEmail(email string) (*UserProfile, error)
func (u *Users) GetByUsername(username string) (*UserProfile, error)
func (u *Users) List(limit, offset int) ([]*UserProfile, error)
func (u *Users) Delete(userID string) error

// Convenience methods on Auth struct
func (a *Auth) Users() *Users
```

### 5. Tokens Component (Renamed from TokenManager)

Enhanced token management with clear, action-oriented method names:

```go
type Tokens struct {
    jwtManager jwtutils.TokenManager
    storage    EnhancedStorage
}

type RefreshResult struct {
    AccessToken  string
    RefreshToken string
}

// Simple, clear method names
func (t *Tokens) Refresh(refreshToken string) (*RefreshResult, error)
func (t *Tokens) Revoke(tokenString string) error
func (t *Tokens) RevokeAll(userID string) error
func (t *Tokens) Validate(tokenString string) (*models.User, error)
func (t *Tokens) IsValid(tokenString string) bool

// Convenience methods on Auth struct
func (a *Auth) Tokens() *Tokens
func (a *Auth) RefreshToken(refreshToken string) (*RefreshResult, error) // Direct access
```

### 6. Configuration Management

Environment-based configuration with validation:

```go
type ConfigManager struct {
    config *Config
}

type EnhancedConfig struct {
    // Database
    DatabaseType string `env:"AUTH_DB_TYPE" default:"sqlite"`
    DatabaseURL  string `env:"AUTH_DB_URL" default:"auth.db"`
    
    // JWT
    JWTAccessSecret  string        `env:"AUTH_JWT_ACCESS_SECRET" required:"true"`
    JWTRefreshSecret string        `env:"AUTH_JWT_REFRESH_SECRET" required:"true"`
    JWTIssuer        string        `env:"AUTH_JWT_ISSUER" default:"go-auth"`
    AccessTokenTTL   time.Duration `env:"AUTH_ACCESS_TOKEN_TTL" default:"15m"`
    RefreshTokenTTL  time.Duration `env:"AUTH_REFRESH_TOKEN_TTL" default:"168h"`
    
    // Security
    PasswordMinLength int  `env:"AUTH_PASSWORD_MIN_LENGTH" default:"8"`
    RequireEmail      bool `env:"AUTH_REQUIRE_EMAIL" default:"true"`
    
    // Logging
    LogLevel string `env:"AUTH_LOG_LEVEL" default:"info"`
}

func LoadConfigFromEnv() (*EnhancedConfig, error)
func (c *EnhancedConfig) Validate() error
```

## Data Models

### Enhanced User Model

```go
type User struct {
    ID           string                 `json:"id"`
    Username     string                 `json:"username"`
    Email        string                 `json:"email"`
    PasswordHash string                 `json:"-"`
    CreatedAt    time.Time             `json:"created_at"`
    UpdatedAt    time.Time             `json:"updated_at"`
    LastLoginAt  *time.Time            `json:"last_login_at,omitempty"`
    IsActive     bool                  `json:"is_active"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type UserProfile struct {
    ID          string                 `json:"id"`
    Username    string                 `json:"username"`
    Email       string                 `json:"email"`
    CreatedAt   time.Time             `json:"created_at"`
    LastLoginAt *time.Time            `json:"last_login_at,omitempty"`
    IsActive    bool                  `json:"is_active"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

### Token Blacklist Model

```go
type BlacklistedToken struct {
    TokenID   string    `json:"token_id"`
    ExpiresAt time.Time `json:"expires_at"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Migration Schema

```go
type Migration struct {
    Version     int       `json:"version"`
    Description string    `json:"description"`
    AppliedAt   time.Time `json:"applied_at"`
}
```

## Error Handling

### Structured Error Types

```go
type AuthError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// Error codes
const (
    ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
    ErrCodeUserExists        = "USER_EXISTS"
    ErrCodeUserNotFound      = "USER_NOT_FOUND"
    ErrCodeInvalidToken      = "INVALID_TOKEN"
    ErrCodeTokenExpired      = "TOKEN_EXPIRED"
    ErrCodeTokenRevoked      = "TOKEN_REVOKED"
    ErrCodeWeakPassword      = "WEAK_PASSWORD"
    ErrCodeDatabaseError     = "DATABASE_ERROR"
    ErrCodeConfigError       = "CONFIG_ERROR"
)

func NewAuthError(code, message string) *AuthError
func (e *AuthError) Error() string
func (e *AuthError) Is(target error) bool
```

### HTTP Error Responses

```go
type HTTPErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}

func WriteErrorResponse(w http.ResponseWriter, err error)
```

## Testing Strategy

### Unit Testing
- Test all new components with comprehensive unit tests
- Mock external dependencies (database, HTTP requests)
- Test error conditions and edge cases
- Maintain >90% code coverage

### Integration Testing
- Test database operations with real database instances
- Test middleware with actual HTTP requests
- Test token lifecycle (generation, validation, refresh, revocation)
- Test configuration loading from environment variables

### Example Testing
- Automated testing of all example code
- Verification that examples work with different storage backends
- Performance testing with realistic workloads

### Security Testing
- Test password hashing and validation
- Test JWT token security (signature validation, expiration)
- Test SQL injection prevention
- Test timing attack prevention

## Migration Strategy

### Database Migrations

```go
type MigrationManager struct {
    storage EnhancedStorage
}

type MigrationStep struct {
    Version int
    Up      func(storage EnhancedStorage) error
    Down    func(storage EnhancedStorage) error
}

func (mm *MigrationManager) Migrate() error
func (mm *MigrationManager) Rollback(targetVersion int) error
func (mm *MigrationManager) GetCurrentVersion() (int, error)
```

### API Naming Philosophy

### Improved Naming Conventions

The new API follows these naming principles:
- **Constructors**: Use simple, clear names like `New()`, `NewSQLite()`, `NewPostgres()`
- **Components**: Use noun names that represent what they manage (`Auth`, `Users`, `Tokens`, `Middleware`)
- **Methods**: Use verb names that clearly describe the action (`Register()`, `Login()`, `Protect()`, `Refresh()`)
- **Types**: Use clear, descriptive names without unnecessary prefixes (`Config` instead of `AuthConfig`)

### Method Naming Examples

```go
// Old naming (current)
authService, err := auth.NewAuthService(cfg)
loginResponse, err := authService.Login(username, password, claims)
userManager := NewUserManager(storage)
tokenManager := NewTokenManager(jwtManager, storage)

// New naming (improved)
auth, err := goauth.New(dbPath, jwtSecret)
tokens, err := auth.Login(username, password, claims)
user, err := auth.Users().Get(userID)
newTokens, err := auth.Tokens().Refresh(refreshToken)
```

## API Compatibility

- Maintain backward compatibility with existing API during transition period
- Provide alias functions that map old names to new names
- Deprecate old methods gradually with clear migration paths
- Provide automated migration tool to help users upgrade from v1 to v2

## Performance Considerations

### Database Optimizations
- Add database indexes for frequently queried fields (username, email, token_id)
- Implement connection pooling for PostgreSQL
- Add query optimization for user listing and search operations

### Caching Strategy
- Optional Redis integration for token blacklist caching
- In-memory caching for frequently accessed user data
- Configurable cache TTL settings

### Memory Management
- Efficient token validation without database hits when possible
- Cleanup routines for expired tokens and sessions
- Configurable limits for concurrent operations

## Security Enhancements

### Additional Security Features
- Rate limiting for authentication attempts
- Account lockout after failed login attempts
- Audit logging for security events
- Optional 2FA support framework

### Token Security
- JWT token ID (jti) claims for revocation support
- Configurable token rotation policies
- Secure token storage recommendations in documentation