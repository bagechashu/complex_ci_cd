package models

import (
	"fmt"
	"time"
)

// ShellTaskExecution 表示 Shell 任务在单个服务器上的一次执行记录
//
// ShellTaskExecution 记录 ShellTask 在具体某台服务器上的一次执行记录。
// 当 ShellTask 执行时，会为每台目标服务器生成一个 ShellTaskExecution，
// 记录该服务器上的执行状态、输出、错误等详细信息。
//
// 字段说明：
//   - ID: 执行记录唯一标识
//   - TaskID: 关联的 ShellTask 任务 ID
//   - ServerID: 目标服务器 ID
//   - CommandID: 执行的命令 ID
//   - Status: 执行状态（pending/running/success/failed）
//   - Output: 命令的标准输出内容
//   - ErrorMessage: 错误信息（如连接失败、命令不存在等）
//   - ExitCode: 命令的退出码（0=成功，非0=失败）
//   - StartedAt: 执行开始时间
//   - CompletedAt: 执行完成时间
//   - CreatedAt: 记录创建时间
//   - UpdatedAt: 记录最后更新时间
//   - TaskName: 任务显示名称（冗余字段）
//   - ServerName: 服务器显示名称（冗余字段）
//   - Command: 命令文本（冗余字段）
//
// 执行状态流转：
//   pending -> running -> success (ExitCode=0) 或 failed (ExitCode≠0)
//   或者
//   pending -> failed (连接失败等未能执行的情况)
//
// 执行时间计算：
//   GetDuration() 返回 (CompletedAt - StartedAt) 的秒数
//   可用于性能分析和执行时间追踪
//
// 示例：
//	exec := &ShellTaskExecution{
//		TaskID:      10,
//		ServerID:    1,
//		CommandID:   5,
//		Status:      "success",
//		Output:      "HTTP/1.1 200 OK",
//		ExitCode:    ptr(0),
//	}
type ShellTaskExecution struct {
	ID           int        `json:"id" db:"id" example:"1"`
	TaskID       int        `json:"task_id" db:"task_id" example:"1" binding:"required"`
	ServerID     int        `json:"server_id" db:"server_id" example:"1" binding:"required"`
	CommandID    int        `json:"command_id" db:"command_id" example:"1" binding:"required"`
	Status       string     `json:"status" db:"status" example:"success" enum:"pending,running,success,failed"`
	Output       *string    `json:"output" db:"output" example:"...command output..."`
	ErrorMessage *string    `json:"error_message" db:"error_message" example:"Connection timeout"`
	CommandParams *string   `json:"command_params" db:"command_params" example:""`
	ExitCode     *int       `json:"exit_code" db:"exit_code" example:"0"`
	StartedAt    *time.Time `json:"started_at" db:"started_at" example:"2026-04-21T10:00:00Z"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at" example:"2026-04-21T10:01:00Z"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at" example:"2026-04-21T10:01:00Z"`

	// 关联关系用于前端显示
	TaskName   string `json:"task_name" db:"-"`
	ServerName string `json:"server_name" db:"-"`
	Command    string `json:"command" db:"-"`
}

func (e *ShellTaskExecution) GetID() int {
	return e.ID
}

// 获取执行耗时（秒）
func (e *ShellTaskExecution) GetDuration() int {
	if e.StartedAt == nil || e.CompletedAt == nil {
		return 0
	}
	return int(e.CompletedAt.Sub(*e.StartedAt).Seconds())
}

// Validate validates the ShellTaskExecution model
func (e *ShellTaskExecution) Validate() error {
	if e.CommandID == 0 {
		return fmt.Errorf("command_id is required")
	}
	if e.ServerID == 0 {
		return fmt.Errorf("server_id is required")
	}
	if e.Status == "" {
		e.Status = "pending"
	}
	// Valid statuses: pending, running, success, failed
	switch e.Status {
	case "pending", "running", "success", "failed":
		// valid
	default:
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	return nil
}
