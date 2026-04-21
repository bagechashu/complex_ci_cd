package models

import "time"

// ReleaseEvent 记录发布过程中的事件
//
// ReleaseEvent 用于追踪发布过程中的关键步骤和状态变化。
// 每次发布的生命周期从 STARTED 开始，经过 DEPLOYING、COMPLETED/FAILED，
// 最后可能进入 ROLLING_BACK/ROLLED_BACK 状态。
//
// 字段说明：
//   - ID: 事件唯一标识
//   - ReleaseID: 关联的发布记录 ID
//   - Type: 事件类型，标记发布阶段（如 STARTED, DEPLOYING, COMPLETED, FAILED）
//   - Message: 人可读的事件描述信息
//   - Details: JSON 格式的扩展信息，可包含错误堆栈、性能指标等
//   - CreatedAt: 事件创建时间
//
// 事件类型说明：
//   - STARTED: 发布开始
//   - DEPLOYING: 部署中（可能有多个此类事件表示各个集群/环节）
//   - COMPLETED: 发布成功
//   - FAILED: 发布失败
//   - ROLLING_BACK: 开始回滚
//   - ROLLED_BACK: 回滚完成
//
// 示例：
//	event := &ReleaseEvent{
//		ReleaseID: 123,
//		Type:      "DEPLOYING",
//		Message:   "Deploying to cluster-prod",
//		Details:   `{"cluster":"prod","namespace":"default"}`,
//	}
type ReleaseEvent struct {
	ID        int       `json:"id" db:"id"`
	ReleaseID int       `json:"release_id" db:"release_id"` // Changed from string to int to match schema
	Type      string    `json:"type" db:"type"`             // Renamed from EventType to match schema
	Message   string    `json:"message" db:"message"`       // Kept, but maps to schema's 'message' field
	Details   string    `json:"details" db:"details"`       // Added to match schema
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
