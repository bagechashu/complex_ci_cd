package models

import (
	"fmt"
	"time"
)

type CommandApproval struct {
	ID              string    `json:"id" db:"id"`
	RequestID       string    `json:"request_id" db:"request_id"`
	ApprovalStatus  string    `json:"approval_status" db:"approval_status"`
	ApprovedBy      string    `json:"approved_by" db:"approved_by"`
	ApprovedAt      *time.Time `json:"approved_at" db:"approved_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

func (c *CommandApproval) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if c.RequestID == "" {
		return fmt.Errorf("request_id cannot be empty")
	}
	if c.ApprovalStatus == "" {
		return fmt.Errorf("approval_status cannot be empty")
	}
	return nil
}

func (c *CommandApproval) GetID() string {
	return c.ID
}
