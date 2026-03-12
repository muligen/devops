// Package migrate provides database migration functionality.
package migrate

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrator handles database migrations.
type Migrator struct {
	db *gorm.DB
}

// New creates a new Migrator instance.
func New(db *gorm.DB) (*Migrator, error) {
	// List all migration files from embedded FS
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Get only up migration files and ensure sorted order
	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	// Log the migration files
	for _, f := range upFiles {
		fmt.Printf("Found migration: %s\n", f)
	}

	return &Migrator{db: db}, nil
}

// Up runs all pending migrations in order.
func (m *Migrator) Up() error {
	// List all migration files from embedded FS
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Get only up migration files and ensure sorted order
	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	// For each migration file in order
	for _, fileName := range upFiles {
		// Extract version number from filename (e.g., "001" from "001_initial_schema.up.sql")
		baseName := strings.TrimSuffix(fileName, ".up.sql")
		parts := strings.SplitN(baseName, "_", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename: %s", fileName)
		}
		version := parts[0]

		// Check if this migration has already been applied
		var count int64
		err := m.db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ? AND dirty = false", version).Scan(&count).Error
		if err == nil && count > 0 {
			fmt.Printf("Migration %s already applied, skipping\n", fileName)
			continue
		}

		// Read migration file content
		fullPath := filepath.ToSlash(filepath.Join("migrations", fileName))
		content, err := migrationsFS.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		// Extract only the Up section (before the Down section)
		migrationSQL := string(content)
		if idx := strings.Index(migrationSQL, "-- +migrate Down"); idx != -1 {
			migrationSQL = migrationSQL[:idx]
		}

		// Execute migration
		if err := m.db.Exec(migrationSQL).Error; err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}

		// Create schema_migrations table if it doesn't exist and record this migration
		m.db.Exec(`
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version BIGINT PRIMARY KEY,
				dirty BOOLEAN NOT NULL DEFAULT false
			)
		`)

		if err := m.db.Exec(`
			INSERT INTO schema_migrations (version, dirty)
			VALUES (?, false)
			ON CONFLICT (version) DO UPDATE SET dirty = false
		`, version).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", fileName, err)
		}

		fmt.Printf("Successfully applied migration: %s\n", fileName)
	}

	return nil
}

// Down rolls back the last migration.
func (m *Migrator) Down() error {
	return fmt.Errorf("manual rollback not implemented. Use database tools to rollback manually")
}

// Steps runs n migrations (positive for up, negative for down).
func (m *Migrator) Steps(n int) error {
	if n > 0 {
		return m.Up()
	}
	return m.Down()
}

// Version returns the current migration version.
func (m *Migrator) Version() (version uint, dirty bool, err error) {
	var v int64
	if err := m.db.Raw("SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = false").Scan(&v).Error; err != nil {
		return 0, false, err
	}

	var d bool
	if err := m.db.Raw("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE dirty = true)").Scan(&d).Error; err != nil {
		return uint(v), false, nil
	}

	return uint(v), d, nil
}

// Close releases resources (no-op for manual migration).
func (m *Migrator) Close() error {
	return nil
}
