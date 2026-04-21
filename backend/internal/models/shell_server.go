package models

import (
	"fmt"
	"time"
)

// ShellServer 表示一个可以执行 Shell 命令的远程服务器（通过 SSH 连接）
//
// ShellServer 代表一个可远程登录并执行 Shell 命令的服务器。
// 支持密码认证（password）和 SSH 密钥认证（key）两种方式，
// 可用于执行运维命令、快速诊断、配置变更等。
//
// 字段说明：
//   - ID: 服务器唯一标识
//   - Name: 服务器显示名称（如 "prod-db-01"）
//   - Host: 服务器 IP 地址或域名
//   - Port: SSH 端口（通常为 22）
//   - Username: SSH 登录用户名
//   - AuthType: 认证方式（password 或 key）
//   - Password: 密码认证的密码（JSON 中隐藏，敏感信息）
//   - PrivateKey: SSH 密钥认证的私钥（JSON 中隐藏，敏感信息）
//   - Status: 连接状态（active/inactive/error）
//   - LastConnected: 最后连接成功的时间
//   - AllowedCommands: 该服务器允许执行的命令列表
//
// 认证类型说明：
//   - password: 使用用户名密码认证，安全性较低，仅建议在内网使用
//   - key: 使用 SSH 密钥认证，安全性更高，生产环境推荐使用
//
// 连接状态说明：
//   - active: 最近连接成功
//   - inactive: 从未连接或长期未连接
//   - error: 最近连接失败
//
// 安全注意：
// Password 和 PrivateKey 字段被标记为 json:"-"，
// API 响应中不会返回这些敏感信息，须在数据库加密保存。
//
// 示例：
//	server := &ShellServer{
//		Name:     "prod-db-01",
//		Host:     "192.168.1.100",
//		Port:     22,
//		Username: "admin",
//		AuthType: "key",
//		Status:   "active",
//	}
type ShellServer struct {
	ID            int        `json:"id" db:"id" example:"1"`
	Name          string     `json:"name" db:"name" example:"prod-server-1" binding:"required"`
	Host          string     `json:"host" db:"host" example:"192.168.1.100" binding:"required"`
	Port          int        `json:"port" db:"port" example:"22" binding:"required"`
	Username      string     `json:"username" db:"username" example:"admin" binding:"required"`
	AuthType      string     `json:"auth_type" db:"auth_type" example:"password" enum:"password,key" binding:"required"`
	Password      string     `json:"-" db:"password"`                                        // 隐藏敏感信息
	PrivateKey    string     `json:"-" db:"private_key"`                                    // 隐藏敏感信息
	Status        string     `json:"status" db:"status" example:"active" enum:"active,inactive,error"`
	LastConnected *time.Time `json:"last_connected" db:"last_connected" example:"2026-04-21T10:00:00Z"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`

	// 关联关系
	AllowedCommands []ShellCommand `json:"allowed_commands" db:"-"` // 该服务器允许的命令
}

func (s *ShellServer) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if s.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if s.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if s.AuthType != "password" && s.AuthType != "key" {
		return fmt.Errorf("auth_type must be 'password' or 'key'")
	}
	if s.AuthType == "password" && s.Password == "" {
		return fmt.Errorf("password is required when auth_type is 'password'")
	}
	if s.AuthType == "key" && s.PrivateKey == "" {
		return fmt.Errorf("private_key is required when auth_type is 'key'")
	}
	return nil
}

func (s *ShellServer) GetID() int {
	return s.ID
}
