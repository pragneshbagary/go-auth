# Framework Integration Examples

This directory contains comprehensive examples showing how to integrate go-auth with popular Go web frameworks.

## Examples Overview

### [gin_integration_example.go](gin_integration_example.go)
Complete Gin framework integration with:
- Built-in middleware for protected routes
- User registration and login endpoints
- Profile management endpoints
- Admin-only routes with role-based access
- Optional authentication endpoints
- Health and metrics endpoints

### [echo_integration_example.go](echo_integration_example.go)
Full Echo framework integration featuring:
- Echo-specific middleware adapters
- Password reset workflow endpoints
- User management API
- Readiness and liveness probes
- Comprehensive error handling

### [fiber_integration_example.go](fiber_integration_example.go)
Complete Fiber framework integration with:
- Fiber-native middleware
- Session management endpoints
- Admin panel functionality
- Token cleanup operations
- Custom error handling

### [middleware_example.go](middleware_example.go)
Framework-agnostic middleware demonstration:
- Standard HTTP middleware
- Framework-specific adapters
- Context injection patterns
- Error response handling

## Key Features Demonstrated

### 1. Automatic Middleware Integration

**Gin Framework:**
```go
r := gin.Default()
middleware := authService.Middleware()

// Protected route group
protected := r.Group("/api/protected")
protected.Use(middleware.Gin())
{
    protected.GET("/profile", func(c *gin.Context) {
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
```

**Echo Framework:**
```go
e := echo.New()
middleware := authService.Middleware()

protected := e.Group("/api/protected")
protected.Use(middleware.Echo())
protected.GET("/profile", func(c echo.Context) error {
    user, ok := auth.GetUserFromEcho(c)
    if !ok {
        return c.JSON(500, map[string]string{"error": "User not found"})
    }
    return c.JSON(200, map[string]interface{}{"user": user})
})
```

**Fiber Framework:**
```go
app := fiber.New()
middleware := authService.Middleware()

protected := app.Group("/api/protected")
protected.Use(middleware.Fiber())
protected.Get("/profile", func(c *fiber.Ctx) error {
    user, ok := auth.GetUserFromFiber(c)
    if !ok {
        return c.Status(500).JSON(fiber.Map{"error": "User not found"})
    }
    return c.JSON(fiber.Map{"user": user})
})
```

### 2. Context Helpers

Each framework has specific context helpers for accessing authenticated user data:

```go
// Gin
user, ok := auth.GetUserFromGin(c)
claims, ok := auth.GetClaimsFromGin(c)

// Echo  
user, ok := auth.GetUserFromEcho(c)
claims, ok := auth.GetClaimsFromEcho(c)

// Fiber
user, ok := auth.GetUserFromFiber(c)
claims, ok := auth.GetClaimsFromFiber(c)

// Standard HTTP
user, ok := auth.GetUserFromContext(r.Context())
claims, ok := auth.GetClaimsFromContext(r.Context())
```

### 3. Complete API Endpoints

Each example includes a full REST API with:

**Public Endpoints:**
- `POST /register` - User registration
- `POST /login` - User authentication
- `POST /refresh` - Token refresh
- `POST /forgot-password` - Password reset request
- `POST /reset-password` - Password reset completion

**Protected Endpoints:**
- `GET /api/protected/profile` - Get user profile
- `PUT /api/protected/profile` - Update user profile
- `POST /api/protected/change-password` - Change password
- `DELETE /api/protected/logout` - Logout (revoke token)
- `DELETE /api/protected/logout-all` - Logout from all devices

**Admin Endpoints:**
- `GET /api/admin/users` - List all users
- `DELETE /api/admin/users/:id` - Delete user
- `PUT /api/admin/users/:id/status` - Update user status
- `GET /api/admin/metrics` - System metrics

**Monitoring Endpoints:**
- `GET /health` - Health check
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe
- `GET /metrics` - Metrics data

### 4. Role-Based Access Control

Examples show how to implement role-based access control:

```go
// Custom middleware for role checking
func requireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, ok := auth.GetClaimsFromGin(c)
        if !ok {
            c.JSON(500, gin.H{"error": "Claims not found"})
            c.Abort()
            return
        }

        userRole, exists := claims["role"]
        if !exists || userRole != role {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// Usage
admin := r.Group("/api/admin")
admin.Use(middleware.Gin(), requireRole("admin"))
```

### 5. Error Handling

Comprehensive error handling patterns:

```go
// Gin error handling
func errorHandler(c *gin.Context, err error) {
    if authErr, ok := err.(*auth.AuthError); ok {
        c.JSON(authErr.HTTPStatus(), gin.H{
            "error": authErr.Code,
            "message": authErr.Message,
        })
        return
    }
    
    c.JSON(500, gin.H{"error": "Internal server error"})
}

// Echo error handling
func errorHandler(err error, c echo.Context) {
    code := http.StatusInternalServerError
    message := "Internal Server Error"

    if authErr, ok := err.(*auth.AuthError); ok {
        code = authErr.HTTPStatus()
        message = authErr.Message
    }

    c.JSON(code, map[string]interface{}{
        "error": message,
        "code": code,
    })
}
```

## Running the Examples

### Prerequisites

Install framework dependencies:

```bash
# For Gin example
go get github.com/gin-gonic/gin

# For Echo example  
go get github.com/labstack/echo/v4

# For Fiber example
go get github.com/gofiber/fiber/v2
```

### Run Individual Examples

```bash
# Gin integration (runs on :8080)
go run examples/gin_integration_example.go

# Echo integration (runs on :8081)
go run examples/echo_integration_example.go

# Fiber integration (runs on :8082)
go run examples/fiber_integration_example.go

# Framework-agnostic middleware
go run examples/middleware_example.go
```

### Test the APIs

Each example includes a demo user. You can test the endpoints:

```bash
# Register a new user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Use the returned access token for protected endpoints
curl -X GET http://localhost:8080/api/protected/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Health check
curl http://localhost:8080/health
```

## API Documentation

### Authentication Flow

1. **Register**: `POST /register`
   ```json
   {
     "username": "john_doe",
     "email": "john@example.com", 
     "password": "secure_password123"
   }
   ```

2. **Login**: `POST /login`
   ```json
   {
     "username": "john_doe",
     "password": "secure_password123",
     "claims": {
       "role": "user"
     }
   }
   ```

3. **Use Token**: Include in Authorization header
   ```
   Authorization: Bearer <access_token>
   ```

4. **Refresh**: `POST /refresh`
   ```json
   {
     "refresh_token": "<refresh_token>"
   }
   ```

### Profile Management

**Get Profile**: `GET /api/protected/profile`
- Returns user profile and claims

**Update Profile**: `PUT /api/protected/profile`
```json
{
  "email": "newemail@example.com",
  "metadata": {
    "role": "admin",
    "department": "engineering"
  }
}
```

**Change Password**: `POST /api/protected/change-password`
```json
{
  "old_password": "current_password",
  "new_password": "new_secure_password"
}
```

### Password Reset

**Request Reset**: `POST /forgot-password`
```json
{
  "email": "user@example.com"
}
```

**Complete Reset**: `POST /reset-password`
```json
{
  "token": "reset_token_from_email",
  "new_password": "new_secure_password"
}
```

### Admin Operations

**List Users**: `GET /api/admin/users?limit=10&offset=0`

**Delete User**: `DELETE /api/admin/users/{user_id}`

**Update User Status**: `PUT /api/admin/users/{user_id}/status`
```json
{
  "is_active": true
}
```

## Production Considerations

### 1. Security Headers

Add security headers to your middleware:

```go
// Gin
r.Use(func(c *gin.Context) {
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("X-Frame-Options", "DENY")
    c.Header("X-XSS-Protection", "1; mode=block")
    c.Next()
})
```

### 2. Rate Limiting

Implement rate limiting for authentication endpoints:

```go
// Example with gin-limiter
import "github.com/ulule/limiter/v3/drivers/middleware/gin"

// Apply to login endpoint
r.POST("/login", limiter.New(store, rate).Handler(), loginHandler)
```

### 3. CORS Configuration

Configure CORS for frontend integration:

```go
// Gin
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://yourdomain.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

### 4. Logging

Add request logging:

```go
// Gin
r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
        param.ClientIP,
        param.TimeStamp.Format(time.RFC1123),
        param.Method,
        param.Path,
        param.Request.Proto,
        param.StatusCode,
        param.Latency,
        param.Request.UserAgent(),
        param.ErrorMessage,
    )
}))
```

### 5. Health Checks

Implement comprehensive health checks:

```go
func healthHandler(authService *auth.Auth) gin.HandlerFunc {
    return func(c *gin.Context) {
        health := map[string]interface{}{
            "status": "healthy",
            "timestamp": time.Now(),
        }

        // Check auth service health
        if err := authService.Health(); err != nil {
            health["status"] = "unhealthy"
            health["error"] = err.Error()
            c.JSON(503, health)
            return
        }

        // Add system info
        health["info"] = authService.GetSystemInfo()
        c.JSON(200, health)
    }
}
```

## Best Practices

### 1. Middleware Order

Apply middleware in the correct order:

```go
r := gin.Default()

// 1. Recovery middleware (first)
r.Use(gin.Recovery())

// 2. Logging middleware
r.Use(gin.Logger())

// 3. CORS middleware
r.Use(cors.Default())

// 4. Security headers
r.Use(securityHeaders())

// 5. Rate limiting (for specific routes)
// 6. Authentication middleware (for protected routes)
```

### 2. Error Responses

Provide consistent error responses:

```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
    Path    string `json:"path,omitempty"`
}

func writeError(c *gin.Context, err error) {
    response := ErrorResponse{
        Path: c.Request.URL.Path,
    }

    if authErr, ok := err.(*auth.AuthError); ok {
        response.Error = authErr.Code
        response.Message = authErr.Message
        response.Code = authErr.HTTPStatus()
        c.JSON(response.Code, response)
        return
    }

    response.Error = "INTERNAL_ERROR"
    response.Message = "Internal server error"
    response.Code = 500
    c.JSON(500, response)
}
```

### 3. Token Storage

Guide clients on secure token storage:

```javascript
// Frontend - Store tokens securely
// Don't use localStorage for sensitive applications
// Consider httpOnly cookies or secure storage

// Example: Store in httpOnly cookie
fetch('/login', {
    method: 'POST',
    credentials: 'include', // Include cookies
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({username, password})
});
```

### 4. Graceful Shutdown

Implement graceful shutdown:

```go
func main() {
    r := setupRouter()
    
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
}
```

## Troubleshooting

### Common Issues

**CORS Errors:**
```go
// Ensure CORS is configured properly
r.Use(cors.New(cors.Config{
    AllowOrigins: []string{"http://localhost:3000"}, // Your frontend URL
    AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
}))
```

**Token Not Found in Context:**
```go
// Ensure middleware is applied to the route
protected.Use(middleware.Gin()) // Apply middleware
protected.GET("/profile", handler) // Then add routes
```

**Database Connection Issues:**
```go
// Check database file permissions for SQLite
// Ensure PostgreSQL connection string is correct
// Use health checks to monitor database connectivity
```

### Debug Mode

Enable debug logging:
```bash
export AUTH_LOG_LEVEL=debug
export GIN_MODE=debug
go run examples/gin_integration_example.go
```

This will show detailed logs of all operations including middleware execution, token validation, and database queries.