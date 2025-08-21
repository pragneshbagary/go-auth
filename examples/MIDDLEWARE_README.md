# HTTP Middleware Guide

The go-auth package provides comprehensive HTTP middleware support for protecting your API endpoints. It includes both framework-agnostic middleware and specific adapters for popular Go web frameworks.

## Features

- **Framework-agnostic**: Standard HTTP middleware that works with any HTTP router
- **Framework-specific adapters**: Native support for Gin, Echo, and Fiber
- **Automatic user injection**: Authenticated user information is automatically added to request context
- **Optional authentication**: Support for endpoints that work with or without authentication
- **Structured error responses**: Consistent JSON error responses for authentication failures
- **Token blacklisting**: Support for revoked token detection

## Quick Start

### Standard HTTP Middleware

```go
package main

import (
    "net/http"
    "github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
    // Initialize auth service
    authService, err := auth.NewInMemory("your-secret-key")
    if err != nil {
        panic(err)
    }

    mux := http.NewServeMux()

    // Protected endpoint
    protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get authenticated user from context
        user, ok := auth.GetUserFromContext(r.Context())
        if !ok {
            http.Error(w, "User not found", http.StatusInternalServerError)
            return
        }
        
        // User is authenticated, proceed with business logic
        fmt.Fprintf(w, "Hello, %s!", user.Username)
    })

    // Apply protection middleware
    mux.Handle("/protected", authService.Protect(protectedHandler))

    // Optional authentication endpoint
    optionalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, authenticated := auth.GetUserFromContext(r.Context())
        if authenticated {
            fmt.Fprintf(w, "Hello, %s!", user.Username)
        } else {
            fmt.Fprint(w, "Hello, anonymous user!")
        }
    })

    mux.Handle("/optional", authService.Optional(optionalHandler))

    http.ListenAndServe(":8080", mux)
}
```

### Gin Framework

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
    authService, _ := auth.NewInMemory("your-secret-key")
    middleware := authService.Middleware()

    r := gin.Default()

    // Protected route group
    protected := r.Group("/api/protected")
    protected.Use(middleware.Gin())
    {
        protected.GET("/profile", func(c *gin.Context) {
            // Get user from Gin context
            user, ok := auth.GetUserFromGin(c)
            if !ok {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
                return
            }

            c.JSON(http.StatusOK, gin.H{
                "message": "Protected endpoint",
                "user":    user,
            })
        })
    }

    // Optional authentication
    r.GET("/api/optional", middleware.GinOptional(), func(c *gin.Context) {
        user, authenticated := auth.GetUserFromGin(c)
        response := gin.H{"authenticated": authenticated}
        if authenticated {
            response["user"] = user
        }
        c.JSON(http.StatusOK, response)
    })

    r.Run(":8080")
}
```

### Echo Framework

```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
    authService, _ := auth.NewInMemory("your-secret-key")
    middleware := authService.Middleware()

    e := echo.New()

    // Protected route group
    protected := e.Group("/api/protected")
    protected.Use(middleware.Echo())
    protected.GET("/profile", func(c echo.Context) error {
        // Get user from Echo context
        user, ok := auth.GetUserFromEcho(c)
        if !ok {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found"})
        }

        return c.JSON(http.StatusOK, map[string]interface{}{
            "message": "Protected endpoint",
            "user":    user,
        })
    })

    // Optional authentication
    e.GET("/api/optional", func(c echo.Context) error {
        user, authenticated := auth.GetUserFromEcho(c)
        response := map[string]interface{}{"authenticated": authenticated}
        if authenticated {
            response["user"] = user
        }
        return c.JSON(http.StatusOK, response)
    }, middleware.EchoOptional())

    e.Start(":8080")
}
```

### Fiber Framework

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
    authService, _ := auth.NewInMemory("your-secret-key")
    middleware := authService.Middleware()

    app := fiber.New()

    // Protected route group
    protected := app.Group("/api/protected")
    protected.Use(middleware.Fiber())
    protected.Get("/profile", func(c *fiber.Ctx) error {
        // Get user from Fiber context
        user, ok := auth.GetUserFromFiber(c)
        if !ok {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found"})
        }

        return c.JSON(fiber.Map{
            "message": "Protected endpoint",
            "user":    user,
        })
    })

    // Optional authentication
    app.Get("/api/optional", middleware.FiberOptional(), func(c *fiber.Ctx) error {
        user, authenticated := auth.GetUserFromFiber(c)
        response := fiber.Map{"authenticated": authenticated}
        if authenticated {
            response["user"] = user
        }
        return c.JSON(response)
    })

    app.Listen(":8080")
}
```

## Authentication Flow

1. **Token Extraction**: The middleware extracts the JWT token from the `Authorization` header in the format `Bearer <token>`
2. **Token Validation**: The token is validated for signature, expiration, and format
3. **Blacklist Check**: If supported by the storage backend, the token is checked against the blacklist
4. **User Retrieval**: The user information is retrieved from storage using the token's subject claim
5. **Context Injection**: User and claims information are injected into the request context
6. **Request Processing**: The request continues to the next handler with authenticated context

## Error Responses

Authentication failures return structured JSON error responses:

```json
{
  "error": "INVALID_TOKEN",
  "message": "Invalid or expired token",
  "code": 401
}
```

Common error codes:
- `MISSING_TOKEN`: No Authorization header provided
- `INVALID_TOKEN`: Token format is invalid or token is malformed
- `TOKEN_EXPIRED`: Token has expired
- `TOKEN_REVOKED`: Token has been blacklisted
- `USER_NOT_FOUND`: User associated with token doesn't exist
- `USER_INACTIVE`: User account is inactive

## Context Helpers

### Standard HTTP Context

```go
// Get authenticated user
user, ok := auth.GetUserFromContext(r.Context())

// Get JWT claims
claims, ok := auth.GetClaimsFromContext(r.Context())
```

### Framework-Specific Helpers

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
```

## Best Practices

1. **Use HTTPS**: Always use HTTPS in production to protect tokens in transit
2. **Token Storage**: Store tokens securely on the client side (avoid localStorage for sensitive apps)
3. **Error Handling**: Handle authentication errors gracefully in your frontend
4. **Token Refresh**: Implement token refresh logic for better user experience
5. **Logging**: Log authentication events for security monitoring
6. **Rate Limiting**: Implement rate limiting on authentication endpoints

## Testing

The middleware includes comprehensive test coverage. Run tests with:

```bash
go test ./pkg/auth -v -run TestMiddleware
```

For a complete working example, see `examples/middleware_example.go`.