package models

import "time"

type Cluster struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Type               string    `json:"type"` // kubernetes, salt, ansible
	KubeconfigPath     string    `json:"kubeconfig_path,omitempty"`
	KubeconfigEncrypted string   `json:"kubeconfig_encrypted,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ClusterRequest struct {
	Name       string `json:"name" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Kubeconfig string `json:"kubeconfig"`
}
