package models

import (
	"fmt"
	"time"
)

type ReleaseRecord struct {
	ID           int        `json:"id" db:"id"`
	AppID        int        `json:"app_id" db:"app_id"`
	EnvID        int        `json:"env_id" db:"env_id"`
	ClusterID    int        `json:"cluster_id" db:"cluster_id"`
	Image        string     `json:"image" db:"image"`
	Status       string     `json:"status" db:"status"`
	PreviousImage *string   `json:"previous_image" db:"previous_image"`
	ErrorMsg     *string    `json:"error_msg" db:"error_msg"`
	TriggeredBy  string     `json:"triggered_by" db:"triggered_by"`
	StartedAt    *time.Time `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

func (r *ReleaseRecord) Validate() error {
	if r.AppID == 0 {
		return fmt.Errorf("app_id cannot be empty")
	}
	if r.ClusterID == 0 {
		return fmt.Errorf("cluster_id cannot be empty")
	}
	return nil
}

func (r *ReleaseRecord) GetID() interface{} {
	return r.ID
}
