package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"built-and-deploy/internal/models"
)

// ReleaseEventRepository 发布事件数据访问接口
type ReleaseEventRepository interface {
	Create(ctx context.Context, event *models.ReleaseEvent) error
	ListByRelease(ctx context.Context, releaseID int) ([]*models.ReleaseEvent, error)
}

// SQLiteReleaseEventRepository SQLite 实现的发布事件仓储
type SQLiteReleaseEventRepository struct {
	db *sql.DB
}

// SQL query constants
const (
	sqlReleaseEventInsert = `
		INSERT INTO release_event (release_id, type, message, details, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	sqlReleaseEventSelect = `
		SELECT id, release_id, type, message, details, created_at
		FROM release_event
	`

	sqlReleaseEventSelectByID = sqlReleaseEventSelect + ` WHERE id = ?`

	sqlReleaseEventSelectByRelease = sqlReleaseEventSelect + ` WHERE release_id = ? ORDER BY created_at DESC`

	sqlReleaseEventDelete = `DELETE FROM release_event WHERE id = ?`
)

// NewSQLiteReleaseEventRepository creates a new SQLite release event repository
func NewSQLiteReleaseEventRepository(db *sql.DB) ReleaseEventRepository {
	return &SQLiteReleaseEventRepository{db: db}
}

// Create creates a new release event
func (r *SQLiteReleaseEventRepository) Create(ctx context.Context, event *models.ReleaseEvent) error {
	if err := validateReleaseEvent(event); err != nil {
		return fmt.Errorf("invalid release event: %w", err)
	}

	now := time.Now()
	result, err := r.db.ExecContext(ctx, sqlReleaseEventInsert,
		event.ReleaseID, event.Type, event.Message, event.Details, now)
	if err != nil {
		return fmt.Errorf("failed to insert release event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get insert id: %w", err)
	}

	event.ID = int(id)
	event.CreatedAt = now
	return nil
}

// ListByRelease lists all events for a release
func (r *SQLiteReleaseEventRepository) ListByRelease(ctx context.Context, releaseID int) ([]*models.ReleaseEvent, error) {
	if releaseID <= 0 {
		return nil, fmt.Errorf("invalid release id: %d", releaseID)
	}

	rows, err := r.db.QueryContext(ctx, sqlReleaseEventSelectByRelease, releaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query release events: %w", err)
	}
	defer rows.Close()

	var events []*models.ReleaseEvent
	for rows.Next() {
		event := &models.ReleaseEvent{}
		err := rows.Scan(&event.ID, &event.ReleaseID, &event.Type, &event.Message, &event.Details, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan release event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating release events: %w", err)
	}

	if len(events) == 0 {
		return []*models.ReleaseEvent{}, nil
	}

	return events, nil
}

// validateReleaseEvent validates a release event
func validateReleaseEvent(event *models.ReleaseEvent) error {
	if event == nil {
		return errors.New("release event cannot be nil")
	}
	if event.ReleaseID <= 0 {
		return fmt.Errorf("invalid release id: %d", event.ReleaseID)
	}
	if event.Type == "" {
		return errors.New("event type is required")
	}
	return nil
}
