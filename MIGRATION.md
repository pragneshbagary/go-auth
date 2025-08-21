# Migration Guide: go-auth v1 to v2

This guide helps you migrate from go-auth v1 to v2, which introduces a significantly improved API while maintaining backward compatibility.

## Overview of Changes

### Key Improvements in v2

1. **Simplified API**: Much easier initialization and usage
2. **Component-Based Architecture**: Organized functionality into logical components
3. **Auto-Migration**: Automatic database setup and schema migration
4. **Environment Configuration**: Full environment variable support
5. **Enhanced Middleware**: Framework-specific adapters for popular web frameworks
6. **Advanced Features**: Password reset, token management, user management
7. **Monitoring**: Built-in logging, metrics, and health checks

### Backward Compatibility

**Important**: v1 API continues to work in v2 through compatibility aliases. You can migrate gradually without breaking existing code.

## Migration Steps

### Step 1: Update Import (if needed)

The import path remains the same:

```go
import "github.com/pragneshbagary/go-auth/pkg/auth"
```

### Step 2: Simplify Initialization

#### v1 Initialization (Complex)

```go
// v1 - Complex configuration
storage := memory.NewInMemoryStorage()

cfg := auth.Config{
    Storage: storage,
    JWT: auth.JWTConfig{
        AccessSecret:    []byte("your-super-secret-access-key"),
        RefreshSecret:   []byte("your-super-secret-refresh-key"),
        Issuer:          "my-awesome-app",
        AccessTokenTTL:  15 * time.Minute,
        RefreshTokenTTL: 7 * 24 * time.Hour,
        SigningMethod:   auth.HS256,
    },
}

authService, err := auth.NewAuthService(cfg)
if err != nil {
    log.Fatalf("Failed to create AuthService: %v", err)
}
```

#### v2 Initialization (Simple)

```go
// v2 - Simple initialization
authService, err := auth.New("auth.db", "your-jwt-secret")
if err != nil {
    log.Fatalf("Failed to create auth service: %v", err)
}

// Or even simpler with SimpleAuth
simpleAuth, err := auth.Quick("your-jwt-secret")
if err != nil {
    log.Fatalf("Failed to create simple auth: %v", err)
}

// Or from environment variables
authService, err := auth.NewFromEnv()
if err != nil {
    log.Fatalf("Failed to create auth service: %v", err)
}
```

### Step 3: Update Registration and Login

#### v1 API

```go
// v1 - Registration
registerPayload := auth.RegisterPayload{
    Username: "testuser",
    Email:    "test@example.com",
    Password: "StrongPassword123!",
}
user, err := authService.Register(registerPayload)

// v1 - Login
customClaims := map[string]interface{}{"role": "admin"}
loginResponse, err := authService.Login("testuser", "StrongPassword123!", customClaims)
```

#### v2 API

```go
// v2 - Registration (cleaner struct name)
user, err := authService.Register(auth.RegisterRequest{
    Username: "testuser",
    Email:    "test@example.com",
    Password: "StrongPassword123!",
})

// v2 - Login (same interface, better naming)
customClaims := map[string]interface{}{"role": "admin"}
loginResult, err := authService.Login("testuser", "StrongPassword123!", customClaims)

// v2 - SimpleAuth (even simpler)
user, err := simpleAuth.Register("testuser", "test@example.com", "StrongPassword123!")
tokens, err := simpleAuth.Login("testuser", "StrongPassword123!")
```

### Step 4: Migrate to Component-Based API

v2 introduces organized components for different functionality:

#### User Management

```go
// v2 - Access Users component
users := authService.Users()

// Get user profile (safe, no sensitive data)
profile, err := users.Get(userID)
profile, err := users.GetByEmail("user@example.com")
profile, err := users.GetByUsername("username")

// Update user profile
err = users.Update(userID, auth.UserUpdate{
    Email: &newEmail,
    Metadata: map[string]interface{}{
        "role": "admin",
        "department": "engineering",
    },
})

// Change password
err = users.ChangePassword(userID, oldPassword, newPassword)

// Password reset workflow
resetToken, err := users.CreateResetToken("user@example.com")
err = users.ResetPassword(resetToken.Token, newPassword)

// List users
userList, err := users.List(10, 0)
err = users.Delete(userID)
```

#### Token Management

```go
// v2 - Access Tokens component
tokens := authService.Tokens()

// Advanced token validation
user, err := tokens.Validate(accessToken)
isValid := tokens.IsValid(accessToken)

// Batch validation for performance
results := tokens.ValidateBatch([]string{token1, token2, token3})

// Token refresh with automatic rotation
refreshResult, err := tokens.Refresh(refreshToken)

// Token revocation
err = tokens.Revoke(accessToken)           // Single token
err = tokens.RevokeAll(userID)             // All user tokens

// Session management
sessions, err := tokens.ListActiveSessions(userID)
sessionInfo, err := tokens.GetSessionInfo(accessToken)

// Cleanup
err = tokens.CleanupExpired()
```

### Step 5: Upgrade Middleware

#### v1 Middleware (Manual)

```go
// v1 - Manual middleware implementation
func authMiddleware(authService *auth.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "No token provided"})
            c.Abort()
            return
        }
        
        // Remove Bearer prefix
        if len(token) > 7 && token[:7] == "Bearer " {
            token = token[7:]
        }
        
        claims, err := authService.ValidateAccessToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Manually set user in context
        c.Set("user_claims", claims)
        c.Next()
    }
}
```

#### v2 Middleware (Built-in)

```go
// v2 - Built-in middleware with framework adapters
middleware := authService.Middleware()

// Gin
protected := r.Group("/api/protected")
protected.Use(middleware.Gin())
{
    protected.GET("/profile", func(c *gin.Context) {
        // User automatically injected into context
        user, ok := auth.GetUserFromGin(c)
        if !ok {
            c.JSON(500, gin.H{"error": "User not found"})
            return
        }
        c.JSON(200, gin.H{"user": user})
    })
}

// Optional authentication
r.GET("/api/optional", middleware.GinOptional(), handler)

// Echo
protected := e.Group("/api/protected")
protected.Use(middleware.Echo())

// Fiber
protected := app.Group("/api/protected")
protected.Use(middleware.Fiber())

// Standard HTTP
mux.Handle("/protected", authService.Protect(handler))
```

### Step 6: Add Environment Configuration

Create a `.env` file or set environment variables:

```bash
# Environment variables for v2
AUTH_JWT_ACCESS_SECRET="your-access-secret"
AUTH_JWT_REFRESH_SECRET="your-refresh-secret"
AUTH_DB_TYPE="sqlite"
AUTH_DB_URL="auth.db"
AUTH_JWT_ISSUER="my-app"
AUTH_ACCESS_TOKEN_TTL="15m"
AUTH_REFRESH_TOKEN_TTL="168h"
AUTH_APP_NAME="My Application"
AUTH_LOG_LEVEL="info"
```

Then use environment-based initialization:

```go
// Automatically loads from environment
authService, err := auth.NewFromEnv()
```

### Step 7: Add Monitoring and Health Checks

```go
// v2 - Built-in monitoring
// Health checks
err := authService.Health()
info := authService.GetSystemInfo()

// Metrics
metrics := authService.GetMetrics()
collector := authService.MetricsCollector()
loginRate := collector.GetLoginSuccessRate()

// Structured logging
logger := authService.Logger()
logger.Info("User operation", map[string]interface{}{
    "user_id": userID,
    "operation": "login",
})

// HTTP endpoints for monitoring
mux := http.NewServeMux()
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    if err := authService.Health(); err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
})
```

## Migration Strategies

### Strategy 1: Gradual Migration (Recommended)

1. **Keep existing v1 code working** - No immediate changes needed
2. **Add new v2 features incrementally** - Use new components for new functionality
3. **Migrate endpoints one by one** - Update middleware and handlers gradually
4. **Update initialization last** - Switch to simpler constructors when ready

### Strategy 2: Complete Migration

1. **Update initialization** - Switch to new constructors
2. **Update all middleware** - Use built-in framework adapters
3. **Migrate to components** - Use Users, Tokens, Middleware components
4. **Add monitoring** - Implement health checks and metrics
5. **Test thoroughly** - Ensure all functionality works

### Strategy 3: New Projects

For new projects, start directly with v2 API:

```go
// New project - use v2 from the start
func main() {
    // Simple initialization
    auth, err := auth.NewFromEnv()
    if err != nil {
        log.Fatal(err)
    }

    // Use components
    users := auth.Users()
    tokens := auth.Tokens()
    middleware := auth.Middleware()

    // Set up web server with built-in middleware
    r := gin.Default()
    protected := r.Group("/api")
    protected.Use(middleware.Gin())
    
    // Add monitoring
    r.GET("/health", func(c *gin.Context) {
        if err := auth.Health(); err != nil {
            c.JSON(503, gin.H{"status": "unhealthy"})
            return
        }
        c.JSON(200, gin.H{"status": "healthy"})
    })
}
```

## Common Migration Issues

### Issue 1: Import Changes

**Problem**: Some internal imports may have changed.

**Solution**: Use the main package import and access components through the main API:

```go
// Instead of importing internal packages
import "github.com/pragneshbagary/go-auth/pkg/auth"

// Access everything through the main auth service
authService, _ := auth.New("auth.db", "secret")
users := authService.Users()
tokens := authService.Tokens()
```

### Issue 2: Configuration Structure

**Problem**: v1 configuration structure is complex.

**Solution**: Use simplified constructors or environment variables:

```go
// v1 complex config
cfg := auth.Config{...}

// v2 simple config
auth, err := auth.New("auth.db", "secret")

// Or environment config
auth, err := auth.NewFromEnv()
```

### Issue 3: Middleware Integration

**Problem**: Custom middleware needs updating.

**Solution**: Use built-in framework adapters:

```go
// Replace custom middleware
middleware := authService.Middleware()
r.Use(middleware.Gin())  // For Gin
e.Use(middleware.Echo()) // For Echo
app.Use(middleware.Fiber()) // For Fiber
```

### Issue 4: Database Migration

**Problem**: Existing database schema needs updating.

**Solution**: v2 handles migrations automatically:

```go
// v2 automatically handles database migrations
authService, err := auth.New("existing.db", "secret")
// Database schema will be updated automatically
```

## Testing Migration

### Unit Tests

```go
func TestMigration(t *testing.T) {
    // Test that v1 and v2 APIs work together
    authService, err := auth.New("test.db", "test-secret")
    require.NoError(t, err)

    // Test registration (both APIs should work)
    user1, err := authService.Register(auth.RegisterRequest{
        Username: "test1",
        Email:    "test1@example.com",
        Password: "password123",
    })
    require.NoError(t, err)

    // Test login
    result, err := authService.Login("test1", "password123", nil)
    require.NoError(t, err)
    require.NotEmpty(t, result.AccessToken)

    // Test new components
    users := authService.Users()
    profile, err := users.Get(user1.ID)
    require.NoError(t, err)
    require.Equal(t, user1.Username, profile.Username)
}
```

### Integration Tests

```go
func TestMiddlewareMigration(t *testing.T) {
    authService, _ := auth.NewInMemory("test-secret")
    
    // Register test user
    user, _ := authService.Register(auth.RegisterRequest{
        Username: "test",
        Email:    "test@example.com", 
        Password: "password123",
    })
    
    // Login to get token
    result, _ := authService.Login("test", "password123", nil)
    
    // Test middleware
    r := gin.New()
    middleware := authService.Middleware()
    
    r.GET("/protected", middleware.Gin(), func(c *gin.Context) {
        user, ok := auth.GetUserFromGin(c)
        if !ok {
            c.JSON(500, gin.H{"error": "User not found"})
            return
        }
        c.JSON(200, gin.H{"username": user.Username})
    })
    
    // Test with token
    req := httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+result.AccessToken)
    
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    
    require.Equal(t, 200, w.Code)
}
```

## Performance Considerations

### Database Connections

v2 includes better connection management:

```go
// v2 automatically handles connection pooling for PostgreSQL
authService, err := auth.NewPostgres("postgres://user:pass@localhost/db")

// Connection pooling is configured automatically
// No manual connection management needed
```

### Token Validation

v2 includes batch validation for better performance:

```go
// v2 - Batch validation for multiple tokens
tokens := authService.Tokens()
results := tokens.ValidateBatch([]string{token1, token2, token3})

// More efficient than individual validations
for _, result := range results {
    if result.Valid {
        // Process valid token
    }
}
```

### Caching

v2 supports optional caching:

```go
// v2 - Optional Redis caching (if configured)
// Caching is handled automatically when Redis is available
// No code changes needed
```

## Rollback Plan

If you need to rollback to v1 behavior:

1. **Keep v1 imports**: Continue using the old API methods
2. **Disable new features**: Don't use new components
3. **Use compatibility mode**: v2 maintains full v1 compatibility

```go
// Rollback approach - use v1 style with v2 package
cfg := auth.Config{
    Storage: storage,
    JWT: auth.JWTConfig{
        AccessSecret: []byte("secret"),
        // ... v1 config
    },
}

// This still works in v2
authService, err := auth.NewAuthService(cfg)
```

## Support

If you encounter issues during migration:

1. **Check Examples**: See `examples/` directory for working code
2. **Read Documentation**: Full API documentation available
3. **Open Issues**: Report problems on GitHub
4. **Community Support**: Ask questions in GitHub Discussions

## Conclusion

The migration from v1 to v2 provides significant benefits:

- **Simpler API**: Much easier to use and understand
- **Better Features**: Advanced user management, token handling, monitoring
- **Production Ready**: Built-in health checks, metrics, logging
- **Framework Integration**: Native middleware for popular frameworks
- **Backward Compatible**: Existing code continues to work

Take your time with the migration and leverage the backward compatibility to migrate gradually. The new API will make your authentication code much cleaner and more maintainable.