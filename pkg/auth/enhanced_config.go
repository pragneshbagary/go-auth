package auth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// EnhancedConfig provides comprehensive configuration with environment variable support
type EnhancedConfig struct {
	// Database configuration
	DatabaseType string `env:"AUTH_DB_TYPE" default:"sqlite"`
	DatabaseURL  string `env:"AUTH_DB_URL" default:"auth.db"`

	// JWT configuration
	JWTAccessSecret  string        `env:"AUTH_JWT_ACCESS_SECRET" required:"true"`
	JWTRefreshSecret string        `env:"AUTH_JWT_REFRESH_SECRET" required:"true"`
	JWTIssuer        string        `env:"AUTH_JWT_ISSUER" default:"go-auth"`
	JWTSigningMethod string        `env:"AUTH_JWT_SIGNING_METHOD" default:"HS256"`
	AccessTokenTTL   time.Duration `env:"AUTH_ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL  time.Duration `env:"AUTH_REFRESH_TOKEN_TTL" default:"168h"`

	// Security configuration
	PasswordMinLength int  `env:"AUTH_PASSWORD_MIN_LENGTH" default:"8"`
	RequireEmail      bool `env:"AUTH_REQUIRE_EMAIL" default:"true"`

	// Application configuration
	AppName     string `env:"AUTH_APP_NAME" default:"go-auth-app"`
	Environment string `env:"AUTH_ENVIRONMENT" default:"development"`

	// Logging configuration
	LogLevel string `env:"AUTH_LOG_LEVEL" default:"info"`

	// Storage instance (set programmatically)
	Storage storage.Storage `env:"-"`
}

// ConfigProfile represents different environment configurations
type ConfigProfile struct {
	Name        string
	Environment string
	Overrides   map[string]interface{}
}

// Predefined configuration profiles
var (
	DevelopmentProfile = ConfigProfile{
		Name:        "development",
		Environment: "development",
		Overrides: map[string]interface{}{
			"AUTH_LOG_LEVEL":           "debug",
			"AUTH_ACCESS_TOKEN_TTL":    "1h",
			"AUTH_REFRESH_TOKEN_TTL":   "24h",
			"AUTH_PASSWORD_MIN_LENGTH": 6,
		},
	}

	StagingProfile = ConfigProfile{
		Name:        "staging",
		Environment: "staging",
		Overrides: map[string]interface{}{
			"AUTH_LOG_LEVEL":           "info",
			"AUTH_ACCESS_TOKEN_TTL":    "30m",
			"AUTH_REFRESH_TOKEN_TTL":   "72h",
			"AUTH_PASSWORD_MIN_LENGTH": 8,
		},
	}

	ProductionProfile = ConfigProfile{
		Name:        "production",
		Environment: "production",
		Overrides: map[string]interface{}{
			"AUTH_LOG_LEVEL":           "warn",
			"AUTH_ACCESS_TOKEN_TTL":    "15m",
			"AUTH_REFRESH_TOKEN_TTL":   "168h",
			"AUTH_PASSWORD_MIN_LENGTH": 10,
		},
	}
)

// LoadConfigFromEnv loads configuration from environment variables with defaults
func LoadConfigFromEnv() (*EnhancedConfig, error) {
	// Start with default configuration
	config := NewEnhancedConfig()

	// Apply profile-based configuration if specified
	if profile := os.Getenv("AUTH_PROFILE"); profile != "" {
		if err := applyProfile(config, profile); err != nil {
			return nil, fmt.Errorf("failed to apply profile %s: %w", profile, err)
		}
	}

	// Load configuration from environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadConfigWithProfile loads configuration with a specific profile
func LoadConfigWithProfile(profileName string) (*EnhancedConfig, error) {
	// Start with default configuration
	config := NewEnhancedConfig()

	// Apply the specified profile
	if err := applyProfile(config, profileName); err != nil {
		return nil, fmt.Errorf("failed to apply profile %s: %w", profileName, err)
	}

	// Load any environment variable overrides
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// NewEnhancedConfig creates a new configuration with defaults
func NewEnhancedConfig() *EnhancedConfig {
	return &EnhancedConfig{
		DatabaseType:      "sqlite",
		DatabaseURL:       "auth.db",
		JWTIssuer:         "go-auth",
		JWTSigningMethod:  "HS256",
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   168 * time.Hour, // 7 days
		PasswordMinLength: 8,
		RequireEmail:      true,
		AppName:           "go-auth-app",
		Environment:       "development",
		LogLevel:          "info",
	}
}

// Validate validates the configuration and returns an error if invalid
func (c *EnhancedConfig) Validate() error {
	var errors []string

	// Validate required fields
	if c.JWTAccessSecret == "" {
		errors = append(errors, "JWT access secret is required")
	}
	if c.JWTRefreshSecret == "" {
		errors = append(errors, "JWT refresh secret is required")
	}

	// Validate JWT signing method
	validSigningMethods := []string{"HS256", "HS384", "HS512", "RS256"}
	if !contains(validSigningMethods, c.JWTSigningMethod) {
		errors = append(errors, fmt.Sprintf("invalid JWT signing method: %s, must be one of %v", c.JWTSigningMethod, validSigningMethods))
	}

	// Validate token TTLs
	if c.AccessTokenTTL <= 0 {
		errors = append(errors, "access token TTL must be positive")
	}
	if c.RefreshTokenTTL <= 0 {
		errors = append(errors, "refresh token TTL must be positive")
	}
	if c.AccessTokenTTL >= c.RefreshTokenTTL {
		errors = append(errors, "access token TTL must be less than refresh token TTL")
	}

	// Validate password requirements
	if c.PasswordMinLength < 4 {
		errors = append(errors, "password minimum length must be at least 4")
	}
	if c.PasswordMinLength > 128 {
		errors = append(errors, "password minimum length must be at most 128")
	}

	// Validate database type
	validDatabaseTypes := []string{"sqlite", "postgres", "memory"}
	if !contains(validDatabaseTypes, c.DatabaseType) {
		errors = append(errors, fmt.Sprintf("invalid database type: %s, must be one of %v", c.DatabaseType, validDatabaseTypes))
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, strings.ToLower(c.LogLevel)) {
		errors = append(errors, fmt.Sprintf("invalid log level: %s, must be one of %v", c.LogLevel, validLogLevels))
	}

	// Validate environment
	validEnvironments := []string{"development", "staging", "production"}
	if !contains(validEnvironments, c.Environment) {
		errors = append(errors, fmt.Sprintf("invalid environment: %s, must be one of %v", c.Environment, validEnvironments))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ToJWTConfig converts EnhancedConfig to the legacy JWTConfig format
func (c *EnhancedConfig) ToJWTConfig() JWTConfig {
	return JWTConfig{
		AccessSecret:    []byte(c.JWTAccessSecret),
		RefreshSecret:   []byte(c.JWTRefreshSecret),
		Issuer:          c.JWTIssuer,
		AccessTokenTTL:  c.AccessTokenTTL,
		RefreshTokenTTL: c.RefreshTokenTTL,
		SigningMethod:   c.JWTSigningMethod,
	}
}

// ToConfig converts EnhancedConfig to the legacy Config format
func (c *EnhancedConfig) ToConfig() Config {
	return Config{
		Storage: c.Storage,
		JWT:     c.ToJWTConfig(),
	}
}

// loadFromEnv loads configuration values from environment variables
func loadFromEnv(config *EnhancedConfig) error {
	// Database configuration
	if val := os.Getenv("AUTH_DB_TYPE"); val != "" {
		config.DatabaseType = val
	}
	if val := os.Getenv("AUTH_DB_URL"); val != "" {
		config.DatabaseURL = val
	}

	// JWT configuration
	if val := os.Getenv("AUTH_JWT_ACCESS_SECRET"); val != "" {
		config.JWTAccessSecret = val
	}
	if val := os.Getenv("AUTH_JWT_REFRESH_SECRET"); val != "" {
		config.JWTRefreshSecret = val
	}
	if val := os.Getenv("AUTH_JWT_ISSUER"); val != "" {
		config.JWTIssuer = val
	}
	if val := os.Getenv("AUTH_JWT_SIGNING_METHOD"); val != "" {
		config.JWTSigningMethod = val
	}

	// Parse duration values
	if val := os.Getenv("AUTH_ACCESS_TOKEN_TTL"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid AUTH_ACCESS_TOKEN_TTL: %w", err)
		}
		config.AccessTokenTTL = duration
	}
	if val := os.Getenv("AUTH_REFRESH_TOKEN_TTL"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid AUTH_REFRESH_TOKEN_TTL: %w", err)
		}
		config.RefreshTokenTTL = duration
	}

	// Security configuration
	if val := os.Getenv("AUTH_PASSWORD_MIN_LENGTH"); val != "" {
		length, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid AUTH_PASSWORD_MIN_LENGTH: %w", err)
		}
		config.PasswordMinLength = length
	}
	if val := os.Getenv("AUTH_REQUIRE_EMAIL"); val != "" {
		require, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid AUTH_REQUIRE_EMAIL: %w", err)
		}
		config.RequireEmail = require
	}

	// Application configuration
	if val := os.Getenv("AUTH_APP_NAME"); val != "" {
		config.AppName = val
	}
	if val := os.Getenv("AUTH_ENVIRONMENT"); val != "" {
		config.Environment = val
	}

	// Logging configuration
	if val := os.Getenv("AUTH_LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}

	return nil
}

// applyProfile applies a configuration profile
func applyProfile(config *EnhancedConfig, profileName string) error {
	var profile ConfigProfile

	switch strings.ToLower(profileName) {
	case "development", "dev":
		profile = DevelopmentProfile
	case "staging", "stage":
		profile = StagingProfile
	case "production", "prod":
		profile = ProductionProfile
	default:
		return fmt.Errorf("unknown profile: %s", profileName)
	}

	// Set environment
	config.Environment = profile.Environment

	// Apply profile overrides directly to config if environment variable is not set
	for key, value := range profile.Overrides {
		// Only apply if environment variable is not already set
		if os.Getenv(key) != "" {
			continue
		}

		// Apply the value directly to the config
		switch key {
		case "AUTH_LOG_LEVEL":
			if v, ok := value.(string); ok {
				config.LogLevel = v
			}
		case "AUTH_ACCESS_TOKEN_TTL":
			if v, ok := value.(string); ok {
				if duration, err := time.ParseDuration(v); err == nil {
					config.AccessTokenTTL = duration
				}
			}
		case "AUTH_REFRESH_TOKEN_TTL":
			if v, ok := value.(string); ok {
				if duration, err := time.ParseDuration(v); err == nil {
					config.RefreshTokenTTL = duration
				}
			}
		case "AUTH_PASSWORD_MIN_LENGTH":
			if v, ok := value.(int); ok {
				config.PasswordMinLength = v
			}
		}
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetAvailableProfiles returns a list of available configuration profiles
func GetAvailableProfiles() []ConfigProfile {
	return []ConfigProfile{
		DevelopmentProfile,
		StagingProfile,
		ProductionProfile,
	}
}

// PrintConfig prints the configuration in a readable format (excluding secrets)
func (c *EnhancedConfig) PrintConfig() {
	fmt.Println("=== Authentication Configuration ===")
	fmt.Printf("Environment: %s\n", c.Environment)
	fmt.Printf("App Name: %s\n", c.AppName)
	fmt.Printf("Database Type: %s\n", c.DatabaseType)
	fmt.Printf("Database URL: %s\n", c.DatabaseURL)
	fmt.Printf("JWT Issuer: %s\n", c.JWTIssuer)
	fmt.Printf("JWT Signing Method: %s\n", c.JWTSigningMethod)
	fmt.Printf("Access Token TTL: %s\n", c.AccessTokenTTL)
	fmt.Printf("Refresh Token TTL: %s\n", c.RefreshTokenTTL)
	fmt.Printf("Password Min Length: %d\n", c.PasswordMinLength)
	fmt.Printf("Require Email: %t\n", c.RequireEmail)
	fmt.Printf("Log Level: %s\n", c.LogLevel)
	fmt.Printf("JWT Access Secret: %s\n", maskSecret(c.JWTAccessSecret))
	fmt.Printf("JWT Refresh Secret: %s\n", maskSecret(c.JWTRefreshSecret))
	fmt.Println("=====================================")
}

// maskSecret masks a secret string for safe printing
func maskSecret(secret string) string {
	if secret == "" {
		return "[NOT SET]"
	}
	if len(secret) <= 8 {
		return "[HIDDEN]"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}