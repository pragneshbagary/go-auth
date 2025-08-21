package auth

import (
	"fmt"
	"sort"

	"github.com/pragneshbagary/go-auth/pkg/models"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// MigrationStep represents a single migration step with up and down functions
type MigrationStep struct {
	Version     int
	Description string
	Up          func(storage.EnhancedStorage) error
	Down        func(storage.EnhancedStorage) error
}

// MigrationManager handles database schema migrations
type MigrationManager struct {
	storage storage.EnhancedStorage
	steps   []MigrationStep
}

// NewMigrationManager creates a new migration manager with the given storage
func NewMigrationManager(storage storage.EnhancedStorage) *MigrationManager {
	mm := &MigrationManager{
		storage: storage,
		steps:   make([]MigrationStep, 0),
	}
	
	// Register built-in migrations
	mm.registerBuiltinMigrations()
	
	return mm
}

// RegisterMigration adds a new migration step to the manager
func (mm *MigrationManager) RegisterMigration(step MigrationStep) {
	mm.steps = append(mm.steps, step)
	// Keep migrations sorted by version
	sort.Slice(mm.steps, func(i, j int) bool {
		return mm.steps[i].Version < mm.steps[j].Version
	})
}

// Migrate runs all pending migrations up to the latest version
func (mm *MigrationManager) Migrate() error {
	currentVersion, err := mm.storage.GetSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Find migrations that need to be applied
	pendingMigrations := make([]MigrationStep, 0)
	for _, step := range mm.steps {
		if step.Version > currentVersion {
			pendingMigrations = append(pendingMigrations, step)
		}
	}

	if len(pendingMigrations) == 0 {
		return nil // No migrations to run
	}

	// Apply each pending migration
	for _, migration := range pendingMigrations {
		if err := mm.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// MigrateToVersion runs migrations up to a specific version
func (mm *MigrationManager) MigrateToVersion(targetVersion int) error {
	currentVersion, err := mm.storage.GetSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	if targetVersion == currentVersion {
		return nil // Already at target version
	}

	if targetVersion > currentVersion {
		// Migrate up
		for _, step := range mm.steps {
			if step.Version > currentVersion && step.Version <= targetVersion {
				if err := mm.applyMigration(step); err != nil {
					return fmt.Errorf("failed to apply migration %d: %w", step.Version, err)
				}
			}
		}
	} else {
		// Migrate down
		// Get migrations in reverse order
		reverseMigrations := make([]MigrationStep, 0)
		for i := len(mm.steps) - 1; i >= 0; i-- {
			step := mm.steps[i]
			if step.Version > targetVersion && step.Version <= currentVersion {
				reverseMigrations = append(reverseMigrations, step)
			}
		}

		for _, migration := range reverseMigrations {
			if err := mm.rollbackMigration(migration); err != nil {
				return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// Rollback rolls back migrations to a specific version
func (mm *MigrationManager) Rollback(targetVersion int) error {
	return mm.MigrateToVersion(targetVersion)
}

// GetCurrentVersion returns the current schema version
func (mm *MigrationManager) GetCurrentVersion() (int, error) {
	return mm.storage.GetSchemaVersion()
}

// GetPendingMigrations returns a list of migrations that haven't been applied yet
func (mm *MigrationManager) GetPendingMigrations() ([]MigrationStep, error) {
	currentVersion, err := mm.storage.GetSchemaVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current schema version: %w", err)
	}

	pending := make([]MigrationStep, 0)
	for _, step := range mm.steps {
		if step.Version > currentVersion {
			pending = append(pending, step)
		}
	}

	return pending, nil
}

// GetAppliedMigrations returns a list of migrations that have been applied
func (mm *MigrationManager) GetAppliedMigrations() ([]models.Migration, error) {
	return mm.storage.GetAppliedMigrations()
}

// applyMigration applies a single migration and records it
func (mm *MigrationManager) applyMigration(migration MigrationStep) error {
	// Apply the migration
	if err := migration.Up(mm.storage); err != nil {
		return err
	}

	// Record the migration in the database
	return mm.recordMigration(migration)
}

// rollbackMigration rolls back a single migration
func (mm *MigrationManager) rollbackMigration(migration MigrationStep) error {
	if migration.Down == nil {
		return fmt.Errorf("migration %d does not support rollback", migration.Version)
	}

	// Apply the rollback
	if err := migration.Down(mm.storage); err != nil {
		return err
	}

	// Remove the migration record from the database
	return mm.removeMigrationRecord(migration)
}

// recordMigration records a migration as applied in the database
func (mm *MigrationManager) recordMigration(migration MigrationStep) error {
	return mm.storage.RecordMigration(migration.Version, migration.Description)
}

// removeMigrationRecord removes a migration record from the database
func (mm *MigrationManager) removeMigrationRecord(migration MigrationStep) error {
	return mm.storage.RemoveMigrationRecord(migration.Version)
}

// registerBuiltinMigrations registers the built-in migrations for the auth system
func (mm *MigrationManager) registerBuiltinMigrations() {
	// Migration 1: Initial schema (this represents the current state)
	mm.RegisterMigration(MigrationStep{
		Version:     1,
		Description: "Initial schema with users, blacklisted_tokens, and migrations tables",
		Up: func(storage storage.EnhancedStorage) error {
			// This migration is essentially a no-op since the init() methods
			// in the storage implementations already create the initial schema
			return nil
		},
		Down: func(storage storage.EnhancedStorage) error {
			// Cannot rollback the initial schema
			return fmt.Errorf("cannot rollback initial schema migration")
		},
	})

	// Example of a future migration
	// mm.RegisterMigration(MigrationStep{
	//     Version:     2,
	//     Description: "Add user preferences table",
	//     Up: func(storage storage.EnhancedStorage) error {
	//         // Implementation would go here
	//         return nil
	//     },
	//     Down: func(storage storage.EnhancedStorage) error {
	//         // Rollback implementation would go here
	//         return nil
	//     },
	// })
}

