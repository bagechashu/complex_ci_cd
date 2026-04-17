package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqShellTaskInsert = "INSERT INTO shell_task (name, description, command_id, execution_method, requires_approval, created_at, updated_at) VALUES (?,?,?,?,?,?,?)"
	sqShellTaskSelect = "SELECT id, name, description, command_id, execution_method, requires_approval, created_at, updated_at FROM shell_task"
	sqShellTaskUpdate = "UPDATE shell_task SET name=?, description=?, command_id=?, execution_method=?, requires_approval=?, updated_at=? WHERE id=?"
	sqShellTaskDelete = "DELETE FROM shell_task WHERE id=?"
	sqShellTaskCount  = "SELECT COUNT(*) FROM shell_task"
)

type ShellTaskRepository interface {
	Create(ctx context.Context, task *models.ShellTask) error
	GetByID(ctx context.Context, id int) (*models.ShellTask, error)
	List(ctx context.Context, offset, limit int) ([]*models.ShellTask, int, error)
	Update(ctx context.Context, task *models.ShellTask) error
	Delete(ctx context.Context, id int) error
	GetServers(ctx context.Context, taskID int) ([]int, error)
	SetServers(ctx context.Context, taskID int, serverIDs []int) error
}

type SQLiteShellTaskRepository struct {
	db *sql.DB
}

func NewSQLiteShellTaskRepository(db *sql.DB) ShellTaskRepository {
	return &SQLiteShellTaskRepository{db: db}
}

func (r *SQLiteShellTaskRepository) Create(ctx context.Context, task *models.ShellTask) error {
	if len(task.ServerIDs) == 0 {
		return errors.New("at least one server must be selected")
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, sqShellTaskInsert,
		task.Name, task.Description, task.CommandID, task.ExecutionMethod, task.RequiresApproval, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return err
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = int(taskID)

	// Set server associations
	return r.SetServers(ctx, task.ID, task.ServerIDs)
}

func (r *SQLiteShellTaskRepository) GetByID(ctx context.Context, id int) (*models.ShellTask, error) {
	var t models.ShellTask
	err := r.db.QueryRowContext(ctx, sqShellTaskSelect+" WHERE id = ?", id).Scan(
		&t.ID, &t.Name, &t.Description, &t.CommandID, &t.ExecutionMethod, &t.RequiresApproval, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("shell task not found")
	}
	if err != nil {
		return nil, err
	}

	// Load server associations
	serverIDs, err := r.GetServers(ctx, id)
	if err != nil {
		return nil, err
	}
	t.ServerIDs = serverIDs

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
			&t.ID, &t.Name, &t.Description, &t.CommandID, &t.ExecutionMethod, &t.RequiresApproval, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}

		// Load server associations
		serverIDs, err := r.GetServers(ctx, t.ID)
		if err != nil {
			return nil, 0, err
		}
		t.ServerIDs = serverIDs

		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *SQLiteShellTaskRepository) Update(ctx context.Context, task *models.ShellTask) error {
	if len(task.ServerIDs) == 0 {
		return errors.New("at least one server must be selected")
	}

	task.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, sqShellTaskUpdate,
		task.Name, task.Description, task.CommandID, task.ExecutionMethod, task.RequiresApproval, task.UpdatedAt, task.ID)
	if err != nil {
		return err
	}

	// Update server associations
	return r.SetServers(ctx, task.ID, task.ServerIDs)
}

func (r *SQLiteShellTaskRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqShellTaskDelete, id)
	return err
}

func (r *SQLiteShellTaskRepository) GetServers(ctx context.Context, taskID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT server_id FROM shell_task_server WHERE task_id = ? ORDER BY created_at", taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var serverIDs []int
	for rows.Next() {
		var serverID int
		if err := rows.Scan(&serverID); err != nil {
			return nil, err
		}
		serverIDs = append(serverIDs, serverID)
	}

	return serverIDs, rows.Err()
}

func (r *SQLiteShellTaskRepository) SetServers(ctx context.Context, taskID int, serverIDs []int) error {
	// Delete existing associations
	_, err := r.db.ExecContext(ctx, "DELETE FROM shell_task_server WHERE task_id = ?", taskID)
	if err != nil {
		return err
	}

	// Insert new associations
	for _, serverID := range serverIDs {
		_, err := r.db.ExecContext(ctx,
			"INSERT INTO shell_task_server (task_id, server_id, created_at) VALUES (?,?,?)",
			taskID, serverID, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}
