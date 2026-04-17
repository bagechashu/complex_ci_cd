package models

import "time"

type WorkloadTarget struct {
	ID             int     `json:"id"`
	AppID          int     `json:"app_id"`
	EnvID          int     `json:"env_id"`
	ClusterID      int     `json:"cluster_id"`
	K8sNamespace   string  `json:"k8s_namespace"`
	K8sWorkload    string  `json:"k8s_workload"`
	ContainerName  *string `json:"container_name"`
	RegistryDomain *string `json:"registry_domain"`
	ImageRepo      *string `json:"image_repo"`
	WorkloadType   string  `json:"workload_type"`
	WorkloadName   string  `json:"workload_name"`
	// Enriched fields from cluster table (not stored in workload_target table)
	ClusterName    string    `json:"cluster_name,omitempty"`
	Environment    string    `json:"environment,omitempty"`
	RegistryPrefix string    `json:"registry_prefix,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type WorkloadTargetRequest struct {
	AppID          int    `json:"app_id" binding:"required"`
	EnvID          int    `json:"env_id" binding:"required"`
	ClusterID      int    `json:"cluster_id" binding:"required"`
	K8sNamespace   string `json:"k8s_namespace" binding:"required"`
	K8sworkload    string `json:"k8s_workload" binding:"required"`
	ContainerName  string `json:"container_name" binding:"required"`
	RegistryDomain string `json:"registry_domain"`
	ImageRepo      string `json:"image_repo" binding:"required"`
}

// workloadInfo combines all necessary information for workload
type WorkloadInfo struct {
	Target  *WorkloadTarget
	App     *Application
	Env     *Environment
	Cluster *Cluster
}
