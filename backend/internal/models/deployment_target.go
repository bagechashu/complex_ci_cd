package models

import "time"

type DeploymentTarget struct {
	ID             int       `json:"id"`
	AppID          int       `json:"app_id"`
	EnvID          int       `json:"env_id"`
	ClusterID      int       `json:"cluster_id"`
	K8sNamespace   string    `json:"k8s_namespace"`
	K8sDeployment  string    `json:"k8s_deployment"`
	ContainerName  *string   `json:"container_name"`
	RegistryDomain *string   `json:"registry_domain"`
	ImageRepo      *string   `json:"image_repo"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DeploymentTargetRequest struct {
	AppID          int    `json:"app_id" binding:"required"`
	EnvID          int    `json:"env_id" binding:"required"`
	ClusterID      int    `json:"cluster_id" binding:"required"`
	K8sNamespace   string `json:"k8s_namespace" binding:"required"`
	K8sDeployment  string `json:"k8s_deployment" binding:"required"`
	ContainerName  string `json:"container_name" binding:"required"`
	RegistryDomain string `json:"registry_domain"`
	ImageRepo      string `json:"image_repo" binding:"required"`
}

// DeploymentInfo combines all necessary information for deployment
type DeploymentInfo struct {
	Target  *DeploymentTarget
	App     *Application
	Env     *Environment
	Cluster *Cluster
}
