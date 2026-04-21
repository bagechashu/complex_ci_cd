package models

import (
	"fmt"
	"time"
)

// ShellTask 表示一个 Shell 任务，包含可在多个服务器上执行的命令
//
// ShellTask 是一个任务容器，关联一个已发布的 ShellCommand，
// 并定义该命令在多个服务器上的执行策略。
// 支持需要批准的命令（如生产环境上的危险操作）。
//
// 字段说明：
//   - ID: 任务唯一标识
//   - Name: 任务显示名称（如 "production-health-check"）
//   - Description: 任务的用途说明
//   - ServerIDs: 执行该任务的服务器 ID 列表（至少包含一个）
//   - CommandID: 关联的已发布 ShellCommand 的 ID
//   - ExecutionMethod: 执行方式（serial 或 parallel）
//   - RequiresApproval: 是否需要审批才能执行
//   - Executions: 该任务的历史执行记录列表（ShellTaskExecution）
//   - Command: 命令文本（冗余字段，便于 API 返回）
//   - CommandDesc: 命令描述（冗余字段，便于 API 返回）
//
// 执行方式说明：
//   - serial: 逐个服务器执行，前一个完成后才执行下一个。用于有依赖关系的操作。
//   - parallel: 同时在所有服务器上执行，用于独立的诊断命令等。
//
// 批准流程（RequiresApproval=true）：
// 1. 任务创建者提交任务
// 2. 等待授权用户批准
// 3. 批准后才能开始执行
// 4. 可用于生产环境的高风险操作
//
// 示例：
//	task := &ShellTask{
//		Name:             "check-cluster-health",
//		Description:      "Health check on production cluster",
//		ServerIDs:        []int{1, 2, 3},
//		CommandID:        10,
//		ExecutionMethod:  "parallel",
//		RequiresApproval: true,
//	}
type ShellTask struct {
	ID               int                `json:"id" db:"id" example:"1"`
	Name             string             `json:"name" db:"name" example:"check-cluster-health" binding:"required"`
	Description      string             `json:"description" db:"description" example:"Health check on all nodes"`
	ServerIDs        []int              `json:"server_ids" db:"-"` // 执行该任务的服务器 ID
	CommandID        int                `json:"command_id" db:"command_id" example:"1" binding:"required"`
	ExecutionMethod  string             `json:"execution_method" db:"execution_method" example:"parallel" enum:"serial,parallel"`
	RequiresApproval bool               `json:"requires_approval" db:"requires_approval" example:"true"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`

	// 关联关系用于前端显示
	Command          string             `json:"command" db:"-"`
	CommandDesc      string             `json:"command_description" db:"-"` // 用于前端显示命令描述
	Executions       []ShellTaskExecution    `json:"executions" db:"-"`           // 执行历史
}

func (t *ShellTask) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if t.CommandID <= 0 {
		return fmt.Errorf("command_id cannot be empty")
	}
	if len(t.ServerIDs) == 0 {
		return fmt.Errorf("at least one server must be selected")
	}
	if t.ExecutionMethod != "serial" && t.ExecutionMethod != "parallel" {
		return fmt.Errorf("execution_method must be 'serial' or 'parallel'")
	}
	return nil
}

func (t *ShellTask) GetID() int {
	return t.ID
}
