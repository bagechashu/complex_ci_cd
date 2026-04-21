package models

import (
	"fmt"
	"time"
)

// ShellCommand 表示允许在服务器上执行的 Shell 命令
//
// ShellCommand 定义了可以在特定服务器上执行的 Shell 命令。
// 每个命令必须明确指定允许执行的服务器（ServerID）。
// 命令需要通过发布流程（IsPublished）才能被任务引用和执行。
//
// 字段说明：
//   - ID: 命令唯一标识
//   - ServerID: 允许执行该命令的服务器 ID
//   - Command: 实际的 Shell 命令字符串（如 "curl http://health-check"）
//   - Description: 命令的用途描述（为了用户理解）
//   - IsPublished: 是否已发布，只有发布的命令才能在任务中被引用
//   - CreatedAt: 命令创建时间
//   - UpdatedAt: 命令最后修改时间
//   - ServerName: 服务器显示名称（冗余字段，便于 API 返回，不存储在数据库）
//
// 发布状态说明：
//   - false: 草稿状态，只有命令创建者可见，无法被任务使用
//   - true: 已发布状态，所有授权用户可见，可在任务中被引用
//
// 安全建议：
// - 命令应当是幂等的（多次执行结果相同）
// - 避免包含硬编码的敏感信息（密码、密钥等）
// - 使用只读命令优先（如 curl、ps、df 等）
// - 命令执行后应记录日志供审计
//
// 示例：
//	cmd := &ShellCommand{
//		ServerID:    1,
//		Command:     "curl -f http://localhost:8080/health",
//		Description: "Health check for API service",
//		IsPublished: true,
//	}
type ShellCommand struct {
	ID          int       `json:"id" db:"id" example:"1"`
	ServerID    int       `json:"server_id" db:"server_id" example:"1" binding:"required"`
	Command     string    `json:"command" db:"command" example:"curl http://health-check" binding:"required"`
	Description string    `json:"description" db:"description" example:"Health check endpoint"`
	IsPublished bool      `json:"is_published" db:"is_published" example:"true"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`

	// 关联关系
	ServerName string `json:"server_name" db:"-"` // 用于前端显示
}

func (c *ShellCommand) Validate() error {
	if c.ServerID <= 0 {
		return fmt.Errorf("server_id must be positive")
	}
	if c.Command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	return nil
}

func (c *ShellCommand) GetID() int {
	return c.ID
}
