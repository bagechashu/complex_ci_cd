package database

import (
	"database/sql"
	"fmt"
	"os"
)

// CurrentSchemaVersion is the current database schema version
const CurrentSchemaVersion = 1

// initSchemaVersion initializes the schema_version table if it doesn't exist
func initSchemaVersion(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS schema_version (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version INTEGER NOT NULL UNIQUE,
		description TEXT,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(schema)
	return err
}

// getSchemaVersion returns the current schema version from database
func getSchemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// recordSchemaVersion records a new schema version in database
func recordSchemaVersion(db *sql.DB, version int, description string) error {
	_, err := db.Exec(
		"INSERT INTO schema_version (version, description) VALUES (?, ?)",
		version, description,
	)
	return err
}

// migrateSchemaV1 creates the initial schema (version 1)
func migrateSchemaV1(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS application (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		image_name TEXT NOT NULL,
		git_repo TEXT,
		build_type TEXT,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS environment (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		rank INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS cluster (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		labels TEXT,
		kubeconfig_path TEXT,
		kubeconfig_encrypted TEXT,
		ansible_hosts TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS deployment_target (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_id INTEGER NOT NULL,
		env_id INTEGER NOT NULL,
		cluster_id INTEGER NOT NULL,
		k8s_namespace TEXT,
		k8s_deployment TEXT,
		container_name TEXT,
		registry_domain TEXT,
		image_repo TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(app_id, env_id, cluster_id),
		FOREIGN KEY (app_id) REFERENCES application(id) ON DELETE CASCADE,
		FOREIGN KEY (env_id) REFERENCES environment(id) ON DELETE CASCADE,
		FOREIGN KEY (cluster_id) REFERENCES cluster(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS release_record (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_id INTEGER NOT NULL,
		env_id INTEGER NOT NULL,
		cluster_id INTEGER NOT NULL,
		image TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		previous_image TEXT,
		error_msg TEXT,
		triggered_by TEXT,
		started_at DATETIME,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (app_id) REFERENCES application(id) ON DELETE RESTRICT,
		FOREIGN KEY (env_id) REFERENCES environment(id) ON DELETE RESTRICT,
		FOREIGN KEY (cluster_id) REFERENCES cluster(id) ON DELETE RESTRICT
	);

	CREATE TABLE IF NOT EXISTS release_event (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		message TEXT,
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (release_id) REFERENCES release_record(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		operation TEXT,
		resource_type TEXT,
		resource_id INTEGER,
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_application_name ON application(name);
	CREATE INDEX IF NOT EXISTS idx_environment_name ON environment(name);
	CREATE INDEX IF NOT EXISTS idx_cluster_name ON cluster(name);
	CREATE INDEX IF NOT EXISTS idx_cluster_type ON cluster(type);
	CREATE INDEX IF NOT EXISTS idx_release_app_env ON release_record(app_id, env_id);
	CREATE INDEX IF NOT EXISTS idx_release_status ON release_record(status);
	CREATE INDEX IF NOT EXISTS idx_release_created ON release_record(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_event_release ON release_event(release_id);
	CREATE INDEX IF NOT EXISTS idx_deployment_target ON deployment_target(app_id, env_id, cluster_id);
	CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at DESC);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return recordSchemaVersion(db, 1, "Initial schema with application, environment, cluster, deployment_target, release_record, release_event, audit_log")
}

// migrateSchemaV2 is a placeholder for future migrations
// Example: adding new fields, creating new tables, etc.
// To use: uncomment and implement when needed
/*
func migrateSchemaV2(db *sql.DB) error {
	// Example: Add new column to application table
	schema := `
	ALTER TABLE application ADD COLUMN image_name TEXT;
	ALTER TABLE cluster ADD COLUMN registry_prefix TEXT;
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return recordSchemaVersion(db, 2, "Added image_name to application and registry_prefix to cluster")
}
*/

// applyMigrations applies all pending migrations based on current schema version
func applyMigrations(db *sql.DB) error {
	currentVersion, err := getSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	if currentVersion >= CurrentSchemaVersion {
		return nil // Already at latest version
	}

	// Apply migrations in order
	for version := currentVersion + 1; version <= CurrentSchemaVersion; version++ {
		switch version {
		case 1:
			if err := migrateSchemaV1(db); err != nil {
				return fmt.Errorf("failed to apply schema v%d: %w", version, err)
			}
		// case 2:
		//	if err := migrateSchemaV2(db); err != nil {
		//		return fmt.Errorf("failed to apply schema v%d: %w", version, err)
		//	}
		default:
			return fmt.Errorf("unknown schema version: %d", version)
		}
	}

	return nil
}

// createTables initializes database schema with version management
func createTables(db *sql.DB) error {
	// First, create schema_version table
	if err := initSchemaVersion(db); err != nil {
		return fmt.Errorf("failed to initialize schema_version table: %w", err)
	}

	// Then apply all pending migrations
	if err := applyMigrations(db); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Check if we should insert initial data
	// Set environment variable INIT_DATA=true to insert test data on first run
	if os.Getenv("INIT_DATA") == "true" {
		// Only insert if this is the first time (schema_version has only version 1)
		currentVersion, err := getSchemaVersion(db)
		if err != nil {
			return fmt.Errorf("failed to check schema version: %w", err)
		}

		// Insert initial data only on first migration (V1 just applied)
		if currentVersion == 1 {
			if err := InsertInitialData(db); err != nil {
				return fmt.Errorf("failed to insert initial data: %w", err)
			}
		}
	}

	return nil
}
