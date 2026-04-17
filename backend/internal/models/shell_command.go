package models

import (
	"fmt"
	"time"
)

// ShellCommand 表示允许在服务器上执行的命令
type ShellCommand struct {
	ID          int       `json:"id" db:"id,primarykey"`
	ServerID    int       `json:"server_id" db:"server_id,notnull"`
	Command     string    `json:"command" db:"command,notnull"` // 实际执行的命令
	Description string    `json:"description" db:"description"`
	IsPublished bool      `json:"is_published" db:"is_published"` // 是否已发布（发布后才允许执行）
	CreatedAt   time.Time `json:"created_at" db:"created_at,notnull"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at,notnull"`

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
