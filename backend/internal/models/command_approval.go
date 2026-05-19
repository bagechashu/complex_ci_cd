package models

import (
	"fmt"
	"time"
)

// CommandApproval 表示 Shell 命令执行的审批记录
//
// CommandApproval 用于管理需要人工批准的 Shell 命令执行请求。
// 当 ShellCommand.RequiresApproval = true 时，执行前必须创建审批记录，
// 由授权用户审批通过后才能进行实际执行。
//
// 审批流程：
// 1. 命令提交者创建 CommandApproval（ApprovalStatus="pending"）
// 2. 授权用户查看并决定是否批准
// 3. 批准（ApprovalStatus="approved"）或拒绝（ApprovalStatus="rejected"）
// 4. 仅批准的请求才能继续执行
//
// 字段说明：
//   - ID: 审批记录唯一标识
//   - RequestID: 审批请求 ID，关联到具体的执行请求
//   - ApprovalStatus: 审批状态（pending/approved/rejected）
//   - ApprovedBy: 审批人的用户名或邮箱
//   - ApprovedAt: 审批时间（拒绝时也会记录）
//   - CreatedAt: 审批请求创建时间
//   - UpdatedAt: 最后更新时间
//
// 审批状态说明：
//   - pending: 等待批准，审批人还未审批
//   - approved: 已批准，允许继续执行
//   - rejected: 已拒绝，该请求无法执行
//
// 安全考虑：
// - 审批记录必须完整保存，用于审计追踪
// - ApprovedBy 应记录真实身份，用于责任追踪
// - 不允许修改已批准的记录（仅允许新建、查询）
//
// 示例：
//	approval := &CommandApproval{
//		RequestID:      "request-456",
//		ApprovalStatus: "approved",
//		ApprovedBy:     "admin@company.com",
//	}
type CommandApproval struct {
	ID              string     `json:"id" db:"id" example:"approval-123"`
	RequestID       string     `json:"request_id" db:"request_id" example:"request-456" binding:"required"`
	ApprovalStatus  string     `json:"approval_status" db:"approval_status" example:"approved" enum:"pending,approved,rejected"`
	ApprovedBy      string     `json:"approved_by" db:"approved_by" example:"admin@company.com"`
	ApprovedAt      *time.Time `json:"approved_at" db:"approved_at" example:"2026-04-21T10:00:00Z"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at" example:"2026-04-21T09:59:00Z"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`
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
