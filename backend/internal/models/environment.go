package models

import "time"

// Environment 表示部署环境如开发、测试、预发布、生产等
//
// Environment 定义了应用可以部署的不同阶段和目标环境。
// 使用 Rank 字段定义环境的发布顺序和优先级：
//   1 = development (开发环境）
//   2 = testing (测试环境)
//   3 = staging (预发布环境)
//   4 = production (生产环境)
//
// Rank 值越高，环境越接近真实用户，权限控制也越严格。
// 这个排序用于 UI 展示、权限检查和发布管道流程。
//
// 字段说明：
//   - ID: 环境唯一标识
//   - Name: 环境名称（如 "production", "dev", "staging"）
//   - Rank: 环境优先级/排序，值越高越接近生产
//   - CreatedAt: 环境创建时间
//   - UpdatedAt: 环境最后修改时间
//
// 示例：
//	prodEnv := &Environment{
//		Name: "production",
//		Rank: 4,
//	}
//	devEnv := &Environment{
//		Name: "development",
//		Rank: 1,
//	}
type Environment struct {
	ID        int       `json:"id" db:"id" example:"1"`
	Name      string    `json:"name" db:"name" example:"production" binding:"required"`
	Rank      int       `json:"rank" db:"rank" example:"4" binding:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`
}

type EnvironmentRequest struct {
	Name string `json:"name" binding:"required"`
	Rank int    `json:"rank"`
}
