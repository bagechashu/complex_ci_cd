package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/op/release-control/internal/models"
)

type ApplicationRepository struct {
	db *sql.DB
}

func NewApplicationRepository(db *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(app *models.Application) (*models.Application, error) {
	result, err := r.db.Exec(
		"INSERT INTO application (name, repo, build_type, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		app.Name, app.Repo, app.BuildType, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get lastInsertId: %w", err)
	}

	app.ID = int(id)
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()
	return app, nil
}

func (r *ApplicationRepository) GetByID(id int) (*models.Application, error) {
	app := &models.Application{}
	err := r.db.QueryRow(
		"SELECT id, name, repo, build_type, created_at, updated_at FROM application WHERE id = ?",
		id,
	).Scan(&app.ID, &app.Name, &app.Repo, &app.BuildType, &app.CreatedAt, &app.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("application not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return app, nil
}

func (r *ApplicationRepository) List(limit int, offset int) ([]*models.Application, error) {
	rows, err := r.db.Query("SELECT id, name, repo, build_type, created_at, updated_at FROM application ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	defer rows.Close()

	var apps []*models.Application
	for rows.Next() {
		app := &models.Application{}
		err := rows.Scan(&app.ID, &app.Name, &app.Repo, &app.BuildType, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application: %w", err)
		}
		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating applications: %w", err)
	}

	return apps, nil
}

func (r *ApplicationRepository) Update(app *models.Application) error {
	app.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		"UPDATE application SET name = ?, repo = ?, build_type = ?, updated_at = ? WHERE id = ?",
		app.Name, app.Repo, app.BuildType, app.UpdatedAt, app.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}
	return nil
}

func (r *ApplicationRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM application WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	return nil
}
