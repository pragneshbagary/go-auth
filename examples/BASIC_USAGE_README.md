# Basic Usage Examples

This directory contains examples demonstrating the basic usage of the go-auth package with the new v2 API.

## Examples Overview

### [basic_usage_example.go](basic_usage_example.go)
Demonstrates the fundamental operations using the full Auth API:
- Simple SQLite setup
- User registration
- Login with tokens
- Token validation
- Token refresh
- User profile retrieval
- Custom claims usage
- Health checks

**Run it:**
```bash
go run examples/basic_usage_example.go
```

### [advanced_usage_example.go](advanced_usage_example.go)
Shows advanced features and configurations:
- Custom configuration setup
- Advanced user management with metadata
- Password reset workflow
- Advanced token management
- User listing and management
- Token revocation and cleanup
- Monitoring and metrics
- System information

**Run it:**
```bash
go run examples/advanced_usage_example.go
```

### [simple_auth_example.go](simple_auth_example.go)
Ultra-simple usage with the SimpleAuth wrapper:
- Quick setup with defaults
- Minimal configuration
- Environment-based setup
- All basic operations with simplified API

**Run it:**
```bash
go run examples/simple_auth_example.go
```

## Key Concepts Demonstrated

### 1. Initialization Patterns

**Simple Initialization:**
```go
// Most common - SQLite with defaults
authService, err := auth.New("auth.db", "your-jwt-secret")

// In-memory for testing
authService, err := auth.NewInMemory("your-jwt-secret")

// PostgreSQL for production
authService, err := auth.NewPostgres("postgres://user:pass@localhost/db", "secret")
```

**Environment-Based:**
```go
// Set environment variables first
os.Setenv("AUTH_JWT_ACCESS_SECRET", "access-secret")
os.Setenv("AUTH_JWT_REFRESH_SECRET", "refresh-secret")

// Then initialize
authService, err := auth.NewFromEnv()
```

**Ultra-Simple:**
```go
// SimpleAuth wrapper for minimal setup
simpleAuth, err := auth.Quick("your-jwt-secret")
```

### 2. Core Operations

**User Registration:**
```go
user, err := authService.Register(auth.RegisterRequest{
    Username: "alice",
    Email:    "alice@example.com",
    Password: "secure_password123",
})
```

**Login and Token Generation:**
```go
// Basic login
loginResult, err := authService.Login("alice", "secure_password123", nil)

// Login with custom claims
loginResult, err := authService.Login("alice", "secure_password123", map[string]interface{}{
    "role": "admin",
    "department": "engineering",
})
```

**Token Operations:**
```go
// Validate access token
claims, err := authService.ValidateAccessToken(loginResult.AccessToken)

// Refresh tokens
refreshResult, err := authService.RefreshToken(loginResult.RefreshToken)
```

### 3. Component-Based API

**User Management:**
```go
users := authService.Users()

// Get user profile (safe, no sensitive data)
profile, err := users.Get(userID)

// Update user
err = users.Update(userID, auth.UserUpdate{
    Email: &newEmail,
    Metadata: map[string]interface{}{"role": "admin"},
})

// Change password
err = users.ChangePassword(userID, oldPassword, newPassword)
```

**Token Management:**
```go
tokens := authService.Tokens()

// Advanced validation
user, err := tokens.Validate(accessToken)
isValid := tokens.IsValid(accessToken)

// Revocation
err = tokens.Revoke(accessToken)
err = tokens.RevokeAll(userID)
```

### 4. Configuration Options

**Custom Configuration:**
```go
config := &auth.AuthConfig{
    JWTSecret:        "your-secret",
    JWTRefreshSecret: "your-refresh-secret",
    JWTIssuer:        "my-app",
    AppName:          "My Application",
    DatabasePath:     "custom.db",
    AccessTokenTTL:   30 * time.Minute,
    RefreshTokenTTL:  7 * 24 * time.Hour,
}

authService, err := auth.NewWithConfig(config)
```

**Environment Variables:**
```bash
export AUTH_JWT_ACCESS_SECRET="your-access-secret"
export AUTH_JWT_REFRESH_SECRET="your-refresh-secret"
export AUTH_DB_TYPE="sqlite"
export AUTH_DB_URL="auth.db"
export AUTH_ACCESS_TOKEN_TTL="15m"
export AUTH_REFRESH_TOKEN_TTL="168h"
```

## Best Practices Demonstrated

### 1. Error Handling
All examples show proper error handling patterns:
```go
if err != nil {
    log.Fatalf("Operation failed: %v", err)
}
```

### 2. Security
- Never hardcode secrets in production
- Use environment variables for configuration
- Validate tokens before using claims
- Handle authentication failures gracefully

### 3. Token Management
- Always refresh tokens before they expire
- Revoke tokens on logout
- Use custom claims for authorization
- Validate tokens on protected endpoints

### 4. User Management
- Use safe user profile methods (no sensitive data)
- Update user profiles with metadata
- Implement password change workflows
- List users with pagination

## Common Patterns

### Registration Flow
```go
// 1. Register user
user, err := authService.Register(auth.RegisterRequest{
    Username: username,
    Email:    email,
    Password: password,
})

// 2. Handle registration errors
if err != nil {
    // Handle user already exists, weak password, etc.
    return err
}

// 3. Optionally login immediately
loginResult, err := authService.Login(username, password, nil)
```

### Authentication Flow
```go
// 1. Login
loginResult, err := authService.Login(username, password, customClaims)
if err != nil {
    // Handle invalid credentials
    return err
}

// 2. Store tokens securely (client-side)
// Store loginResult.AccessToken and loginResult.RefreshToken

// 3. Use access token for API calls
claims, err := authService.ValidateAccessToken(accessToken)
if err != nil {
    // Token invalid or expired, try refresh
    refreshResult, err := authService.RefreshToken(refreshToken)
    if err != nil {
        // Refresh failed, require re-login
        return err
    }
    // Use new tokens
}
```

### Profile Management Flow
```go
users := authService.Users()

// 1. Get current profile
profile, err := users.Get(userID)

// 2. Update profile
err = users.Update(userID, auth.UserUpdate{
    Email: &newEmail,
    Metadata: map[string]interface{}{
        "role": "admin",
        "last_updated": time.Now(),
    },
})

// 3. Change password if needed
err = users.ChangePassword(userID, oldPassword, newPassword)
```

## Running the Examples

### Prerequisites
```bash
# Install dependencies
go mod download

# Set environment variables (optional)
export AUTH_JWT_ACCESS_SECRET="your-access-secret"
export AUTH_JWT_REFRESH_SECRET="your-refresh-secret"
```

### Run Individual Examples
```bash
# Basic usage
go run examples/basic_usage_example.go

# Advanced usage
go run examples/advanced_usage_example.go

# SimpleAuth wrapper
go run examples/simple_auth_example.go
```

### Expected Output
Each example will:
1. Initialize the auth service
2. Register test users
3. Perform authentication operations
4. Demonstrate various features
5. Show success/failure messages
6. Clean up resources

## Next Steps

After understanding the basic usage:

1. **Framework Integration**: Check out framework-specific examples (Gin, Echo, Fiber)
2. **Advanced Features**: Explore password reset and token management examples
3. **Production Setup**: Review monitoring and logging examples
4. **Middleware**: Learn about HTTP middleware integration

## Troubleshooting

### Common Issues

**Database Locked (SQLite):**
```go
// Ensure database file is not in use by another process
// Use different database files for different examples
authService, err := auth.New("example_specific.db", "secret")
```

**Token Validation Errors:**
```go
// Ensure you're using the correct token type
claims, err := authService.ValidateAccessToken(accessToken) // Not refresh token
```

**User Already Exists:**
```go
// Handle registration errors gracefully
user, err := authService.Register(req)
if err != nil {
    if strings.Contains(err.Error(), "already exists") {
        // User exists, maybe try login instead
        return handleExistingUser(username)
    }
    return err
}
```

### Debug Mode

Enable debug logging to see detailed operations:
```bash
export AUTH_LOG_LEVEL=debug
go run examples/basic_usage_example.go
```

This will show detailed logs of all authentication operations, database queries, and token operations.