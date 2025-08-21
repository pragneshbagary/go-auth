package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Initialize auth service
	authService, err := auth.NewInMemory("echo-example-secret")
	if err != nil {
		panic("Failed to initialize auth service: " + err.Error())
	}

	// Register a demo user
	_, err = authService.Register(auth.RegisterRequest{
		Username: "demo",
		Email:    "demo@example.com",
		Password: "demo123",
	})
	if err != nil {
		panic("Failed to register demo user: " + err.Error())
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Get auth middleware
	authMiddleware := authService.Middleware()

	// Public routes
	e.POST("/register", registerHandler(authService))
	e.POST("/login", loginHandler(authService))
	e.POST("/refresh", refreshHandler(authService))
	e.POST("/forgot-password", forgotPasswordHandler(authService))
	e.POST("/reset-password", resetPasswordHandler(authService))

	// Protected routes group
	protected := e.Group("/api/protected")
	protected.Use(authMiddleware.Echo())
	{
		protected.GET("/profile", getProfileHandler())
		protected.PUT("/profile", updateProfileHandler(authService))
		protected.POST("/change-password", changePasswordHandler(authService))
		protected.DELETE("/logout", logoutHandler(authService))
		protected.DELETE("/logout-all", logoutAllHandler(authService))
	}

	// Optional authentication routes
	e.GET("/api/optional", optionalAuthHandler(), authMiddleware.EchoOptional())

	// Admin routes
	admin := e.Group("/api/admin")
	admin.Use(authMiddleware.Echo(), requireRole("admin"))
	{
		admin.GET("/users", listUsersHandler(authService))
		admin.DELETE("/users/:id", deleteUserHandler(authService))
		admin.PUT("/users/:id/status", updateUserStatusHandler(authService))
		admin.GET("/metrics", adminMetricsHandler(authService))
	}

	// Health and monitoring
	e.GET("/health", healthHandler(authService))
	e.GET("/health/ready", readinessHandler(authService))
	e.GET("/health/live", livenessHandler(authService))

	// Start server
	e.Logger.Fatal(e.Start(":8081"))
}

// Public handlers
func registerHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req auth.RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		user, err := authService.Register(req)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "User registered successfully",
			"user": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	}
}

func loginHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Username string                 `json:"username"`
			Password string                 `json:"password"`
			Claims   map[string]interface{} `json:"claims,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		if req.Username == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and password are required"})
		}

		result, err := authService.Login(req.Username, req.Password, req.Claims)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":       "Login successful",
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"user": map[string]interface{}{
				"id":       result.User.ID,
				"username": result.User.Username,
				"email":    result.User.Email,
			},
		})
	}
}

func refreshHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		if req.RefreshToken == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token is required"})
		}

		result, err := authService.RefreshToken(req.RefreshToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":       "Token refreshed successfully",
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		})
	}
}

func forgotPasswordHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Email string `json:"email"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		if req.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
		}

		users := authService.Users()
		resetToken, err := users.CreateResetToken(req.Email)
		if err != nil {
			// Don't reveal if email exists or not for security
			return c.JSON(http.StatusOK, map[string]string{
				"message": "If the email exists, a reset token has been sent",
			})
		}

		// In a real application, you would send this token via email
		// For demo purposes, we'll return it in the response
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":     "Reset token created",
			"reset_token": resetToken.Token, // Don't do this in production!
			"expires_at":  resetToken.ExpiresAt,
		})
	}
}

func resetPasswordHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		if req.Token == "" || req.NewPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Token and new password are required"})
		}

		users := authService.Users()
		err := users.ResetPassword(req.Token, req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Password reset successfully"})
	}
}

// Protected handlers
func getProfileHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := auth.GetUserFromEcho(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found in context"})
		}

		claims, ok := auth.GetClaimsFromEcho(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Claims not found in context"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"user":   user,
			"claims": claims,
		})
	}
}

func updateProfileHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := auth.GetUserFromEcho(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found in context"})
		}

		var req struct {
			Email    *string                `json:"email,omitempty"`
			Metadata map[string]interface{} `json:"metadata,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		users := authService.Users()
		err := users.Update(user.ID, auth.UserUpdate{
			Email:    req.Email,
			Metadata: req.Metadata,
		})
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Profile updated successfully"})
	}
}

func changePasswordHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := auth.GetUserFromEcho(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found in context"})
		}

		var req struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		if req.OldPassword == "" || req.NewPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Old and new passwords are required"})
		}

		users := authService.Users()
		err := users.ChangePassword(user.ID, req.OldPassword, req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Password changed successfully"})
	}
}

func logoutHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get token from header
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "No token provided"})
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		tokens := authService.Tokens()
		err := tokens.Revoke(token)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to logout"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
	}
}

func logoutAllHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := auth.GetUserFromEcho(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found in context"})
		}

		tokens := authService.Tokens()
		err := tokens.RevokeAll(user.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to logout from all devices"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Logged out from all devices successfully"})
	}
}

// Optional authentication handler
func optionalAuthHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		user, authenticated := auth.GetUserFromEcho(c)

		response := map[string]interface{}{
			"message":       "This endpoint works with or without authentication",
			"authenticated": authenticated,
		}

		if authenticated {
			response["user"] = map[string]interface{}{
				"username": user.Username,
				"email":    user.Email,
			}
		}

		return c.JSON(http.StatusOK, response)
	}
}

// Admin handlers
func listUsersHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit := 10
		offset := 0

		if l := c.QueryParam("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := c.QueryParam("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		users := authService.Users()
		userList, err := users.List(limit, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list users"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"users": userList,
			"count": len(userList),
		})
	}
}

func deleteUserHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("id")
		if userID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
		}

		users := authService.Users()
		err := users.Delete(userID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
	}
}

func updateUserStatusHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("id")
		if userID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
		}

		var req struct {
			IsActive bool `json:"is_active"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
		}

		users := authService.Users()
		err := users.Update(userID, auth.UserUpdate{
			Metadata: map[string]interface{}{
				"is_active": req.IsActive,
			},
		})
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		status := "deactivated"
		if req.IsActive {
			status = "activated"
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "User " + status + " successfully"})
	}
}

func adminMetricsHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		metrics := authService.GetMetrics()
		collector := authService.MetricsCollector()

		return c.JSON(http.StatusOK, map[string]interface{}{
			"metrics": metrics,
			"rates": map[string]interface{}{
				"login_success_rate":             collector.GetLoginSuccessRate(),
				"registration_success_rate":      collector.GetRegistrationSuccessRate(),
				"token_validation_success_rate":  collector.GetTokenValidationSuccessRate(),
			},
			"system": authService.GetSystemInfo(),
		})
	}
}

// Health handlers
func healthHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := authService.Health(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "healthy",
			"info":   authService.GetSystemInfo(),
		})
	}
}

func readinessHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := authService.Health(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "not ready",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status": "ready",
		})
	}
}

func livenessHandler(authService *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "alive",
		})
	}
}

// Middleware helpers
func requireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := auth.GetClaimsFromEcho(c)
			if !ok {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Claims not found"})
			}

			userRole, exists := claims["role"]
			if !exists || userRole != role {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
			}

			return next(c)
		}
	}
}