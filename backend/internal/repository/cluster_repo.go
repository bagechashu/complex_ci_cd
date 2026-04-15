package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/op/release-control/internal/models"
)

type ClusterRepository interface {
	Create(ctx context.Context, cluster *models.Cluster) error
	GetByID(ctx context.Context, id int) (*models.Cluster, error)
	List(ctx context.Context, offset, limit int) ([]*models.Cluster, int, error)
	Update(ctx context.Context, cluster *models.Cluster) error
	Delete(ctx context.Context, id int) error
}

type SQLiteClusterRepository struct {
	db *sql.DB
}

func NewSQLiteClusterRepository(db *sql.DB) ClusterRepository {
	return &SQLiteClusterRepository{db: db}
}

func (r *SQLiteClusterRepository) Create(ctx context.Context, cluster *models.Cluster) error {
	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO cluster (name, type, labels, kubeconfig_path, kubeconfig_encrypted, ansible_hosts, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?)",
		cluster.Name, cluster.Type, cluster.Labels, cluster.KubeconfigPath, cluster.KubeconfigEncrypted, cluster.AnsibleHosts, cluster.CreatedAt, cluster.UpdatedAt)
	return err
}

func (r *SQLiteClusterRepository) GetByID(ctx context.Context, id int) (*models.Cluster, error) {
	var c models.Cluster
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, type, labels, kubeconfig_path, kubeconfig_encrypted, ansible_hosts, created_at, updated_at FROM cluster WHERE id = ?",
		id).Scan(&c.ID, &c.Name, &c.Type, &c.Labels, &c.KubeconfigPath, &c.KubeconfigEncrypted, &c.AnsibleHosts, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("cluster not found")
	}
	return &c, err
}

func (r *SQLiteClusterRepository) List(ctx context.Context, offset, limit int) ([]*models.Cluster, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cluster").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, type, labels, kubeconfig_path, kubeconfig_encrypted, ansible_hosts, created_at, updated_at FROM cluster ORDER BY created_at DESC LIMIT ? OFFSET ?",
		limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clusters []*models.Cluster
	for rows.Next() {
		var c models.Cluster
		err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Labels, &c.KubeconfigPath, &c.KubeconfigEncrypted, &c.AnsibleHosts, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		clusters = append(clusters, &c)
	}
	return clusters, total, rows.Err()
}

func (r *SQLiteClusterRepository) Update(ctx context.Context, cluster *models.Cluster) error {
	cluster.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		"UPDATE cluster SET name=?, type=?, labels=?, kubeconfig_path=?, kubeconfig_encrypted=?, ansible_hosts=?, updated_at=? WHERE id=?",
		cluster.Name, cluster.Type, cluster.Labels, cluster.KubeconfigPath, cluster.KubeconfigEncrypted, cluster.AnsibleHosts, cluster.UpdatedAt, cluster.ID)
	return err
}

func (r *SQLiteClusterRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM cluster WHERE id=?", id)
	return err
}
