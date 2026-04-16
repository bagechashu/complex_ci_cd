package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/op/release-control/internal/models"
)

const (
	sqDeploymentTargetInsert = "INSERT INTO deployment_target (app_id, env_id, cluster_id, k8s_namespace, k8s_deployment, container_name, registry_domain, image_repo, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	sqDeploymentTargetSelect = "SELECT id, app_id, env_id, cluster_id, k8s_namespace, k8s_deployment, container_name, registry_domain, image_repo, created_at, updated_at FROM deployment_target"
	sqDeploymentTargetUpdate = "UPDATE deployment_target SET app_id = ?, env_id = ?, cluster_id = ?, k8s_namespace = ?, k8s_deployment = ?, container_name = ?, registry_domain = ?, image_repo = ?, updated_at = ? WHERE id = ?"
	sqDeploymentTargetDelete = "DELETE FROM deployment_target WHERE id = ?"
)

type DeploymentTargetRepository struct {
	db *sql.DB
}

func NewDeploymentTargetRepository(db *sql.DB) *DeploymentTargetRepository {
	return &DeploymentTargetRepository{db: db}
}

func (r *DeploymentTargetRepository) Create(target *models.DeploymentTarget) (*models.DeploymentTarget, error) {
	result, err := r.db.Exec(
		sqDeploymentTargetInsert,
		target.AppID, target.EnvID, target.ClusterID, target.K8sNamespace, target.K8sDeployment, target.ContainerName, target.RegistryDomain, target.ImageRepo, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment target: %w", err)
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

func (r *DeploymentTargetRepository) GetByID(id int) (*models.DeploymentTarget, error) {
	target := &models.DeploymentTarget{}
	err := r.db.QueryRow(
		sqDeploymentTargetSelect+" WHERE id = ?",
		id,
	).Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sDeployment, &target.ContainerName, &target.RegistryDomain, &target.ImageRepo, &target.CreatedAt, &target.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("deployment target not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment target: %w", err)
	}

	return target, nil
}

func (r *DeploymentTargetRepository) GetByAppEnvCluster(appID, envID, clusterID int) (*models.DeploymentTarget, error) {
	target := &models.DeploymentTarget{}
	err := r.db.QueryRow(
		sqDeploymentTargetSelect+" WHERE app_id = ? AND env_id = ? AND cluster_id = ?",
		appID, envID, clusterID,
	).Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sDeployment, &target.ContainerName, &target.RegistryDomain, &target.ImageRepo, &target.CreatedAt, &target.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("deployment target not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment target: %w", err)
	}

	return target, nil
}

func (r *DeploymentTargetRepository) List(limit int, offset int) ([]*models.DeploymentTarget, error) {
	rows, err := r.db.Query(sqDeploymentTargetSelect+" ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployment targets: %w", err)
	}
	defer rows.Close()

	var targets []*models.DeploymentTarget
	for rows.Next() {
		target := &models.DeploymentTarget{}
		err := rows.Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sDeployment, &target.ContainerName, &target.RegistryDomain, &target.ImageRepo, &target.CreatedAt, &target.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment target: %w", err)
		}
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployment targets: %w", err)
	}

	return targets, nil
}

func (r *DeploymentTargetRepository) ListByAppAndEnv(appID int, envID int) ([]*models.DeploymentTarget, error) {
	rows, err := r.db.Query(sqDeploymentTargetSelect+" WHERE app_id = ? AND env_id = ? ORDER BY id DESC", appID, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployment targets: %w", err)
	}
	defer rows.Close()

	var targets []*models.DeploymentTarget
	for rows.Next() {
		target := &models.DeploymentTarget{}
		err := rows.Scan(&target.ID, &target.AppID, &target.EnvID, &target.ClusterID, &target.K8sNamespace, &target.K8sDeployment, &target.ContainerName, &target.RegistryDomain, &target.ImageRepo, &target.CreatedAt, &target.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment target: %w", err)
		}
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployment targets: %w", err)
	}

	return targets, nil
}

func (r *DeploymentTargetRepository) Update(target *models.DeploymentTarget) error {
	target.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		sqDeploymentTargetUpdate,
		target.AppID, target.EnvID, target.ClusterID, target.K8sNamespace, target.K8sDeployment, target.ContainerName, target.RegistryDomain, target.ImageRepo, target.UpdatedAt, target.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment target: %w", err)
	}
	return nil
}

func (r *DeploymentTargetRepository) Delete(id int) error {
	_, err := r.db.Exec(sqDeploymentTargetDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete deployment target: %w", err)
	}
	return nil
}
