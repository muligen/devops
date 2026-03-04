// Package migrate provides database migration functionality.
package migrate

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrator handles database migrations.
type Migrator struct {
	migrate *migrate.Migrate
}

// New creates a new Migrator instance.
func New(db *gorm.DB) (*Migrator, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

// Up runs all pending migrations.
func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Down rolls back the last migration.
func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	return nil
}

// Steps runs n migrations (positive for up, negative for down).
func (m *Migrator) Steps(n int) error {
	if err := m.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration steps: %w", err)
	}
	return nil
}

// Version returns the current migration version.
func (m *Migrator) Version() (version uint, dirty bool, err error) {
	return m.migrate.Version()
}

// Close closes the migrator and releases resources.
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}
