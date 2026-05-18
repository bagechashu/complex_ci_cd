/**
 * error-handler.ts - 错误处理工具函数
 * 根据后端业务码转换为用户友好的错误消息
 */

import { BusinessError } from '@/api/request'
import { ErrorCode } from '@/types/api'

/**
 * 根据业务错误码获取用户友好的错误消息
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof BusinessError) {
    return getBusinessErrorMessage(error.code, error.message)
  }

  if (error instanceof Error) {
    return error.message
  }

  return '发生未知错误，请稍后重试'
}

/**
 * 根据业务码返回用户友好的消息
 */
function getBusinessErrorMessage(code: number, defaultMessage: string): string {
  // 成功
  if (code === ErrorCode.SUCCESS) {
    return '操作成功'
  }

  // 1000-1999: 资源不存在
  if (code >= 1000 && code < 2000) {
    switch (code) {
      case ErrorCode.NOT_FOUND:
      case ErrorCode.APP_NOT_FOUND:
        return '应用不存在'
      case ErrorCode.CLUSTER_NOT_FOUND:
        return '集群不存在'
      case ErrorCode.ENVIRONMENT_NOT_FOUND:
        return '环境不存在'
      case ErrorCode.SHELL_SERVER_NOT_FOUND:
        return 'Shell 服务器不存在'
      case ErrorCode.SHELL_COMMAND_NOT_FOUND:
        return 'Shell 命令不存在'
      case ErrorCode.RELEASE_NOT_FOUND:
        return '发布记录不存在'
      default:
        return `资源不存在: ${defaultMessage}`
    }
  }

  // 2000-2999: 业务冲突
  if (code >= 2000 && code < 3000) {
    switch (code) {
      case ErrorCode.DUPLICATE_RESOURCE:
        return '资源已存在，请勿重复创建'
      default:
        return `操作冲突: ${defaultMessage}`
    }
  }

  // 3000-3999: 参数/验证错误
  if (code >= 3000 && code < 4000) {
    switch (code) {
      case ErrorCode.INVALID_REQUEST:
        return '请求格式错误'
      case ErrorCode.VALIDATION_ERROR:
        return `验证失败: ${defaultMessage}`
      case ErrorCode.MISSING_REQUIRED_FIELD:
        return '缺少必填字段'
      case ErrorCode.INVALID_PARAMETER:
        return `参数错误: ${defaultMessage}`
      default:
        return `输入错误: ${defaultMessage}`
    }
  }

  // 4000-4999: 权限/认证错误
  if (code >= 4000 && code < 5000) {
    switch (code) {
      case ErrorCode.UNAUTHORIZED:
        return '未授权，请重新登录'
      case ErrorCode.FORBIDDEN:
        return '权限不足，无法执行此操作'
      default:
        return `权限错误: ${defaultMessage}`
    }
  }

  // 5000-5999: 业务状态错误
  if (code >= 5000 && code < 6000) {
    switch (code) {
      case ErrorCode.INVALID_STATE:
        return '资源状态错误，无法执行此操作'
      case ErrorCode.OPERATION_NOT_ALLOWED:
        return '该操作不被允许'
      case ErrorCode.RELEASE_FAILED:
        return '发布失败'
      default:
        return `操作失败: ${defaultMessage}`
    }
  }

  // 9999: 服务器内部错误
  if (code === ErrorCode.INTERNAL_ERROR) {
    return '服务器内部错误，请稍后重试'
  }

  return defaultMessage || '发生错误，请稍后重试'
}

/**
 * 检查错误是否是特定类型
 */
export function isBusinessError(error: unknown, code?: number): error is BusinessError {
  if (!(error instanceof BusinessError)) {
    return false
  }

  if (code === undefined) {
    return true
  }

  return error.code === code
}

/**
 * 检查错误是否是资源不存在
 */
export function isNotFoundError(error: unknown): error is BusinessError {
  return isBusinessError(error) && error.code >= 1000 && error.code < 2000
}

/**
 * 检查错误是否是参数验证错误
 */
export function isValidationError(error: unknown): error is BusinessError {
  return isBusinessError(error) && error.code >= 3000 && error.code < 4000
}

/**
 * 检查错误是否是权限错误
 */
export function isPermissionError(error: unknown): error is BusinessError {
  return isBusinessError(error) && error.code >= 4000 && error.code < 5000
}

/**
 * 检查错误是否是业务状态错误
 */
export function isStateError(error: unknown): error is BusinessError {
  return isBusinessError(error) && error.code >= 5000 && error.code < 6000
}

/**
 * 检查错误是否是服务器错误
 */
export function isServerError(error: unknown): error is BusinessError {
  return isBusinessError(error) && error.code === ErrorCode.INTERNAL_ERROR
}
