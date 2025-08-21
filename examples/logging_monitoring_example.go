package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	// Create auth service with logging enabled
	authService, err := auth.NewSQLite("example_auth.db", "your-secret-key")
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Demonstrate logging and monitoring features
	demonstrateLoggingAndMonitoring(authService)

	// Set up HTTP server with monitoring endpoints
	setupMonitoringServer(authService)
}

func demonstrateLoggingAndMonitoring(authService *auth.Auth) {
	fmt.Println("=== Logging and Monitoring Demo ===")

	// Get logger for custom logging
	logger := authService.Logger()
	logger.Info("Starting authentication operations demo")

	// Register a user (this will be logged automatically)
	user, err := authService.Register(auth.RegisterRequest{
		Username: "demouser",
		Email:    "demo@example.com",
		Password: "securepassword123",
	})
	if err != nil {
		logger.Error("Registration failed", map[string]interface{}{
			"error": err,
		})
	} else {
		logger.Info("User registered successfully", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
		})
	}

	// Login (this will be logged automatically)
	loginResult, err := authService.Login("demouser", "securepassword123", nil)
	if err != nil {
		logger.Error("Login failed", map[string]interface{}{
			"error": err,
		})
	} else {
		logger.Info("Login successful")
	}

	// Validate token (this will be logged automatically)
	if loginResult != nil {
		_, err = authService.ValidateAccessToken(loginResult.AccessToken)
		if err != nil {
			logger.Error("Token validation failed", map[string]interface{}{
				"error": err,
			})
		}
	}

	// Refresh token (this will be logged automatically)
	if loginResult != nil {
		_, err = authService.RefreshToken(loginResult.RefreshToken)
		if err != nil {
			logger.Error("Token refresh failed", map[string]interface{}{
				"error": err,
			})
		}
	}

	// Display current metrics
	displayMetrics(authService)

	// Display system health
	displaySystemHealth(authService)

	// Display system information
	displaySystemInfo(authService)
}

func displayMetrics(authService *auth.Auth) {
	fmt.Println("\n=== Current Metrics ===")
	metrics := authService.GetMetrics()

	fmt.Printf("Registration Attempts: %d (Success: %d, Failures: %d)\n",
		metrics.RegistrationAttempts, metrics.RegistrationSuccess, metrics.RegistrationFailures)
	fmt.Printf("Login Attempts: %d (Success: %d, Failures: %d)\n",
		metrics.LoginAttempts, metrics.LoginSuccess, metrics.LoginFailures)
	fmt.Printf("Token Operations: Generated: %d, Refreshes: %d, Validations: %d, Revocations: %d\n",
		metrics.TokensGenerated, metrics.TokenRefreshes, metrics.TokenValidations, metrics.TokenRevocations)
	fmt.Printf("Users: Created: %d, Updated: %d, Deleted: %d\n",
		metrics.UsersCreated, metrics.UsersUpdated, metrics.UsersDeleted)
	fmt.Printf("Errors: Database: %d, Validation: %d, Authentication: %d\n",
		metrics.DatabaseErrors, metrics.ValidationErrors, metrics.AuthenticationErrors)

	// Display success rates
	collector := authService.MetricsCollector()
	fmt.Printf("Success Rates: Login: %.1f%%, Registration: %.1f%%, Token Validation: %.1f%%\n",
		collector.GetLoginSuccessRate(),
		collector.GetRegistrationSuccessRate(),
		collector.GetTokenValidationSuccessRate())

	// Display performance metrics
	fmt.Printf("Average Durations: Login: %v, Token Refresh: %v, Token Validation: %v\n",
		metrics.AverageLoginDuration,
		metrics.AverageRefreshDuration,
		metrics.AverageValidationDuration)

	fmt.Printf("System Uptime: %v\n", collector.GetUptime())
}

func displaySystemHealth(authService *auth.Auth) {
	fmt.Println("\n=== System Health ===")
	health := authService.GetSystemHealth()

	fmt.Printf("Overall Status: %s\n", health.Status)
	fmt.Printf("Uptime: %s\n", health.Uptime)
	fmt.Printf("Version: %s\n", health.Version)

	fmt.Println("Component Health:")
	for _, component := range health.Components {
		fmt.Printf("  - %s: %s (%s)\n", component.Name, component.Status, component.Duration)
		if component.Message != "" {
			fmt.Printf("    Message: %s\n", component.Message)
		}
	}
}

func displaySystemInfo(authService *auth.Auth) {
	fmt.Println("\n=== System Information ===")
	info := authService.GetSystemInfo()

	fmt.Printf("Application: %s v%s\n", info.Name, info.Version)
	fmt.Printf("Go Version: %s\n", info.GoVersion)
	fmt.Printf("CPUs: %d, Goroutines: %d\n", info.NumCPU, info.NumGoroutine)
	fmt.Printf("Memory: Alloc: %d KB, Sys: %d KB, GC Runs: %d\n",
		info.MemoryAlloc/1024, info.MemorySys/1024, info.MemoryNumGC)
	fmt.Printf("Database: Type: %s, Status: %s\n", info.DatabaseType, info.DatabaseStatus)
	fmt.Printf("Uptime: %s\n", info.Uptime)
}

func setupMonitoringServer(authService *auth.Auth) {
	fmt.Println("\n=== Setting up Monitoring Server ===")

	// Create HTTP server with monitoring endpoints
	mux := http.NewServeMux()

	// Register monitoring endpoints
	monitor := authService.Monitor()
	monitor.RegisterHTTPHandlers(mux)

	// Add a custom endpoint that shows formatted metrics
	mux.HandleFunc("/metrics/formatted", func(w http.ResponseWriter, r *http.Request) {
		metrics := authService.GetMetrics()
		collector := authService.MetricsCollector()

		response := map[string]interface{}{
			"metrics": metrics,
			"success_rates": map[string]float64{
				"login":             collector.GetLoginSuccessRate(),
				"registration":      collector.GetRegistrationSuccessRate(),
				"token_validation":  collector.GetTokenValidationSuccessRate(),
			},
			"uptime": collector.GetUptime().String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Add a custom endpoint for authentication operations
	mux.HandleFunc("/demo/auth", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			// Simulate authentication operations for demo
			performDemoOperations(authService, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Monitoring endpoints available:")
	fmt.Println("  GET  /health          - Health check")
	fmt.Println("  GET  /health/ready    - Readiness check")
	fmt.Println("  GET  /health/live     - Liveness check")
	fmt.Println("  GET  /metrics         - Raw metrics")
	fmt.Println("  GET  /metrics/formatted - Formatted metrics with success rates")
	fmt.Println("  GET  /info            - System information")
	fmt.Println("  POST /demo/auth       - Demo authentication operations")

	fmt.Println("\nStarting server on :8080...")
	fmt.Println("Try these commands:")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println("  curl http://localhost:8080/metrics/formatted")
	fmt.Println("  curl -X POST http://localhost:8080/demo/auth")

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func performDemoOperations(authService *auth.Auth, w http.ResponseWriter, r *http.Request) {
	logger := authService.Logger()
	results := []string{}

	// Simulate various authentication operations
	operations := []func() string{
		func() string {
			username := fmt.Sprintf("user_%d", time.Now().Unix())
			_, err := authService.Register(auth.RegisterRequest{
				Username: username,
				Email:    fmt.Sprintf("%s@example.com", username),
				Password: "password123",
			})
			if err != nil {
				logger.Error("Demo registration failed", map[string]interface{}{"error": err})
				return fmt.Sprintf("Registration failed: %v", err)
			}
			return fmt.Sprintf("User %s registered successfully", username)
		},
		func() string {
			// Try to login with a non-existent user (will fail)
			_, err := authService.Login("nonexistent", "wrongpassword", nil)
			if err != nil {
				return "Login failed as expected (invalid credentials)"
			}
			return "Login unexpectedly succeeded"
		},
		func() string {
			// Validate an invalid token (will fail)
			_, err := authService.ValidateAccessToken("invalid.token.here")
			if err != nil {
				return "Token validation failed as expected (invalid token)"
			}
			return "Token validation unexpectedly succeeded"
		},
	}

	for i, operation := range operations {
		result := operation()
		results = append(results, fmt.Sprintf("Operation %d: %s", i+1, result))
		time.Sleep(100 * time.Millisecond) // Small delay between operations
	}

	// Return results and current metrics
	response := map[string]interface{}{
		"operations": results,
		"metrics":    authService.GetMetrics(),
		"health":     authService.GetSystemHealth(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}