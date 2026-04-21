package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqShellCommandInsert = "INSERT INTO shell_command (server_id, command, description, is_published, created_at, updated_at) VALUES (?,?,?,?,?,?)"
	sqShellCommandSelect = "SELECT id, server_id, command, description, is_published, created_at, updated_at FROM shell_command"
	sqShellCommandUpdate = "UPDATE shell_command SET command=?, description=?, is_published=?, updated_at=? WHERE id=?"
	sqShellCommandDelete = "DELETE FROM shell_command WHERE id=?"
	sqShellCommandCount  = "SELECT COUNT(*) FROM shell_command"
)

type ShellCommandRepository interface {
	Create(ctx context.Context, command *models.ShellCommand) error
	GetByID(ctx context.Context, id int) (*models.ShellCommand, error)
	List(ctx context.Context, offset, limit int) ([]*models.ShellCommand, int, error)
	ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellCommand, int, error)
	Update(ctx context.Context, command *models.ShellCommand) error
	Delete(ctx context.Context, id int) error
	Publish(ctx context.Context, id int) error
	Unpublish(ctx context.Context, id int) error
}

type SQLiteShellCommandRepository struct {
	db *sql.DB
}

func NewSQLiteShellCommandRepository(db *sql.DB) ShellCommandRepository {
	return &SQLiteShellCommandRepository{db: db}
}

func (r *SQLiteShellCommandRepository) Create(ctx context.Context, command *models.ShellCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}

	now := time.Now()
	command.CreatedAt = now
	command.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, sqShellCommandInsert,
		command.ServerID, command.Command, command.Description, command.IsPublished, command.CreatedAt, command.UpdatedAt)
	return err
}

func (r *SQLiteShellCommandRepository) GetByID(ctx context.Context, id int) (*models.ShellCommand, error) {
	var c models.ShellCommand
	err := r.db.QueryRowContext(ctx, sqShellCommandSelect+" WHERE id = ?", id).Scan(
		&c.ID, &c.ServerID, &c.Command, &c.Description, &c.IsPublished, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("shell command not found")
	}
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *SQLiteShellCommandRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellCommand, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellCommandCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellCommandSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var commands []*models.ShellCommand
	for rows.Next() {
		var c models.ShellCommand
		if err := rows.Scan(
			&c.ID, &c.ServerID, &c.Command, &c.Description, &c.IsPublished, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		commands = append(commands, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return commands, total, nil
}

func (r *SQLiteShellCommandRepository) ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellCommand, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shell_command WHERE server_id = ?", serverID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellCommandSelect+" WHERE server_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", serverID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var commands []*models.ShellCommand
	for rows.Next() {
		var c models.ShellCommand
		if err := rows.Scan(
			&c.ID, &c.ServerID, &c.Command, &c.Description, &c.IsPublished, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		commands = append(commands, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return commands, total, nil
}

func (r *SQLiteShellCommandRepository) Update(ctx context.Context, command *models.ShellCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}

	command.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, sqShellCommandUpdate,
		command.Command, command.Description, command.IsPublished, command.UpdatedAt, command.ID)
	return err
}

func (r *SQLiteShellCommandRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqShellCommandDelete, id)
	return err
}

func (r *SQLiteShellCommandRepository) Publish(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE shell_command SET is_published=1, updated_at=? WHERE id=?",
		time.Now(), id)
	return err
}

func (r *SQLiteShellCommandRepository) Unpublish(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE shell_command SET is_published=0, updated_at=? WHERE id=?",
		time.Now(), id)
	return err
}
