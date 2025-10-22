package migration

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// RunMigrations runs all pending migrations using Goose
func RunMigrations(dbURL string, migrationsDir string) error {
	db, err := getDB(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Run migrations
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CreateMigration creates a new migration file
func CreateMigration(name string, migrationsDir string) error {
	// Create the migration files
	if err := goose.Create(nil, migrationsDir, name, "sql"); err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	return nil
}

// MigrateUp runs the next migration
func MigrateUp(dbURL string, migrationsDir string) error {
	db, err := getDB(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.UpByOne(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run up migration: %w", err)
	}

	return nil
}

// MigrateDown rolls back the last migration
func MigrateDown(dbURL string, migrationsDir string) error {
	db, err := getDB(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Down(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run down migration: %w", err)
	}

	return nil
}

// MigrateStatus shows the status of all migrations
func MigrateStatus(dbURL string, migrationsDir string) error {
	db, err := getDB(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Status(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// MigrateVersion shows the current migration version
func MigrateVersion(dbURL string, migrationsDir string) error {
	db, err := getDB(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	fmt.Printf("Current migration version: %d\n", version)
	return nil
}

// ResetMigrations rolls back all migrations
func ResetMigrations(dbURL string, migrationsDir string) error {
	db, err := getDB(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Reset(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	return nil
}

// Helper functions
func getDB(dbURL string) (*sql.DB, error) {

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	// Add TLS configuration to handle certificate issues
	// dbURL = dbURL

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

func getMigrationsDir() string {
	migrationsDir := filepath.Join("service", "db", "migrations")
	absPath, _ := filepath.Abs(migrationsDir)
	return absPath
}
