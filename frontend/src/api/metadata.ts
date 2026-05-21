/**
 * metadata.ts - 应用、环境、集群等元数据 API 调用
 */

import type {
  Application,
  Environment,
  Cluster,
  WorkloadTarget,
  PaginatedResponse
} from '@/types/api'
import request from './request'

// ==================== 应用 ====================

export const getApplications = async (page: number = 1, pageSize: number = 10, search: string = ''): Promise<PaginatedResponse<Application>> => {
  const params = new URLSearchParams()
  params.append('page', page.toString())
  params.append('pageSize', pageSize.toString())
  if (search) {
    params.append('search', search)
  }
  
  try {
    const response: any = await request.get(`/v1/applications?${params.toString()}`)
    console.log('[DEBUG] getApplications raw response:', response)
    
    const result = {
      page: response?.page || 1,
      pageSize: response?.pageSize || 10,
      total: response?.total || 0,
      totalPages: response?.totalPages || 1,
      data: response?.data || []
    }
    
    console.log('[DEBUG] getApplications final result:', result)
    console.log('[DEBUG] applications data:', result.data)
    console.log('[DEBUG] applications count:', result.data?.length || 0)
    
    return result
  } catch (error) {
    console.error('[DEBUG] getApplications error:', error)
    throw error
  }
}

/**
 * 获取单个应用
 */
export const getApplication = (appId: number): Promise<Application> => {
  return request.get(`/v1/applications/${appId}`)
}

/**
 * 创建应用
 */
export const createApplication = (app: Partial<Application>): Promise<Application> => {
  return request.post('/v1/applications', app)
}

/**
 * 更新应用
 */
export const updateApplication = (appId: number, app: Partial<Application>): Promise<Application> => {
  return request.put(`/v1/applications/${appId}`, app)
}

/**
 * 删除应用
 */
export const deleteApplication = (appId: number): Promise<void> => {
  return request.delete(`/v1/applications/${appId}`)
}

// ==================== 环境 ====================

/**
 * 获取环境列表
 */
export const getEnvironments = async (): Promise<Environment[]> => {
  const response: any = await request.get('/v1/environments')
  console.log('[DEBUG] getEnvironments response:', response)
  // 响应拦截器已经返回了 data 字段，对于直接数组响应，response 就是数组
  if (Array.isArray(response)) {
    console.log('[DEBUG] getEnvironments is array, length:', response.length)
    return response
  }
  console.warn('[DEBUG] getEnvironments unexpected format:', typeof response)
  return response?.data || []
}

/**
 * 获取单个环境
 */
export const getEnvironment = (envId: number): Promise<Environment> => {
  return request.get(`/v1/environments/${envId}`)
}

// ==================== 集群 ====================

/**
 * 获取集群列表
 */
export const getClusters = async (): Promise<Cluster[]> => {
  const response: any = await request.get('/v1/clusters')
  console.log('[DEBUG] getClusters response:', response)
  // 响应拦截器已经返回了 data 字段，对于直接数组响应，response 就是数组
  if (Array.isArray(response)) {
    console.log('[DEBUG] getClusters is array, length:', response.length)
    return response
  }
  console.warn('[DEBUG] getClusters unexpected format:', typeof response)
  return response?.data || []
}

/**
 * 获取单个集群
 */
export const getCluster = (clusterId: number): Promise<Cluster> => {
  return request.get(`/v1/clusters/${clusterId}`)
}

/**
 * 创建集群
 */
export const createCluster = (cluster: Partial<Cluster>): Promise<Cluster> => {
  return request.post('/v1/clusters', cluster)
}

/**
 * 更新集群
 */
export const updateCluster = (clusterId: number, cluster: Partial<Cluster>): Promise<Cluster> => {
  return request.put(`/v1/clusters/${clusterId}`, cluster)
}

/**
 * 删除集群
 */
export const deleteCluster = (clusterId: number): Promise<void> => {
  return request.delete(`/v1/clusters/${clusterId}`)
}

/**
 * 测试集群连接状态
 */
export const testClusterConnection = (clusterId: number): Promise<{ id: number; status: string; message: string }> => {
  return request.post(`/v1/clusters/${clusterId}/test-connection`, {})
}

// ==================== 部署目标 ====================

/**
 * 获取所有部署目标
 */
export const getWorkloadTargets = async (): Promise<WorkloadTarget[]> => {
  const response: any = await request.get('/v1/workload-targets')
  console.log('[DEBUG] getWorkloadTargets response:', response)
  // 响应拦截器已经返回了 data 字段，对于直接数组响应，response 就是数组
  if (Array.isArray(response)) {
    console.log('[DEBUG] getWorkloadTargets is array, length:', response.length)
    return response
  }
  console.warn('[DEBUG] getWorkloadTargets unexpected format:', typeof response)
  return response?.data || []
}

/**
 * 获取指定应用在指定环境下的集群列表
 * @param appId 应用 ID
 * @param envId 环境 ID
 */
export const getClustersByAppAndEnv = (
  appId: number,
  envId: number
): Promise<WorkloadTarget[]> => {
  return request.get(
    `/v1/workload-targets/app/${appId}/env/${envId}`
  )
}

/**
 * 获取与指定集群关联的应用列表
 * @param clusterId 集群 ID
 */
export const getApplicationsByCluster = async (
  clusterId: number | string
): Promise<Application[]> => {
  try {
    const allTargets = await getWorkloadTargets()
    const appIds = new Set<number>()
    
    // 从部署目标中提取与该集群关联的应用 ID
    allTargets.forEach(target => {
      if (target.cluster_id === clusterId || target.cluster_id === parseInt(String(clusterId))) {
        appIds.add(target.app_id)
      }
    })
    
    // 如果没有应用，返回空数组
    if (appIds.size === 0) {
      return []
    }
    
    // 获取所有应用
    const allAppsResponse = await getApplications()
    
    // 过滤出在集群中的应用
    return allAppsResponse.data.filter(app => appIds.has(app.id))
  } catch (error) {
    console.error('Failed to fetch applications for cluster:', error)
    return []
  }
}

// ==================== Shell 服务器 ====================

/**
 * 获取所有 Shell 服务器
 */
export const getShellServers = async (): Promise<any[]> => {
  const response: any = await request.get('/v1/shell-servers')
  return response?.data || []
}

/**
 * 获取单个 Shell 服务器
 */
export const getShellServer = (serverId: number): Promise<any> => {
  return request.get(`/v1/shell-servers/${serverId}`)
}

/**
 * 创建 Shell 服务器
 */
export const createShellServer = (server: any): Promise<any> => {
  return request.post('/v1/shell-servers', server)
}

/**
 * 更新 Shell 服务器
 */
export const updateShellServer = (serverId: number, server: any): Promise<any> => {
  return request.put(`/v1/shell-servers/${serverId}`, server)
}

/**
 * 删除 Shell 服务器
 */
export const deleteShellServer = (serverId: number): Promise<void> => {
  return request.delete(`/v1/shell-servers/${serverId}`)
}

// ==================== 发布 ====================

/**
 * 获取所有发布记录
 */
export const getReleases = async (): Promise<any[]> => {
  const response: any = await request.get('/v1/releases')
  return response?.data || []
}

/**
 * 获取单个发布记录
 */
export const getRelease = (releaseId: number): Promise<any> => {
  return request.get(`/v1/releases/${releaseId}`)
}

/**
 * 创建发布
 */
export const createRelease = (payload: any): Promise<any> => {
  return request.post('/v1/releases', payload)
}

/**
 * 获取发布事件
 */
export const getReleaseEvents = async (releaseId: number): Promise<any[]> => {
  const response: any = await request.get(`/v1/releases/${releaseId}/events`)
  return response?.data || []
}
// ==================== 执行历史 ====================

/**
 * 获取所有执行历史
 */
export const getExecutionHistories = async (): Promise<any[]> => {
  const response: any = await request.get('/v1/shell-execution-history')
  return response?.data || []
}
export const metadataAPI = {
  getApplications,
  getApplication,
  getEnvironments,
  getEnvironment,
  getClusters,
  getCluster,
  createCluster,
  updateCluster,
  deleteCluster,
  getWorkloadTargets,
  getClustersByAppAndEnv,
  getApplicationsByCluster,
  getShellServers,
  getShellServer,
  createShellServer,
  updateShellServer,
  deleteShellServer,
  getReleases,
  getRelease,
  createRelease,
  getReleaseEvents,
  getExecutionHistories
}
