/**
 * cluster-mapping.ts - 应用与集群的映射关系 API 调用
 */

import type { ClusterMapping } from '@/types/api'
import request from './request'

// ==================== 集群映射 ====================

/**
 * 获取所有集群映射
 */
export const getClusterMappings = async (): Promise<ClusterMapping[]> => {
  try {
    const response: any = await request.get('/v1/workload-targets')
    console.log('[DEBUG] getClusterMappings response:', response)
    // 响应拦截器已经返回了 data 字段，对于直接数组响应，response 就是数组
    if (Array.isArray(response)) {
      console.log('[DEBUG] getClusterMappings is array, length:', response.length)
      return response
    }
    console.warn('[DEBUG] getClusterMappings unexpected format:', typeof response)
    return response?.data || []
  } catch (error) {
    console.error('Failed to fetch cluster mappings:', error)
    throw error
  }
}

/**
 * 获取指定应用的集群映射列表
 * @param appId 应用 ID
 */
export const getClusterMappingsByApp = async (appId: number): Promise<ClusterMapping[]> => {
  try {
    const response: any = await request.get(`/v1/workload-targets/by-app/${appId}`)
    console.log('[DEBUG] getClusterMappingsByApp response:', response)
    // 响应拦截器已经返回了 data 字段，对于直接数组响应，response 就是数组
    if (Array.isArray(response)) {
      console.log('[DEBUG] getClusterMappingsByApp is array, length:', response.length)
      return response
    }
    console.warn('[DEBUG] getClusterMappingsByApp unexpected format:', typeof response)
    return response?.data || []
  } catch (error) {
    console.error(`Failed to fetch cluster mappings for app ${appId}:`, error)
    throw error
  }
}

/**
 * 获取单个集群映射
 * @param mappingId 映射 ID
 */
export const getClusterMapping = async (mappingId: number): Promise<ClusterMapping> => {
  try {
    const response: any = await request.get(`/v1/workload-targets/${mappingId}`)
    console.log('[DEBUG] getClusterMapping response:', response)
    // 响应拦截器已经返回了 data 字段，response 应该直接是对象
    return response
  } catch (error) {
    console.error(`Failed to fetch cluster mapping ${mappingId}:`, error)
    throw error
  }
}

/**
 * 创建新的集群映射
 * @param data 集群映射数据
 */
export const createClusterMapping = async (data: Partial<ClusterMapping>): Promise<ClusterMapping> => {
  try {
    // Validate required fields
    if (!data.app_id) throw new Error('app_id is required')
    if (!data.env_id) throw new Error('env_id is required')
    if (!data.cluster_id) throw new Error('cluster_id is required')
    if (!data.k8s_namespace) throw new Error('k8s_namespace is required')
    if (!data.workload_name) throw new Error('workload_name is required')
    if (!data.workload_type) throw new Error('workload_type is required')

    const response: any = await request.post('/v1/workload-targets', data)
    return response
  } catch (error) {
    console.error('Failed to create cluster mapping:', error)
    throw error
  }
}

/**
 * 更新集群映射
 * @param mappingId 映射 ID
 * @param data 更新的数据
 */
export const updateClusterMapping = async (
  mappingId: number,
  data: Partial<ClusterMapping>
): Promise<ClusterMapping> => {
  try {
    const response: any = await request.put(`/v1/workload-targets/${mappingId}`, data)
    return response
  } catch (error) {
    console.error(`Failed to update cluster mapping ${mappingId}:`, error)
    throw error
  }
}

/**
 * 删除集群映射
 * @param mappingId 映射 ID
 */
export const deleteClusterMapping = async (mappingId: number): Promise<void> => {
  try {
    await request.delete(`/v1/workload-targets/${mappingId}`)
  } catch (error) {
    console.error(`Failed to delete cluster mapping ${mappingId}:`, error)
    throw error
  }
}

/**
 * 批量删除集群映射
 * @param mappingIds 映射 ID 列表
 */
export const deleteClusterMappings = async (mappingIds: number[]): Promise<void> => {
  try {
    await Promise.all(mappingIds.map(id => deleteClusterMapping(id)))
  } catch (error) {
    console.error('Failed to delete cluster mappings:', error)
    throw error
  }
}
