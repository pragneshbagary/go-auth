package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pragneshbagary/go-auth/pkg/models"
)

// Middleware provides HTTP middleware functionality for authentication.
// It supports both framework-agnostic HTTP middleware and framework-specific adapters.
type Middleware struct {
	auth *Auth
}

// UserContextKey is the key used to store user information in request context
type UserContextKey string

const (
	// UserKey is the context key for storing authenticated user information
	UserKey UserContextKey = "auth_user"
	// ClaimsKey is the context key for storing JWT claims
	ClaimsKey UserContextKey = "auth_claims"
)



// extractTokenFromHeader extracts the JWT token from the Authorization header
func extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingToken()
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", NewAuthErrorWithDetails(ErrCodeMalformedToken, 
			"Authorization header must be in format 'Bearer <token>'",
			"Expected format: Authorization: Bearer <jwt-token>")
	}

	return parts[1], nil
}

// validateTokenAndGetUser validates a token and retrieves the associated user
func (m *Middleware) validateTokenAndGetUser(tokenString string) (*models.UserProfile, jwt.MapClaims, error) {
	// Validate the access token
	claims, err := m.auth.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, nil, ErrInvalidToken()
	}

	// Extract user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, nil, NewAuthErrorWithDetails(ErrCodeInvalidToken, 
			"Token missing user ID", "Token must contain a valid 'sub' claim")
	}

	// Check if token is blacklisted (if storage supports it)
	if jti, exists := claims["jti"].(string); exists {
		if blacklisted, err := m.auth.storage.IsTokenBlacklisted(jti); err == nil && blacklisted {
			return nil, nil, ErrTokenRevoked()
		}
	}

	// Get user information
	user, err := m.auth.GetUser(userID)
	if err != nil {
		return nil, nil, ErrUserNotFound()
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, ErrUserInactive()
	}

	return user, claims, nil
}

// Protect is a generic HTTP middleware that requires authentication.
// It validates the JWT token and injects user information into the request context.
func (m *Middleware) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenString, err := extractTokenFromHeader(r)
		if err != nil {
			WriteJSONError(w, err)
			return
		}

		// Validate token and get user
		user, claims, err := m.validateTokenAndGetUser(tokenString)
		if err != nil {
			WriteJSONError(w, err)
			return
		}

		// Add user and claims to request context
		ctx := context.WithValue(r.Context(), UserKey, user)
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Optional is a generic HTTP middleware that optionally validates authentication.
// If a valid token is provided, it injects user information into the request context.
// If no token or an invalid token is provided, it continues without authentication.
func (m *Middleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token from Authorization header
		tokenString, err := extractTokenFromHeader(r)
		if err != nil {
			// No token provided or invalid format, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate token and get user
		user, claims, err := m.validateTokenAndGetUser(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Add user and claims to request context
		ctx := context.WithValue(r.Context(), UserKey, user)
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(ctx context.Context) (*models.UserProfile, bool) {
	user, ok := ctx.Value(UserKey).(*models.UserProfile)
	return user, ok
}

// GetClaimsFromContext retrieves the JWT claims from the request context
func GetClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(jwt.MapClaims)
	return claims, ok
}

// Middleware returns a Middleware instance for the Auth service
func (a *Auth) Middleware() *Middleware {
	return &Middleware{
		auth: a,
	}
}

// Protect provides direct access to the Protect middleware from the Auth instance
func (a *Auth) Protect(next http.Handler) http.Handler {
	return a.Middleware().Protect(next)
}

// Optional provides direct access to the Optional middleware from the Auth instance
func (a *Auth) Optional(next http.Handler) http.Handler {
	return a.Middleware().Optional(next)
}