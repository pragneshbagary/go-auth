package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Initialize auth service
	authService, err := auth.NewInMemory("gin-example-secret")
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

	// Create Gin router
	r := gin.Default()

	// Get middleware component
	middleware := authService.Middleware()

	// Public routes
	r.POST("/register", registerHandler(authService))
	r.POST("/login", loginHandler(authService))
	r.POST("/refresh", refreshHandler(authService))

	// Protected routes group
	protected := r.Group("/api/protected")
	protected.Use(middleware.Gin())
	{
		protected.GET("/profile", getProfileHandler())
		protected.PUT("/profile", updateProfileHandler(authService))
		protected.POST("/change-password", changePasswordHandler(authService))
		protected.DELETE("/logout", logoutHandler(authService))
		protected.GET("/admin", adminOnlyHandler(), requireRole("admin"))
	}

	// Optional authentication routes
	r.GET("/api/optional", middleware.GinOptional(), optionalAuthHandler())

	// User management routes (admin only)
	admin := r.Group("/api/admin")
	admin.Use(middleware.Gin(), requireRole("admin"))
	{
		admin.GET("/users", listUsersHandler(authService))
		admin.DELETE("/users/:id", deleteUserHandler(authService))
		admin.PUT("/users/:id/activate", activateUserHandler(authService))
	}

	// Health and monitoring
	r.GET("/health", healthHandler(authService))
	r.GET("/metrics", metricsHandler(authService))

	// Start server
	r.Run(":8080")
}

// Public handlers
func registerHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		user, err := authService.Register(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	}
}

func loginHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string                 `json:"username" binding:"required"`
			Password string                 `json:"password" binding:"required"`
			Claims   map[string]interface{} `json:"claims,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		result, err := authService.Login(req.Username, req.Password, req.Claims)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Login successful",
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"user": gin.H{
				"id":       result.User.ID,
				"username": result.User.Username,
				"email":    result.User.Email,
			},
		})
	}
}

func refreshHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		result, err := authService.RefreshToken(req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Token refreshed successfully",
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		})
	}
}

// Protected handlers
func getProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.GetUserFromGin(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
			return
		}

		claims, ok := auth.GetClaimsFromGin(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Claims not found in context"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":   user,
			"claims": claims,
		})
	}
}

func updateProfileHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.GetUserFromGin(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
			return
		}

		var req struct {
			Email    *string                `json:"email,omitempty"`
			Metadata map[string]interface{} `json:"metadata,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		users := authService.Users()
		err := users.Update(user.ID, auth.UserUpdate{
			Email:    req.Email,
			Metadata: req.Metadata,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}

func changePasswordHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.GetUserFromGin(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
			return
		}

		var req struct {
			OldPassword string `json:"old_password" binding:"required"`
			NewPassword string `json:"new_password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		users := authService.Users()
		err := users.ChangePassword(user.ID, req.OldPassword, req.NewPassword)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}

func logoutHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.GetUserFromGin(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
			return
		}

		// Get token from header
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No token provided"})
			return
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		tokens := authService.Tokens()
		err := tokens.Revoke(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

func adminOnlyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to admin area",
			"data":    "This is admin-only content",
		})
	}
}

// Optional authentication handler
func optionalAuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, authenticated := auth.GetUserFromGin(c)

		response := gin.H{
			"message":       "This endpoint works with or without authentication",
			"authenticated": authenticated,
		}

		if authenticated {
			response["user"] = gin.H{
				"username": user.Username,
				"email":    user.Email,
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

// Admin handlers
func listUsersHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"users": userList,
			"count": len(userList),
		})
	}
}

func deleteUserHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		users := authService.Users()
		err := users.Delete(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

func activateUserHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		users := authService.Users()
		isActive := true
		err := users.Update(userID, auth.UserUpdate{
			Metadata: map[string]interface{}{
				"is_active": isActive,
			},
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
	}
}

// Utility handlers
func healthHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := authService.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"info":   authService.GetSystemInfo(),
		})
	}
}

func metricsHandler(authService *auth.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := authService.GetMetrics()
		collector := authService.MetricsCollector()

		c.JSON(http.StatusOK, gin.H{
			"metrics": metrics,
			"rates": gin.H{
				"login_success_rate":        collector.GetLoginSuccessRate(),
				"registration_success_rate": collector.GetRegistrationSuccessRate(),
				"token_validation_success_rate": collector.GetTokenValidationSuccessRate(),
			},
		})
	}
}

// Middleware helpers
func requireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := auth.GetClaimsFromGin(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Claims not found"})
			c.Abort()
			return
		}

		userRole, exists := claims["role"]
		if !exists || userRole != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}