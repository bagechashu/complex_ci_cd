package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Init initializes the database connection with connection pooling and pragmas
func Init(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Configure connection pool for SQLite
	db.SetMaxOpenConns(25)                 // maximum open connections
	db.SetMaxIdleConns(5)                  // idle connections to keep open
	db.SetConnMaxLifetime(5 * time.Minute) // force recycle connections
	db.SetConnMaxIdleTime(10 * time.Second) // close idle connections after 10s

	// Set SQLite pragmas for performance
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = 10000",
		"PRAGMA foreign_keys = ON",
		"PRAGMA temp_store = MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return nil, fmt.Errorf("setting pragma: %w", err)
		}
	}

	// Test connection with timeout
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Initialize schema with version management
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("creating tables: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
