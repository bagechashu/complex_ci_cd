/**
 * api.ts - API 的请求和响应类型定义
 */

// 发布请求
export interface ReleaseRequest {
  app_id: number
  env_id: number
  cluster_id: number
  image: string
  user: string
}

// 发布响应
export interface ReleaseResponse {
  id: number
  app_id: number
  env_id: number
  cluster_id: number
  image: string
  status: ReleaseStatus
  previous_image: string | null
  error_msg: string | null
  triggered_by: string
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

// 发布事件
export interface ReleaseEvent {
  id: number
  release_id: number
  type: string
  message: string
  details: string | null
  created_at: string
}

// 发布列表响应
export interface ListReleasesResponse {
  data: ReleaseResponse[]
  total: number
  limit: number
  offset: number
}

// 应用
export interface Application {
  id: number
  name: string
  image_name: string
  git_repo?: string | null
  repo?: string | null
  build_type?: string | null
  description?: string
  created_at?: string
  updated_at?: string
}

// 环境
export interface Environment {
  id: number
  name: string
  rank: number
  created_at: string
  updated_at: string
}

// 集群
export interface Cluster {
  id: number | string
  name: string
  type?: string // kubernetes | salt | ansible
  environment?: string
  registry_prefix?: string
  k8s_connection_status?: string // "connected" | "disconnected" | "unknown"
  created_at?: string
  updated_at?: string
}

// 应用与集群的映射（聚合视图 - 包含关联表及其关联的应用/集群信息）
export interface ClusterMapping {
  id: number
  app_id: number
  env_id?: number
  cluster_id: number | string
  cluster_name?: string
  environment?: string
  registry_prefix?: string
  k8s_namespace?: string
  k8s_deployment?: string
  workload_type?: string
  workload_name?: string
  container_name?: string
  current_image?: string
  created_at?: string
  updated_at?: string
}

// 部署目标
export interface DeploymentTarget {
  id: number
  app_id: number
  env_id: number
  cluster_id: number
  k8s_namespace: string | null
  k8s_deployment: string | null
  container_name: string | null
  registry_domain: string | null
  image_repo: string | null
  created_at: string
  updated_at: string
}

// 通用错误响应
export interface ErrorResponse {
  error: {
    code: string
    message: string
  }
}

// 发布状态枚举
export type ReleaseStatus =
  | 'pending'
  | 'validating'
  | 'deploying'
  | 'success'
  | 'failed'
  | 'rolled_back'
