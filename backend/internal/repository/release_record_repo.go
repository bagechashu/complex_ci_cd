package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/op/release-control/internal/models"
)

const (
	sqReleaseRecordInsert = "INSERT INTO release_record (app_id, env_id, cluster_id, image, status, previous_image, error_msg, triggered_by, started_at, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
	sqReleaseRecordSelect = "SELECT id, app_id, env_id, cluster_id, image, status, previous_image, error_msg, triggered_by, started_at, completed_at, created_at, updated_at FROM release_record"
	sqReleaseRecordUpdate = "UPDATE release_record SET status=?, completed_at=?, updated_at=? WHERE id=?"
	sqReleaseRecordDelete = "DELETE FROM release_record WHERE id=?"
	sqReleaseRecordCount  = "SELECT COUNT(*) FROM release_record"
)

type ReleaseRecordRepository interface {
	Create(ctx context.Context, rr *models.ReleaseRecord) error
	GetByID(ctx context.Context, id int) (*models.ReleaseRecord, error)
	List(ctx context.Context, offset, limit int) ([]*models.ReleaseRecord, int, error)
	GetByApplicationAndCluster(ctx context.Context, appID, clusterID int) ([]*models.ReleaseRecord, error)
	Update(ctx context.Context, rr *models.ReleaseRecord) error
	Delete(ctx context.Context, id int) error
}

type SQLiteReleaseRecordRepository struct {
	db *sql.DB
}

func NewSQLiteReleaseRecordRepository(db *sql.DB) ReleaseRecordRepository {
	return &SQLiteReleaseRecordRepository{db: db}
}

func (r *SQLiteReleaseRecordRepository) Create(ctx context.Context, rr *models.ReleaseRecord) error {
	now := time.Now()
	rr.CreatedAt = now
	rr.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, sqReleaseRecordInsert,
		rr.AppID, rr.EnvID, rr.ClusterID, rr.Image, rr.Status, rr.PreviousImage, rr.ErrorMsg, rr.TriggeredBy, rr.StartedAt, rr.CreatedAt, rr.UpdatedAt)
	return err
}

func (r *SQLiteReleaseRecordRepository) GetByID(ctx context.Context, id int) (*models.ReleaseRecord, error) {
	var rr models.ReleaseRecord
	err := r.db.QueryRowContext(ctx, sqReleaseRecordSelect+" WHERE id = ?", id).Scan(&rr.ID, &rr.AppID, &rr.EnvID, &rr.ClusterID, &rr.Image, &rr.Status, &rr.PreviousImage, &rr.ErrorMsg, &rr.TriggeredBy, &rr.StartedAt, &rr.CompletedAt, &rr.CreatedAt, &rr.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("release not found")
	}
	return &rr, err
}

func (r *SQLiteReleaseRecordRepository) List(ctx context.Context, offset, limit int) ([]*models.ReleaseRecord, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqReleaseRecordCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqReleaseRecordSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*models.ReleaseRecord
	for rows.Next() {
		var rr models.ReleaseRecord
		err := rows.Scan(&rr.ID, &rr.AppID, &rr.EnvID, &rr.ClusterID, &rr.Image, &rr.Status, &rr.PreviousImage, &rr.ErrorMsg, &rr.TriggeredBy, &rr.StartedAt, &rr.CompletedAt, &rr.CreatedAt, &rr.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, &rr)
	}
	return records, total, rows.Err()
}

func (r *SQLiteReleaseRecordRepository) GetByApplicationAndCluster(ctx context.Context, appID, clusterID int) ([]*models.ReleaseRecord, error) {
	rows, err := r.db.QueryContext(ctx, sqReleaseRecordSelect+" WHERE app_id = ? AND cluster_id = ? ORDER BY created_at DESC", appID, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.ReleaseRecord
	for rows.Next() {
		var rr models.ReleaseRecord
		err := rows.Scan(&rr.ID, &rr.AppID, &rr.EnvID, &rr.ClusterID, &rr.Image, &rr.Status, &rr.PreviousImage, &rr.ErrorMsg, &rr.TriggeredBy, &rr.StartedAt, &rr.CompletedAt, &rr.CreatedAt, &rr.UpdatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, &rr)
	}
	return records, rows.Err()
}

func (r *SQLiteReleaseRecordRepository) Update(ctx context.Context, rr *models.ReleaseRecord) error {
	rr.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, sqReleaseRecordUpdate,
		rr.Status, rr.CompletedAt, rr.UpdatedAt, rr.ID)
	return err
}

func (r *SQLiteReleaseRecordRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqReleaseRecordDelete, id)
	return err
}
