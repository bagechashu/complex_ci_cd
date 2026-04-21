//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"built-and-deploy/internal/database"
	"built-and-deploy/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ReleaseEventWorkflow tests a complete release event workflow
func TestIntegration_ReleaseEventWorkflow(t *testing.T) {
	// Setup
	dbPath := "test_integration.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	require.NoError(t, err)
	defer database.Close(db)

	repo := NewSQLiteReleaseEventRepository(db)

	// Create multiple events
	events := []*models.ReleaseEvent{
		{
			ReleaseID: 1,
			Type:      "STARTED",
			Message:   "Release started",
			Details:   "{\"version\": \"1.0.0\"}",
		},
		{
			ReleaseID: 1,
			Type:      "DEPLOYING",
			Message:   "Deploying to cluster A",
			Details:   "{\"cluster\": \"cluster-1\"}",
		},
		{
			ReleaseID: 1,
			Type:      "COMPLETED",
			Message:   "Release completed successfully",
			Details:   "{\"duration\": 120}",
		},
	}

	// Create events
	for _, event := range events {
		err := repo.Create(context.Background(), event)
		assert.NoError(t, err)
		assert.NotZero(t, event.ID)
	}

	// List events
	retrieved, err := repo.ListByRelease(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 3, len(retrieved))

	// Verify order (should be DESC by created_at)
	assert.Equal(t, "COMPLETED", retrieved[0].Type)
	assert.Equal(t, "DEPLOYING", retrieved[1].Type)
	assert.Equal(t, "STARTED", retrieved[2].Type)

	// Verify data
	assert.Equal(t, events[2].Message, retrieved[0].Message)
}

// TestIntegration_MultipleReleases tests events from different releases
func TestIntegration_MultipleReleases(t *testing.T) {
	dbPath := "test_multi_releases.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	require.NoError(t, err)
	defer database.Close(db)

	repo := NewSQLiteReleaseEventRepository(db)

	// Create events for release 1
	for i := 1; i <= 3; i++ {
		err := repo.Create(context.Background(), &models.ReleaseEvent{
			ReleaseID: 1,
			Type:      "STEP",
			Message:   "Step completed",
		})
		assert.NoError(t, err)
	}

	// Create events for release 2
	for i := 1; i <= 2; i++ {
		err := repo.Create(context.Background(), &models.ReleaseEvent{
			ReleaseID: 2,
			Type:      "STEP",
			Message:   "Step completed",
		})
		assert.NoError(t, err)
	}

	// List events for release 1
	events1, err := repo.ListByRelease(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 3, len(events1))

	// List events for release 2
	events2, err := repo.ListByRelease(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(events2))

	// Verify isolation
	for _, e := range events1 {
		assert.Equal(t, 1, e.ReleaseID)
	}
	for _, e := range events2 {
		assert.Equal(t, 2, e.ReleaseID)
	}
}
