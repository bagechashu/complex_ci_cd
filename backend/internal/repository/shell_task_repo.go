package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqShellTaskInsert = "INSERT INTO shell_task (name, description, server_id, command_id, requires_approval, created_at, updated_at) VALUES (?,?,?,?,?,?,?)"
	sqShellTaskSelect = "SELECT id, name, description, server_id, command_id, requires_approval, created_at, updated_at FROM shell_task"
	sqShellTaskUpdate = "UPDATE shell_task SET name=?, description=?, server_id=?, command_id=?, requires_approval=?, updated_at=? WHERE id=?"
	sqShellTaskDelete = "DELETE FROM shell_task WHERE id=?"
	sqShellTaskCount  = "SELECT COUNT(*) FROM shell_task"
)

type ShellTaskRepository interface {
	Create(ctx context.Context, task *models.ShellTask) error
	GetByID(ctx context.Context, id int) (*models.ShellTask, error)
	List(ctx context.Context, offset, limit int) ([]*models.ShellTask, int, error)
	Update(ctx context.Context, task *models.ShellTask) error
	Delete(ctx context.Context, id int) error
	ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellTask, int, error)
}

type SQLiteShellTaskRepository struct {
	db *sql.DB
}

func NewSQLiteShellTaskRepository(db *sql.DB) ShellTaskRepository {
	return &SQLiteShellTaskRepository{db: db}
}

func (r *SQLiteShellTaskRepository) Create(ctx context.Context, task *models.ShellTask) error {
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, sqShellTaskInsert,
		task.Name, task.Description, task.ServerID, task.CommandID, task.RequiresApproval, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return err
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = int(taskID)
	return nil
}

func (r *SQLiteShellTaskRepository) GetByID(ctx context.Context, id int) (*models.ShellTask, error) {
	var t models.ShellTask
	err := r.db.QueryRowContext(ctx, sqShellTaskSelect+" WHERE id = ?", id).Scan(
		&t.ID, &t.Name, &t.Description, &t.ServerID, &t.CommandID, &t.RequiresApproval, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("shell task not found")
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *SQLiteShellTaskRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellTask, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellTaskCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellTaskSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*models.ShellTask
	for rows.Next() {
		var t models.ShellTask
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.ServerID, &t.CommandID, &t.RequiresApproval, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}

		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *SQLiteShellTaskRepository) Update(ctx context.Context, task *models.ShellTask) error {
	task.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, sqShellTaskUpdate,
		task.Name, task.Description, task.ServerID, task.CommandID, task.RequiresApproval, task.UpdatedAt, task.ID)
	return err
}

func (r *SQLiteShellTaskRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqShellTaskDelete, id)
	return err
}

// ListByServer 返回指定服务器上的所有 shell 任务
func (r *SQLiteShellTaskRepository) ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellTask, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shell_task WHERE server_id = ?", serverID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		sqShellTaskSelect+" WHERE server_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		serverID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*models.ShellTask
	for rows.Next() {
		var t models.ShellTask
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.ServerID, &t.CommandID, &t.RequiresApproval, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}

		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}
