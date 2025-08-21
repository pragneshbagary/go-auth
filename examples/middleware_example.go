package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Initialize auth service
	authService, err := auth.NewInMemory("your-secret-key")
	if err != nil {
		log.Fatal("Failed to initialize auth service:", err)
	}

	// Register a test user
	user, err := authService.Register(auth.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		log.Fatal("Failed to register user:", err)
	}
	fmt.Printf("Registered user: %s\n", user.Username)

	// Login to get tokens
	loginResult, err := authService.Login("testuser", "password123", map[string]interface{}{
		"role": "user",
	})
	if err != nil {
		log.Fatal("Failed to login:", err)
	}
	fmt.Printf("Access token: %s\n", loginResult.AccessToken)

	// Example 1: Standard HTTP middleware
	fmt.Println("\n=== Standard HTTP Middleware Example ===")
	runStandardHTTPExample(authService)

	// Example 2: Gin middleware
	fmt.Println("\n=== Gin Middleware Example ===")
	runGinExample(authService)

	// Example 3: Echo middleware
	fmt.Println("\n=== Echo Middleware Example ===")
	runEchoExample(authService)

	// Example 4: Fiber middleware
	fmt.Println("\n=== Fiber Middleware Example ===")
	runFiberExample(authService)
}

func runStandardHTTPExample(authService *auth.Auth) {
	mux := http.NewServeMux()

	// Protected route using direct middleware access
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}

		// Get claims from context
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Claims not found in context", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"message": "Access granted to protected resource",
			"user":    user,
			"claims":  claims,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Apply protection middleware
	mux.Handle("/protected", authService.Protect(protectedHandler))

	// Optional authentication route
	optionalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, authenticated := auth.GetUserFromContext(r.Context())
		
		response := map[string]interface{}{
			"message":       "This endpoint works with or without authentication",
			"authenticated": authenticated,
		}
		
		if authenticated {
			response["user"] = user
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.Handle("/optional", authService.Optional(optionalHandler))

	fmt.Println("Standard HTTP server would be running on :8080")
	fmt.Println("Protected endpoint: GET /protected (requires Authorization: Bearer <token>)")
	fmt.Println("Optional endpoint: GET /optional (works with or without token)")
}

func runGinExample(authService *auth.Auth) {
	r := gin.New()
	middleware := authService.Middleware()

	// Protected route group
	protected := r.Group("/api/protected")
	protected.Use(middleware.Gin())
	{
		protected.GET("/profile", func(c *gin.Context) {
			// Get user from Gin context
			user, ok := auth.GetUserFromGin(c)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
				return
			}

			// Get claims from Gin context
			claims, ok := auth.GetClaimsFromGin(c)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Claims not found in context"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Protected profile endpoint",
				"user":    user,
				"claims":  claims,
			})
		})
	}

	// Optional authentication route
	r.GET("/api/optional", middleware.GinOptional(), func(c *gin.Context) {
		user, authenticated := auth.GetUserFromGin(c)
		
		response := gin.H{
			"message":       "Optional authentication endpoint",
			"authenticated": authenticated,
		}
		
		if authenticated {
			response["user"] = user
		}

		c.JSON(http.StatusOK, response)
	})

	fmt.Println("Gin server would be running on :8081")
	fmt.Println("Protected endpoint: GET /api/protected/profile")
	fmt.Println("Optional endpoint: GET /api/optional")
}

func runEchoExample(authService *auth.Auth) {
	e := echo.New()
	middleware := authService.Middleware()

	// Protected route group
	protected := e.Group("/api/protected")
	protected.Use(middleware.Echo())
	protected.GET("/profile", func(c echo.Context) error {
		// Get user from Echo context
		user, ok := auth.GetUserFromEcho(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found in context"})
		}

		// Get claims from Echo context
		claims, ok := auth.GetClaimsFromEcho(c)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Claims not found in context"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Protected profile endpoint",
			"user":    user,
			"claims":  claims,
		})
	})

	// Optional authentication route
	e.GET("/api/optional", func(c echo.Context) error {
		user, authenticated := auth.GetUserFromEcho(c)
		
		response := map[string]interface{}{
			"message":       "Optional authentication endpoint",
			"authenticated": authenticated,
		}
		
		if authenticated {
			response["user"] = user
		}

		return c.JSON(http.StatusOK, response)
	}, middleware.EchoOptional())

	fmt.Println("Echo server would be running on :8082")
	fmt.Println("Protected endpoint: GET /api/protected/profile")
	fmt.Println("Optional endpoint: GET /api/optional")
}

func runFiberExample(authService *auth.Auth) {
	app := fiber.New()
	middleware := authService.Middleware()

	// Protected route group
	protected := app.Group("/api/protected")
	protected.Use(middleware.Fiber())
	protected.Get("/profile", func(c *fiber.Ctx) error {
		// Get user from Fiber context
		user, ok := auth.GetUserFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found in context"})
		}

		// Get claims from Fiber context
		claims, ok := auth.GetClaimsFromFiber(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Claims not found in context"})
		}

		return c.JSON(fiber.Map{
			"message": "Protected profile endpoint",
			"user":    user,
			"claims":  claims,
		})
	})

	// Optional authentication route
	app.Get("/api/optional", middleware.FiberOptional(), func(c *fiber.Ctx) error {
		user, authenticated := auth.GetUserFromFiber(c)
		
		response := fiber.Map{
			"message":       "Optional authentication endpoint",
			"authenticated": authenticated,
		}
		
		if authenticated {
			response["user"] = user
		}

		return c.JSON(response)
	})

	fmt.Println("Fiber server would be running on :8083")
	fmt.Println("Protected endpoint: GET /api/protected/profile")
	fmt.Println("Optional endpoint: GET /api/optional")
}