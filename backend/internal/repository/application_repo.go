package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/op/release-control/internal/models"
)

type ApplicationRepository interface {
	Create(ctx context.Context, app *models.Application) error
	GetByID(ctx context.Context, id int) (*models.Application, error)
	List(ctx context.Context, offset, limit int) ([]*models.Application, int, error)
	Update(ctx context.Context, app *models.Application) error
	Delete(ctx context.Context, id int) error
}

type SQLiteApplicationRepository struct {
	db *sql.DB
}

func NewSQLiteApplicationRepository(db *sql.DB) ApplicationRepository {
	return &SQLiteApplicationRepository{db: db}
}

func (r *SQLiteApplicationRepository) Create(ctx context.Context, app *models.Application) error {
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO application (name, image_name, git_repo, build_type, description, created_at, updated_at) VALUES (?,?,?,?,?,?,?)",
		app.Name, app.ImageName, app.GitRepo, app.BuildType, app.Description, app.CreatedAt, app.UpdatedAt)
	return err
}

func (r *SQLiteApplicationRepository) GetByID(ctx context.Context, id int) (*models.Application, error) {
	var app models.Application
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, image_name, git_repo, build_type, description, created_at, updated_at FROM application WHERE id = ?",
		id).Scan(&app.ID, &app.Name, &app.ImageName, &app.GitRepo, &app.BuildType, &app.Description, &app.CreatedAt, &app.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("application not found")
	}
	return &app, err
}

func (r *SQLiteApplicationRepository) List(ctx context.Context, offset, limit int) ([]*models.Application, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM application").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, image_name, git_repo, build_type, description, created_at, updated_at FROM application ORDER BY created_at DESC LIMIT ? OFFSET ?",
		limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var apps []*models.Application
	for rows.Next() {
		var app models.Application
		err := rows.Scan(&app.ID, &app.Name, &app.ImageName, &app.GitRepo, &app.BuildType, &app.Description, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		apps = append(apps, &app)
	}
	return apps, total, rows.Err()
}

func (r *SQLiteApplicationRepository) Update(ctx context.Context, app *models.Application) error {
	app.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		"UPDATE application SET name=?, image_name=?, git_repo=?, build_type=?, description=?, updated_at=? WHERE id=?",
		app.Name, app.ImageName, app.GitRepo, app.BuildType, app.Description, app.UpdatedAt, app.ID)
	return err
}

func (r *SQLiteApplicationRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM application WHERE id=?", id)
	return err
}
