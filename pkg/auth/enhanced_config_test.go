package auth

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnhancedConfig(t *testing.T) {
	config := NewEnhancedConfig()

	assert.Equal(t, "sqlite", config.DatabaseType)
	assert.Equal(t, "auth.db", config.DatabaseURL)
	assert.Equal(t, "go-auth", config.JWTIssuer)
	assert.Equal(t, "HS256", config.JWTSigningMethod)
	assert.Equal(t, 15*time.Minute, config.AccessTokenTTL)
	assert.Equal(t, 168*time.Hour, config.RefreshTokenTTL)
	assert.Equal(t, 8, config.PasswordMinLength)
	assert.True(t, config.RequireEmail)
	assert.Equal(t, "go-auth-app", config.AppName)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "info", config.LogLevel)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		envVars := []string{
			"AUTH_DB_TYPE", "AUTH_DB_URL", "AUTH_JWT_ACCESS_SECRET", "AUTH_JWT_REFRESH_SECRET",
			"AUTH_JWT_ISSUER", "AUTH_JWT_SIGNING_METHOD", "AUTH_ACCESS_TOKEN_TTL", "AUTH_REFRESH_TOKEN_TTL",
			"AUTH_PASSWORD_MIN_LENGTH", "AUTH_REQUIRE_EMAIL", "AUTH_APP_NAME", "AUTH_ENVIRONMENT", "AUTH_LOG_LEVEL",
		}
		for _, env := range envVars {
			os.Unsetenv(env)
		}
	}()

	// Set environment variables
	os.Setenv("AUTH_DB_TYPE", "postgres")
	os.Setenv("AUTH_DB_URL", "postgres://localhost/test")
	os.Setenv("AUTH_JWT_ACCESS_SECRET", "test-access-secret")
	os.Setenv("AUTH_JWT_REFRESH_SECRET", "test-refresh-secret")
	os.Setenv("AUTH_JWT_ISSUER", "test-issuer")
	os.Setenv("AUTH_JWT_SIGNING_METHOD", "HS384")
	os.Setenv("AUTH_ACCESS_TOKEN_TTL", "30m")
	os.Setenv("AUTH_REFRESH_TOKEN_TTL", "72h")
	os.Setenv("AUTH_PASSWORD_MIN_LENGTH", "10")
	os.Setenv("AUTH_REQUIRE_EMAIL", "false")
	os.Setenv("AUTH_APP_NAME", "test-app")
	os.Setenv("AUTH_ENVIRONMENT", "staging")
	os.Setenv("AUTH_LOG_LEVEL", "debug")

	config, err := LoadConfigFromEnv()
	require.NoError(t, err)

	assert.Equal(t, "postgres", config.DatabaseType)
	assert.Equal(t, "postgres://localhost/test", config.DatabaseURL)
	assert.Equal(t, "test-access-secret", config.JWTAccessSecret)
	assert.Equal(t, "test-refresh-secret", config.JWTRefreshSecret)
	assert.Equal(t, "test-issuer", config.JWTIssuer)
	assert.Equal(t, "HS384", config.JWTSigningMethod)
	assert.Equal(t, 30*time.Minute, config.AccessTokenTTL)
	assert.Equal(t, 72*time.Hour, config.RefreshTokenTTL)
	assert.Equal(t, 10, config.PasswordMinLength)
	assert.False(t, config.RequireEmail)
	assert.Equal(t, "test-app", config.AppName)
	assert.Equal(t, "staging", config.Environment)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestLoadConfigFromEnvMissingRequired(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("AUTH_JWT_ACCESS_SECRET")
		os.Unsetenv("AUTH_JWT_REFRESH_SECRET")
	}()

	// Don't set required secrets
	_, err := LoadConfigFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT access secret is required")
	assert.Contains(t, err.Error(), "JWT refresh secret is required")
}

func TestLoadConfigWithProfile(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("AUTH_JWT_ACCESS_SECRET")
		os.Unsetenv("AUTH_JWT_REFRESH_SECRET")
	}()

	// Set required secrets
	os.Setenv("AUTH_JWT_ACCESS_SECRET", "test-access-secret")
	os.Setenv("AUTH_JWT_REFRESH_SECRET", "test-refresh-secret")

	tests := []struct {
		profile     string
		environment string
		logLevel    string
		accessTTL   time.Duration
		refreshTTL  time.Duration
		minLength   int
	}{
		{
			profile:     "development",
			environment: "development",
			logLevel:    "debug",
			accessTTL:   1 * time.Hour,
			refreshTTL:  24 * time.Hour,
			minLength:   6,
		},
		{
			profile:     "staging",
			environment: "staging",
			logLevel:    "info",
			accessTTL:   30 * time.Minute,
			refreshTTL:  72 * time.Hour,
			minLength:   8,
		},
		{
			profile:     "production",
			environment: "production",
			logLevel:    "warn",
			accessTTL:   15 * time.Minute,
			refreshTTL:  168 * time.Hour,
			minLength:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			config, err := LoadConfigWithProfile(tt.profile)
			require.NoError(t, err)

			assert.Equal(t, tt.environment, config.Environment)
			assert.Equal(t, tt.logLevel, config.LogLevel)
			assert.Equal(t, tt.accessTTL, config.AccessTokenTTL)
			assert.Equal(t, tt.refreshTTL, config.RefreshTokenTTL)
			assert.Equal(t, tt.minLength, config.PasswordMinLength)
		})
	}
}

func TestLoadConfigWithInvalidProfile(t *testing.T) {
	_, err := LoadConfigWithProfile("invalid-profile")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown profile: invalid-profile")
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *EnhancedConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &EnhancedConfig{
				DatabaseType:      "sqlite",
				DatabaseURL:       "test.db",
				JWTAccessSecret:   "access-secret",
				JWTRefreshSecret:  "refresh-secret",
				JWTIssuer:         "test",
				JWTSigningMethod:  "HS256",
				AccessTokenTTL:    15 * time.Minute,
				RefreshTokenTTL:   24 * time.Hour,
				PasswordMinLength: 8,
				RequireEmail:      true,
				AppName:           "test-app",
				Environment:       "development",
				LogLevel:          "info",
			},
			expectError: false,
		},
		{
			name: "missing access secret",
			config: &EnhancedConfig{
				DatabaseType:      "sqlite",
				JWTRefreshSecret:  "refresh-secret",
				JWTSigningMethod:  "HS256",
				AccessTokenTTL:    15 * time.Minute,
				RefreshTokenTTL:   24 * time.Hour,
				PasswordMinLength: 8,
				Environment:       "development",
				LogLevel:          "info",
			},
			expectError: true,
			errorMsg:    "JWT access secret is required",
		},
		{
			name: "invalid signing method",
			config: &EnhancedConfig{
				DatabaseType:      "sqlite",
				JWTAccessSecret:   "access-secret",
				JWTRefreshSecret:  "refresh-secret",
				JWTSigningMethod:  "INVALID",
				AccessTokenTTL:    15 * time.Minute,
				RefreshTokenTTL:   24 * time.Hour,
				PasswordMinLength: 8,
				Environment:       "development",
				LogLevel:          "info",
			},
			expectError: true,
			errorMsg:    "invalid JWT signing method",
		},
		{
			name: "access token TTL greater than refresh token TTL",
			config: &EnhancedConfig{
				DatabaseType:      "sqlite",
				JWTAccessSecret:   "access-secret",
				JWTRefreshSecret:  "refresh-secret",
				JWTSigningMethod:  "HS256",
				AccessTokenTTL:    24 * time.Hour,
				RefreshTokenTTL:   15 * time.Minute,
				PasswordMinLength: 8,
				Environment:       "development",
				LogLevel:          "info",
			},
			expectError: true,
			errorMsg:    "access token TTL must be less than refresh token TTL",
		},
		{
			name: "password min length too short",
			config: &EnhancedConfig{
				DatabaseType:      "sqlite",
				JWTAccessSecret:   "access-secret",
				JWTRefreshSecret:  "refresh-secret",
				JWTSigningMethod:  "HS256",
				AccessTokenTTL:    15 * time.Minute,
				RefreshTokenTTL:   24 * time.Hour,
				PasswordMinLength: 2,
				Environment:       "development",
				LogLevel:          "info",
			},
			expectError: true,
			errorMsg:    "password minimum length must be at least 4",
		},
		{
			name: "invalid database type",
			config: &EnhancedConfig{
				DatabaseType:      "invalid",
				JWTAccessSecret:   "access-secret",
				JWTRefreshSecret:  "refresh-secret",
				JWTSigningMethod:  "HS256",
				AccessTokenTTL:    15 * time.Minute,
				RefreshTokenTTL:   24 * time.Hour,
				PasswordMinLength: 8,
				Environment:       "development",
				LogLevel:          "info",
			},
			expectError: true,
			errorMsg:    "invalid database type",
		},
		{
			name: "invalid log level",
			config: &EnhancedConfig{
				DatabaseType:      "sqlite",
				JWTAccessSecret:   "access-secret",
				JWTRefreshSecret:  "refresh-secret",
				JWTSigningMethod:  "HS256",
				AccessTokenTTL:    15 * time.Minute,
				RefreshTokenTTL:   24 * time.Hour,
				PasswordMinLength: 8,
				Environment:       "development",
				LogLevel:          "invalid",
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToJWTConfig(t *testing.T) {
	enhanced := &EnhancedConfig{
		JWTAccessSecret:   "access-secret",
		JWTRefreshSecret:  "refresh-secret",
		JWTIssuer:         "test-issuer",
		JWTSigningMethod:  "HS384",
		AccessTokenTTL:    30 * time.Minute,
		RefreshTokenTTL:   72 * time.Hour,
	}

	jwtConfig := enhanced.ToJWTConfig()

	assert.Equal(t, []byte("access-secret"), jwtConfig.AccessSecret)
	assert.Equal(t, []byte("refresh-secret"), jwtConfig.RefreshSecret)
	assert.Equal(t, "test-issuer", jwtConfig.Issuer)
	assert.Equal(t, "HS384", jwtConfig.SigningMethod)
	assert.Equal(t, 30*time.Minute, jwtConfig.AccessTokenTTL)
	assert.Equal(t, 72*time.Hour, jwtConfig.RefreshTokenTTL)
}

func TestGetAvailableProfiles(t *testing.T) {
	profiles := GetAvailableProfiles()
	assert.Len(t, profiles, 3)

	profileNames := make([]string, len(profiles))
	for i, profile := range profiles {
		profileNames[i] = profile.Name
	}

	assert.Contains(t, profileNames, "development")
	assert.Contains(t, profileNames, "staging")
	assert.Contains(t, profileNames, "production")
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "[NOT SET]"},
		{"short", "[HIDDEN]"},
		{"12345678", "[HIDDEN]"},
		{"123456789", "1234****6789"},
		{"very-long-secret-key", "very****-key"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskSecret(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadFromEnvInvalidValues(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("AUTH_ACCESS_TOKEN_TTL")
		os.Unsetenv("AUTH_PASSWORD_MIN_LENGTH")
		os.Unsetenv("AUTH_REQUIRE_EMAIL")
	}()

	config := NewEnhancedConfig()

	// Test invalid duration
	os.Setenv("AUTH_ACCESS_TOKEN_TTL", "invalid-duration")
	err := loadFromEnv(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AUTH_ACCESS_TOKEN_TTL")

	// Test invalid integer
	os.Unsetenv("AUTH_ACCESS_TOKEN_TTL")
	os.Setenv("AUTH_PASSWORD_MIN_LENGTH", "not-a-number")
	err = loadFromEnv(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AUTH_PASSWORD_MIN_LENGTH")

	// Test invalid boolean
	os.Unsetenv("AUTH_PASSWORD_MIN_LENGTH")
	os.Setenv("AUTH_REQUIRE_EMAIL", "not-a-boolean")
	err = loadFromEnv(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AUTH_REQUIRE_EMAIL")
}