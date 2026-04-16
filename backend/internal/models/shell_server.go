package models

import (
	"fmt"
	"time"
)

type ShellServer struct {
	ID                int       `json:"id" db:"id,primarykey"`
	Name              string    `json:"name" db:"name,notnull,unique"`
	Host              string    `json:"host" db:"host,notnull"`
	Port              int       `json:"port" db:"port,notnull"`
	Username          string    `json:"username" db:"username"`
	AuthType          string    `json:"auth_type" db:"auth_type"` // password or key
	Password          string    `json:"password,omitempty" db:"password"`
	PrivateKey        string    `json:"private_key,omitempty" db:"private_key"`
	AllowedCommands   []string  `json:"allowed_commands" db:"-"`
	Status            string    `json:"status" db:"status"` // active, inactive, error
	LastConnected     *time.Time `json:"last_connected" db:"last_connected"`
	CreatedAt         time.Time `json:"created_at" db:"created_at,notnull"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at,notnull"`
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
