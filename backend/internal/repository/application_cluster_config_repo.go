package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/op/release-control/internal/models"
)

const (
	sqAppClusterConfigInsert = "INSERT INTO application_cluster_config (id, application_id, cluster_id, labels, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	sqAppClusterConfigSelect = "SELECT id, application_id, cluster_id, labels, created_at, updated_at FROM application_cluster_config"
	sqAppClusterConfigUpdate = "UPDATE application_cluster_config SET labels = ?, updated_at = ? WHERE id = ?"
	sqAppClusterConfigDelete = "DELETE FROM application_cluster_config WHERE id = ?"
	sqAppClusterConfigDeleteByAppCluster = "DELETE FROM application_cluster_config WHERE application_id = ? AND cluster_id = ?"
	sqAppClusterConfigCount = "SELECT COUNT(*) FROM application_cluster_config"
)

// ApplicationClusterConfigRepository 定义应用-集群配置的数据访问接口
type ApplicationClusterConfigRepository interface {
	Create(ctx context.Context, config *models.ApplicationClusterConfig) error
	GetByID(ctx context.Context, id string) (*models.ApplicationClusterConfig, error)
	GetByApplicationAndCluster(ctx context.Context, appID, clusterID string) (*models.ApplicationClusterConfig, error)
	GetByApplication(ctx context.Context, applicationID string) ([]*models.ApplicationClusterConfig, error)
	GetByCluster(ctx context.Context, clusterID string) ([]*models.ApplicationClusterConfig, error)
	Update(ctx context.Context, config *models.ApplicationClusterConfig) error
	Delete(ctx context.Context, id string) error
	DeleteByApplicationAndCluster(ctx context.Context, appID, clusterID string) error
	List(ctx context.Context, offset, limit int) ([]*models.ApplicationClusterConfig, int, error)
}

// SQLiteApplicationClusterConfigRepository SQLite 实现
type SQLiteApplicationClusterConfigRepository struct {
	db *sql.DB
}

// NewSQLiteApplicationClusterConfigRepository 创建新的 SQLite Repository
func NewSQLiteApplicationClusterConfigRepository(db *sql.DB) ApplicationClusterConfigRepository {
	return &SQLiteApplicationClusterConfigRepository{db: db}
}

// Create 实现 Create 方法
func (r *SQLiteApplicationClusterConfigRepository) Create(ctx context.Context, config *models.ApplicationClusterConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config data: %w", err)
	}

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, sqAppClusterConfigInsert,
		config.ID,
		config.ApplicationID,
		config.ClusterID,
		config.Labels,
		config.CreatedAt,
		config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert config: %w", err)
	}

	return nil
}

// GetByID 实现 GetByID 方法
func (r *SQLiteApplicationClusterConfigRepository) GetByID(ctx context.Context, id string) (*models.ApplicationClusterConfig, error) {
	config := &models.ApplicationClusterConfig{}
	err := r.db.QueryRowContext(ctx, sqAppClusterConfigSelect+" WHERE id = ?", id).Scan(
		&config.ID,
		&config.ApplicationID,
		&config.ClusterID,
		&config.Labels,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query config: %w", err)
	}

	return config, nil
}

// GetByApplicationAndCluster 实现 GetByApplicationAndCluster 方法
func (r *SQLiteApplicationClusterConfigRepository) GetByApplicationAndCluster(ctx context.Context, appID, clusterID string) (*models.ApplicationClusterConfig, error) {
	config := &models.ApplicationClusterConfig{}
	err := r.db.QueryRowContext(ctx, sqAppClusterConfigSelect+" WHERE application_id = ? AND cluster_id = ? LIMIT 1", appID, clusterID).Scan(
		&config.ID,
		&config.ApplicationID,
		&config.ClusterID,
		&config.Labels,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query config: %w", err)
	}

	return config, nil
}

// GetByApplication 实现 GetByApplication 方法
func (r *SQLiteApplicationClusterConfigRepository) GetByApplication(ctx context.Context, applicationID string) ([]*models.ApplicationClusterConfig, error) {
	rows, err := r.db.QueryContext(ctx, sqAppClusterConfigSelect+" WHERE application_id = ? ORDER BY created_at DESC", applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.ApplicationClusterConfig
	for rows.Next() {
		config := &models.ApplicationClusterConfig{}
		err := rows.Scan(
			&config.ID,
			&config.ApplicationID,
			&config.ClusterID,
			&config.Labels,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}

	return configs, nil
}

// GetByCluster 实现 GetByCluster 方法
func (r *SQLiteApplicationClusterConfigRepository) GetByCluster(ctx context.Context, clusterID string) ([]*models.ApplicationClusterConfig, error) {
	rows, err := r.db.QueryContext(ctx, sqAppClusterConfigSelect+" WHERE cluster_id = ? ORDER BY created_at DESC", clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to query configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.ApplicationClusterConfig
	for rows.Next() {
		config := &models.ApplicationClusterConfig{}
		err := rows.Scan(
			&config.ID,
			&config.ApplicationID,
			&config.ClusterID,
			&config.Labels,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %w", err)
	}

	return configs, nil
}

// Update 实现 Update 方法
func (r *SQLiteApplicationClusterConfigRepository) Update(ctx context.Context, config *models.ApplicationClusterConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config data: %w", err)
	}

	config.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, sqAppClusterConfigUpdate,
		config.Labels,
		config.UpdatedAt,
		config.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("config not found")
	}

	return nil
}

// Delete 实现 Delete 方法
func (r *SQLiteApplicationClusterConfigRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, sqAppClusterConfigDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("config not found")
	}

	return nil
}

// DeleteByApplicationAndCluster 实现 DeleteByApplicationAndCluster 方法
func (r *SQLiteApplicationClusterConfigRepository) DeleteByApplicationAndCluster(ctx context.Context, appID, clusterID string) error {
	_, err := r.db.ExecContext(ctx, sqAppClusterConfigDeleteByAppCluster, appID, clusterID)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	return nil
}

// List 实现 List 方法
func (r *SQLiteApplicationClusterConfigRepository) List(ctx context.Context, offset, limit int) ([]*models.ApplicationClusterConfig, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqAppClusterConfigCount).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count configs: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqAppClusterConfigSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.ApplicationClusterConfig
	for rows.Next() {
		config := &models.ApplicationClusterConfig{}
		err := rows.Scan(
			&config.ID,
			&config.ApplicationID,
			&config.ClusterID,
			&config.Labels,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row error: %w", err)
	}

	return configs, total, nil
}
