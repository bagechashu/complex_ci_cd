/**
 * release.ts - 发布相关的 API 调用
 */

import type {
  ReleaseRequest,
  ReleaseResponse,
  ReleaseEvent,
  ListReleasesResponse
} from '@/types/api'
import request from './request'

/**
 * 发起发布
 * @param data 发布请求
 * @returns 发布响应 (202 Accepted)
 */
export const createRelease = (data: ReleaseRequest): Promise<ReleaseResponse> => {
  return request.post('/api/v1/releases', data)
}

/**
 * 查询发布状态
 * @param releaseId 发布 ID
 * @returns 发布详情及事件
 */
export const getRelease = (releaseId: number): Promise<ReleaseResponse> => {
  return request.get(`/api/v1/releases/${releaseId}`)
}

/**
 * 获取发布列表
 * @param limit 分页大小
 * @param offset 分页偏移
 * @returns 发布列表
 */
export const listReleases = (
  limit: number = 20,
  offset: number = 0
): Promise<ListReleasesResponse> => {
  return request.get('/api/v1/releases', {
    params: { limit, offset }
  })
}

/**
 * 获取发布事件流
 * @param releaseId 发布 ID
 * @returns 事件列表
 */
export const getReleaseEvents = (releaseId: number): Promise<ReleaseEvent[]> => {
  return request.get(`/api/v1/releases/${releaseId}/events`)
}

/**
 * 回滚发布
 * @param releaseId 发布 ID
 * @returns 新的发布响应
 */
export const rollbackRelease = (releaseId: number): Promise<ReleaseResponse> => {
  return request.post(`/api/v1/releases/${releaseId}/rollback`)
}

export const releaseAPI = {
  createRelease,
  getRelease,
  listReleases,
  getReleaseEvents,
  rollbackRelease
}
