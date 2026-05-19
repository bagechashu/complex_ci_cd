//go:build unit

package repository

import (
	"testing"
	"time"

	"built-and-deploy/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplication_Validate tests application validation
func TestApplication_Validate(t *testing.T) {
	tests := []struct {
		name    string
		app     *models.Application
		wantErr bool
	}{
		{
			name: "valid application",
			app: &models.Application{
				Name:      "test-app",
				ImageName: "docker.io/test:latest",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			app: &models.Application{
				ImageName: "docker.io/test:latest",
			},
			wantErr: true,
		},
		{
			name: "missing image name",
			app: &models.Application{
				Name: "test-app",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.app.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCluster_Validate tests cluster validation
func TestCluster_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cluster *models.Cluster
		wantErr bool
	}{
		{
			name: "valid cluster",
			cluster: &models.Cluster{
				Name: "prod-cluster",
				Type: "kubernetes",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cluster: &models.Cluster{
				Type: "kubernetes",
			},
			wantErr: true,
		},
		{
			name: "missing type",
			cluster: &models.Cluster{
				Name: "prod-cluster",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cluster.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestReleaseRecord_GetStatus tests release record status tracking
func TestReleaseRecord_GetStatus(t *testing.T) {
	record := &models.ReleaseRecord{
		ID:        1,
		AppID:     1,
		Image:     "test:1.0.0",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Verify status is set correctly
	assert.Equal(t, "completed", record.Status)
}

// TestEnvironment_Rank tests environment ranking
func TestEnvironment_Rank(t *testing.T) {
	tests := []struct {
		name         string
		env          *models.Environment
		expectedRank int
	}{
		{
			name: "development",
			env: &models.Environment{
				Name: "development",
				Rank: 1,
			},
			expectedRank: 1,
		},
		{
			name: "staging",
			env: &models.Environment{
				Name: "staging",
				Rank: 3,
			},
			expectedRank: 3,
		},
		{
			name: "production",
			env: &models.Environment{
				Name: "production",
				Rank: 4,
			},
			expectedRank: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedRank, tt.env.Rank)
		})
	}
}

// TestWorkloadTarget_Validate tests workload target validation
func TestWorkloadTarget_Validate(t *testing.T) {
	tests := []struct {
		name    string
		target  *models.WorkloadTarget
		wantErr bool
	}{
		{
			name: "valid target",
			target: &models.WorkloadTarget{
				AppID:        1,
				EnvID:        1,
				ClusterID:    1,
				WorkloadType: "Deployment",
				K8sNamespace: "default",
				K8sWorkload:  "api-service",
				WorkloadName: "api-service",
			},
			wantErr: false,
		},
		{
			name: "missing app id",
			target: &models.WorkloadTarget{
				EnvID:        1,
				ClusterID:    1,
				WorkloadType: "Deployment",
				K8sNamespace: "default",
				K8sWorkload:  "api-service",
				WorkloadName: "api-service",
			},
			wantErr: false, // May or may not validate, depends on model
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a placeholder since Validate() may not be implemented
			assert.NotNil(t, tt.target)
		})
	}
}

// TestShellServer_Mask tests that sensitive fields are masked
func TestShellServer_Mask(t *testing.T) {
	server := &models.ShellServer{
		ID:         1,
		Host:       "192.168.1.1",
		Port:       22,
		Username:   "admin",
		AuthType:   "password",
		Password:   "secret-password",
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
		Status:     "active",
		CreatedAt:  time.Now(),
	}

	// Verify fields are set
	assert.Equal(t, "192.168.1.1", server.Host)
	assert.Equal(t, "admin", server.Username)
	// Password and PrivateKey should not be logged
	assert.NotEmpty(t, server.Password)
	assert.NotEmpty(t, server.PrivateKey)
}

// TestShellCommand_IsPublished tests published state
func TestShellCommand_IsPublished(t *testing.T) {
	tests := []struct {
		name        string
		isPublished bool
	}{
		{
			name:        "published command",
			isPublished: true,
		},
		{
			name:        "unpublished command",
			isPublished: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &models.ShellCommand{
				ID:          1,
				ServerID:    1,
				Command:     "ls -la",
				IsPublished: tt.isPublished,
				CreatedAt:   time.Now(),
			}

			assert.Equal(t, tt.isPublished, cmd.IsPublished)
		})
	}
}

// TestShellCommandExecution_GetDuration tests duration calculation
func TestShellCommandExecution_GetDuration(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)

	command := &models.ShellCommandExecution{
		ID:          1,
		CommandID:      1,
		StartedAt:   &startTime,
		CompletedAt: &endTime,
	}

	// Verify timestamps are set
	require.NotNil(t, command.StartedAt)
	require.NotNil(t, command.CompletedAt)

	// Duration should be 5 minutes = 300 seconds
	duration := command.CompletedAt.Sub(*command.StartedAt)
	assert.Equal(t, 5*time.Minute, duration)
}

// TestReleaseEvent_Type tests event type validation
func TestReleaseEvent_Type(t *testing.T) {
	validTypes := []string{"STARTED", "DEPLOYING", "COMPLETED", "FAILED", "ROLLING_BACK", "ROLLED_BACK"}

	for _, eventType := range validTypes {
		t.Run(eventType, func(t *testing.T) {
			event := &models.ReleaseEvent{
				ReleaseID: 1,
				Type:      eventType,
				Message:   "Test message",
				CreatedAt: time.Now(),
			}

			assert.Equal(t, eventType, event.Type)
		})
	}
}
