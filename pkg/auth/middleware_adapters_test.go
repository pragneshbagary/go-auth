package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
	"github.com/pragneshbagary/go-auth/pkg/models"
)

func TestMiddleware_Gin(t *testing.T) {
	// Create an in-memory auth instance for testing
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Register a test user and get token
	_, err = auth.Register(RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginResult, err := auth.Login("testuser", "password123", nil)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	middleware := auth.Middleware()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + loginResult.AccessToken,
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Gin router
			r := gin.New()
			r.GET("/test", middleware.Gin(), func(c *gin.Context) {
				user, ok := GetUserFromGin(c)
				if tt.expectUser {
					if !ok {
						t.Error("Expected user in Gin context, but not found")
						return
					}
					if user.Username != "testuser" {
						t.Errorf("Expected username 'testuser', got '%s'", user.Username)
						return
					}
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			r.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestMiddleware_Echo(t *testing.T) {
	// Create an in-memory auth instance for testing
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Register a test user and get token
	_, err = auth.Register(RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginResult, err := auth.Login("testuser", "password123", nil)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	middleware := auth.Middleware()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + loginResult.AccessToken,
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo instance
			e := echo.New()
			e.GET("/test", func(c echo.Context) error {
				user, ok := GetUserFromEcho(c)
				if tt.expectUser {
					if !ok {
						t.Error("Expected user in Echo context, but not found")
						return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found"})
					}
					if user.Username != "testuser" {
						t.Errorf("Expected username 'testuser', got '%s'", user.Username)
						return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Wrong username"})
					}
				}
				return c.JSON(http.StatusOK, map[string]string{"message": "success"})
			}, middleware.Echo())

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			e.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestMiddleware_Fiber(t *testing.T) {
	// Create an in-memory auth instance for testing
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Register a test user and get token
	_, err = auth.Register(RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginResult, err := auth.Login("testuser", "password123", nil)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	middleware := auth.Middleware()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + loginResult.AccessToken,
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Fiber app
			app := fiber.New()
			app.Get("/test", middleware.Fiber(), func(c *fiber.Ctx) error {
				user, ok := GetUserFromFiber(c)
				if tt.expectUser {
					if !ok {
						t.Error("Expected user in Fiber context, but not found")
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found"})
					}
					if user.Username != "testuser" {
						t.Errorf("Expected username 'testuser', got '%s'", user.Username)
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Wrong username"})
					}
				}
				return c.JSON(fiber.Map{"message": "success"})
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Execute request
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestGetUserFromFrameworks(t *testing.T) {
	// Test Gin helper functions
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	
	// Test without user
	_, ok := GetUserFromGin(c)
	if ok {
		t.Error("Expected no user in empty Gin context")
	}
	
	// Test with user
	testUser := &models.UserProfile{Username: "testuser"}
	c.Set("user", testUser)
	user, ok := GetUserFromGin(c)
	if !ok {
		t.Error("Expected user in Gin context")
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	// Test Echo helper functions
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	echoCtx := e.NewContext(req, rec)
	
	// Test without user
	_, ok = GetUserFromEcho(echoCtx)
	if ok {
		t.Error("Expected no user in empty Echo context")
	}
	
	// Test with user
	echoCtx.Set("user", testUser)
	user, ok = GetUserFromEcho(echoCtx)
	if !ok {
		t.Error("Expected user in Echo context")
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	// Note: Fiber context testing is more complex due to its internal structure
	// The Fiber middleware functionality is tested in the integration test above
}