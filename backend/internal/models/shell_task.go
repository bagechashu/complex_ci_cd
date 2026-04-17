package models

import (
	"fmt"
	"time"
)

type ShellTask struct {
	ID               int                  `json:"id" db:"id,primarykey"`
	Name             string               `json:"name" db:"name,notnull"`
	Description      string               `json:"description" db:"description"`
	ServerIDs        []int                `json:"server_ids" db:"-"` // 执行该任务的服务器 ID
	CommandID        int                  `json:"command_id" db:"command_id,notnull"` // 关联的 ShellCommand
	ExecutionMethod  string               `json:"execution_method" db:"execution_method"` // serial or parallel
	RequiresApproval bool                 `json:"requires_approval" db:"requires_approval"`
	CreatedAt        time.Time            `json:"created_at" db:"created_at,notnull"`
	UpdatedAt        time.Time            `json:"updated_at" db:"updated_at,notnull"`

	// 关联关系用于前端显示
	Command          string               `json:"command" db:"-"`
	CommandDesc      string               `json:"command_description" db:"-"` // 用于前端显示命令描述
	Executions       []ShellTaskExecution `json:"executions" db:"-"`          // 执行历史
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
