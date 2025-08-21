
package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
	_ "github.com/mattn/go-sqlite3" // Import the sqlite3 driver
)

// SQLiteStorage is a SQLite implementation of the storage.EnhancedStorage interface.
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance and initializes the database schema.
func NewSQLiteStorage(dataSourceName string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return storage, nil
}

// init creates the required tables if they don't exist.
func (s *SQLiteStorage) init() error {
	// Create migrations table first
	migrationsQuery := `
    CREATE TABLE IF NOT EXISTS migrations (
        version INTEGER PRIMARY KEY,
        description TEXT NOT NULL,
        applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := s.db.Exec(migrationsQuery); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Create users table with enhanced fields
	usersQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        username TEXT UNIQUE NOT NULL,
        email TEXT NOT NULL,
        password_hash TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        last_login_at DATETIME,
        is_active BOOLEAN NOT NULL DEFAULT 1,
        metadata TEXT
    );`
	if _, err := s.db.Exec(usersQuery); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create blacklisted_tokens table
	tokensQuery := `
    CREATE TABLE IF NOT EXISTS blacklisted_tokens (
        token_id TEXT PRIMARY KEY,
        expires_at DATETIME NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := s.db.Exec(tokensQuery); err != nil {
		return fmt.Errorf("failed to create blacklisted_tokens table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at);",
	}

	for _, indexQuery := range indexes {
		if _, err := s.db.Exec(indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// CreateUser saves a new user to the database.
func (s *SQLiteStorage) CreateUser(user models.User) error {
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
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, user.ID, user.Username, user.Email, user.PasswordHash, 
		user.CreatedAt, user.UpdatedAt, user.LastLoginAt, user.IsActive, metadataJSON)
	return err
}

// GetUserByUsername retrieves a user by their username.
func (s *SQLiteStorage) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users WHERE username = ?`
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
func (s *SQLiteStorage) UpdateUser(userID string, updates storage.UserUpdates) error {
	setParts := []string{"updated_at = ?"}
	args := []interface{}{time.Now()}

	if updates.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *updates.Email)
	}
	if updates.Username != nil {
		setParts = append(setParts, "username = ?")
		args = append(args, *updates.Username)
	}
	if updates.Metadata != nil {
		metadataJSON, err := json.Marshal(updates.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setParts = append(setParts, "metadata = ?")
		args = append(args, metadataJSON)
	}

	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(setParts, ", "))
	
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
func (s *SQLiteStorage) DeleteUser(userID string) error {
	result, err := s.db.Exec("DELETE FROM users WHERE id = ?", userID)
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
func (s *SQLiteStorage) GetUserByID(userID string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users WHERE id = ?`
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
func (s *SQLiteStorage) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users WHERE email = ?`
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
func (s *SQLiteStorage) ListUsers(limit, offset int) ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, is_active, metadata 
              FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`
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
func (s *SQLiteStorage) UpdatePassword(userID string, passwordHash string) error {
	result, err := s.db.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", 
		passwordHash, time.Now(), userID)
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
func (s *SQLiteStorage) BlacklistToken(tokenID string, expiresAt time.Time) error {
	query := "INSERT INTO blacklisted_tokens (token_id, expires_at) VALUES (?, ?)"
	_, err := s.db.Exec(query, tokenID, expiresAt)
	return err
}

// IsTokenBlacklisted checks if a token is blacklisted.
func (s *SQLiteStorage) IsTokenBlacklisted(tokenID string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM blacklisted_tokens WHERE token_id = ? AND expires_at > ?"
	err := s.db.QueryRow(query, tokenID, time.Now()).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CleanupExpiredTokens removes expired tokens from the blacklist.
func (s *SQLiteStorage) CleanupExpiredTokens() error {
	_, err := s.db.Exec("DELETE FROM blacklisted_tokens WHERE expires_at <= ?", time.Now())
	return err
}

// Ping checks the database connection.
func (s *SQLiteStorage) Ping() error {
	return s.db.Ping()
}

// Migrate runs database migrations.
func (s *SQLiteStorage) Migrate() error {
	// For now, just ensure the current schema is up to date
	// In a real implementation, this would run incremental migrations
	return s.init()
}

// GetSchemaVersion returns the current schema version.
func (s *SQLiteStorage) GetSchemaVersion() (int, error) {
	var version int
	query := "SELECT COALESCE(MAX(version), 0) FROM migrations"
	err := s.db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// RecordMigration records a migration as applied in the database.
func (s *SQLiteStorage) RecordMigration(version int, description string) error {
	query := "INSERT INTO migrations (version, description, applied_at) VALUES (?, ?, ?)"
	_, err := s.db.Exec(query, version, description, time.Now())
	return err
}

// RemoveMigrationRecord removes a migration record from the database.
func (s *SQLiteStorage) RemoveMigrationRecord(version int) error {
	query := "DELETE FROM migrations WHERE version = ?"
	_, err := s.db.Exec(query, version)
	return err
}

// GetAppliedMigrations returns all applied migrations from the database.
func (s *SQLiteStorage) GetAppliedMigrations() ([]models.Migration, error) {
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
