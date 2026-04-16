package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/op/release-control/internal/models"
)

const (
	sqCommandApprovalInsert = "INSERT INTO command_approval (id, request_id, approval_status, approved_by, approved_at, created_at, updated_at) VALUES (?,?,?,?,?,?,?)"
	sqCommandApprovalSelect = "SELECT id, request_id, approval_status, approved_by, approved_at, created_at, updated_at FROM command_approval"
	sqCommandApprovalUpdate = "UPDATE command_approval SET request_id=?, approval_status=?, approved_by=?, approved_at=?, updated_at=? WHERE id=?"
	sqCommandApprovalDelete = "DELETE FROM command_approval WHERE id=?"
	sqCommandApprovalCount  = "SELECT COUNT(*) FROM command_approval"
)

type CommandApprovalRepository interface {
	Create(ctx context.Context, ca *models.CommandApproval) error
	GetByID(ctx context.Context, id string) (*models.CommandApproval, error)
	GetByRequestID(ctx context.Context, requestID string) (*models.CommandApproval, error)
	List(ctx context.Context, offset, limit int) ([]*models.CommandApproval, int, error)
	Update(ctx context.Context, ca *models.CommandApproval) error
	Delete(ctx context.Context, id string) error
}

type SQLiteCommandApprovalRepository struct {
	db *sql.DB
}

func NewSQLiteCommandApprovalRepository(db *sql.DB) CommandApprovalRepository {
	return &SQLiteCommandApprovalRepository{db: db}
}

func (r *SQLiteCommandApprovalRepository) Create(ctx context.Context, ca *models.CommandApproval) error {
	if err := ca.Validate(); err != nil {
		return err
	}
	now := time.Now()
	ca.CreatedAt = now
	ca.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, sqCommandApprovalInsert,
		ca.ID, ca.RequestID, ca.ApprovalStatus, ca.ApprovedBy, ca.ApprovedAt, ca.CreatedAt, ca.UpdatedAt)
	return err
}

func (r *SQLiteCommandApprovalRepository) GetByID(ctx context.Context, id string) (*models.CommandApproval, error) {
	var ca models.CommandApproval
	err := r.db.QueryRowContext(ctx, sqCommandApprovalSelect+" WHERE id = ?", id).Scan(&ca.ID, &ca.RequestID, &ca.ApprovalStatus, &ca.ApprovedBy, &ca.ApprovedAt, &ca.CreatedAt, &ca.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("approval not found")
	}
	return &ca, err
}

func (r *SQLiteCommandApprovalRepository) GetByRequestID(ctx context.Context, requestID string) (*models.CommandApproval, error) {
	var ca models.CommandApproval
	err := r.db.QueryRowContext(ctx, sqCommandApprovalSelect+" WHERE request_id = ?", requestID).Scan(&ca.ID, &ca.RequestID, &ca.ApprovalStatus, &ca.ApprovedBy, &ca.ApprovedAt, &ca.CreatedAt, &ca.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("approval not found")
	}
	return &ca, err
}

func (r *SQLiteCommandApprovalRepository) List(ctx context.Context, offset, limit int) ([]*models.CommandApproval, int, error) {
	var total int
	r.db.QueryRowContext(ctx, sqCommandApprovalCount).Scan(&total)

	rows, err := r.db.QueryContext(ctx, sqCommandApprovalSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var cas []*models.CommandApproval
	for rows.Next() {
		var ca models.CommandApproval
		rows.Scan(&ca.ID, &ca.RequestID, &ca.ApprovalStatus, &ca.ApprovedBy, &ca.ApprovedAt, &ca.CreatedAt, &ca.UpdatedAt)
		cas = append(cas, &ca)
	}
	return cas, total, rows.Err()
}

func (r *SQLiteCommandApprovalRepository) Update(ctx context.Context, ca *models.CommandApproval) error {
	ca.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, sqCommandApprovalUpdate,
		ca.RequestID, ca.ApprovalStatus, ca.ApprovedBy, ca.ApprovedAt, ca.UpdatedAt, ca.ID)
	return err
}

func (r *SQLiteCommandApprovalRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, sqCommandApprovalDelete, id)
	return err
}
