// Package repository provides data persistence abstractions and implementations.
//
// This package defines interfaces for all domain entities and provides SQLite
// implementations of those interfaces. Each entity has:
//   - Interface definition for abstraction
//   - SQLite implementation using database/sql
//   - Standard CRUD operations (Create, Read, Update, Delete)
//   - List operations with pagination
//   - Search operations where applicable
//
// Key repositories:
//   - ApplicationRepository: Application management
//   - ClusterRepository: Cluster management
//   - ReleaseRecordRepository: Release tracking
//   - ReleaseEventRepository: Release event audit trail
//   - ShellServerRepository: Target servers for command execution
//   - ShellCommandRepository: Shell command templates
//   - ShellCommandExecutionRepository: Shell command execution tracking
//   - WorkloadTargetRepository: Workload deployment targets
//   - EnvironmentRepository: Environment configurations
//
// All operations accept context.Context for cancellation and timeout support.
//
// See:
//   - ApplicationRepository for application CRUD operations
//   - ClusterRepository for cluster management
//   - ReleaseRecordRepository for deployment records
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
)

const (
	sqApplicationInsert = "INSERT INTO application (name, image_name, owner, git_repo, build_type, description, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?)"
	sqApplicationSelect = "SELECT id, name, image_name, owner, git_repo, build_type, description, created_at, updated_at FROM application"
	sqApplicationUpdate = "UPDATE application SET name=?, image_name=?, owner=?, git_repo=?, build_type=?, description=?, updated_at=? WHERE id=?"
	sqApplicationDelete = "DELETE FROM application WHERE id=?"
	sqApplicationCount  = "SELECT COUNT(*) FROM application"
)

// ApplicationRepository defines the data access interface for Application entities.
//
// Applications represent deployable units in the Release Control System.
// This interface abstracts database operations to allow for testing and
// alternative implementations (e.g., PostgreSQL, MongoDB).
//
// All methods accept context.Context for cancellation and timeout support.
// All methods are thread-safe for concurrent access.
type ApplicationRepository interface {
	// Create inserts a new application into persistent storage.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline
	//   - app: Application model to insert
	//
	// Returns:
	//   - error: Non-nil if insert fails (e.g., unique constraint violation)
	//
	// Behavior:
	//   - Sets CreatedAt and UpdatedAt to current time
	//   - Generates ID in the underlying database
	//   - Is idempotent if retry logic is used (transaction must handle uniqueness)
	//
	// Example:
	//
	//	app := &models.Application{
	//	    Name:      "api-service",
	//	    ImageName: "api:v1.0.0",
	//	}
	//	err := repo.Create(ctx, app)
	Create(ctx context.Context, app *models.Application) error

	// GetByID retrieves a single application by its ID.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline
	//   - id: The numeric ID of the application
	//
	// Returns:
	//   - *models.Application: The application if found, nil if not found
	//   - error: Non-nil if query fails (not if app doesn't exist)
	//
	// Note:
	//   - Returns nil application and no error if application doesn't exist
	//   - Only returns error on database/query problems
	//
	// Example:
	//
	//	app, err := repo.GetByID(ctx, 42)
	//	if app == nil {
	//	    // Application doesn't exist
	//	}
	GetByID(ctx context.Context, id int) (*models.Application, error)

	// List retrieves a paginated list of applications.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline
	//   - offset: Starting position (zero-based)
	//   - limit: Maximum number of results
	//
	// Returns:
	//   - []*models.Application: Slice of applications, empty if none exist
	//   - int: Total count of applications in database
	//   - error: Non-nil if query fails
	//
	// Pagination:
	//   - Results ordered by creation time (newest first)
	//   - offset=0, limit=10 returns first 10 results
	//   - offset=10, limit=10 returns items 11-20
	//
	// Example:
	//
	//	apps, total, err := repo.List(ctx, 0, 20)
	//	pageCount := (total + 19) / 20  // Ceiling division
	List(ctx context.Context, offset, limit int) ([]*models.Application, int, error)

	// ListWithSearch retrieves applications matching a search query.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline
	//   - offset: Starting position (zero-based)
	//   - limit: Maximum number of results
	//   - search: Search term to match against name and image_name
	//
	// Returns:
	//   - []*models.Application: Matching applications
	//   - int: Total count of matching applications
	//   - error: Non-nil if query fails
	//
	// Search:
	//   - Searches name and image_name fields using LIKE %pattern%
	//   - Case-insensitive (depends on SQLite collation)
	//   - Empty search string returns all applications
	//
	// Example:
	//
	//	apps, total, err := repo.ListWithSearch(ctx, 0, 20, "payment")
	ListWithSearch(ctx context.Context, offset, limit int, search string) ([]*models.Application, int, error)

	// Update modifies an existing application.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline
	//   - app: Application model with updated values
	//
	// Returns:
	//   - error: Non-nil if update fails
	//
	// Behavior:
	//   - Updates application with matching ID
	//   - Sets UpdatedAt to current time
	//   - Returns error if application doesn't exist (0 rows affected)
	//
	// Example:
	//
	//	app.Name = "new-name"
	//	err := repo.Update(ctx, app)
	Update(ctx context.Context, app *models.Application) error

	// Delete removes an application from persistent storage.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline
	//   - id: The ID of the application to delete
	//
	// Returns:
	//   - error: Non-nil if delete fails
	//
	// Warning:
	//   - This is a hard delete - data cannot be recovered
	//   - Cascading deletes may remove related records
	//
	// Example:
	//
	//	err := repo.Delete(ctx, 42)
	Delete(ctx context.Context, id int) error
}

// SQLiteApplicationRepository implements ApplicationRepository using SQLite.
//
// This implementation uses prepared SQL statements for efficiency and
// prepared statement caching in the sql.DB connection pool.
//
// Concurrency:
//   - All operations are thread-safe
//   - sql.DB handles connection pooling internally
//   - Concurrent queries are serialized by SQLite WAL mode
type SQLiteApplicationRepository struct {
	db *sql.DB
}

// NewSQLiteApplicationRepository creates a new SQLiteApplicationRepository.
//
// Parameters:
//   - db: Configured sql.DB connection to SQLite database
//
// Returns:
//   - ApplicationRepository: The configured repository
//
// Example:
//
//	db, err := sql.Open("sqlite3", "file:test.db")
//	repo := NewSQLiteApplicationRepository(db)
func NewSQLiteApplicationRepository(db *sql.DB) ApplicationRepository {
	return &SQLiteApplicationRepository{db: db}
}

func (r *SQLiteApplicationRepository) Create(ctx context.Context, app *models.Application) error {
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, sqApplicationInsert,
		app.Name, app.ImageName, app.Owner, app.GitRepo, app.BuildType, app.Description, app.CreatedAt, app.UpdatedAt)
	if err != nil {
		return err
	}

	// Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	app.ID = int(id)
	return nil
}

func (r *SQLiteApplicationRepository) GetByID(ctx context.Context, id int) (*models.Application, error) {
	var app models.Application
	err := r.db.QueryRowContext(ctx, sqApplicationSelect+" WHERE id = ?", id).Scan(&app.ID, &app.Name, &app.ImageName, &app.Owner, &app.GitRepo, &app.BuildType, &app.Description, &app.CreatedAt, &app.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("application not found")
	}
	return &app, err
}

func (r *SQLiteApplicationRepository) List(ctx context.Context, offset, limit int) ([]*models.Application, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqApplicationCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqApplicationSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var apps []*models.Application
	for rows.Next() {
		var app models.Application
		err := rows.Scan(&app.ID, &app.Name, &app.ImageName, &app.Owner, &app.GitRepo, &app.BuildType, &app.Description, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		apps = append(apps, &app)
	}
	return apps, total, rows.Err()
}

func (r *SQLiteApplicationRepository) ListWithSearch(ctx context.Context, offset, limit int, search string) ([]*models.Application, int, error) {
	whereClause := ""
	countQuery := sqApplicationCount
	selectQuery := sqApplicationSelect

	if search != "" {
		whereClause = " WHERE name LIKE ? OR image_name LIKE ?"
		countQuery = countQuery + whereClause
		selectQuery = selectQuery + whereClause
	}

	var total int
	if search != "" {
		searchPattern := "%" + search + "%"
		err := r.db.QueryRowContext(ctx, countQuery, searchPattern, searchPattern).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	query := selectQuery + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	var rows *sql.Rows
	var err error

	if search != "" {
		searchPattern := "%" + search + "%"
		rows, err = r.db.QueryContext(ctx, query, searchPattern, searchPattern, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, query, limit, offset)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var apps []*models.Application
	for rows.Next() {
		var app models.Application
		err := rows.Scan(&app.ID, &app.Name, &app.ImageName, &app.Owner, &app.GitRepo, &app.BuildType, &app.Description, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		apps = append(apps, &app)
	}
	return apps, total, rows.Err()
}

func (r *SQLiteApplicationRepository) Update(ctx context.Context, app *models.Application) error {
	app.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, sqApplicationUpdate,
		app.Name, app.ImageName, app.Owner, app.GitRepo, app.BuildType, app.Description, app.UpdatedAt, app.ID)
	return err
}

func (r *SQLiteApplicationRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqApplicationDelete, id)
	return err
}
