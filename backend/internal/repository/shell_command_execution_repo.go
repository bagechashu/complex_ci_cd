package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqShellCommandExecutionInsert = "INSERT INTO shell_command_execution (server_id, command_id, status, output, error_message, command_params, exit_code, started_at, completed_at, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)"
	sqShellCommandExecutionSelect = "SELECT id, server_id, command_id, status, output, error_message, command_params, exit_code, started_at, completed_at, created_at, updated_at FROM shell_command_execution"
	// Select with joins to get related names
	sqShellCommandExecutionSelectWithJoins = `
		SELECT 
		  e.id, e.server_id, e.command_id, e.status,
		  e.output, e.error_message, e.command_params, e.exit_code,
		  e.started_at, e.completed_at, e.created_at, e.updated_at,
		  COALESCE(t.name, ''), COALESCE(s.name, ''), COALESCE(c.command, '')
		FROM shell_command_execution e
		LEFT JOIN shell_server s ON e.server_id = s.id
		LEFT JOIN shell_command c ON e.command_id = c.id
	`
	sqShellCommandExecutionUpdate = "UPDATE shell_command_execution SET status=?, output=?, error_message=?, exit_code=?, started_at=?, completed_at=?, updated_at=? WHERE id=?"
	sqShellCommandExecutionDelete = "DELETE FROM shell_command_execution WHERE id=?"
	sqShellCommandExecutionCount  = "SELECT COUNT(*) FROM shell_command_execution"

	sqShellCommandExecutionCountByServer   = "SELECT COUNT(*) FROM shell_command_execution WHERE server_id = ?"
	sqShellCommandExecutionSelectByServer  = sqShellCommandExecutionSelectWithJoins + " WHERE e.server_id = ? ORDER BY e.created_at DESC LIMIT ? OFFSET ?"
	sqShellCommandExecutionCountByCommand  = "SELECT COUNT(*) FROM shell_command_execution WHERE command_id = ?"
	sqShellCommandExecutionSelectByCommand = sqShellCommandExecutionSelectWithJoins + " WHERE e.command_id = ? ORDER BY e.created_at DESC LIMIT ? OFFSET ?"
	sqShellCommandExecutionSelectLatest    = sqShellCommandExecutionSelectWithJoins + " WHERE e.server_id = ? ORDER BY e.created_at DESC LIMIT 1"
)

type ShellCommandExecutionRepository interface {
	Create(ctx context.Context, execution *models.ShellCommandExecution) error
	GetByID(ctx context.Context, id int) (*models.ShellCommandExecution, error)
	List(ctx context.Context, offset, limit int) ([]*models.ShellCommandExecution, int, error)
	ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellCommandExecution, int, error)
	ListByCommand(ctx context.Context, commandID int, offset, limit int) ([]*models.ShellCommandExecution, int, error)
	Update(ctx context.Context, execution *models.ShellCommandExecution) error
	Delete(ctx context.Context, id int) error
}

type SQLiteShellCommandExecutionRepository struct {
	db *sql.DB
}

func NewSQLiteShellCommandExecutionRepository(db *sql.DB) ShellCommandExecutionRepository {
	return &SQLiteShellCommandExecutionRepository{db: db}
}

func (r *SQLiteShellCommandExecutionRepository) Create(ctx context.Context, execution *models.ShellCommandExecution) error {
	now := time.Now()
	execution.CreatedAt = now
	execution.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, sqShellCommandExecutionInsert,
		execution.ServerID, execution.CommandID, execution.Status,
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

func (r *SQLiteShellCommandExecutionRepository) GetByID(ctx context.Context, id int) (*models.ShellCommandExecution, error) {
	var e models.ShellCommandExecution
	err := r.db.QueryRowContext(ctx, sqShellCommandExecutionSelectWithJoins+" WHERE e.id = ?", id).Scan(
		&e.ID, &e.ServerID, &e.CommandID, &e.Status,
		&e.Output, &e.ErrorMessage, &e.CommandParams, &e.ExitCode,
		&e.StartedAt, &e.CompletedAt, &e.CreatedAt, &e.UpdatedAt,
		&e.ServerName, &e.Command)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("shell command execution not found")
	}
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (r *SQLiteShellCommandExecutionRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellCommandExecution, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellCommandExecutionCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellCommandExecutionSelectWithJoins+" ORDER BY e.created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanExecutionsWithJoins(rows, total)
}

func (r *SQLiteShellCommandExecutionRepository) ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellCommandExecution, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellCommandExecutionCountByServer, serverID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellCommandExecutionSelectByServer, serverID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanExecutionsWithJoins(rows, total)
}

func (r *SQLiteShellCommandExecutionRepository) ListByCommand(ctx context.Context, commandID int, offset, limit int) ([]*models.ShellCommandExecution, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellCommandExecutionCountByCommand, commandID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellCommandExecutionSelectByCommand, commandID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanExecutionsWithJoins(rows, total)
}

func (r *SQLiteShellCommandExecutionRepository) Update(ctx context.Context, execution *models.ShellCommandExecution) error {
	execution.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, sqShellCommandExecutionUpdate,
		execution.Status, execution.Output, execution.ErrorMessage, execution.ExitCode,
		execution.StartedAt, execution.CompletedAt, execution.UpdatedAt, execution.ID)
	return err
}

func (r *SQLiteShellCommandExecutionRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqShellCommandExecutionDelete, id)
	return err
}

func scanExecutions(rows *sql.Rows, total int) ([]*models.ShellCommandExecution, int, error) {
	var executions []*models.ShellCommandExecution
	for rows.Next() {
		var e models.ShellCommandExecution
		if err := rows.Scan(
			&e.ID, &e.ServerID, &e.CommandID, &e.Status,
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

func scanExecutionsWithJoins(rows *sql.Rows, total int) ([]*models.ShellCommandExecution, int, error) {
	var executions []*models.ShellCommandExecution
	for rows.Next() {
		var e models.ShellCommandExecution
		if err := rows.Scan(
			&e.ID, &e.ServerID, &e.CommandID, &e.Status,
			&e.Output, &e.ErrorMessage, &e.CommandParams, &e.ExitCode,
			&e.StartedAt, &e.CompletedAt, &e.CreatedAt, &e.UpdatedAt,
			&e.ServerName, &e.Command); err != nil {
			return nil, 0, err
		}
		executions = append(executions, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}
