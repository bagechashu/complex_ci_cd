package models

import (
	"time"
)

// ShellTaskExecution 表示 Shell 任务的一次执行记录
type ShellTaskExecution struct {
	ID           int        `json:"id" db:"id,primarykey"`
	TaskID       int        `json:"task_id" db:"task_id,notnull"`
	ServerID     int        `json:"server_id" db:"server_id,notnull"`
	CommandID    int        `json:"command_id" db:"command_id,notnull"` // 关联的 ShellCommand
	Status       string     `json:"status" db:"status"`                 // pending, running, success, failed
	Output       string     `json:"output" db:"output"`                 // 执行输出
	ErrorMessage string     `json:"error_message" db:"error_message"`
	ExitCode     *int       `json:"exit_code" db:"exit_code"` // Unix 退出码
	StartedAt    *time.Time `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at,notnull"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at,notnull"`

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
