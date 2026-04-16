package models

import (
	"fmt"
	"time"
)

type Cluster struct {
	ID                   int       `json:"id" db:"id"`
	Name                 string    `json:"name" db:"name"`
	Type                 string    `json:"type" db:"type"`
	Environment          string    `json:"environment" db:"environment"`
	RegistryPrefix       string    `json:"registry_prefix" db:"registry_prefix"`
	Labels               *string   `json:"labels" db:"labels"`
	KubeconfigPath       *string   `json:"kubeconfig_path" db:"kubeconfig_path"`
	Kubeconfig           *string   `json:"-" db:"kubeconfig"`                         // Hidden from JSON response
	KubeconfigEncrypted  *string   `json:"-" db:"kubeconfig_encrypted"`               // Hidden from JSON response
	K8sConnectionStatus  string    `json:"k8s_connection_status" db:"k8s_connection_status"` // "connected", "disconnected", "unknown"
	AnsibleHosts         *string   `json:"ansible_hosts" db:"ansible_hosts"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

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
