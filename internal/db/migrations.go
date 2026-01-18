package db

import (
	"embed"
	"fmt"

	"github.com/steveredden/KindredCard/internal/logger"
)

//go:embed all:migrations/*.sql
var migrationFiles embed.FS

func (d *Database) Migrate() error {
	// 1. Ensure the schema_migrations table exists
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// 2. Read migration files from the embedded filesystem
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}

	// 3. Get currently applied version
	var currentVersion int
	err = d.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return err
	}

	// 4. Apply new migrations in order
	for _, entry := range entries {
		var version int
		_, err := fmt.Sscanf(entry.Name(), "%d_", &version)
		if err != nil || version <= currentVersion {
			continue
		}

		logger.Info("[DATABASE] Applying migration: %s", entry.Name())
		content, _ := migrationFiles.ReadFile("migrations/" + entry.Name())

		tx, err := d.db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("error in %s: %w", entry.Name(), err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}
