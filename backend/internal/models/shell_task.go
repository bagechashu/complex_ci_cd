package models

import (
	"fmt"
	"time"
)

type ShellTask struct {
	ID              int       `json:"id" db:"id,primarykey"`
	Name            string    `json:"name" db:"name,notnull"`
	Description     string    `json:"description" db:"description"`
	Command         string    `json:"command" db:"command,notnull"`
	Servers         []int     `json:"servers" db:"-"` // Server IDs
	Method          string    `json:"method" db:"method"` // serial or parallel
	RequiresApproval bool     `json:"requires_approval" db:"requires_approval"`
	CreatedAt       time.Time `json:"created_at" db:"created_at,notnull"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at,notnull"`
}

func (t *ShellTask) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if t.Command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if len(t.Servers) == 0 {
		return fmt.Errorf("at least one server must be selected")
	}
	if t.Method != "serial" && t.Method != "parallel" {
		return fmt.Errorf("method must be 'serial' or 'parallel'")
	}
	return nil
}

func (t *ShellTask) GetID() int {
	return t.ID
}
