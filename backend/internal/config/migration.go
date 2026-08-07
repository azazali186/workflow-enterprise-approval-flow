package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// MigrationRunner handles database migrations
type MigrationRunner struct {
	cfg    *Config
	m      *migrate.Migrate
	logger *zap.Logger
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(cfg *Config, migrationsPath string) (*MigrationRunner, error) {
	logger := cfg.Logger

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory not found: %s", migrationsPath)
	}

	// Create migrate instance with file source
	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath))
	m, err := migrate.New(sourceURL, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	logger.Info("migration runner initialized",
		zap.String("migrations_path", migrationsPath),
		zap.String("database", redactURL(cfg.DatabaseURL)),
	)

	return &MigrationRunner{
		cfg:    cfg,
		m:      m,
		logger: logger,
	}, nil
}

// Up runs all pending migrations
func (mr *MigrationRunner) Up() error {
	mr.logger.Info("running database migrations")

	err := mr.m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			mr.logger.Info("no pending migrations")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	mr.logger.Info("migrations completed successfully")
	return nil
}

// Down rolls back the last migration
func (mr *MigrationRunner) Down() error {
	mr.logger.Info("rolling back last migration")

	if err := mr.m.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			mr.logger.Info("no migrations to rollback")
			return nil
		}
		return fmt.Errorf("rollback failed: %w", err)
	}

	mr.logger.Info("rollback completed successfully")
	return nil
}

// Rollback rolls back the specified number of migrations
func (mr *MigrationRunner) Rollback(steps int) error {
	mr.logger.Info("rolling back migrations", zap.Int("steps", steps))

	if err := mr.m.Steps(-steps); err != nil {
		if err == migrate.ErrNoChange {
			mr.logger.Info("no migrations to rollback")
			return nil
		}
		return fmt.Errorf("rollback failed: %w", err)
	}

	mr.logger.Info("rollback completed successfully")
	return nil
}

// Version returns the current migration version
func (mr *MigrationRunner) Version() (uint, bool, error) {
	return mr.m.Version()
}

// Close closes the migration instance
func (mr *MigrationRunner) Close() (source error, db error) {
	return mr.m.Close()
}

// redactURL redacts password from database URL for logging
func redactURL(url string) string {
	for i := 0; i < len(url); i++ {
		if i+1 < len(url) && url[i] == ':' && url[i+1] == '/' {
			// Find the @ sign
			for j := i + 2; j < len(url); j++ {
				if url[j] == '@' {
					return url[:i+3] + "***@" + url[j+1:]
				}
			}
		}
	}
	return url
}

// RunMigrationsFromMain is the entry point for running migrations from main.go
func RunMigrationsFromMain(cfg *Config, migrationsDir string) error {
	// Check if the directory contains migration files
	var hasMigrations bool
	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".sql" {
			hasMigrations = true
		}
		return nil
	})
	if err != nil {
		cfg.Warn("failed to check migrations directory", zap.Error(err))
		// Try AutoMigrate as fallback
		return nil
	}

	if !hasMigrations {
		cfg.Info("no migration files found, skipping golang-migrate")
		return nil
	}

	runner, err := NewMigrationRunner(cfg, migrationsDir)
	if err != nil {
		cfg.Error("failed to create migration runner, falling back to AutoMigrate", zap.Error(err))
		return nil
	}
	defer runner.Close()

	version, dirty, err := runner.Version()
	if err != nil && err != migrate.ErrNilVersion {
		cfg.Warn("failed to get migration version", zap.Error(err))
	} else {
		cfg.Info("current migration version",
			zap.Uint("version", version),
			zap.Bool("dirty", dirty),
		)
	}

	if err := runner.Up(); err != nil {
		cfg.Error("migration failed", zap.Error(err))
		return err
	}

	return nil
}
