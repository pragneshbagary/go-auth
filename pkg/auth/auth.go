package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pragneshbagary/go-auth/internal/jwtutils"
	"github.com/pragneshbagary/go-auth/internal/storage/memory"
	"github.com/pragneshbagary/go-auth/internal/storage/postgres"
	"github.com/pragneshbagary/go-auth/internal/storage/sqlite"
	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// Auth provides high-level authentication operations with improved naming and usability.
// It orchestrates user storage, password hashing, and token generation.
type Auth struct {
	storage          storage.EnhancedStorage
	jwtManager       jwtutils.TokenManager
	config           *AuthConfig
	migrationManager *MigrationManager
	logger           *Logger
	eventLogger      *AuthEventLogger
	metricsCollector *MetricsCollector
	monitor          *Monitor
}

// AuthConfig holds the configuration for the Auth service.
type AuthConfig struct {
	// Database configuration
	DatabasePath string // For SQLite
	DatabaseURL  string // For PostgreSQL
	
	// JWT configuration
	JWTSecret       string
	JWTRefreshSecret string
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	
	// Application configuration
	AppName string
	Version string
	
	// Logging configuration
	LogLevel string
}

// New creates a new Auth instance with SQLite storage using the provided database path and JWT secret.
// This is the simplest constructor for getting started quickly.
func New(databasePath string, jwtSecret string) (*Auth, error) {
	return NewSQLite(databasePath, jwtSecret)
}

// NewSQLite creates a new Auth instance with SQLite storage.
func NewSQLite(databasePath string, jwtSecret string) (*Auth, error) {
	storage, err := sqlite.NewSQLiteStorage(databasePath)
	if err != nil {
		return nil, WrapDatabaseError(err)
	}

	config := &AuthConfig{
		DatabasePath:    databasePath,
		JWTSecret:       jwtSecret,
		JWTRefreshSecret: jwtSecret + "_refresh", // Default refresh secret
		JWTIssuer:       "go-auth",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		AppName:         "go-auth-app",
		Version:         "2.0.0",
		LogLevel:        "info",
	}

	return newAuthWithStorage(storage, config)
}

// NewPostgres creates a new Auth instance with PostgreSQL storage.
func NewPostgres(connectionString string, jwtSecret string) (*Auth, error) {
	storage, err := postgres.NewPostgresStorage(connectionString)
	if err != nil {
		return nil, WrapDatabaseError(err)
	}

	config := &AuthConfig{
		DatabaseURL:     connectionString,
		JWTSecret:       jwtSecret,
		JWTRefreshSecret: jwtSecret + "_refresh", // Default refresh secret
		JWTIssuer:       "go-auth",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		AppName:         "go-auth-app",
		Version:         "2.0.0",
		LogLevel:        "info",
	}

	return newAuthWithStorage(storage, config)
}

// NewInMemory creates a new Auth instance with in-memory storage.
// This is useful for testing and development.
func NewInMemory(jwtSecret string) (*Auth, error) {
	storage := memory.NewInMemoryStorage()

	config := &AuthConfig{
		JWTSecret:       jwtSecret,
		JWTRefreshSecret: jwtSecret + "_refresh", // Default refresh secret
		JWTIssuer:       "go-auth",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		AppName:         "go-auth-app",
		Version:         "2.0.0",
		LogLevel:        "info",
	}

	return newAuthWithStorage(storage, config)
}

// NewWithConfig creates a new Auth instance with custom configuration.
// This provides the most flexibility for advanced use cases.
func NewWithConfig(config *AuthConfig) (*Auth, error) {
	if config == nil {
		return nil, ErrConfigError("config")
	}

	var storageImpl storage.EnhancedStorage
	var err error

	// Determine storage type based on config
	if config.DatabaseURL != "" {
		// PostgreSQL
		storageImpl, err = postgres.NewPostgresStorage(config.DatabaseURL)
		if err != nil {
			return nil, WrapDatabaseError(err)
		}
	} else if config.DatabasePath != "" {
		// SQLite
		storageImpl, err = sqlite.NewSQLiteStorage(config.DatabasePath)
		if err != nil {
			return nil, WrapDatabaseError(err)
		}
	} else {
		// Default to in-memory
		storageImpl = memory.NewInMemoryStorage()
	}

	return newAuthWithStorage(storageImpl, config)
}

// newAuthWithStorage is an internal helper to create Auth with storage and config.
func newAuthWithStorage(storageImpl storage.EnhancedStorage, config *AuthConfig) (*Auth, error) {
	// Validate required configuration
	if config.JWTSecret == "" {
		return nil, ErrConfigError("JWT secret")
	}

	// Set defaults for optional fields
	if config.JWTRefreshSecret == "" {
		config.JWTRefreshSecret = config.JWTSecret + "_refresh"
	}
	if config.JWTIssuer == "" {
		config.JWTIssuer = "go-auth"
	}
	if config.AccessTokenTTL == 0 {
		config.AccessTokenTTL = 15 * time.Minute
	}
	if config.RefreshTokenTTL == 0 {
		config.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	if config.AppName == "" {
		config.AppName = "go-auth-app"
	}
	if config.Version == "" {
		config.Version = "2.0.0"
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	// Create JWT manager
	jwtManager := jwtutils.NewJWTManager(jwtutils.JWTConfig{
		AccessSecret:    []byte(config.JWTSecret),
		RefreshSecret:   []byte(config.JWTRefreshSecret),
		Issuer:          config.JWTIssuer,
		AccessTokenTTL:  config.AccessTokenTTL,
		RefreshTokenTTL: config.RefreshTokenTTL,
		SigningMethod:   HS256, // Default to HS256
	})

	// Create migration manager
	migrationManager := NewMigrationManager(storageImpl)

	// Create logger
	logger := NewLogger(ParseLogLevel(config.LogLevel), nil)
	eventLogger := NewAuthEventLogger(logger)

	// Create metrics collector
	metricsCollector := NewMetricsCollector()

	auth := &Auth{
		storage:          storageImpl,
		jwtManager:       jwtManager,
		config:           config,
		migrationManager: migrationManager,
		logger:           logger,
		eventLogger:      eventLogger,
		metricsCollector: metricsCollector,
	}

	// Create monitor
	auth.monitor = NewMonitor(storageImpl, metricsCollector, logger, config.AppName, config.Version)

	// Automatic database initialization and migration on startup
	if err := auth.initializeDatabase(); err != nil {
		return nil, WrapError(err, ErrCodeMigrationError, "Failed to initialize database")
	}

	return auth, nil
}

// initializeDatabase performs automatic database initialization and migration.
func (a *Auth) initializeDatabase() error {
	// Check database connectivity
	if err := a.storage.Ping(); err != nil {
		return WrapError(err, ErrCodeConnectionError, "Database connectivity check failed")
	}

	// Run basic storage initialization first
	if err := a.storage.Migrate(); err != nil {
		return WrapError(err, ErrCodeMigrationError, "Database initialization failed")
	}

	// Run managed migrations
	if err := a.migrationManager.Migrate(); err != nil {
		return WrapError(err, ErrCodeMigrationError, "Database migration failed")
	}

	return nil
}

// RegisterRequest defines the data required to register a new user.
type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

// Register creates a new user, hashes their password, and saves them to storage.
// It returns the newly created user.
func (a *Auth) Register(payload RegisterRequest) (*models.User, error) {
	start := time.Now()
	var userID string
	var success bool
	var err error

	defer func() {
		// Log the registration event
		a.eventLogger.LogRegistration(userID, payload.Username, payload.Email, "", "", success, err)
		// Record metrics
		a.metricsCollector.RecordRegistrationAttempt(success)
	}()

	// Basic validation
	if payload.Username == "" {
		err = ErrValidationError("username")
		a.metricsCollector.RecordValidationError()
		return nil, err
	}
	if payload.Password == "" {
		err = ErrValidationError("password")
		a.metricsCollector.RecordValidationError()
		return nil, err
	}

	a.logger.Debug("Starting user registration", map[string]interface{}{
		"username": payload.Username,
		"email":    payload.Email,
	})

	// Check if user already exists by username
	if _, err := a.storage.GetUserByUsername(payload.Username); err == nil {
		err = ErrUserExists("username")
		a.logger.Warn("Registration failed: username already exists", map[string]interface{}{
			"username": payload.Username,
		})
		return nil, err
	}

	// Check if user already exists by email (if email is provided)
	if payload.Email != "" {
		if _, err := a.storage.GetUserByEmail(payload.Email); err == nil {
			err = ErrUserExists("email")
			a.logger.Warn("Registration failed: email already exists", map[string]interface{}{
				"email": payload.Email,
			})
			return nil, err
		}
	}

	passwordHash, hashErr := HashPassword(payload.Password)
	if hashErr != nil {
		err = WrapError(hashErr, ErrCodeInternalError, "Failed to hash password")
		a.logger.Error("Password hashing failed", map[string]interface{}{
			"username": payload.Username,
			"error":    hashErr,
		})
		return nil, err
	}

	newUser := models.User{
		ID:           uuid.New().String(),
		Username:     payload.Username,
		Email:        payload.Email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	userID = newUser.ID

	if createErr := a.storage.CreateUser(newUser); createErr != nil {
		err = WrapDatabaseError(createErr)
		a.metricsCollector.RecordDatabaseError()
		a.logger.Error("Failed to create user in storage", map[string]interface{}{
			"username": payload.Username,
			"user_id":  userID,
			"error":    createErr,
		})
		return nil, err
	}

	success = true
	a.logger.Info("User registered successfully", map[string]interface{}{
		"username": payload.Username,
		"user_id":  userID,
		"duration": time.Since(start),
	})

	return &newUser, nil
}

// LoginResult is the structure returned upon a successful login.
type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login authenticates a user and returns an access and refresh token pair.
// It accepts customClaims to be embedded in the access token for authorization purposes.
func (a *Auth) Login(username, password string, customClaims map[string]interface{}) (*LoginResult, error) {
	start := time.Now()
	var userID string
	var success bool
	var err error

	defer func() {
		duration := time.Since(start)
		// Log the login event
		a.eventLogger.LogLogin(userID, username, "", "", success, duration, err)
		// Record metrics
		a.metricsCollector.RecordLoginAttempt(success, duration)
	}()

	a.logger.Debug("Starting user login", map[string]interface{}{
		"username": username,
	})

	user, getUserErr := a.storage.GetUserByUsername(username)
	if getUserErr != nil {
		err = ErrInvalidCredentials() // Generic error for security
		a.logger.Warn("Login failed: user not found", map[string]interface{}{
			"username": username,
		})
		return nil, err
	}

	userID = user.ID

	if !user.IsActive {
		err = ErrUserInactive()
		a.logger.Warn("Login failed: user inactive", map[string]interface{}{
			"username": username,
			"user_id":  userID,
		})
		return nil, err
	}

	match, checkErr := CheckPasswordHash(password, user.PasswordHash)
	if checkErr != nil {
		err = ErrInvalidCredentials()
		a.logger.Error("Login failed: password check error", map[string]interface{}{
			"username": username,
			"user_id":  userID,
			"error":    checkErr,
		})
		return nil, err
	}
	if !match {
		err = ErrInvalidCredentials()
		a.logger.Warn("Login failed: invalid password", map[string]interface{}{
			"username": username,
			"user_id":  userID,
		})
		return nil, err
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	// Note: We could update this in storage, but for now we'll keep it simple

	// Add standard claims
	claims := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"user_id":  user.ID,
	}
	// Merge with custom claims
	for k, v := range customClaims {
		claims[k] = v
	}

	accessToken, tokenErr := a.jwtManager.GenerateAccessToken(user.ID, claims)
	if tokenErr != nil {
		err = WrapError(tokenErr, ErrCodeInternalError, "Failed to generate access token")
		a.logger.Error("Failed to generate access token", map[string]interface{}{
			"username": username,
			"user_id":  userID,
			"error":    tokenErr,
		})
		return nil, err
	}

	refreshToken, refreshErr := a.jwtManager.GenerateRefreshToken(user.ID)
	if refreshErr != nil {
		err = WrapError(refreshErr, ErrCodeInternalError, "Failed to generate refresh token")
		a.logger.Error("Failed to generate refresh token", map[string]interface{}{
			"username": username,
			"user_id":  userID,
			"error":    refreshErr,
		})
		return nil, err
	}

	success = true
	a.logger.Info("User logged in successfully", map[string]interface{}{
		"username": username,
		"user_id":  userID,
		"duration": time.Since(start),
	})

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken validates an access token string.
// It returns the claims if the token is valid, otherwise an error.
func (a *Auth) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	start := time.Now()
	claims, err := a.jwtManager.ValidateAccessToken(tokenString)
	duration := time.Since(start)

	var userID string
	if claims != nil {
		if uid, ok := claims["user_id"].(string); ok {
			userID = uid
		}
	}

	success := err == nil
	a.eventLogger.LogTokenValidation(userID, "", "", success, duration, err)
	a.metricsCollector.RecordTokenValidation(success, duration)

	if err != nil {
		a.logger.Debug("Token validation failed", map[string]interface{}{
			"user_id":  userID,
			"duration": duration,
			"error":    err,
		})
	} else {
		a.logger.Debug("Token validated successfully", map[string]interface{}{
			"user_id":  userID,
			"duration": duration,
		})
	}

	return claims, err
}

// ValidateRefreshToken validates a refresh token string.
// It returns the claims if the token is valid, otherwise an error.
func (a *Auth) ValidateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	return a.jwtManager.ValidateRefreshToken(tokenString)
}

// GetUser retrieves a user by their ID, returning a safe UserProfile.
func (a *Auth) GetUser(userID string) (*models.UserProfile, error) {
	user, err := a.storage.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user.ToUserProfile(), nil
}

// GetUserByUsername retrieves a user by their username, returning a safe UserProfile.
func (a *Auth) GetUserByUsername(username string) (*models.UserProfile, error) {
	user, err := a.storage.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return user.ToUserProfile(), nil
}

// GetUserByEmail retrieves a user by their email, returning a safe UserProfile.
func (a *Auth) GetUserByEmail(email string) (*models.UserProfile, error) {
	user, err := a.storage.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	return user.ToUserProfile(), nil
}

// Users returns a Users component for enhanced user management operations.
func (a *Auth) Users() *Users {
	return &Users{
		storage:          a.storage,
		eventLogger:      a.eventLogger,
		metricsCollector: a.metricsCollector,
	}
}

// Tokens returns a Tokens component for enhanced token management operations.
func (a *Auth) Tokens() *Tokens {
	return &Tokens{
		jwtManager:       a.jwtManager,
		storage:          a.storage,
		eventLogger:      a.eventLogger,
		metricsCollector: a.metricsCollector,
	}
}

// RefreshToken provides direct access to token refresh functionality.
// This is a convenience method that delegates to the Tokens component.
func (a *Auth) RefreshToken(refreshToken string) (*RefreshResult, error) {
	return a.Tokens().Refresh(refreshToken)
}

// Health checks the health of the Auth service and its dependencies.
func (a *Auth) Health() error {
	return a.storage.Ping()
}

// Migrations returns the migration manager for advanced migration operations.
func (a *Auth) Migrations() *MigrationManager {
	return a.migrationManager
}

// GetSchemaVersion returns the current database schema version.
func (a *Auth) GetSchemaVersion() (int, error) {
	return a.migrationManager.GetCurrentVersion()
}

// MigrateToVersion migrates the database to a specific version.
func (a *Auth) MigrateToVersion(version int) error {
	return a.migrationManager.MigrateToVersion(version)
}

// RollbackToVersion rolls back the database to a specific version.
func (a *Auth) RollbackToVersion(version int) error {
	return a.migrationManager.Rollback(version)
}

// Logger returns the logger instance for custom logging
func (a *Auth) Logger() *Logger {
	return a.logger
}

// EventLogger returns the authentication event logger
func (a *Auth) EventLogger() *AuthEventLogger {
	return a.eventLogger
}

// MetricsCollector returns the metrics collector
func (a *Auth) MetricsCollector() *MetricsCollector {
	return a.metricsCollector
}

// Monitor returns the monitoring component
func (a *Auth) Monitor() *Monitor {
	return a.monitor
}

// GetMetrics returns the current authentication metrics
func (a *Auth) GetMetrics() Metrics {
	return a.metricsCollector.GetMetrics()
}

// GetSystemHealth returns the current system health status
func (a *Auth) GetSystemHealth() SystemHealth {
	return a.monitor.CheckHealth()
}

// GetSystemInfo returns detailed system information
func (a *Auth) GetSystemInfo() SystemInfo {
	return a.monitor.GetSystemInfo()
}