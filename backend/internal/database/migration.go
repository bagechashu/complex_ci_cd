package database

import (
	"database/sql"
	"fmt"
)

// InsertInitialData inserts sample data for testing
func InsertInitialData(db *sql.DB) error {
	// Insert sample applications
	apps := []map[string]interface{}{
		{"name": "api-service", "repo": "git@github.com:company/api-service.git", "build_type": "docker"},
		{"name": "web-ui", "repo": "git@github.com:company/web-ui.git", "build_type": "docker"},
		{"name": "data-processor", "repo": "git@github.com:company/data-processor.git", "build_type": "docker"},
	}

	for _, app := range apps {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO application (name, repo, build_type) VALUES (?, ?, ?)",
			app["name"], app["repo"], app["build_type"],
		)
		if err != nil {
			return err
		}
	}

	// Insert sample environments
	envs := []map[string]interface{}{
		{"name": "development", "rank": 1},
		{"name": "staging", "rank": 2},
		{"name": "production", "rank": 3},
	}

	for _, env := range envs {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO environment (name, rank) VALUES (?, ?)",
			env["name"], env["rank"],
		)
		if err != nil {
			return err
		}
	}

	// Insert sample clusters
	clusters := []map[string]interface{}{
		{"name": "cluster-dev", "type": "kubernetes"},
		{"name": "cluster-staging", "type": "kubernetes"},
		{"name": "cluster-prod-1", "type": "kubernetes"},
		{"name": "cluster-prod-2", "type": "kubernetes"},
	}

	for _, cluster := range clusters {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO cluster (name, type) VALUES (?, ?)",
			cluster["name"], cluster["type"],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetConnectionInfo returns database info
func GetConnectionInfo(db *sql.DB) (map[string]interface{}, error) {
	info := map[string]interface{}{}

	// Get application count
	var appCount int
	err := db.QueryRow("SELECT COUNT(*) FROM application").Scan(&appCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications: %w", err)
	}
	info["applications"] = appCount

	// Get environment count
	var envCount int
	err = db.QueryRow("SELECT COUNT(*) FROM environment").Scan(&envCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count environments: %w", err)
	}
	info["environments"] = envCount

	// Get cluster count
	var clusterCount int
	err = db.QueryRow("SELECT COUNT(*) FROM cluster").Scan(&clusterCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count clusters: %w", err)
	}
	info["clusters"] = clusterCount

	// Get release count
	var releaseCount int
	err = db.QueryRow("SELECT COUNT(*) FROM release_record").Scan(&releaseCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count releases: %w", err)
	}
	info["releases"] = releaseCount

	return info, nil
}
