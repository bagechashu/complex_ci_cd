package models

import (
	"fmt"
	"time"
)

// ShellTask 表示在特定服务器上执行预设命令的一次操作
//
// ShellTask 是对预设命令在某个特定服务器上的执行操作的记录。
// 每个任务绑定到一个服务器和该服务器上的一个预设命令。
// 任务可选择需要批准（如生产环境的高风险命令）。
//
// 字段说明：
//   - ID: 任务唯一标识
//   - Name: 任务显示名称（如 "prod-db-health-check"）
//   - Description: 任务的用途说明
//   - ServerID: 执行该任务的服务器 ID（单个）
//   - CommandID: 关联的预设 ShellCommand 的 ID
//   - RequiresApproval: 是否需要审批才能执行（生产环境高风险命令）
//   - Executions: 该任务的历史执行记录列表（ShellTaskExecution）
//   - Command: 命令文本（冗余字段，便于 API 返回）
//   - CommandDesc: 命令描述（冗余字段，便于 API 返回）
//   - ServerName: 服务器名称（冗余字段，便于 API 返回）
//
// 批准流程（RequiresApproval=true）：
// 1. 用户提交任务执行请求
// 2. 等待授权用户批准
// 3. 批准后才能开始执行
// 4. 可用于生产环境的高风险操作
//
// 示例：
//	task := &ShellTask{
//		Name:             "check-db-health",
//		Description:      "Health check on production database",
//		ServerID:         1,
//		CommandID:        10,
//		RequiresApproval: true,
//	}
type ShellTask struct {
	ID               int                    `json:"id" db:"id" example:"1"`
	Name             string                 `json:"name" db:"name" example:"check-db-health" binding:"required"`
	Description      string                 `json:"description" db:"description" example:"Health check on database"`
	ServerID         int                    `json:"server_id" db:"server_id" example:"1" binding:"required"`
	CommandID        int                    `json:"command_id" db:"command_id" example:"1" binding:"required"`
	RequiresApproval bool                   `json:"requires_approval" db:"requires_approval" example:"true"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`

	// 关联关系用于前端显示
	Command          string                 `json:"command" db:"-"`
	CommandDesc      string                 `json:"command_description" db:"-"` // 用于前端显示命令描述
	ServerName       string                 `json:"server_name" db:"-"`         // 用于前端显示服务器名称
	Executions       []ShellTaskExecution   `json:"executions" db:"-"`          // 执行历史
}

func (t *ShellTask) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if t.ServerID <= 0 {
		return fmt.Errorf("server_id must be positive")
	}
	if t.CommandID <= 0 {
		return fmt.Errorf("command_id must be positive")
	}
	return nil
}

func (t *ShellTask) GetID() int {
	return t.ID
}
