/**
 * metadata.ts - 应用、环境、集群等元数据 API 调用
 */

import type {
  Application,
  Environment,
  Cluster,
  DeploymentTarget
} from '@/types/api'
import request from './request'

// ==================== 应用 ====================

/**
 * 获取应用列表
 */
export const getApplications = async (): Promise<Application[]> => {
  const response: any = await request.get('/v1/applications')
  return response?.data || []
}

/**
 * 获取单个应用
 */
export const getApplication = (appId: number): Promise<Application> => {
  return request.get(`/v1/applications/${appId}`)
}

// ==================== 环境 ====================

/**
 * 获取环境列表
 */
export const getEnvironments = async (): Promise<Environment[]> => {
  const response: any = await request.get('/v1/environments')
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
  return response?.data || []
}

/**
 * 获取单个集群
 */
export const getCluster = (clusterId: number): Promise<Cluster> => {
  return request.get(`/v1/clusters/${clusterId}`)
}

// ==================== 部署目标 ====================

/**
 * 获取所有部署目标
 */
export const getDeploymentTargets = async (): Promise<DeploymentTarget[]> => {
  const response: any = await request.get('/v1/deployment-targets')
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
): Promise<DeploymentTarget[]> => {
  return request.get(
    `/v1/deployment-targets/app/${appId}/env/${envId}`
  )
}

export const metadataAPI = {
  getApplications,
  getApplication,
  getEnvironments,
  getEnvironment,
  getClusters,
  getCluster,
  getDeploymentTargets,
  getClustersByAppAndEnv
}
 
 
