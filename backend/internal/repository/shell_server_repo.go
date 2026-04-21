package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/pkg/utils"
)

const (
	sqShellServerInsert = "INSERT INTO shell_server (name, host, port, username, auth_type, password, private_key, status, last_connected, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
	sqShellServerSelect = "SELECT id, name, host, port, username, auth_type, password, private_key, status, last_connected, created_at, updated_at FROM shell_server"
	sqShellServerUpdate = "UPDATE shell_server SET name=?, host=?, port=?, username=?, auth_type=?, password=?, private_key=?, status=?, last_connected=?, updated_at=? WHERE id=?"
	sqShellServerDelete = "DELETE FROM shell_server WHERE id=?"
	sqShellServerCount  = "SELECT COUNT(*) FROM shell_server"
)

type ShellServerRepository interface {
	Create(ctx context.Context, server *models.ShellServer) error
	GetByID(ctx context.Context, id int) (*models.ShellServer, error)
	List(ctx context.Context, offset, limit int) ([]*models.ShellServer, int, error)
	Update(ctx context.Context, server *models.ShellServer) error
	Delete(ctx context.Context, id int) error
	GetByName(ctx context.Context, name string) (*models.ShellServer, error)
}

type SQLiteShellServerRepository struct {
	db            *sql.DB
	encryptionKey string
}

func NewSQLiteShellServerRepository(db *sql.DB, encryptionKey string) ShellServerRepository {
	return &SQLiteShellServerRepository{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

func (r *SQLiteShellServerRepository) Create(ctx context.Context, server *models.ShellServer) error {
	if err := server.Validate(); err != nil {
		return err
	}

	now := time.Now()
	server.CreatedAt = now
	server.UpdatedAt = now

	// Encrypt sensitive information
	var encryptedPassword *string
	var encryptedPrivateKey *string

	if server.Password != "" {
		ciphertext, err := utils.EncryptAES(server.Password, r.encryptionKey)
		if err != nil {
			return err
		}
		encryptedPassword = &ciphertext
	}

	if server.PrivateKey != "" {
		ciphertext, err := utils.EncryptAES(server.PrivateKey, r.encryptionKey)
		if err != nil {
			return err
		}
		encryptedPrivateKey = &ciphertext
	}

	_, err := r.db.ExecContext(ctx, sqShellServerInsert,
		server.Name, server.Host, server.Port, server.Username,
		server.AuthType, encryptedPassword, encryptedPrivateKey,
		server.Status, server.LastConnected, server.CreatedAt, server.UpdatedAt)
	return err
}

func (r *SQLiteShellServerRepository) GetByID(ctx context.Context, id int) (*models.ShellServer, error) {
	var s models.ShellServer
	var encryptedPassword *string
	var encryptedPrivateKey *string

	err := r.db.QueryRowContext(ctx, sqShellServerSelect+" WHERE id = ?", id).Scan(
		&s.ID, &s.Name, &s.Host, &s.Port, &s.Username,
		&s.AuthType, &encryptedPassword, &encryptedPrivateKey,
		&s.Status, &s.LastConnected, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("shell server not found")
	}
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive information
	if encryptedPassword != nil && *encryptedPassword != "" {
		plaintext, err := utils.DecryptAES(*encryptedPassword, r.encryptionKey)
		if err != nil {
			return nil, err
		}
		s.Password = plaintext
	}

	if encryptedPrivateKey != nil && *encryptedPrivateKey != "" {
		plaintext, err := utils.DecryptAES(*encryptedPrivateKey, r.encryptionKey)
		if err != nil {
			return nil, err
		}
		s.PrivateKey = plaintext
	}

	return &s, nil
}

func (r *SQLiteShellServerRepository) GetByName(ctx context.Context, name string) (*models.ShellServer, error) {
	var s models.ShellServer
	var encryptedPassword *string
	var encryptedPrivateKey *string

	err := r.db.QueryRowContext(ctx, sqShellServerSelect+" WHERE name = ?", name).Scan(
		&s.ID, &s.Name, &s.Host, &s.Port, &s.Username,
		&s.AuthType, &encryptedPassword, &encryptedPrivateKey,
		&s.Status, &s.LastConnected, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("shell server not found")
	}
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive information
	if encryptedPassword != nil && *encryptedPassword != "" {
		plaintext, err := utils.DecryptAES(*encryptedPassword, r.encryptionKey)
		if err != nil {
			return nil, err
		}
		s.Password = plaintext
	}

	if encryptedPrivateKey != nil && *encryptedPrivateKey != "" {
		plaintext, err := utils.DecryptAES(*encryptedPrivateKey, r.encryptionKey)
		if err != nil {
			return nil, err
		}
		s.PrivateKey = plaintext
	}

	return &s, nil
}

func (r *SQLiteShellServerRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellServer, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, sqShellServerCount).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqShellServerSelect+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var servers []*models.ShellServer
	for rows.Next() {
		var s models.ShellServer
		var encryptedPassword *string
		var encryptedPrivateKey *string

		if err := rows.Scan(
			&s.ID, &s.Name, &s.Host, &s.Port, &s.Username,
			&s.AuthType, &encryptedPassword, &encryptedPrivateKey,
			&s.Status, &s.LastConnected, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}

		// Decrypt sensitive information
		if encryptedPassword != nil && *encryptedPassword != "" {
			plaintext, err := utils.DecryptAES(*encryptedPassword, r.encryptionKey)
			if err != nil {
				return nil, 0, err
			}
			s.Password = plaintext
		}

		if encryptedPrivateKey != nil && *encryptedPrivateKey != "" {
			plaintext, err := utils.DecryptAES(*encryptedPrivateKey, r.encryptionKey)
			if err != nil {
				return nil, 0, err
			}
			s.PrivateKey = plaintext
		}

		servers = append(servers, &s)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return servers, total, nil
}

func (r *SQLiteShellServerRepository) Update(ctx context.Context, server *models.ShellServer) error {
	if err := server.Validate(); err != nil {
		return err
	}

	server.UpdatedAt = time.Now()

	// Encrypt sensitive information
	var encryptedPassword *string
	var encryptedPrivateKey *string

	if server.Password != "" {
		ciphertext, err := utils.EncryptAES(server.Password, r.encryptionKey)
		if err != nil {
			return err
		}
		encryptedPassword = &ciphertext
	}

	if server.PrivateKey != "" {
		ciphertext, err := utils.EncryptAES(server.PrivateKey, r.encryptionKey)
		if err != nil {
			return err
		}
		encryptedPrivateKey = &ciphertext
	}

	_, err := r.db.ExecContext(ctx, sqShellServerUpdate,
		server.Name, server.Host, server.Port, server.Username,
		server.AuthType, encryptedPassword, encryptedPrivateKey,
		server.Status, server.LastConnected, server.UpdatedAt, server.ID)
	return err
}

func (r *SQLiteShellServerRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, sqShellServerDelete, id)
	return err
}
