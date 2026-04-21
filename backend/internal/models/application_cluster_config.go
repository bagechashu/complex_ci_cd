package models

import (
	"fmt"
	"time"
)

// ApplicationClusterConfig 表示应用在特定集群上的自定义配置
//
// ApplicationClusterConfig 允许为同一应用在不同集群上设置不同的标签和参数。
// 这是应用-集群绑定关系的扩展，用于集群级别的定制化配置。
//
// 字段说明：
//   - ID: 配置唯一标识（通常格式为 "app-{appID}-cluster-{clusterID}"）
//   - ApplicationID: 应用 ID
//   - ClusterID: 集群 ID
//   - Labels: 标签信息，以逗号分隔的键值对（如 "env=prod,zone=us-west,tier=critical"）
//     标签用于 Kubernetes 选择器、监控分组、权限控制等
//   - CreatedAt: 配置创建时间
//   - UpdatedAt: 配置最后修改时间
//
// 标签用途：
// - Kubernetes: Pod 选择器、资源隔离
// - 监控告警: 按标签分组、定制告警策略
// - 权限控制: 基于标签的访问控制
// - 成本分配: 按标签汇总成本
// - 业务追踪: 标记业务所有者、成本中心等
//
// 示例：
//	config := &ApplicationClusterConfig{
//		ApplicationID: "1",
//		ClusterID:     "1",
//		Labels:        "env=prod,zone=us-west,tier=critical,owner=backend-team",
//	}
type ApplicationClusterConfig struct {
	ID            string    `json:"id" db:"id" example:"app-1-cluster-1"`
	ApplicationID string    `json:"application_id" db:"application_id" example:"1" binding:"required"`
	ClusterID     string    `json:"cluster_id" db:"cluster_id" example:"1" binding:"required"`
	Labels        string    `json:"labels" db:"labels" example:"env=prod,zone=us-west"`
	CreatedAt     time.Time `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`
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
