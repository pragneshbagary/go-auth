package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
	fmt.Println("=== Enhanced Configuration Examples ===")
	fmt.Println()

	// Example 1: Create configuration with defaults
	fmt.Println("1. Creating configuration with defaults:")
	defaultConfig := auth.NewEnhancedConfig()
	defaultConfig.JWTAccessSecret = "example-access-secret"
	defaultConfig.JWTRefreshSecret = "example-refresh-secret"
	fmt.Printf("Default config created with environment: %s\n", defaultConfig.Environment)
	fmt.Printf("Default access token TTL: %s\n", defaultConfig.AccessTokenTTL)
	fmt.Printf("Default log level: %s\n\n", defaultConfig.LogLevel)

	// Example 2: Load configuration from environment variables
	fmt.Println("2. Loading configuration from environment variables:")
	
	// Set some environment variables for demonstration
	os.Setenv("AUTH_JWT_ACCESS_SECRET", "env-access-secret")
	os.Setenv("AUTH_JWT_REFRESH_SECRET", "env-refresh-secret")
	os.Setenv("AUTH_DB_TYPE", "postgres")
	os.Setenv("AUTH_DB_URL", "postgres://localhost:5432/myapp")
	os.Setenv("AUTH_JWT_ISSUER", "my-app")
	os.Setenv("AUTH_ACCESS_TOKEN_TTL", "30m")
	os.Setenv("AUTH_LOG_LEVEL", "debug")

	envConfig, err := auth.LoadConfigFromEnv()
	if err != nil {
		log.Printf("Error loading config from env: %v\n", err)
	} else {
		fmt.Printf("Environment config loaded successfully\n")
		fmt.Printf("Database type: %s\n", envConfig.DatabaseType)
		fmt.Printf("Database URL: %s\n", envConfig.DatabaseURL)
		fmt.Printf("JWT Issuer: %s\n", envConfig.JWTIssuer)
		fmt.Printf("Access Token TTL: %s\n", envConfig.AccessTokenTTL)
		fmt.Printf("Log Level: %s\n\n", envConfig.LogLevel)
	}

	// Example 3: Using configuration profiles
	fmt.Println("3. Using configuration profiles:")
	
	// Clear environment variables to see profile defaults
	os.Unsetenv("AUTH_LOG_LEVEL")
	os.Unsetenv("AUTH_ACCESS_TOKEN_TTL")
	
	profiles := []string{"development", "staging", "production"}
	for _, profile := range profiles {
		config, err := auth.LoadConfigWithProfile(profile)
		if err != nil {
			log.Printf("Error loading %s profile: %v\n", profile, err)
			continue
		}
		
		fmt.Printf("%s profile:\n", profile)
		fmt.Printf("  Environment: %s\n", config.Environment)
		fmt.Printf("  Log Level: %s\n", config.LogLevel)
		fmt.Printf("  Access Token TTL: %s\n", config.AccessTokenTTL)
		fmt.Printf("  Password Min Length: %d\n", config.PasswordMinLength)
		fmt.Println()
	}

	// Example 4: Configuration validation
	fmt.Println("4. Configuration validation:")
	
	// Create an invalid configuration
	invalidConfig := auth.NewEnhancedConfig()
	invalidConfig.JWTAccessSecret = "" // Missing required field
	invalidConfig.JWTSigningMethod = "INVALID" // Invalid signing method
	invalidConfig.AccessTokenTTL = 0 // Invalid TTL
	
	if err := invalidConfig.Validate(); err != nil {
		fmt.Printf("Validation failed as expected: %v\n\n", err)
	}

	// Example 5: Converting to legacy config format
	fmt.Println("5. Converting to legacy config format:")
	enhanced := auth.NewEnhancedConfig()
	enhanced.JWTAccessSecret = "access-secret"
	enhanced.JWTRefreshSecret = "refresh-secret"
	enhanced.JWTIssuer = "legacy-example"
	
	legacyJWT := enhanced.ToJWTConfig()
	fmt.Printf("Legacy JWT config created with issuer: %s\n", legacyJWT.Issuer)
	fmt.Printf("Access secret length: %d bytes\n", len(legacyJWT.AccessSecret))
	fmt.Printf("Refresh secret length: %d bytes\n\n", len(legacyJWT.RefreshSecret))

	// Example 6: Available profiles
	fmt.Println("6. Available configuration profiles:")
	availableProfiles := auth.GetAvailableProfiles()
	for _, profile := range availableProfiles {
		fmt.Printf("- %s (environment: %s)\n", profile.Name, profile.Environment)
	}
	fmt.Println()

	// Example 7: Print configuration (with masked secrets)
	fmt.Println("7. Printing configuration safely:")
	if envConfig != nil {
		envConfig.PrintConfig()
	}

	// Clean up environment variables
	envVars := []string{
		"AUTH_JWT_ACCESS_SECRET", "AUTH_JWT_REFRESH_SECRET", "AUTH_DB_TYPE", 
		"AUTH_DB_URL", "AUTH_JWT_ISSUER", "AUTH_ACCESS_TOKEN_TTL", "AUTH_LOG_LEVEL",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}

	fmt.Println("\n=== Examples completed ===")
}