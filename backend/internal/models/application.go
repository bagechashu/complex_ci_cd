package models

import (
	"fmt"
	"time"
)

// Application 表示要部署的应用程序单元
//
// Application 是系统中的核心概念，代表一个可以部署的独立服务或应用。
// 它包含应用的基本信息（名称、镜像、代码仓库等）以及元数据。
//
// 字段说明：
//   - ID: 数据库主键
//   - Name: 应用名称，全局唯一，用于标识应用
//   - ImageName: Docker 镜像名称，用于容器化部署
//   - Owner: 应用负责人，用于权限控制和责任追踪
//   - GitRepo: 应用的 Git 仓库地址（可选）
//   - BuildType: 构建类型（Docker、JAR、Go Binary 等）
//   - Description: 应用的详细描述
//   - CreatedAt: 应用创建时间
//   - UpdatedAt: 应用最后更新时间
//
// 示例：
//	app := &Application{
//		Name:      "api-service",
//		ImageName: "api-service:latest",
//		Owner:     "backend-team",
//		GitRepo:   ptr("https://github.com/example/api-service"),
//		BuildType: ptr("Docker"),
//	}
type Application struct {
	ID          int        `db:"id" json:"id" example:"1"`
	Name        string     `db:"name" json:"name" example:"api-service" binding:"required"`
	ImageName   string     `db:"image_name" json:"image_name" example:"api-service:v1.0.0" binding:"required"`
	Owner       string     `db:"owner" json:"owner" example:"backend-team" binding:"required"`
	GitRepo     *string    `db:"git_repo" json:"git_repo" example:"https://github.com/example/api-service"`
	BuildType   *string    `db:"build_type" json:"build_type" example:"Docker"`
	Description *string    `db:"description" json:"description" example:"RESTful API service for order management"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at" example:"2026-04-21T10:00:00Z"`
}

// Validate 验证 Application 实例的必填字段
//
// 检查以下内容：
//   - Name 不能为空
//   - ImageName 不能为空
//   - Owner 不能为空
//
// 返回非 nil 的 error 表示验证失败
func (a *Application) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("application name is required")
	}
	if a.ImageName == "" {
		return fmt.Errorf("image_name is required")
	}
	if a.Owner == "" {
		return fmt.Errorf("owner is required")
	}
	return nil
}

func (a *Application) GetID() interface{} {
	return a.ID
}
