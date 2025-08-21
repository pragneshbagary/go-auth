
package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
	_ "github.com/lib/pq" // Import the postgres driver
)

// PostgresStorage is a PostgreSQL implementation of the storage.EnhancedStorage interface.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance and initializes the database schema.
func NewPostgresStorage(dataSourceName string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}
	if err := storage.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return storage, nil
}

// init creates the required tables if they don't exist.
func (s *PostgresStorage) init() error {
	// Create migrations table first
	migrationsQuery := `
    CREATE TABLE IF NOT EXISTS migrations (
        version INTEGER PRIMARY KEY,
        description TEXT NOT NULL,
        applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );`
	if _, err := s.db.Exec(migrationsQuery); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Create users table with enhanced fields
	usersQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        email VARCHAR(255) NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
        last_login_at TIMESTAMP WITH TIME ZONE,
        is_active BOOLEAN NOT NULL DEFAULT TRUE,
        metadata JSONB
    );`
	if _, err := s.db.Exec(usersQuery); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create blacklisted_tokens table
	tokensQuery := `
    CREATE TABLE IF NOT EXISTS blacklisted_tokens (
        token_id TEXT PRIMARY KEY,
        expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );`
	if _, err := s.db.Exec(tokensQuery); err != nil {
		return fmt.Errorf("failed to create blacklisted_tokens table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_users_metadata ON users USING GIN(metadata);",
	}

	for _, indexQuery := range indexes {
		if _, err := s.db.Exec(indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// CreateUser saves a new user to the database.
func (s *PostgresStorage) CreateUser(user models.User) error {
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	var metadataJSON []byte
	if user.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(user.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `INSERT INTO users (id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := s.db.Exec(query, user.ID, user.Username, user.Email, user.PasswordHash, 
		user.CreatedAt, user.UpdatedAt, user.LastLoginAt, user.IsActive, metadataJSON)
	return err
}

// GetUserByUsername retrieves a user by their username.
func (s *PostgresStorage) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users WHERE username = $1`
	err := s.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return user, nil
}

// UpdateUser updates user information.
func (s *PostgresStorage) UpdateUser(userID string, updates storage.UserUpdates) error {
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 1

	if updates.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *updates.Email)
		argIndex++
	}
	if updates.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *updates.Username)
		argIndex++
	}
	if updates.Metadata != nil {
		metadataJSON, err := json.Marshal(updates.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser removes a user from the database.
func (s *PostgresStorage) DeleteUser(userID string) error {
	result, err := s.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetUserByID retrieves a user by their ID.
func (s *PostgresStorage) GetUserByID(userID string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users WHERE id = $1`
	err := s.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email.
func (s *PostgresStorage) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return user, nil
}

// ListUsers retrieves a paginated list of users.
func (s *PostgresStorage) ListUsers(limit, offset int) ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var metadataJSON []byte
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.IsActive, &metadataJSON)
		if err != nil {
			return nil, err
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		users = append(users, user)
	}

	return users, rows.Err()
}

// UpdatePassword updates a user's password hash.
func (s *PostgresStorage) UpdatePassword(userID string, passwordHash string) error {
	result, err := s.db.Exec("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", 
		passwordHash, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// BlacklistToken adds a token to the blacklist.
func (s *PostgresStorage) BlacklistToken(tokenID string, expiresAt time.Time) error {
	query := "INSERT INTO blacklisted_tokens (token_id, expires_at) VALUES ($1, $2)"
	_, err := s.db.Exec(query, tokenID, expiresAt)
	return err
}

// IsTokenBlacklisted checks if a token is blacklisted.
func (s *PostgresStorage) IsTokenBlacklisted(tokenID string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM blacklisted_tokens WHERE token_id = $1 AND expires_at > NOW()"
	err := s.db.QueryRow(query, tokenID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CleanupExpiredTokens removes expired tokens from the blacklist.
func (s *PostgresStorage) CleanupExpiredTokens() error {
	_, err := s.db.Exec("DELETE FROM blacklisted_tokens WHERE expires_at <= NOW()")
	return err
}

// Ping checks the database connection.
func (s *PostgresStorage) Ping() error {
	return s.db.Ping()
}

// Migrate runs database migrations.
func (s *PostgresStorage) Migrate() error {
	// For now, just ensure the current schema is up to date
	// In a real implementation, this would run incremental migrations
	return s.init()
}

// GetSchemaVersion returns the current schema version.
func (s *PostgresStorage) GetSchemaVersion() (int, error) {
	var version int
	query := "SELECT COALESCE(MAX(version), 0) FROM migrations"
	err := s.db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// RecordMigration records a migration as applied in the database.
func (s *PostgresStorage) RecordMigration(version int, description string) error {
	query := "INSERT INTO migrations (version, description, applied_at) VALUES ($1, $2, $3)"
	_, err := s.db.Exec(query, version, description, time.Now())
	return err
}

// RemoveMigrationRecord removes a migration record from the database.
func (s *PostgresStorage) RemoveMigrationRecord(version int) error {
	query := "DELETE FROM migrations WHERE version = $1"
	_, err := s.db.Exec(query, version)
	return err
}

// GetAppliedMigrations returns all applied migrations from the database.
func (s *PostgresStorage) GetAppliedMigrations() ([]models.Migration, error) {
	query := "SELECT version, description, applied_at FROM migrations ORDER BY version"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []models.Migration
	for rows.Next() {
		var migration models.Migration
		err := rows.Scan(&migration.Version, &migration.Description, &migration.AppliedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, rows.Err()
}
