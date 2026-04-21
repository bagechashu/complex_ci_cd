package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/pkg/utils"
)

const (
	sqClusterInsert = "INSERT INTO cluster (name, type, environment, registry_prefix, labels, kubeconfig, k8s_connection_status, ansible_hosts, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)"
	sqClusterSelect = "SELECT id, name, type, environment, registry_prefix, labels, kubeconfig, k8s_connection_status, ansible_hosts, created_at, updated_at FROM cluster"
	sqClusterUpdate = "UPDATE cluster SET name=?, type=?, environment=?, registry_prefix=?, labels=?, kubeconfig=?, k8s_connection_status=?, ansible_hosts=?, updated_at=? WHERE id=?"
	sqClusterDelete = "DELETE FROM cluster WHERE id=?"
	sqClusterCount  = "SELECT COUNT(*) FROM cluster"
)

type ClusterRepository interface {
	Create(ctx context.Context, cluster *models.Cluster) error
	GetByID(ctx context.Context, id int) (*models.Cluster, error)
	List(ctx context.Context, offset, limit int) ([]*models.Cluster, int, error)
	Update(ctx context.Context, cluster *models.Cluster) error
	Delete(ctx context.Context, id int) error
}

type SQLiteClusterRepository struct {
	db            *sql.DB
	encryptionKey string
}

func NewSQLiteClusterRepository(db *sql.DB, encryptionKey string) ClusterRepository {
	return &SQLiteClusterRepository{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

func (r *SQLiteClusterRepository) Create(ctx context.Context, cluster *models.Cluster) error {
	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now

	// Encrypt kubeconfig before storing
	var encryptedKubeconfig *string
	if cluster.Kubeconfig != nil && *cluster.Kubeconfig != "" {
		ciphertext, err := utils.EncryptAES(*cluster.Kubeconfig, r.encryptionKey)
		if err != nil {
			return err
		}
		encryptedKubeconfig = &ciphertext
	}

	_, err := r.db.ExecContext(ctx, sqClusterInsert,
		cluster.Name, cluster.Type, cluster.Environment, cluster.RegistryPrefix,
		cluster.Labels, encryptedKubeconfig,
		cluster.K8sConnectionStatus, cluster.AnsibleHosts, cluster.CreatedAt, cluster.UpdatedAt)
	return err
}

func (r *SQLiteClusterRepository) GetByID(ctx context.Context, id int) (*models.Cluster, error) {
	var c models.Cluster
	var encryptedKubeconfig *string
	err := r.db.QueryRowContext(ctx, sqClusterSelect+" WHERE id = ?", id).Scan(
		&c.ID, &c.Name, &c.Type, &c.Environment, &c.RegistryPrefix,
		&c.Labels, &encryptedKubeconfig,
		&c.K8sConnectionStatus, &c.AnsibleHosts, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("cluster not found")
	}
	if err != nil {
		return nil, err
	}

	// Decrypt kubeconfig
	if encryptedKubeconfig != nil && *encryptedKubeconfig != "" {
		plaintext, err := utils.DecryptAES(*encryptedKubeconfig, r.encryptionKey)
		if err != nil {
			return nil, err
		}
		c.Kubeconfig = &plaintext
	}

	return &c, nil
}

func (r *SQLiteClusterRepository) List(ctx context.Context, offset, limit int) ([]*models.Cluster, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqClusterCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqClusterSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clusters []*models.Cluster
	for rows.Next() {
		var c models.Cluster
		var encryptedKubeconfig *string
		err := rows.Scan(
			&c.ID, &c.Name, &c.Type, &c.Environment, &c.RegistryPrefix,
			&c.Labels, &encryptedKubeconfig,
			&c.K8sConnectionStatus, &c.AnsibleHosts, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}

		// Decrypt kubeconfig
		if encryptedKubeconfig != nil && *encryptedKubeconfig != "" {
			plaintext, err := utils.DecryptAES(*encryptedKubeconfig, r.encryptionKey)
			if err != nil {
				return nil, 0, err
			}
			c.Kubeconfig = &plaintext
		}

		clusters = append(clusters, &c)
	}
	return clusters, total, rows.Err()
}

func (r *SQLiteClusterRepository) Update(ctx context.Context, cluster *models.Cluster) error {
	cluster.UpdatedAt = time.Now()

	// Encrypt kubeconfig before storing
	var encryptedKubeconfig *string
	if cluster.Kubeconfig != nil && *cluster.Kubeconfig != "" {
		ciphertext, err := utils.EncryptAES(*cluster.Kubeconfig, r.encryptionKey)
		if err != nil {
			return err
		}
		encryptedKubeconfig = &ciphertext
	}

	_, err := r.db.ExecContext(ctx, sqClusterUpdate,
		cluster.Name, cluster.Type, cluster.Environment, cluster.RegistryPrefix,
		cluster.Labels, encryptedKubeconfig,
		cluster.K8sConnectionStatus, cluster.AnsibleHosts, cluster.UpdatedAt, cluster.ID)
	return err
}

func (r *SQLiteClusterRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqClusterDelete, id)
	return err
}
