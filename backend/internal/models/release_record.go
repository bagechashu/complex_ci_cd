package models

import "time"

const (
	StatusPending    = "pending"
	StatusValidating = "validating"
	StatusDeploying  = "deploying"
	StatusSuccess    = "success"
	StatusFailed     = "failed"
	StatusRolledBack = "rolled_back"
)

type ReleaseRecord struct {
	ID            int       `json:"id"`
	AppID         int       `json:"app_id"`
	EnvID         int       `json:"env_id"`
	ClusterID     int       `json:"cluster_id"`
	Image         string    `json:"image"`
	Status        string    `json:"status"`
	PreviousImage string    `json:"previous_image,omitempty"`
	ErrorMsg      string    `json:"error_msg,omitempty"`
	TriggeredBy   string    `json:"triggered_by,omitempty"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ReleaseRequest struct {
	AppID  int    `json:"app_id" binding:"required"`
	EnvID  int    `json:"env_id" binding:"required"`
	Image  string `json:"image" binding:"required"`
	User   string `json:"user"`
}

type ReleaseEvent struct {
	ID        int       `json:"id"`
	ReleaseID int       `json:"release_id"`
	Type      string    `json:"type"` // started, validating, deploying, success, failed, error
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
