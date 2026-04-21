package models

import (
	"fmt"
	"time"
)

// ReleaseRecord 表示一次应用发布的完整记录
// 记录了应用部署的全生命周期：pending → deploying → completed/failed
type ReleaseRecord struct {
	ID            int        `json:"id" db:"id" example:"1"`
	AppID         int        `json:"app_id" db:"app_id" example:"1" binding:"required"`
	EnvID         int        `json:"env_id" db:"env_id" example:"1" binding:"required"`
	ClusterID     int        `json:"cluster_id" db:"cluster_id" example:"1" binding:"required"`
	Image         string     `json:"image" db:"image" example:"api-service:v1.0.0" binding:"required"`
	Status        string     `json:"status" db:"status" example:"completed" enum:"pending,deploying,completed,failed,rolling_back,rolled_back"`
	PreviousImage *string    `json:"previous_image" db:"previous_image" example:"api-service:v0.9.9"`
	ErrorMsg      *string    `json:"error_msg" db:"error_msg"`
	TriggeredBy   string     `json:"triggered_by" db:"triggered_by" example:"admin" binding:"required"`
	StartedAt     *time.Time `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`
}

// Validate 验证 ReleaseRecord 的必填字段
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
