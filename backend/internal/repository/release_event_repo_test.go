//go:build unit

package repository

import (
	"context"
	"database/sql"
	"testing"

	"built-and-deploy/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReleaseEventRepository_Create tests the Create method
func TestReleaseEventRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		event   *models.ReleaseEvent
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: &models.ReleaseEvent{
				ReleaseID: 1,
				Type:      "STARTED",
				Message:   "Release started",
				Details:   "{}",
			},
			wantErr: false,
		},
		{
			name: "missing release id",
			event: &models.ReleaseEvent{
				ReleaseID: 0,
				Type:      "STARTED",
				Message:   "Release started",
			},
			wantErr: true,
			errMsg:  "invalid release id",
		},
		{
			name: "missing event type",
			event: &models.ReleaseEvent{
				ReleaseID: 1,
				Type:      "",
				Message:   "Release started",
			},
			wantErr: true,
			errMsg:  "event type is required",
		},
		{
			name:    "nil event",
			event:   nil,
			wantErr: true,
			errMsg:  "cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			repo := NewSQLiteReleaseEventRepository(db)
			err := repo.Create(context.Background(), tt.event)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.event.ID)
				assert.NotZero(t, tt.event.CreatedAt)
			}
		})
	}
}

// TestReleaseEventRepository_ListByRelease tests the ListByRelease method
func TestReleaseEventRepository_ListByRelease(t *testing.T) {
	tests := []struct {
		name      string
		releaseID int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "existing release with events",
			releaseID: 1,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "release with no events",
			releaseID: 999,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "invalid release id",
			releaseID: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			repo := NewSQLiteReleaseEventRepository(db)

			// Seed test data
			if tt.releaseID == 1 {
				_ = repo.Create(context.Background(), &models.ReleaseEvent{
					ReleaseID: 1,
					Type:      "STARTED",
					Message:   "Release started",
				})
				_ = repo.Create(context.Background(), &models.ReleaseEvent{
					ReleaseID: 1,
					Type:      "COMPLETED",
					Message:   "Release completed",
				})
			}

			events, err := repo.ListByRelease(context.Background(), tt.releaseID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(events))
			}
		})
	}
}

// Helper function to set up test database
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create schema
	schema := `
		CREATE TABLE IF NOT EXISTS release_event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			release_id INTEGER NOT NULL,
			type TEXT NOT NULL,
			message TEXT,
			details TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

// Benchmark
func BenchmarkReleaseEventRepository_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	repo := NewSQLiteReleaseEventRepository(db)
	event := &models.ReleaseEvent{
		ReleaseID: 1,
		Type:      "STARTED",
		Message:   "Release started",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event.ID = 0 // Reset ID for next iteration
		_ = repo.Create(context.Background(), event)
	}
}
