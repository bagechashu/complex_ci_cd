package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/op/release-control/internal/models"
)

type ClusterRepository struct {
	db *sql.DB
}

func NewClusterRepository(db *sql.DB) *ClusterRepository {
	return &ClusterRepository{db: db}
}

func (r *ClusterRepository) Create(cluster *models.Cluster) (*models.Cluster, error) {
	result, err := r.db.Exec(
		"INSERT INTO cluster (name, type, kubeconfig_path, kubeconfig_encrypted, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		cluster.Name, cluster.Type, cluster.KubeconfigPath, cluster.KubeconfigEncrypted, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get lastInsertId: %w", err)
	}

	cluster.ID = int(id)
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	return cluster, nil
}

func (r *ClusterRepository) GetByID(id int) (*models.Cluster, error) {
	cluster := &models.Cluster{}
	err := r.db.QueryRow(
		"SELECT id, name, type, kubeconfig_path, kubeconfig_encrypted, created_at, updated_at FROM cluster WHERE id = ?",
		id,
	).Scan(&cluster.ID, &cluster.Name, &cluster.Type, &cluster.KubeconfigPath, &cluster.KubeconfigEncrypted, &cluster.CreatedAt, &cluster.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cluster not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	return cluster, nil
}

func (r *ClusterRepository) List(limit int, offset int) ([]*models.Cluster, error) {
	rows, err := r.db.Query("SELECT id, name, type, kubeconfig_path, kubeconfig_encrypted, created_at, updated_at FROM cluster ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}
	defer rows.Close()

	var clusters []*models.Cluster
	for rows.Next() {
		cluster := &models.Cluster{}
		err := rows.Scan(&cluster.ID, &cluster.Name, &cluster.Type, &cluster.KubeconfigPath, &cluster.KubeconfigEncrypted, &cluster.CreatedAt, &cluster.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cluster: %w", err)
		}
		clusters = append(clusters, cluster)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clusters: %w", err)
	}

	return clusters, nil
}

func (r *ClusterRepository) Update(cluster *models.Cluster) error {
	cluster.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		"UPDATE cluster SET name = ?, type = ?, kubeconfig_path = ?, kubeconfig_encrypted = ?, updated_at = ? WHERE id = ?",
		cluster.Name, cluster.Type, cluster.KubeconfigPath, cluster.KubeconfigEncrypted, cluster.UpdatedAt, cluster.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update cluster: %w", err)
	}
	return nil
}

func (r *ClusterRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM cluster WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}
	return nil
}
