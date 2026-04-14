/**
 * models.ts - 业务模型类型定义
 */

import type { ReleaseStatus, ReleaseEvent } from './api'

// 发布记录（UI 展示用）
export interface ReleaseRecord {
  id: number
  app_id: number
  app_name: string
  env_id: number
  env_name: string
  cluster_id: number
  cluster_name: string
  image: string
  status: ReleaseStatus
  previous_image: string | null
  error_msg: string | null
  triggered_by: string
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
  events?: ReleaseEvent[]
  duration?: number // 毫秒
}

// 发布流程表单
export interface ReleaseFlowForm {
  app_id: number | null
  env_id: number | null
  cluster_id: number | null
  image: string
}

// 表格行数据
export interface TableReleaseRecord extends ReleaseRecord {
  key: string | number
}

// 进度计算结果
export interface ProgressInfo {
  percentage: number
  status: ReleaseStatus
  message: string
}
