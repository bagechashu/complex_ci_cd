package models

import (
	"fmt"
	"time"
)

type ApplicationClusterConfig struct {
	ID            string    `json:"id" db:"id"`
	ApplicationID string    `json:"application_id" db:"application_id"`
	ClusterID     string    `json:"cluster_id" db:"cluster_id"`
	Labels        string    `json:"labels" db:"labels"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func (a *ApplicationClusterConfig) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if a.ApplicationID == "" {
		return fmt.Errorf("application_id cannot be empty")
	}
	if a.ClusterID == "" {
		return fmt.Errorf("cluster_id cannot be empty")
	}
	return nil
}

func (a *ApplicationClusterConfig) GetID() string {
	return a.ID
}
