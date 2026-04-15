package database

import (
	"database/sql"
	"fmt"
	"time"
)

// InsertInitialData inserts comprehensive sample data for testing
func InsertInitialData(db *sql.DB) error {
	// 1. Insert sample applications
	apps := []map[string]interface{}{
		{"name": "api-service", "image_name": "api-service", "git_repo": "git@github.com:company/api-service.git", "build_type": "docker"},
		{"name": "web-ui", "image_name": "web-ui", "git_repo": "git@github.com:company/web-ui.git", "build_type": "docker"},
		{"name": "data-processor", "image_name": "data-processor", "git_repo": "git@github.com:company/data-processor.git", "build_type": "docker"},
	}

	appIDs := make(map[string]int64)
	for _, app := range apps {
		result, err := db.Exec(
			"INSERT OR IGNORE INTO application (name, image_name, git_repo, build_type) VALUES (?, ?, ?, ?)",
			app["name"], app["image_name"], app["git_repo"], app["build_type"],
		)
		if err != nil {
			return fmt.Errorf("failed to insert application: %w", err)
		}
		id, _ := result.LastInsertId()
		if id == 0 {
			// Already exists, fetch from database
			var existingID int64
			db.QueryRow("SELECT id FROM application WHERE name = ?", app["name"]).Scan(&existingID)
			appIDs[app["name"].(string)] = existingID
		} else {
			appIDs[app["name"].(string)] = id
		}
	}

	// 2. Insert sample environments
	envs := []map[string]interface{}{
		{"name": "development", "rank": 1},
		{"name": "staging", "rank": 2},
		{"name": "production", "rank": 3},
	}

	envIDs := make(map[string]int64)
	for _, env := range envs {
		result, err := db.Exec(
			"INSERT OR IGNORE INTO environment (name, rank) VALUES (?, ?)",
			env["name"], env["rank"],
		)
		if err != nil {
			return fmt.Errorf("failed to insert environment: %w", err)
		}
		id, _ := result.LastInsertId()
		if id == 0 {
			var existingID int64
			db.QueryRow("SELECT id FROM environment WHERE name = ?", env["name"]).Scan(&existingID)
			envIDs[env["name"].(string)] = existingID
		} else {
			envIDs[env["name"].(string)] = id
		}
	}

	// 3. Insert sample clusters
	clusters := []map[string]interface{}{
		{"name": "k8s-dev", "type": "kubernetes"},
		{"name": "k8s-staging", "type": "kubernetes"},
		{"name": "k8s-prod-cn1", "type": "kubernetes"},
		{"name": "k8s-prod-cn2", "type": "kubernetes"},
	}

	clusterIDs := make(map[string]int64)
	for _, cluster := range clusters {
		result, err := db.Exec(
			"INSERT OR IGNORE INTO cluster (name, type) VALUES (?, ?)",
			cluster["name"], cluster["type"],
		)
		if err != nil {
			return fmt.Errorf("failed to insert cluster: %w", err)
		}
		id, _ := result.LastInsertId()
		if id == 0 {
			var existingID int64
			db.QueryRow("SELECT id FROM cluster WHERE name = ?", cluster["name"]).Scan(&existingID)
			clusterIDs[cluster["name"].(string)] = existingID
		} else {
			clusterIDs[cluster["name"].(string)] = id
		}
	}

	// 4. Insert deployment targets (app-env-cluster mappings)
	deploymentTargets := []map[string]interface{}{
		// api-service
		{"app": "api-service", "env": "development", "cluster": "k8s-dev", "namespace": "development", "deployment": "api-service", "container": "api"},
		{"app": "api-service", "env": "staging", "cluster": "k8s-staging", "namespace": "staging", "deployment": "api-service", "container": "api"},
		{"app": "api-service", "env": "production", "cluster": "k8s-prod-cn1", "namespace": "production", "deployment": "api-service", "container": "api-prod"},
		{"app": "api-service", "env": "production", "cluster": "k8s-prod-cn2", "namespace": "production", "deployment": "api-service", "container": "api-prod"},
		// web-ui
		{"app": "web-ui", "env": "development", "cluster": "k8s-dev", "namespace": "development", "deployment": "web-ui", "container": "web"},
		{"app": "web-ui", "env": "staging", "cluster": "k8s-staging", "namespace": "staging", "deployment": "web-ui", "container": "web"},
		{"app": "web-ui", "env": "production", "cluster": "k8s-prod-cn1", "namespace": "production", "deployment": "web-ui", "container": "web-prod"},
		// data-processor
		{"app": "data-processor", "env": "development", "cluster": "k8s-dev", "namespace": "development", "deployment": "data-processor", "container": "processor"},
		{"app": "data-processor", "env": "staging", "cluster": "k8s-staging", "namespace": "staging", "deployment": "data-processor", "container": "processor"},
		{"app": "data-processor", "env": "production", "cluster": "k8s-prod-cn1", "namespace": "production", "deployment": "data-processor", "container": "processor-prod"},
	}

	releaseIDs := make([]int64, 0)
	for _, dt := range deploymentTargets {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO deployment_target (app_id, env_id, cluster_id, k8s_namespace, k8s_deployment, container_name) VALUES (?, ?, ?, ?, ?, ?)",
			appIDs[dt["app"].(string)],
			envIDs[dt["env"].(string)],
			clusterIDs[dt["cluster"].(string)],
			dt["namespace"], dt["deployment"], dt["container"],
		)
		if err != nil {
			return fmt.Errorf("failed to insert deployment target: %w", err)
		}
	}

	// 5. Insert release records
	releaseRecords := []map[string]interface{}{
		{
			"app": "api-service", "env": "development", "cluster": "k8s-dev",
			"image": "api-service:v1.2.3", "status": "success",
			"previous_image": "api-service:v1.2.2",
			"triggered_by": "user@example.com",
		},
		{
			"app": "api-service", "env": "staging", "cluster": "k8s-staging",
			"image": "api-service:v1.2.3", "status": "success",
			"previous_image": "api-service:v1.2.2",
			"triggered_by": "CI/CD Pipeline",
		},
		{
			"app": "api-service", "env": "production", "cluster": "k8s-prod-cn1",
			"image": "api-service:v1.2.3", "status": "deploying",
			"previous_image": "api-service:v1.2.1",
			"triggered_by": "admin@example.com",
		},
		{
			"app": "web-ui", "env": "development", "cluster": "k8s-dev",
			"image": "web-ui:v2.1.0", "status": "success",
			"previous_image": "web-ui:v2.0.9",
			"triggered_by": "developer@example.com",
		},
		{
			"app": "web-ui", "env": "production", "cluster": "k8s-prod-cn1",
			"image": "web-ui:v2.1.0", "status": "failed",
			"previous_image": "web-ui:v2.0.8",
			"error_msg": "Failed to pull image from registry",
			"triggered_by": "CI/CD Pipeline",
		},
		{
			"app": "data-processor", "env": "production", "cluster": "k8s-prod-cn1",
			"image": "data-processor:v3.5.0", "status": "success",
			"previous_image": "data-processor:v3.4.9",
			"triggered_by": "scheduler@system",
		},
	}

	now := time.Now()
	startTime := now.Add(-72 * time.Hour)

	for i, rr := range releaseRecords {
		completedAt := startTime.Add(time.Duration(i) * 24 * time.Hour)
		result, err := db.Exec(
			"INSERT OR IGNORE INTO release_record (app_id, env_id, cluster_id, image, status, previous_image, error_msg, triggered_by, started_at, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			appIDs[rr["app"].(string)],
			envIDs[rr["env"].(string)],
			clusterIDs[rr["cluster"].(string)],
			rr["image"], rr["status"], rr["previous_image"],
			rr["error_msg"], rr["triggered_by"],
			completedAt, completedAt.Add(5*time.Minute),
		)
		if err != nil {
			return fmt.Errorf("failed to insert release record: %w", err)
		}
		id, _ := result.LastInsertId()
		if id > 0 {
			releaseIDs = append(releaseIDs, id)
		}
	}

	// 6. Insert release events
	if len(releaseIDs) > 0 {
		eventTypes := []string{"started", "deploying", "success", "failed"}
		messages := []string{
			"Release deployment initialized",
			"Pulling image from registry",
			"Applying Kubernetes manifests",
			"Waiting for pods to be ready",
			"Health check passed",
			"Deployment completed successfully",
			"Image pull failed",
			"Invalid image tag",
		}

		for i, releaseID := range releaseIDs {
			// Add 3-5 events per release
			eventCount := 3 + (i % 3)
			eventTime := startTime.Add(time.Duration(i) * 24 * time.Hour)

			for j := 0; j < eventCount; j++ {
				eventType := eventTypes[j%len(eventTypes)]
				message := messages[j%len(messages)]
				eventTime = eventTime.Add(1 * time.Minute)

				_, err := db.Exec(
					"INSERT INTO release_event (release_id, type, message, created_at) VALUES (?, ?, ?, ?)",
					releaseID, eventType, message, eventTime,
				)
				if err != nil {
					return fmt.Errorf("failed to insert release event: %w", err)
				}
			}
		}
	}

	// 7. Insert audit logs
	auditOps := []map[string]interface{}{
		{"user": "admin@example.com", "operation": "CREATE", "resource_type": "application", "resource_id": appIDs["api-service"]},
		{"user": "admin@example.com", "operation": "CREATE", "resource_type": "environment", "resource_id": envIDs["production"]},
		{"user": "developer@example.com", "operation": "DEPLOY", "resource_type": "release", "resource_id": 1},
		{"user": "CI/CD Pipeline", "operation": "AUTO_DEPLOY", "resource_type": "release", "resource_id": 2},
		{"user": "admin@example.com", "operation": "UPDATE", "resource_type": "deployment_target", "resource_id": 1},
		{"user": "developer@example.com", "operation": "VIEW", "resource_type": "release_history", "resource_id": 1},
		{"user": "admin@example.com", "operation": "DELETE", "resource_type": "application", "resource_id": 0},
		{"user": "scheduler@system", "operation": "AUTO_DEPLOY", "resource_type": "release", "resource_id": 3},
	}

	auditTime := startTime
	for _, audit := range auditOps {
		auditTime = auditTime.Add(2 * time.Hour)
		_, err := db.Exec(
			"INSERT INTO audit_log (user_id, operation, resource_type, resource_id, created_at) VALUES (?, ?, ?, ?, ?)",
			audit["user"], audit["operation"], audit["resource_type"], audit["resource_id"], auditTime,
		)
		if err != nil {
			return fmt.Errorf("failed to insert audit log: %w", err)
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

	// Get deployment target count
	var deploymentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM deployment_target").Scan(&deploymentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count deployment targets: %w", err)
	}
	info["deployment_targets"] = deploymentCount

	// Get release count
	var releaseCount int
	err = db.QueryRow("SELECT COUNT(*) FROM release_record").Scan(&releaseCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count releases: %w", err)
	}
	info["releases"] = releaseCount

	// Get release event count
	var eventCount int
	err = db.QueryRow("SELECT COUNT(*) FROM release_event").Scan(&eventCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count release events: %w", err)
	}
	info["release_events"] = eventCount

	// Get audit log count
	var auditCount int
	err = db.QueryRow("SELECT COUNT(*) FROM audit_log").Scan(&auditCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}
	info["audit_logs"] = auditCount

	return info, nil
}
