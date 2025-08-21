package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Initialize auth service
	authService, err := auth.NewInMemory("fiber-example-secret")
	if err != nil {
		panic("Failed to initialize auth service: " + err.Error())
	}

	// Register a demo user with admin role
	_, err = authService.Register(auth.RegisterRequest{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "admin123",
	})
	if err != nil {
		panic("Failed to register admin user: " + err.Error())
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Get auth middleware
	authMiddleware := authService.Middleware()

	// Public routes
	app.Post("/register", registerHandler(authService))
	app.Post("/login", loginHandler(authService))
	app.Post("/refresh", refreshHandler(authService))
	app.Post("/forgot-password", forgotPasswordHandler(authService))
	app.Post("/reset-password", resetPasswordHandler(authService))

	// Protected routes group
	protected := app.Group("/api/protected")
	protected.Use(authMiddleware.Fiber())
	{
		protected.Get("/profile", getProfileHandler())
		protected.Put("/profile", updateProfileHandler(authService))
		protected.Post("/change-password", changePasswordHandler(authService))
		protected.Delete("/logout", logoutHandler(authService))
		protected.Delete("/logout-all", logoutAllHandler(authService))
		protected.Get("/sessions", getSessionsHandler(authService))
	}

	// Optional authentication routes
	app.Get("/api/optional", authMiddleware.FiberOptional(), optionalAuthHandler())
	app.Get("/api/public", publicHandler())

	// Admin routes
	admin := app.Group("/api/admin")
	admin.Use(authMiddleware.Fiber(), requireRole("admin"))
	{
		admin.Get("/users", listUsersHandler(authService))
		admin.Delete("/users/:id", deleteUserHandler(authService))
		admin.Put("/users/:id/status", updateUserStatusHandler(authService))
		admin.Get("/metrics", adminMetricsHandler(authService))
		admin.Post("/cleanup", cleanupHandler(authService))
	}

	// Health and monitoring
	app.Get("/health", healthHandler(authService))
	app.Get("/health/ready", readinessHandler(authService))
	app.Get("/health/live", livenessHandler())
	app.Get("/metrics", metricsHandler(authService))

	// Start server
	app.Listen(":8082")
}

// Error handler
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   message,
		"code":    code,
		"path":    c.Path(),
		"method":  c.Method(),
	})
}

// Public handlers
func registerHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req auth.RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		user, err := authService.Register(req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "User registered successfully",
			"user": fiber.Map{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	}
}

func loginHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string                 `json:"username"`
			Password string                 `json:"password"`
			Claims   map[string]interface{} `json:"claims,omitempty"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Add default role if not specified
		if req.Claims == nil {
			req.Claims = make(map[string]interface{})
		}
		if _, exists := req.Claims["role"]; !exists {
			if req.Username == "admin" {
				req.Claims["role"] = "admin"
			} else {
				req.Claims["role"] = "user"
			}
		}

		result, err := authService.Login(req.Username, req.Password, req.Claims)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		return c.JSON(fiber.Map{
			"message":       "Login successful",
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"user": fiber.Map{
				"id":       result.User.ID,
				"username": result.User.Username,
				"email":    result.User.Email,
			},
		})
	}
}

func refreshHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		if req.RefreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Refresh token is required",
			})
		}

		result, err := authService.RefreshToken(req.RefreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid refresh token",
			})
		}

		return c.JSON(fiber.Map{
			"message":       "Token refreshed successfully",
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		})
	}
}

func forgotPasswordHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Email string `json:"email"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		if req.Email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email is required",
			})
		}

		users := authService.Users()
		resetToken, err := users.CreateResetToken(req.Email)
		if err != nil {
			// Don't reveal if email exists or not for security
			return c.JSON(fiber.Map{
				"message": "If the email exists, a reset token has been sent",
			})
		}

		// In a real application, you would send this token via email
		return c.JSON(fiber.Map{
			"message":     "Reset token created",
			"reset_token": resetToken.Token, // Don't do this in production!
			"expires_at":  resetToken.ExpiresAt,
		})
	}
}

func resetPasswordHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		if req.Token == "" || req.NewPassword == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Token and new password are required",
			})
		}

		users := authService.Users()
		err := users.ResetPassword(req.Token, req.NewPassword)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Password reset successfully",
		})
	}
}

// Protected handlers
func getProfileHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := auth.GetUserFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "User not found in context",
			})
		}

		claims, ok := auth.GetClaimsFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Claims not found in context",
			})
		}

		return c.JSON(fiber.Map{
			"user":   user,
			"claims": claims,
		})
	}
}

func updateProfileHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := auth.GetUserFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "User not found in context",
			})
		}

		var req struct {
			Email    *string                `json:"email,omitempty"`
			Metadata map[string]interface{} `json:"metadata,omitempty"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		users := authService.Users()
		err := users.Update(user.ID, auth.UserUpdate{
			Email:    req.Email,
			Metadata: req.Metadata,
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Profile updated successfully",
		})
	}
}

func changePasswordHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := auth.GetUserFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "User not found in context",
			})
		}

		var req struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		if req.OldPassword == "" || req.NewPassword == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Old and new passwords are required",
			})
		}

		users := authService.Users()
		err := users.ChangePassword(user.ID, req.OldPassword, req.NewPassword)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Password changed successfully",
		})
	}
}

func logoutHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No token provided",
			})
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		tokens := authService.Tokens()
		err := tokens.Revoke(token)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to logout",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	}
}

func logoutAllHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := auth.GetUserFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "User not found in context",
			})
		}

		tokens := authService.Tokens()
		err := tokens.RevokeAll(user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to logout from all devices",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Logged out from all devices successfully",
		})
	}
}

func getSessionsHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := auth.GetUserFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "User not found in context",
			})
		}

		tokens := authService.Tokens()
		sessions, err := tokens.ListActiveSessions(user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list sessions",
			})
		}

		return c.JSON(fiber.Map{
			"sessions": sessions,
			"count":    len(sessions),
		})
	}
}

// Public handlers
func publicHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "This is a public endpoint",
			"data":    "Anyone can access this",
		})
	}
}

func optionalAuthHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, authenticated := auth.GetUserFromFiber(c)

		response := fiber.Map{
			"message":       "This endpoint works with or without authentication",
			"authenticated": authenticated,
		}

		if authenticated {
			response["user"] = fiber.Map{
				"username": user.Username,
				"email":    user.Email,
			}
		}

		return c.JSON(response)
	}
}

// Admin handlers
func listUsersHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := 10
		offset := 0

		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		users := authService.Users()
		userList, err := users.List(limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list users",
			})
		}

		return c.JSON(fiber.Map{
			"users": userList,
			"count": len(userList),
		})
	}
}

func deleteUserHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("id")
		if userID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User ID is required",
			})
		}

		users := authService.Users()
		err := users.Delete(userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "User deleted successfully",
		})
	}
}

func updateUserStatusHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("id")
		if userID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User ID is required",
			})
		}

		var req struct {
			IsActive bool `json:"is_active"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		users := authService.Users()
		err := users.Update(userID, auth.UserUpdate{
			Metadata: map[string]interface{}{
				"is_active": req.IsActive,
			},
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		status := "deactivated"
		if req.IsActive {
			status = "activated"
		}

		return c.JSON(fiber.Map{
			"message": "User " + status + " successfully",
		})
	}
}

func adminMetricsHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		metrics := authService.GetMetrics()
		collector := authService.MetricsCollector()

		return c.JSON(fiber.Map{
			"metrics": metrics,
			"rates": fiber.Map{
				"login_success_rate":             collector.GetLoginSuccessRate(),
				"registration_success_rate":      collector.GetRegistrationSuccessRate(),
				"token_validation_success_rate":  collector.GetTokenValidationSuccessRate(),
			},
			"system": authService.GetSystemInfo(),
		})
	}
}

func cleanupHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokens := authService.Tokens()
		err := tokens.CleanupExpired()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to cleanup expired tokens",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Expired tokens cleaned up successfully",
		})
	}
}

// Health handlers
func healthHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := authService.Health(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status": "healthy",
			"info":   authService.GetSystemInfo(),
		})
	}
}

func readinessHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := authService.Health(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not ready",
			})
		}

		return c.JSON(fiber.Map{
			"status": "ready",
		})
	}
}

func livenessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "alive",
		})
	}
}

func metricsHandler(authService *auth.Auth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		metrics := authService.GetMetrics()
		return c.JSON(metrics)
	}
}

// Middleware helpers
func requireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := auth.GetClaimsFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Claims not found",
			})
		}

		userRole, exists := claims["role"]
		if !exists || userRole != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}