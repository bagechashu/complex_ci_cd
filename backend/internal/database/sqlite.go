package database

import (
	"database/sql"
	"fmt"

	"built-and-deploy/pkg/logger"
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

// migrateSchemaV1 creates the complete initial schema with all tables and indexes
func migrateSchemaV1(db *sql.DB) error {
	schema := `
	-- Core application management tables
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
		environment TEXT DEFAULT '',
		registry_prefix TEXT DEFAULT '',
		labels TEXT,
		kubeconfig TEXT,
		kubernetes_version TEXT DEFAULT NULL,
		k8s_connection_status TEXT DEFAULT 'unknown',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS workload_target (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_id INTEGER NOT NULL,
		env_id INTEGER NOT NULL,
		cluster_id INTEGER NOT NULL,
		k8s_namespace TEXT,
		k8s_workload TEXT,
		container_name TEXT,
		registry_domain TEXT,
		image_repo TEXT,
		workload_type TEXT DEFAULT 'Deployment',
		workload_name TEXT,
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

	-- Shell server and command management tables
	CREATE TABLE IF NOT EXISTS shell_server (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		username TEXT NOT NULL,
		auth_type TEXT NOT NULL,
		password TEXT,
		private_key TEXT,
		status TEXT DEFAULT 'active',
		last_connected DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS shell_command (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL,
		command TEXT NOT NULL,
		description TEXT,
		is_published BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES shell_server(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS shell_command_execution (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL,
		command_id INTEGER NOT NULL,
		status TEXT DEFAULT 'pending',
		output TEXT,
		error_message TEXT,
		command_params TEXT,
		exit_code INTEGER,
		started_at DATETIME,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES shell_server(id) ON DELETE CASCADE,
		FOREIGN KEY (command_id) REFERENCES shell_command(id) ON DELETE CASCADE
	);

	-- Core indexes
	CREATE INDEX IF NOT EXISTS idx_application_name ON application(name);
	CREATE INDEX IF NOT EXISTS idx_environment_name ON environment(name);
	CREATE INDEX IF NOT EXISTS idx_cluster_name ON cluster(name);
	CREATE INDEX IF NOT EXISTS idx_cluster_type ON cluster(type);

	-- Release record indexes
	CREATE INDEX IF NOT EXISTS idx_release_app_env ON release_record(app_id, env_id);
	CREATE INDEX IF NOT EXISTS idx_release_status ON release_record(status);
	CREATE INDEX IF NOT EXISTS idx_release_created ON release_record(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_release_status_created ON release_record(status, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_release_app_status ON release_record(app_id, status);
	CREATE INDEX IF NOT EXISTS idx_release_app_cluster ON release_record(app_id, cluster_id);

	-- Release event indexes
	CREATE INDEX IF NOT EXISTS idx_event_release ON release_event(release_id);

	-- Workload target indexes
	CREATE INDEX IF NOT EXISTS idx_workload_target ON workload_target(app_id, env_id, cluster_id);
	CREATE INDEX IF NOT EXISTS idx_workload_app_env_cluster ON workload_target(app_id, env_id, cluster_id);
	CREATE INDEX IF NOT EXISTS idx_workload_app ON workload_target(app_id);
	CREATE INDEX IF NOT EXISTS idx_workload_cluster ON workload_target(cluster_id);

	-- Audit log indexes
	CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_log_resource ON audit_log(resource_type, resource_id);
	CREATE INDEX IF NOT EXISTS idx_audit_log_operation ON audit_log(operation, created_at DESC);

	-- Shell server and command execution indexes
	CREATE INDEX IF NOT EXISTS idx_shell_server_name ON shell_server(name);
	CREATE INDEX IF NOT EXISTS idx_shell_server_status ON shell_server(status);
	CREATE INDEX IF NOT EXISTS idx_shell_command_server ON shell_command(server_id);
	CREATE INDEX IF NOT EXISTS idx_shell_command_published ON shell_command(is_published);
	CREATE INDEX IF NOT EXISTS idx_shell_command_execution_status ON shell_command_execution(status);
	CREATE INDEX IF NOT EXISTS idx_shell_command_execution_created ON shell_command_execution(created_at DESC);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return recordSchemaVersion(db, 1, "Complete initial schema: application, environment, cluster, workload_target, release_record, release_event, audit_log, shell_server, shell_command, shell_command_execution")
}

// applyMigrations applies all pending migrations based on current schema version
func applyMigrations(db *sql.DB) error {
	log := logger.GetLogger()
	currentVersion, err := getSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	if currentVersion >= CurrentSchemaVersion {
		return nil // Already at latest version
	}

	// Apply migrations in order
	if err := migrateSchemaV1(db); err != nil {
		return fmt.Errorf("failed to apply schema v1: %w", err)
	}
	log.Info("Applied schema version 1: Complete initial schema created")

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

	// Note: Initial data should be loaded manually using backend/db/init_data.sql
	// Run: sqlite3 create-sample-data.db < backend/db/init_data.sql
	logger.GetLogger().Info("Database tables created successfully", "note", "Run: sqlite3 create-sample-data.db < backend/db/init_data.sql to load sample data")

	return nil
}
