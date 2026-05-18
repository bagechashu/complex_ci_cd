/**
 * api.ts - API 的请求和响应类型定义
 */

// 发布状态枚举
export type ReleaseStatus =
  | 'pending'
  | 'validating'
  | 'deploying'
  | 'success'
  | 'failed'
  | 'rolled_back'

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
  description?: string | null
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
  kubeconfig?: string // K8s kubeconfig content (encrypted in database)
  labels?: string // Cluster labels
  k8s_connection_status?: string // "connected" | "disconnected" | "unknown"
  created_at?: string
  updated_at?: string
}

// WorkloadTarget - Application to Cluster mapping with K8s workload configuration
// Represents deployment target (Deployment/StatefulSet/DaemonSet)
export interface WorkloadTarget {
  id: number
  app_id: number
  env_id: number
  cluster_id: number
  k8s_namespace: string
  k8s_workload: string
  container_name?: string | null
  registry_domain?: string | null
  image_repo?: string | null
  workload_type: string // Deployment, StatefulSet, DaemonSet, etc.
  workload_name: string
  // Enriched fields from related tables (not stored in workload_target table)
  cluster_name?: string
  environment?: string
  registry_prefix?: string
  current_image?: string
  created_at: string
  updated_at: string
}

// ClusterMapping - aggregated view of workload target with additional context
export interface ClusterMapping extends WorkloadTarget {
  current_image?: string
}

// 通用错误响应
export interface ErrorResponse {
  error: {
    code: string
    message: string
  }
}

// ============ Shell Server Types ============
export interface ShellServer {
  id: number
  name: string
  host: string
  port: number
  username: string
  auth_type: 'password' | 'key'
  password?: string // 前端不显示
  private_key?: string // 前端不显示
  status: 'active' | 'inactive' | 'error'
  last_connected?: string | null
  created_at: string
  updated_at: string
  allowed_commands?: ShellCommand[]
}

export interface ShellCommand {
  id: number
  server_id: number
  command: string
  description?: string
  is_published: boolean
  created_at: string
  updated_at: string
  server_name?: string // 用于前端显示
}

export interface ShellTask {
  id: number
  name: string
  description?: string
  server_id: number
  command_id: number
  requires_approval: boolean
  created_at: string
  updated_at: string
  command?: string // 用于前端显示
  command_description?: string
  server_name?: string // 用于前端显示服务器名称
  executions?: ShellTaskExecution[]
}

export interface ShellTaskExecution {
  id: number
  task_id: number
  server_id: number
  command_id: number
  status: 'pending' | 'running' | 'success' | 'failed'
  output?: string
  error_message?: string
  exit_code?: number | null
  started_at?: string | null
  completed_at?: string | null
  created_at: string
  updated_at: string
  task_name?: string
  server_name?: string
  command?: string
}

// 分页响应
export interface PaginatedResponse<T> {
  page: number
  pageSize: number
  total: number
  totalPages: number
  data: T[]
}
