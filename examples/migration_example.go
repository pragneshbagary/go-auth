package main

import (
	"fmt"
	"log"

	"github.com/pragneshbagary/go-auth/pkg/auth"
	"github.com/pragneshbagary/go-auth/pkg/storage"
)

func main() {
	// Create auth service with SQLite
	authService, err := auth.NewSQLite("migration_example.db", "your-secret-key")
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	fmt.Println("=== Database Migration System Example ===")

	// Get current schema version
	version, err := authService.GetSchemaVersion()
	if err != nil {
		log.Fatalf("Failed to get schema version: %v", err)
	}
	fmt.Printf("Current schema version: %d\n", version)

	// Access the migration manager for advanced operations
	migrationManager := authService.Migrations()

	// Register a custom migration
	customMigration := auth.MigrationStep{
		Version:     2,
		Description: "Add user preferences table",
		Up: func(s storage.EnhancedStorage) error {
			fmt.Println("  Executing migration 2: Adding user preferences table...")
			// In a real scenario, you would execute SQL here
			// For this example, we'll just simulate the migration
			return nil
		},
		Down: func(s storage.EnhancedStorage) error {
			fmt.Println("  Rolling back migration 2: Dropping user preferences table...")
			// In a real scenario, you would execute rollback SQL here
			return nil
		},
	}

	migrationManager.RegisterMigration(customMigration)

	// Register another migration
	anotherMigration := auth.MigrationStep{
		Version:     3,
		Description: "Add user activity log table",
		Up: func(s storage.EnhancedStorage) error {
			fmt.Println("  Executing migration 3: Adding user activity log table...")
			return nil
		},
		Down: func(s storage.EnhancedStorage) error {
			fmt.Println("  Rolling back migration 3: Dropping user activity log table...")
			return nil
		},
	}

	migrationManager.RegisterMigration(anotherMigration)

	// Check for pending migrations
	pending, err := migrationManager.GetPendingMigrations()
	if err != nil {
		log.Fatalf("Failed to get pending migrations: %v", err)
	}

	fmt.Printf("\nPending migrations: %d\n", len(pending))
	for _, migration := range pending {
		fmt.Printf("  - Version %d: %s\n", migration.Version, migration.Description)
	}

	// Run all pending migrations
	if len(pending) > 0 {
		fmt.Println("\nRunning migrations...")
		err = migrationManager.Migrate()
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations completed successfully!")
	}

	// Check new version
	version, err = authService.GetSchemaVersion()
	if err != nil {
		log.Fatalf("Failed to get schema version: %v", err)
	}
	fmt.Printf("New schema version: %d\n", version)

	// Show applied migrations
	applied, err := migrationManager.GetAppliedMigrations()
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	fmt.Printf("\nApplied migrations: %d\n", len(applied))
	for _, migration := range applied {
		fmt.Printf("  - Version %d: %s (applied at %s)\n", 
			migration.Version, migration.Description, migration.AppliedAt.Format("2006-01-02 15:04:05"))
	}

	// Demonstrate targeted migration
	fmt.Println("\n=== Targeted Migration Example ===")
	fmt.Println("Rolling back to version 2...")
	err = authService.RollbackToVersion(2)
	if err != nil {
		log.Fatalf("Rollback failed: %v", err)
	}

	version, err = authService.GetSchemaVersion()
	if err != nil {
		log.Fatalf("Failed to get schema version: %v", err)
	}
	fmt.Printf("Schema version after rollback: %d\n", version)

	// Migrate back up to latest
	fmt.Println("\nMigrating back to latest version...")
	err = migrationManager.Migrate()
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	version, err = authService.GetSchemaVersion()
	if err != nil {
		log.Fatalf("Failed to get schema version: %v", err)
	}
	fmt.Printf("Final schema version: %d\n", version)

	// Demonstrate migration to specific version
	fmt.Println("\n=== Migration to Specific Version ===")
	fmt.Println("Migrating to version 2...")
	err = authService.MigrateToVersion(2)
	if err != nil {
		log.Fatalf("Migration to version 2 failed: %v", err)
	}

	version, err = authService.GetSchemaVersion()
	if err != nil {
		log.Fatalf("Failed to get schema version: %v", err)
	}
	fmt.Printf("Schema version after targeted migration: %d\n", version)

	fmt.Println("\n=== Migration System Features ===")
	fmt.Println("✓ Automatic migration execution on service startup")
	fmt.Println("✓ Custom migration registration")
	fmt.Println("✓ Up and down migration support")
	fmt.Println("✓ Migration versioning and tracking")
	fmt.Println("✓ Rollback functionality")
	fmt.Println("✓ Targeted migration to specific versions")
	fmt.Println("✓ Pending and applied migration queries")
	fmt.Println("✓ Error handling and transaction safety")

	fmt.Println("\nMigration system example completed successfully!")
}