package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/op/release-control/internal/models"
)

type ReleaseRecordRepository struct {
	db *sql.DB
}

func NewReleaseRecordRepository(db *sql.DB) *ReleaseRecordRepository {
	return &ReleaseRecordRepository{db: db}
}

func (r *ReleaseRecordRepository) Create(release *models.ReleaseRecord) (*models.ReleaseRecord, error) {
	result, err := r.db.Exec(
		"INSERT INTO release_record (app_id, env_id, cluster_id, image, status, previous_image, error_msg, triggered_by, started_at, completed_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		release.AppID, release.EnvID, release.ClusterID, release.Image, release.Status, release.PreviousImage, release.ErrorMsg, release.TriggeredBy, release.StartedAt, release.CompletedAt, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create release record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get lastInsertId: %w", err)
	}

	release.ID = int(id)
	release.CreatedAt = time.Now()
	release.UpdatedAt = time.Now()
	return release, nil
}

func (r *ReleaseRecordRepository) GetByID(id int) (*models.ReleaseRecord, error) {
	release := &models.ReleaseRecord{}
	err := r.db.QueryRow(
		"SELECT id, app_id, env_id, cluster_id, image, status, previous_image, error_msg, triggered_by, started_at, completed_at, created_at, updated_at FROM release_record WHERE id = ?",
		id,
	).Scan(&release.ID, &release.AppID, &release.EnvID, &release.ClusterID, &release.Image, &release.Status, &release.PreviousImage, &release.ErrorMsg, &release.TriggeredBy, &release.StartedAt, &release.CompletedAt, &release.CreatedAt, &release.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("release record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get release record: %w", err)
	}

	return release, nil
}

func (r *ReleaseRecordRepository) List(limit int, offset int) ([]*models.ReleaseRecord, error) {
	rows, err := r.db.Query(
		"SELECT id, app_id, env_id, cluster_id, image, status, previous_image, error_msg, triggered_by, started_at, completed_at, created_at, updated_at FROM release_record ORDER BY id DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list release records: %w", err)
	}
	defer rows.Close()

	var releases []*models.ReleaseRecord
	for rows.Next() {
		release := &models.ReleaseRecord{}
		err := rows.Scan(&release.ID, &release.AppID, &release.EnvID, &release.ClusterID, &release.Image, &release.Status, &release.PreviousImage, &release.ErrorMsg, &release.TriggeredBy, &release.StartedAt, &release.CompletedAt, &release.CreatedAt, &release.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan release record: %w", err)
		}
		releases = append(releases, release)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating release records: %w", err)
	}

	return releases, nil
}

func (r *ReleaseRecordRepository) Update(release *models.ReleaseRecord) error {
	release.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		"UPDATE release_record SET app_id = ?, env_id = ?, cluster_id = ?, image = ?, status = ?, previous_image = ?, error_msg = ?, triggered_by = ?, started_at = ?, completed_at = ?, updated_at = ? WHERE id = ?",
		release.AppID, release.EnvID, release.ClusterID, release.Image, release.Status, release.PreviousImage, release.ErrorMsg, release.TriggeredBy, release.StartedAt, release.CompletedAt, release.UpdatedAt, release.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update release record: %w", err)
	}
	return nil
}

func (r *ReleaseRecordRepository) CreateEvent(event *models.ReleaseEvent) error {
	_, err := r.db.Exec(
		"INSERT INTO release_event (release_id, type, message, details, created_at) VALUES (?, ?, ?, ?, ?)",
		event.ReleaseID, event.Type, event.Message, event.Details, time.Now(),
	)
	return err
}

func (r *ReleaseRecordRepository) GetEvents(releaseID int) ([]*models.ReleaseEvent, error) {
	rows, err := r.db.Query(
		"SELECT id, release_id, type, message, details, created_at FROM release_event WHERE release_id = ? ORDER BY created_at ASC",
		releaseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	var events []*models.ReleaseEvent
	for rows.Next() {
		event := &models.ReleaseEvent{}
		err := rows.Scan(&event.ID, &event.ReleaseID, &event.Type, &event.Message, &event.Details, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}
