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
export const getApplications = (): Promise<Application[]> => {
  return request.get('/api/v1/applications')
}

/**
 * 获取单个应用
 */
export const getApplication = (appId: number): Promise<Application> => {
  return request.get(`/api/v1/applications/${appId}`)
}

// ==================== 环境 ====================

/**
 * 获取环境列表
 */
export const getEnvironments = (): Promise<Environment[]> => {
  return request.get('/api/v1/environments')
}

/**
 * 获取单个环境
 */
export const getEnvironment = (envId: number): Promise<Environment> => {
  return request.get(`/api/v1/environments/${envId}`)
}

// ==================== 集群 ====================

/**
 * 获取集群列表
 */
export const getClusters = (): Promise<Cluster[]> => {
  return request.get('/api/v1/clusters')
}

/**
 * 获取单个集群
 */
export const getCluster = (clusterId: number): Promise<Cluster> => {
  return request.get(`/api/v1/clusters/${clusterId}`)
}

// ==================== 部署目标 ====================

/**
 * 获取所有部署目标
 */
export const getDeploymentTargets = (): Promise<DeploymentTarget[]> => {
  return request.get('/api/v1/deployment-targets')
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
    `/api/v1/deployment-targets/app/${appId}/env/${envId}`
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
