package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqShellTaskExecutionInsert = "INSERT INTO shell_task_execution (task_id, server_id, command_id, status, output, error_message, command_params, exit_code, started_at, completed_at, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)"
	sqShellTaskExecutionSelect = "SELECT id, task_id, server_id, command_id, status, output, error_message, command_params, exit_code, started_at, completed_at, created_at, updated_at FROM shell_task_execution"
	// Select with joins to get related names
	sqShellTaskExecutionSelectWithJoins = `
		SELECT 
		  e.id, e.task_id, e.server_id, e.command_id, e.status,
		  e.output, e.error_message, e.command_params, e.exit_code,
		  e.started_at, e.completed_at, e.created_at, e.updated_at,
		  COALESCE(t.name, ''), COALESCE(s.name, ''), COALESCE(c.command, '')
		FROM shell_task_execution e
		LEFT JOIN shell_task t ON e.task_id = t.id
		LEFT JOIN shell_server s ON e.server_id = s.id
		LEFT JOIN shell_command c ON e.command_id = c.id
	`
	sqShellTaskExecutionUpdate = "UPDATE shell_task_execution SET status=?, output=?, error_message=?, exit_code=?, started_at=?, completed_at=?, updated_at=? WHERE id=?"
	sqShellTaskExecutionDelete = "DELETE FROM shell_task_execution WHERE id=?"
	sqShellTaskExecutionCount  = "SELECT COUNT(*) FROM shell_task_execution"
	
	sqShellTaskExecutionCountByTask    = "SELECT COUNT(*) FROM shell_task_execution WHERE task_id = ?"
	sqShellTaskExecutionSelectByTask   = sqShellTaskExecutionSelectWithJoins + " WHERE e.task_id = ? ORDER BY e.created_at DESC LIMIT ? OFFSET ?"
	sqShellTaskExecutionCountByServer  = "SELECT COUNT(*) FROM shell_task_execution WHERE server_id = ?"
	sqShellTaskExecutionSelectByServer = sqShellTaskExecutionSelectWithJoins + " WHERE e.server_id = ? ORDER BY e.created_at DESC LIMIT ? OFFSET ?"
	sqShellTaskExecutionCountByCommand = "SELECT COUNT(*) FROM shell_task_execution WHERE command_id = ?"
	sqShellTaskExecutionSelectByCommand = sqShellTaskExecutionSelectWithJoins + " WHERE e.command_id = ? ORDER BY e.created_at DESC LIMIT ? OFFSET ?"
	sqShellTaskExecutionSelectLatest   = sqShellTaskExecutionSelectWithJoins + " WHERE e.task_id = ? AND e.server_id = ? ORDER BY e.created_at DESC LIMIT 1"
)

type ShellTaskExecutionRepository interface {
	Create(ctx context.Context, execution *models.ShellTaskExecution) error
	GetByID(ctx context.Context, id int) (*models.ShellTaskExecution, error)
	List(ctx context.Context, offset, limit int) ([]*models.ShellTaskExecution, int, error)
	ListByTask(ctx context.Context, taskID int, offset, limit int) ([]*models.ShellTaskExecution, int, error)
	ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellTaskExecution, int, error)
	ListByCommand(ctx context.Context, commandID int, offset, limit int) ([]*models.ShellTaskExecution, int, error)
	Update(ctx context.Context, execution *models.ShellTaskExecution) error
	Delete(ctx context.Context, id int) error
	GetLatestByTaskAndServer(ctx context.Context, taskID, serverID int) (*models.ShellTaskExecution, error)
}

type SQLiteShellTaskExecutionRepository struct {
	db *sql.DB
}

func NewSQLiteShellTaskExecutionRepository(db *sql.DB) ShellTaskExecutionRepository {
	return &SQLiteShellTaskExecutionRepository{db: db}
}

func (r *SQLiteShellTaskExecutionRepository) Create(ctx context.Context, execution *models.ShellTaskExecution) error {
	now := time.Now()
	execution.CreatedAt = now
	execution.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, sqShellTaskExecutionInsert,
		execution.TaskID, execution.ServerID, execution.CommandID, execution.Status,
		execution.Output, execution.ErrorMessage, execution.CommandParams, execution.ExitCode,
		execution.StartedAt, execution.CompletedAt, execution.CreatedAt, execution.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	execution.ID = int(id)
	return nil
}

func (r *SQLiteShellTaskExecutionRepository) GetByID(ctx context.Context, id int) (*models.ShellTaskExecution, error) {
	var e models.ShellTaskExecution
	err := r.db.QueryRowContext(ctx, sqShellTaskExecutionSelectWithJoins+" WHERE e.id = ?", id).Scan(
		&e.ID, &e.TaskID, &e.ServerID, &e.CommandID, &e.Status,
		&e.Output, &e.ErrorMessage, &e.CommandParams, &e.ExitCode,
		&e.StartedAt, &e.CompletedAt, &e.CreatedAt, &e.UpdatedAt,
		&e.TaskName, &e.ServerName, &e.Command)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("shell task execution not found")
	}
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (r *SQLiteShellTaskExecutionRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellTaskExecution, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellTaskExecutionCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellTaskExecutionSelectWithJoins+" ORDER BY e.created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanExecutionsWithJoins(rows, total)
}

func (r *SQLiteShellTaskExecutionRepository) ListByTask(ctx context.Context, taskID int, offset, limit int) ([]*models.ShellTaskExecution, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellTaskExecutionCountByTask, taskID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellTaskExecutionSelectByTask, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanExecutionsWithJoins(rows, total)
}

func (r *SQLiteShellTaskExecutionRepository) ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellTaskExecution, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellTaskExecutionCountByServer, serverID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellTaskExecutionSelectByServer, serverID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanExecutionsWithJoins(rows, total)
}

func (r *SQLiteShellTaskExecutionRepository) ListByCommand(ctx context.Context, commandID int, offset, limit int) ([]*models.ShellTaskExecution, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellTaskExecutionCountByCommand, commandID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellTaskExecutionSelectByCommand, commandID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanExecutionsWithJoins(rows, total)
}

func (r *SQLiteShellTaskExecutionRepository) Update(ctx context.Context, execution *models.ShellTaskExecution) error {
	execution.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, sqShellTaskExecutionUpdate,
		execution.Status, execution.Output, execution.ErrorMessage, execution.ExitCode,
		execution.StartedAt, execution.CompletedAt, execution.UpdatedAt, execution.ID)
	return err
}

func (r *SQLiteShellTaskExecutionRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqShellTaskExecutionDelete, id)
	return err
}

func (r *SQLiteShellTaskExecutionRepository) GetLatestByTaskAndServer(ctx context.Context, taskID, serverID int) (*models.ShellTaskExecution, error) {
	var e models.ShellTaskExecution
	err := r.db.QueryRowContext(ctx, sqShellTaskExecutionSelectLatest, taskID, serverID).Scan(
		&e.ID, &e.TaskID, &e.ServerID, &e.CommandID, &e.Status,
		&e.Output, &e.ErrorMessage, &e.CommandParams, &e.ExitCode,
		&e.StartedAt, &e.CompletedAt, &e.CreatedAt, &e.UpdatedAt,
		&e.TaskName, &e.ServerName, &e.Command)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func scanExecutions(rows *sql.Rows, total int) ([]*models.ShellTaskExecution, int, error) {
	var executions []*models.ShellTaskExecution
	for rows.Next() {
		var e models.ShellTaskExecution
		if err := rows.Scan(
			&e.ID, &e.TaskID, &e.ServerID, &e.CommandID, &e.Status,
			&e.Output, &e.ErrorMessage, &e.CommandParams, &e.ExitCode,
			&e.StartedAt, &e.CompletedAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		executions = append(executions, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

func scanExecutionsWithJoins(rows *sql.Rows, total int) ([]*models.ShellTaskExecution, int, error) {
	var executions []*models.ShellTaskExecution
	for rows.Next() {
		var e models.ShellTaskExecution
		if err := rows.Scan(
			&e.ID, &e.TaskID, &e.ServerID, &e.CommandID, &e.Status,
			&e.Output, &e.ErrorMessage, &e.CommandParams, &e.ExitCode,
			&e.StartedAt, &e.CompletedAt, &e.CreatedAt, &e.UpdatedAt,
			&e.TaskName, &e.ServerName, &e.Command); err != nil {
			return nil, 0, err
		}
		executions = append(executions, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}
