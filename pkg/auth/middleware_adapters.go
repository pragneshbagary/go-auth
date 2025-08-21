package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
	"github.com/pragneshbagary/go-auth/pkg/models"
)

// Framework-specific middleware adapters

// Gin returns a Gin middleware function that requires authentication
func (m *Middleware) Gin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		tokenString, err := extractTokenFromHeader(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, HTTPErrorResponse{
				Error:   err.(*AuthError).Code,
				Message: err.(*AuthError).Message,
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate token and get user
		user, claims, err := m.validateTokenAndGetUser(tokenString)
		if err != nil {
			var authErr *AuthError
			if errors.As(err, &authErr) {
				c.JSON(http.StatusUnauthorized, HTTPErrorResponse{
					Error:   authErr.Code,
					Message: authErr.Message,
					Code:    http.StatusUnauthorized,
				})
			} else {
				c.JSON(http.StatusUnauthorized, HTTPErrorResponse{
					Error:   "INTERNAL_ERROR",
					Message: "An internal error occurred",
					Code:    http.StatusUnauthorized,
				})
			}
			c.Abort()
			return
		}

		// Store user and claims in Gin context
		c.Set("user", user)
		c.Set("claims", claims)

		c.Next()
	}
}

// GinOptional returns a Gin middleware function that optionally validates authentication
func (m *Middleware) GinOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to extract token from Authorization header
		tokenString, err := extractTokenFromHeader(c.Request)
		if err != nil {
			// No token provided or invalid format, continue without authentication
			c.Next()
			return
		}

		// Try to validate token and get user
		user, claims, err := m.validateTokenAndGetUser(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Store user and claims in Gin context
		c.Set("user", user)
		c.Set("claims", claims)

		c.Next()
	}
}

// Echo returns an Echo middleware function that requires authentication
func (m *Middleware) Echo() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a wrapper handler
			handler := m.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Update the Echo context with the modified request
				c.SetRequest(r)

				// Get user and claims from context
				if user, ok := GetUserFromContext(r.Context()); ok {
					c.Set("user", user)
				}
				if claims, ok := GetClaimsFromContext(r.Context()); ok {
					c.Set("claims", claims)
				}

				// Call the next Echo handler
				if err := next(c); err != nil {
					c.Error(err)
				}
			}))

			// Serve the request through our middleware
			handler.ServeHTTP(c.Response(), c.Request())
			return nil
		}
	}
}

// EchoOptional returns an Echo middleware function that optionally validates authentication
func (m *Middleware) EchoOptional() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a wrapper handler
			handler := m.Optional(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Update the Echo context with the modified request
				c.SetRequest(r)

				// Get user and claims from context if available
				if user, ok := GetUserFromContext(r.Context()); ok {
					c.Set("user", user)
				}
				if claims, ok := GetClaimsFromContext(r.Context()); ok {
					c.Set("claims", claims)
				}

				// Call the next Echo handler
				if err := next(c); err != nil {
					c.Error(err)
				}
			}))

			// Serve the request through our middleware
			handler.ServeHTTP(c.Response(), c.Request())
			return nil
		}
	}
}

// Fiber returns a Fiber middleware function that requires authentication
func (m *Middleware) Fiber() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(HTTPErrorResponse{
				Error:   ErrCodeMissingToken,
				Message: "Authorization header is required",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Remove "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(HTTPErrorResponse{
				Error:   ErrCodeInvalidToken,
				Message: "Authorization header must be in format 'Bearer <token>'",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Validate token and get user
		user, claims, err := m.validateTokenAndGetUser(tokenString)
		if err != nil {
			var authErr *AuthError
			if errors.As(err, &authErr) {
				return c.Status(fiber.StatusUnauthorized).JSON(HTTPErrorResponse{
					Error:   authErr.Code,
					Message: authErr.Message,
					Code:    fiber.StatusUnauthorized,
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(HTTPErrorResponse{
				Error:   "INTERNAL_ERROR",
				Message: "An internal error occurred",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Store user and claims in Fiber context
		c.Locals("user", user)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// FiberOptional returns a Fiber middleware function that optionally validates authentication
func (m *Middleware) FiberOptional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Remove "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			// Invalid format, continue without authentication
			return c.Next()
		}

		// Try to validate token and get user
		user, claims, err := m.validateTokenAndGetUser(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			return c.Next()
		}

		// Store user and claims in Fiber context
		c.Locals("user", user)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// Helper functions for framework-specific contexts

// GetUserFromGin retrieves the authenticated user from Gin context
func GetUserFromGin(c *gin.Context) (*models.UserProfile, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	userProfile, ok := user.(*models.UserProfile)
	return userProfile, ok
}

// GetClaimsFromGin retrieves the JWT claims from Gin context
func GetClaimsFromGin(c *gin.Context) (map[string]interface{}, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	claimsMap, ok := claims.(map[string]interface{})
	return claimsMap, ok
}

// GetUserFromEcho retrieves the authenticated user from Echo context
func GetUserFromEcho(c echo.Context) (*models.UserProfile, bool) {
	user := c.Get("user")
	if user == nil {
		return nil, false
	}
	userProfile, ok := user.(*models.UserProfile)
	return userProfile, ok
}

// GetClaimsFromEcho retrieves the JWT claims from Echo context
func GetClaimsFromEcho(c echo.Context) (map[string]interface{}, bool) {
	claims := c.Get("claims")
	if claims == nil {
		return nil, false
	}
	claimsMap, ok := claims.(map[string]interface{})
	return claimsMap, ok
}

// GetUserFromFiber retrieves the authenticated user from Fiber context
func GetUserFromFiber(c *fiber.Ctx) (*models.UserProfile, bool) {
	user := c.Locals("user")
	if user == nil {
		return nil, false
	}
	userProfile, ok := user.(*models.UserProfile)
	return userProfile, ok
}

// GetClaimsFromFiber retrieves the JWT claims from Fiber context
func GetClaimsFromFiber(c *fiber.Ctx) (map[string]interface{}, bool) {
	claims := c.Locals("claims")
	if claims == nil {
		return nil, false
	}
	claimsMap, ok := claims.(map[string]interface{})
	return claimsMap, ok
}