# SimpleAuth - Ultra-Simple Authentication

SimpleAuth is a wrapper around the main Auth component that provides ultra-simple usage with minimal configuration requirements. It's designed for developers who want to get started with authentication in just a few lines of code.

## Quick Start

### 1. Basic Usage (SQLite with defaults)

```go
package main

import (
    "log"
    "github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
    // Create SimpleAuth with default SQLite database
    simpleAuth, err := auth.Quick("your-jwt-secret-key")
    if err != nil {
        log.Fatal(err)
    }

    // Register a user
    user, err := simpleAuth.Register("username", "user@example.com", "password123")
    if err != nil {
        log.Fatal(err)
    }

    // Login and get tokens
    tokens, err := simpleAuth.Login("username", "password123")
    if err != nil {
        log.Fatal(err)
    }

    // Validate token
    claims, err := simpleAuth.ValidateToken(tokens.AccessToken)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use claims...
}
```

### 2. Environment-Based Configuration

Set environment variables:
```bash
export AUTH_JWT_ACCESS_SECRET="your-access-secret"
export AUTH_JWT_REFRESH_SECRET="your-refresh-secret"
export AUTH_DB_TYPE="sqlite"  # or "postgres", "memory"
export AUTH_DB_URL="auth.db"
export AUTH_ACCESS_TOKEN_TTL="15m"
export AUTH_REFRESH_TOKEN_TTL="168h"
```

Then use:
```go
simpleAuth, err := auth.QuickFromEnv()
if err != nil {
    log.Fatal(err)
}
```

## Constructor Functions

### Quick Functions

- **`Quick(jwtSecret)`** - SQLite with default path (`auth.db`)
- **`QuickSQLite(dbPath, jwtSecret)`** - SQLite with custom path
- **`QuickPostgres(dbURL, jwtSecret)`** - PostgreSQL database
- **`QuickInMemory(jwtSecret)`** - In-memory storage (testing/dev)
- **`QuickFromEnv()`** - Load all config from environment variables

## Core Methods

### User Management
```go
// Register a new user
user, err := simpleAuth.Register("username", "email@example.com", "password")

// Get user information (safe, no sensitive data)
profile, err := simpleAuth.GetUser(userID)
profile, err := simpleAuth.GetUserByUsername("username")
profile, err := simpleAuth.GetUserByEmail("email@example.com")
```

### Authentication
```go
// Login with standard claims
tokens, err := simpleAuth.Login("username", "password")

// Login with custom claims
customClaims := map[string]interface{}{
    "role": "admin",
    "permissions": []string{"read", "write"},
}
tokens, err := simpleAuth.LoginWithClaims("username", "password", customClaims)
```

### Token Management
```go
// Validate access token
claims, err := simpleAuth.ValidateToken(accessToken)

// Refresh tokens
newTokens, err := simpleAuth.RefreshToken(refreshToken)
```

### Health Check
```go
// Check if service is healthy
err := simpleAuth.Health()
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `AUTH_JWT_ACCESS_SECRET` | JWT access token secret | - | Yes |
| `AUTH_JWT_REFRESH_SECRET` | JWT refresh token secret | - | Yes |
| `AUTH_DB_TYPE` | Database type (`sqlite`, `postgres`, `memory`) | `sqlite` | No |
| `AUTH_DB_URL` | Database URL/path | `auth.db` | No |
| `AUTH_JWT_ISSUER` | JWT issuer | `go-auth` | No |
| `AUTH_ACCESS_TOKEN_TTL` | Access token TTL | `15m` | No |
| `AUTH_REFRESH_TOKEN_TTL` | Refresh token TTL | `168h` | No |
| `AUTH_APP_NAME` | Application name | `go-auth-app` | No |

## Advanced Usage

If you need more advanced features, you can access the underlying Auth instance:

```go
simpleAuth, _ := auth.Quick("secret")

// Get the underlying Auth instance for advanced features
advancedAuth := simpleAuth.GetAuth()

// Now you can use advanced features like custom middleware, etc.
```

## Examples

- **[simple_auth_example.go](simple_auth_example.go)** - Complete basic usage example
- **[simple_auth_env_example.go](simple_auth_env_example.go)** - Environment configuration example

## Sensible Defaults

SimpleAuth provides sensible defaults for all configuration options:

- **Database**: SQLite with `auth.db` file
- **JWT Issuer**: `go-auth`
- **Access Token TTL**: 15 minutes
- **Refresh Token TTL**: 7 days (168 hours)
- **App Name**: `go-auth-app`
- **Automatic**: Database initialization and migration

## Error Handling

All methods return clear error messages. Common errors include:

- Empty JWT secret
- Invalid credentials
- Token validation failures
- Database connectivity issues
- User already exists
- User not found

## Security Features

- Automatic password hashing with Argon2id
- JWT token validation with proper expiration
- Secure token refresh mechanism
- Database migration and initialization
- Input validation and sanitization