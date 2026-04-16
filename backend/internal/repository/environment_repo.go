package repository

import (
	"database/sql"
	"fmt"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqEnvironmentInsert = "INSERT INTO environment (name, rank, created_at, updated_at) VALUES (?, ?, ?, ?)"
	sqEnvironmentSelect = "SELECT id, name, rank, created_at, updated_at FROM environment"
	sqEnvironmentUpdate = "UPDATE environment SET name = ?, rank = ?, updated_at = ? WHERE id = ?"
	sqEnvironmentDelete = "DELETE FROM environment WHERE id = ?"
)

type EnvironmentRepository struct {
	db *sql.DB
}

func NewEnvironmentRepository(db *sql.DB) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

func (r *EnvironmentRepository) Create(env *models.Environment) (*models.Environment, error) {
	result, err := r.db.Exec(
		sqEnvironmentInsert,
		env.Name, env.Rank, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get lastInsertId: %w", err)
	}

	env.ID = int(id)
	env.CreatedAt = time.Now()
	env.UpdatedAt = time.Now()
	return env, nil
}

func (r *EnvironmentRepository) GetByID(id int) (*models.Environment, error) {
	env := &models.Environment{}
	err := r.db.QueryRow(
		sqEnvironmentSelect+" WHERE id = ?",
		id,
	).Scan(&env.ID, &env.Name, &env.Rank, &env.CreatedAt, &env.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("environment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}

	return env, nil
}

func (r *EnvironmentRepository) List(limit int, offset int) ([]*models.Environment, error) {
	rows, err := r.db.Query(sqEnvironmentSelect+" ORDER BY rank ASC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}
	defer rows.Close()

	var envs []*models.Environment
	for rows.Next() {
		env := &models.Environment{}
		err := rows.Scan(&env.ID, &env.Name, &env.Rank, &env.CreatedAt, &env.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		envs = append(envs, env)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environments: %w", err)
	}

	return envs, nil
}

func (r *EnvironmentRepository) Update(env *models.Environment) error {
	env.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		sqEnvironmentUpdate,
		env.Name, env.Rank, env.UpdatedAt, env.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update environment: %w", err)
	}
	return nil
}

func (r *EnvironmentRepository) Delete(id int) error {
	_, err := r.db.Exec(sqEnvironmentDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}
	return nil
}
