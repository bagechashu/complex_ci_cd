package repository

import (
	"database/sql"
	"fmt"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqWorkloadTargetInsert = "INSERT INTO workload_target (app_id, env_id, cluster_id, k8s_namespace, k8s_workload, container_name, registry_domain, image_repo, workload_type, workload_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	sqWorkloadTargetSelect = "SELECT id, app_id, env_id, cluster_id, k8s_namespace, k8s_workload, container_name, registry_domain, image_repo, workload_type, workload_name, created_at, updated_at FROM workload_target"
	sqWorkloadTargetUpdate = "UPDATE workload_target SET app_id = ?, env_id = ?, cluster_id = ?, k8s_namespace = ?, k8s_workload = ?, container_name = ?, registry_domain = ?, image_repo = ?, workload_type = ?, workload_name = ?, updated_at = ? WHERE id = ?"
	sqWorkloadTargetDelete = "DELETE FROM workload_target WHERE id = ?"
)

type WorkloadTargetRepository struct {
	db *sql.DB
}

func NewWorkloadTargetRepository(db *sql.DB) *WorkloadTargetRepository {
	return &WorkloadTargetRepository{db: db}
}

func (r *WorkloadTargetRepository) Create(target *models.WorkloadTarget) (*models.WorkloadTarget, error) {
	result, err := r.db.Exec(
		sqWorkloadTargetInsert,
		target.AppID, target.EnvID, target.ClusterID, target.K8sNamespace, target.K8sWorkload, target.ContainerName, target.RegistryDomain, target.ImageRepo, target.WorkloadType, target.WorkloadName, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workload target: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get lastInsertId: %w", err)
	}

	target.ID = int(id)
	target.CreatedAt = time.Now()
	target.UpdatedAt = time.Now()
	return target, nil
}

func (r *WorkloadTargetRepository) GetByID(id int) (*models.WorkloadTarget, error) {
	target := &models.WorkloadTarget{}
	var containerName, registryDomain, imageRepo *string
	err := r.db.QueryRow(
		sqWorkloadTargetSelect+" WHERE id = ?",
		id,
	).Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sWorkload, &containerName, &registryDomain, &imageRepo, &target.WorkloadType, &target.WorkloadName, &target.CreatedAt, &target.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workload target not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workload target: %w", err)
	}

	target.ContainerName = containerName
	target.RegistryDomain = registryDomain
	target.ImageRepo = imageRepo

	return target, nil
}

func (r *WorkloadTargetRepository) GetByAppEnvCluster(appID, envID, clusterID int) (*models.WorkloadTarget, error) {
	target := &models.WorkloadTarget{}
	var containerName, registryDomain, imageRepo *string
	err := r.db.QueryRow(
		sqWorkloadTargetSelect+" WHERE app_id = ? AND env_id = ? AND cluster_id = ?",
		appID, envID, clusterID,
	).Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sWorkload, &containerName, &registryDomain, &imageRepo, &target.WorkloadType, &target.WorkloadName, &target.CreatedAt, &target.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workload target not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workload target: %w", err)
	}

	target.ContainerName = containerName
	target.RegistryDomain = registryDomain
	target.ImageRepo = imageRepo

	return target, nil
}

func (r *WorkloadTargetRepository) List(limit int, offset int) ([]*models.WorkloadTarget, error) {
	rows, err := r.db.Query(sqWorkloadTargetSelect+" ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list workload targets: %w", err)
	}
	defer rows.Close()

	var targets []*models.WorkloadTarget
	for rows.Next() {
		target := &models.WorkloadTarget{}
		var containerName, registryDomain, imageRepo *string
		err := rows.Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sWorkload, &containerName, &registryDomain, &imageRepo, &target.WorkloadType, &target.WorkloadName, &target.CreatedAt, &target.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workload target: %w", err)
		}
		target.ContainerName = containerName
		target.RegistryDomain = registryDomain
		target.ImageRepo = imageRepo
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workload targets: %w", err)
	}

	return targets, nil
}

func (r *WorkloadTargetRepository) GetByApp(appID int) ([]*models.WorkloadTarget, error) {
	rows, err := r.db.Query(sqWorkloadTargetSelect+" WHERE app_id = ? ORDER BY id DESC", appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workload targets by app: %w", err)
	}
	defer rows.Close()

	var targets []*models.WorkloadTarget
	for rows.Next() {
		target := &models.WorkloadTarget{}
		var containerName, registryDomain, imageRepo *string
		err := rows.Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sWorkload, &containerName, &registryDomain, &imageRepo, &target.WorkloadType, &target.WorkloadName, &target.CreatedAt, &target.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workload target: %w", err)
		}
		target.ContainerName = containerName
		target.RegistryDomain = registryDomain
		target.ImageRepo = imageRepo
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workload targets: %w", err)
	}

	return targets, nil
}

func (r *WorkloadTargetRepository) ListByAppAndEnv(appID int, envID int) ([]*models.WorkloadTarget, error) {
	rows, err := r.db.Query(sqWorkloadTargetSelect+" WHERE app_id = ? AND env_id = ? ORDER BY id DESC", appID, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workload targets: %w", err)
	}
	defer rows.Close()

	var targets []*models.WorkloadTarget
	for rows.Next() {
		target := &models.WorkloadTarget{}
		var containerName, registryDomain, imageRepo *string
		err := rows.Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sWorkload, &containerName, &registryDomain, &imageRepo, &target.WorkloadType, &target.WorkloadName, &target.CreatedAt, &target.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workload target: %w", err)
		}
		target.ContainerName = containerName
		target.RegistryDomain = registryDomain
		target.ImageRepo = imageRepo
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workload targets: %w", err)
	}

	return targets, nil
}

func (r *WorkloadTargetRepository) Update(target *models.WorkloadTarget) error {
	target.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		sqWorkloadTargetUpdate,
		target.AppID, target.EnvID, target.ClusterID, target.K8sNamespace, target.K8sWorkload, target.ContainerName, target.RegistryDomain, target.ImageRepo, target.WorkloadType, target.WorkloadName, target.UpdatedAt, target.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workload target: %w", err)
	}
	return nil
}

func (r *WorkloadTargetRepository) Delete(id int) error {
	_, err := r.db.Exec(sqWorkloadTargetDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete workload target: %w", err)
	}
	return nil
}
