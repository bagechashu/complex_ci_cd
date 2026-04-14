package database

import "database/sql"

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS application (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		repo TEXT,
		build_type TEXT,
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
		kubeconfig_path TEXT,
		kubeconfig_encrypted TEXT,
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
		FOREIGN KEY (app_id) REFERENCES application(id),
		FOREIGN KEY (env_id) REFERENCES environment(id),
		FOREIGN KEY (cluster_id) REFERENCES cluster(id)
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
		FOREIGN KEY (app_id) REFERENCES application(id),
		FOREIGN KEY (env_id) REFERENCES environment(id),
		FOREIGN KEY (cluster_id) REFERENCES cluster(id)
	);

	CREATE TABLE IF NOT EXISTS release_event (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		message TEXT,
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (release_id) REFERENCES release_record(id)
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

	CREATE INDEX IF NOT EXISTS idx_release_app_env ON release_record(app_id, env_id);
	CREATE INDEX IF NOT EXISTS idx_release_status ON release_record(status);
	CREATE INDEX IF NOT EXISTS idx_event_release ON release_event(release_id);
	CREATE INDEX IF NOT EXISTS idx_deployment_target ON deployment_target(app_id, env_id, cluster_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return nil
}
