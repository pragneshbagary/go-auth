# Go-Auth: A Secure and Modern Authentication Package for Go

## Description

Go-Auth is a production-ready Go package designed to simplify user authentication. It provides a high-level API for secure user registration and login, built on modern cryptographic standards like Argon2id and JWT.

This package is designed to be both easy to use for beginners and flexible enough for experienced developers. It handles the complexities of password hashing, token generation, and token refreshment, allowing you to focus on your application's core logic.

---

## Features

- **Secure Password Hashing**: Uses **Argon2id**, the modern, recommended standard for password hashing.
- **JSON Web Tokens (JWT)**: Implements a robust JWT system with short-lived access tokens and long-lived refresh tokens.
- **Clean, High-Level API**: Offers simple `Register` and `Login` functions that abstract away complexity.
- **Database Agnostic**: Uses a `Storage` interface, allowing you to plug in any database backend.
- **Extensible and Configurable**: Easily add custom claims to JWTs and configure token lifespans.
- **Thoroughly Tested**: Includes a comprehensive test suite to ensure reliability and security.

---

## Getting Started

### Installation

```bash
go get github.com/pragneshbagary/go-Auth
```

### Usage

Using Go-Auth involves three main steps: configuring the services, registering a user, and logging in.

**1. Initialization**

```go
// Configure and create the required services.
storage := memory.NewInMemoryStorage() // Use your own DB implementation here
jwtManager := jwtutils.NewJWTManager(jwtConfig) // See Configuration section below
authService := auth.NewAuthService(storage, jwtManager)
```

**2. Register a New User**

```go
registerPayload := auth.RegisterPayload{
	Username: "testuser",
	Email:    "test@example.com",
	Password: "StrongPassword123!",
}
user, err := authService.Register(registerPayload)
```

**3. Log In**

```go
customClaims := map[string]interface{}{"role": "admin"}
loginResponse, err := authService.Login("testuser", "StrongPassword123!", customClaims)
```

---
## Example

A complete, runnable example demonstrating the full registration and login flow is available in the `main.go` file in the root of the repository.

---

## Configuration

To use the JWT manager, you must provide a `JWTConfig` struct. This allows you to configure the secrets and token lifespans.

```go
jwtConfig := jwtutils.JWTConfig{
    // A secret key for signing access tokens. Keep this private.
    AccessSecret:    []byte("your-super-secret-access-key"),

    // A separate secret key for signing refresh tokens.
    RefreshSecret:   []byte("your-super-secret-refresh-key"),

    // The issuer name for your application (e.g., "my-awesome-app").
    Issuer:          "my-awesome-app",

    // The lifespan of an access token (e.g., 15 minutes).
    AccessTokenTTL:  15 * time.Minute,

    // The lifespan of a refresh token (e.g., 7 days).
    RefreshTokenTTL: 7 * 24 * time.Hour,

    // The signing method to use (e.g., HS256, RS256).
    SigningMethod:   jwtutils.HS256,
}
```
⚠️ Note: The credentials and secrets are for demonstration purposes only. Never use hardcoded or weak secrets in production.

---

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you would like to contribute code, please fork the repository and create a pull request.

---

## License

This project is licensed under the [MIT License](LICENSE).
