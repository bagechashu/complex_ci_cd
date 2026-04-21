package models

import "time"

// WorkloadTarget 表示应用在特定集群中的部署配置
//
// WorkloadTarget 是连接应用（Application）、环境（Environment）和集群（Cluster）的中间层，
// 定义了一个应用如何在某个集群的某个环境中部署，包括 Kubernetes 命名空间、工作负载类型等。
//
// 每个 WorkloadTarget 记录了部署的具体目标和配置，是发布决策的重要参考。
//
// 字段说明：
//   - ID: 部署配置唯一标识
//   - AppID: 应用 ID，关联 Application
//   - EnvID: 环境 ID，关联 Environment
//   - ClusterID: 集群 ID，关联 Cluster
//   - K8sNamespace: Kubernetes 命名空间名称（如 "default", "production"）
//   - K8sWorkload: Kubernetes 工作负载资源名称（部署对象的 metadata.name）
//   - ContainerName: 容器名称（Pod 中的容器标识符）
//   - RegistryDomain: 容器镜像仓库域名（如 "docker.io"）
//   - ImageRepo: 镜像仓库路径（如 "my-org/api-service"）
//   - WorkloadType: 工作负载类型 (Deployment/StatefulSet/DaemonSet/Job)
//   - WorkloadName: 工作负载显示名称
//   - ClusterName: 集群名称（冗余字段，便于 API 返回）
//   - Environment: 环境名称（冗余字段，便于 API 返回）
//   - RegistryPrefix: 镜像仓库前缀（冗余字段，便于 API 返回）
//
// 工作负载类型说明：
//   - Deployment: 无状态应用（默认类型）
//   - StatefulSet: 有状态应用（数据库等）
//   - DaemonSet: 节点代理（日志收集、监控等）
//   - Job: 一次性任务（数据迁移等）
//
// 示例：
//	wt := &WorkloadTarget{
//		AppID:        1,
//		EnvID:        4,              // production
//		ClusterID:    2,
//		K8sNamespace: "production",
//		K8sWorkload:  "api-service-deployment",
//		WorkloadType: "Deployment",
//		WorkloadName: "api-service",
//	}
type WorkloadTarget struct {
	ID             int       `json:"id" db:"id" example:"1"`
	AppID          int       `json:"app_id" db:"app_id" example:"1" binding:"required"`
	EnvID          int       `json:"env_id" db:"env_id" example:"1" binding:"required"`
	ClusterID      int       `json:"cluster_id" db:"cluster_id" example:"1" binding:"required"`
	K8sNamespace   string    `json:"k8s_namespace" db:"k8s_namespace" example:"default" binding:"required"`
	K8sWorkload    string    `json:"k8s_workload" db:"k8s_workload" example:"api-service-deployment" binding:"required"`
	ContainerName  *string   `json:"container_name" db:"container_name" example:"api-service"`
	RegistryDomain *string   `json:"registry_domain" db:"registry_domain" example:"docker.io"`
	ImageRepo      *string   `json:"image_repo" db:"image_repo" example:"my-org/api-service"`
	WorkloadType   string    `json:"workload_type" db:"workload_type" example:"Deployment" enum:"Deployment,StatefulSet,DaemonSet,Job"`
	WorkloadName   string    `json:"workload_name" db:"workload_name" example:"api-service"`
	// Enriched fields from cluster table (not stored in workload_target table)
	ClusterName    string    `json:"cluster_name,omitempty"`
	Environment    string    `json:"environment,omitempty"`
	RegistryPrefix string    `json:"registry_prefix,omitempty"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" example:"2026-04-21T10:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at" example:"2026-04-21T10:00:00Z"`
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
