package auth

import (
	"testing"

	"github.com/pragneshbagary/go-auth/internal/storage/memory"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

type testStorage = storage.EnhancedStorage

func TestMigrationManager_NewMigrationManager(t *testing.T) {
	storage := memory.NewInMemoryStorage()
	mm := NewMigrationManager(storage)

	if mm == nil {
		t.Fatal("Expected migration manager to be created")
	}

	if mm.storage != storage {
		t.Error("Expected storage to be set correctly")
	}

	// Should have at least the initial migration registered
	if len(mm.steps) == 0 {
		t.Error("Expected at least one built-in migration to be registered")
	}
}

func TestMigrationManager_RegisterMigration(t *testing.T) {
	storage := memory.NewInMemoryStorage()
	mm := NewMigrationManager(storage)

	initialCount := len(mm.steps)

	// Register a new migration
	testMigration := MigrationStep{
		Version:     10,
		Description: "Test migration",
		Up: func(s testStorage) error {
			return nil
		},
		Down: func(s testStorage) error {
			return nil
		},
	}

	mm.RegisterMigration(testMigration)

	if len(mm.steps) != initialCount+1 {
		t.Errorf("Expected %d migrations, got %d", initialCount+1, len(mm.steps))
	}

	// Check that migrations are sorted by version
	for i := 1; i < len(mm.steps); i++ {
		if mm.steps[i-1].Version >= mm.steps[i].Version {
			t.Error("Migrations should be sorted by version")
		}
	}
}

func TestMigrationManager_GetCurrentVersion(t *testing.T) {
	storage := memory.NewInMemoryStorage()
	mm := NewMigrationManager(storage)

	version, err := mm.GetCurrentVersion()
	if err != nil {
		t.Fatalf("Failed to get current version: %v", err)
	}

	// In-memory storage starts with version 1
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
}

func TestMigrationManager_GetPendingMigrations(t *testing.T) {
	storage := memory.NewInMemoryStorage()
	mm := NewMigrationManager(storage)

	// Add a migration with version higher than current
	testMigration := MigrationStep{
		Version:     5,
		Description: "Test pending migration",
		Up: func(s testStorage) error {
			return nil
		},
	}
	mm.RegisterMigration(testMigration)

	pending, err := mm.GetPendingMigrations()
	if err != nil {
		t.Fatalf("Failed to get pending migrations: %v", err)
	}

	// Should have at least the test migration as pending
	found := false
	for _, migration := range pending {
		if migration.Version == 5 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected test migration to be in pending list")
	}
}

func TestMigrationManager_GetAppliedMigrations(t *testing.T) {
	storage := memory.NewInMemoryStorage()
	mm := NewMigrationManager(storage)

	applied, err := mm.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	// Should have at least the initial migration
	if len(applied) == 0 {
		t.Error("Expected at least one applied migration")
	}
}

func TestMigrationManager_Migrate(t *testing.T) {
	storage := memory.NewInMemoryStorage()
	mm := NewMigrationManager(storage)

	// Add a test migration
	migrationExecuted := false
	testMigration := MigrationStep{
		Version:     2,
		Description: "Test migration execution",
		Up: func(s testStorage) error {
			migrationExecuted = true
			return nil
		},
	}
	mm.RegisterMigration(testMigration)

	// Run migrations
	err := mm.Migrate()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if !migrationExecuted {
		t.Error("Expected test migration to be executed")
	}

	// Check that version was updated
	version, err := mm.GetCurrentVersion()
	if err != nil {
		t.Fatalf("Failed to get current version: %v", err)
	}

	if version < 2 {
		t.Errorf("Expected version to be at least 2, got %d", version)
	}
}

func TestMigrationManager_Integration_WithAuth(t *testing.T) {
	// Test that Auth service properly integrates with migration manager
	auth, err := NewInMemory("test-secret")
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	// Test that migration manager is accessible
	mm := auth.Migrations()
	if mm == nil {
		t.Fatal("Expected migration manager to be accessible from Auth service")
	}

	// Test schema version access
	version, err := auth.GetSchemaVersion()
	if err != nil {
		t.Fatalf("Failed to get schema version from Auth service: %v", err)
	}

	if version < 1 {
		t.Errorf("Expected schema version to be at least 1, got %d", version)
	}

	// Test that we can add custom migrations
	customMigrationExecuted := false
	customMigration := MigrationStep{
		Version:     10,
		Description: "Custom test migration",
		Up: func(s testStorage) error {
			customMigrationExecuted = true
			return nil
		},
	}

	mm.RegisterMigration(customMigration)

	// Run migration
	err = auth.MigrateToVersion(10)
	if err != nil {
		t.Fatalf("Failed to run custom migration: %v", err)
	}

	if !customMigrationExecuted {
		t.Error("Expected custom migration to be executed")
	}

	// Verify version
	version, err = auth.GetSchemaVersion()
	if err != nil {
		t.Fatalf("Failed to get schema version: %v", err)
	}

	if version != 10 {
		t.Errorf("Expected version 10, got %d", version)
	}
}