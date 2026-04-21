package models

import (
	"fmt"
	"time"
)

// Cluster 表示一个 Kubernetes 集群或传统主机集群，用作应用部署目标
type Cluster struct {
	ID                  int       `json:"id" db:"id" example:"1"`
	Name                string    `json:"name" db:"name" example:"prod-k8s" binding:"required"`
	Type                string    `json:"type" db:"type" example:"kubernetes" enum:"kubernetes,ssh,ansible" binding:"required"`
	Environment         string    `json:"environment" db:"environment" example:"production"`
	RegistryPrefix      string    `json:"registry_prefix" db:"registry_prefix" example:"docker.io"`
	Labels              *string   `json:"labels" db:"labels" example:"zone=us-west,env=prod"`
	Kubeconfig          *string   `json:"-" db:"kubeconfig"`                                    // Hidden from JSON (K8s only)
	K8sConnectionStatus string    `json:"k8s_connection_status" db:"k8s_connection_status" enum:"connected,disconnected,unknown"`
	AnsibleHosts        *string   `json:"ansible_hosts" db:"ansible_hosts"`                    // Ansible/SaltStack inventory
	CreatedAt           time.Time `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`
}

// Validate 验证 Cluster 的必填字段
func (c *Cluster) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if c.Type == "" {
		return fmt.Errorf("type cannot be empty")
	}
	return nil
}

func (c *Cluster) GetID() interface{} {
	return c.ID
}
