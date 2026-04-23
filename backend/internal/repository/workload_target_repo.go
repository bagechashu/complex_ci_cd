package repository

import (
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
)

const (
    sqWorkloadTargetInsert = "INSERT INTO workload_target (app_id, env_id, cluster_id, k8s_namespace, k8s_workload, container_name, registry_domain, image_repo, workload_type, workload_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
    sqWorkloadTargetSelect = "SELECT id, app_id, env_id, cluster_id, k8s_namespace, k8s_workload, container_name, registry_domain, image_repo, workload_type, workload_name, created_at, updated_at FROM workload_target"
    sqWorkloadTargetUpdate = "UPDATE workload_target SET app_id = ?, env_id = ?, cluster_id = ?, k8s_namespace = ?, k8s_workload = ?, container_name = ?, registry_domain = ?, image_repo = ?, workload_type = ?, workload_name = ?, updated_at = ? WHERE id = ?"
    sqWorkloadTargetDelete = "DELETE FROM workload_target WHERE id = ?"
    sqWorkloadTargetCount  = "SELECT COUNT(*) FROM workload_target"
)

type WorkloadTargetRepository interface {
    Create(target *models.WorkloadTarget) (*models.WorkloadTarget, error)
    GetByID(id int) (*models.WorkloadTarget, error)
    GetByApp(appID int) ([]*models.WorkloadTarget, error)
    List(limit, offset int) ([]*models.WorkloadTarget, error)
    Update(target *models.WorkloadTarget) error
    Delete(id int) error
}

type SQLiteWorkloadTargetRepository struct {
    db *sql.DB
}

func NewWorkloadTargetRepository(db *sql.DB) WorkloadTargetRepository {
    return &SQLiteWorkloadTargetRepository{db: db}
}

func (r *SQLiteWorkloadTargetRepository) Create(target *models.WorkloadTarget) (*models.WorkloadTarget, error) {
    now := time.Now()
    target.CreatedAt = now
    target.UpdatedAt = now

    // Using dummy environmental ID for now
    envID := 1

    res, err := r.db.Exec(sqWorkloadTargetInsert,
        target.AppID, envID, target.ClusterID, target.K8sNamespace, "", "", "", "", "", "", target.CreatedAt, target.UpdatedAt)
    if err != nil {
        return nil, err
    }
    id, err := res.LastInsertId()
    if err != nil {
        return nil, err
    }
    target.ID = int(id)
    return target, nil
}

func (r *SQLiteWorkloadTargetRepository) GetByID(id int) (*models.WorkloadTarget, error) {
    var t models.WorkloadTarget
    var envID int
    var containerName, registryDomain, imageRepo sql.NullString
    var workloadType, workloadName string
    err := r.db.QueryRow(sqWorkloadTargetSelect+" WHERE id = ?", id).Scan(
        &t.ID, &t.AppID, &envID, &t.ClusterID, &t.K8sNamespace, &t.K8sWorkload, &containerName, &registryDomain, &imageRepo, &workloadType, &workloadName, &t.CreatedAt, &t.UpdatedAt)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, errors.New("not found")
    }
    if err != nil {
        return nil, err
    }
    t.EnvID = envID
    t.WorkloadType = workloadType
    t.WorkloadName = workloadName
    if containerName.Valid {
        t.ContainerName = &containerName.String
    }
    if registryDomain.Valid {
        t.RegistryDomain = &registryDomain.String
    }
    if imageRepo.Valid {
        t.ImageRepo = &imageRepo.String
    }
    return &t, nil
}

func (r *SQLiteWorkloadTargetRepository) GetByApp(appID int) ([]*models.WorkloadTarget, error) {
    rows, err := r.db.Query(sqWorkloadTargetSelect+" WHERE app_id = ?", appID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []*models.WorkloadTarget
    for rows.Next() {
        var t models.WorkloadTarget
        var envID int
        var containerName, registryDomain, imageRepo sql.NullString
        var workloadType, workloadName string
        if err := rows.Scan(&t.ID, &t.AppID, &envID, &t.ClusterID, &t.K8sNamespace, &t.K8sWorkload, &containerName, &registryDomain, &imageRepo, &workloadType, &workloadName, &t.CreatedAt, &t.UpdatedAt); err != nil {
            return nil, err
        }
        t.EnvID = envID
        t.WorkloadType = workloadType
        t.WorkloadName = workloadName
        if containerName.Valid {
            t.ContainerName = &containerName.String
        }
        if registryDomain.Valid {
            t.RegistryDomain = &registryDomain.String
        }
        if imageRepo.Valid {
            t.ImageRepo = &imageRepo.String
        }
        result = append(result, &t)
    }
    return result, nil
}

func (r *SQLiteWorkloadTargetRepository) List(limit, offset int) ([]*models.WorkloadTarget, error) {
    rows, err := r.db.Query(sqWorkloadTargetSelect+" LIMIT ? OFFSET ?", limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []*models.WorkloadTarget
    for rows.Next() {
        var t models.WorkloadTarget
        var envID int
        var containerName, registryDomain, imageRepo sql.NullString
        var workloadType, workloadName string
        if err := rows.Scan(&t.ID, &t.AppID, &envID, &t.ClusterID, &t.K8sNamespace, &t.K8sWorkload, &containerName, &registryDomain, &imageRepo, &workloadType, &workloadName, &t.CreatedAt, &t.UpdatedAt); err != nil {
            return nil, err
        }
        t.EnvID = envID
        t.WorkloadType = workloadType
        t.WorkloadName = workloadName
        if containerName.Valid {
            t.ContainerName = &containerName.String
        }
        if registryDomain.Valid {
            t.RegistryDomain = &registryDomain.String
        }
        if imageRepo.Valid {
            t.ImageRepo = &imageRepo.String
        }
        result = append(result, &t)
    }
    return result, nil
}

func (r *SQLiteWorkloadTargetRepository) Update(target *models.WorkloadTarget) error {
    target.UpdatedAt = time.Now()
    envID := 1
    _, err := r.db.Exec(sqWorkloadTargetUpdate,
        target.AppID, envID, target.ClusterID, target.K8sNamespace, "", "", "", "", "", "", target.UpdatedAt, target.ID)
    return err
}

func (r *SQLiteWorkloadTargetRepository) Delete(id int) error {
    _, err := r.db.Exec(sqWorkloadTargetDelete, id)
    return err
}
