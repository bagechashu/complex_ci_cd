/**
 * request.ts - Axios 请求拦截器和实例
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { APIResponse, ErrorCode } from '@/types/api'

// 在开发模式下使用相对路径（通过Vite代理），生产模式下使用完整URL
const baseURL = import.meta.env.DEV 
  ? '/api'
  : (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api')

// 创建 Axios 实例
const request: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 添加请求 ID（链路追踪）
    const requestId = generateUUID()
    config.headers['X-Request-ID'] = requestId

    // 添加认证信息
    const token = getAuthToken()
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }

    // 调试模式下输出请求
    if (import.meta.env.DEV) {
      console.log(`[${requestId}] ${config.method?.toUpperCase()} ${config.url}`)
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 获取认证令牌
function getAuthToken(): string | null {
  // 从 localStorage 获取 token
  const token = localStorage.getItem('auth_token')
  if (token) {
    return token
  }

  // 从 sessionStorage 获取 token（临时会话）
  const sessionToken = sessionStorage.getItem('auth_token')
  if (sessionToken) {
    return sessionToken
  }

  return null
}

// 设置认证令牌
export function setAuthToken(token: string, persistent: boolean = false): void {
  if (persistent) {
    localStorage.setItem('auth_token', token)
    sessionStorage.removeItem('auth_token')
  } else {
    sessionStorage.setItem('auth_token', token)
    localStorage.removeItem('auth_token')
  }
}

// 清除认证令牌
export function clearAuthToken(): void {
  localStorage.removeItem('auth_token')
  sessionStorage.removeItem('auth_token')
}

// 业务错误类
export class BusinessError extends Error {
  constructor(
    public code: number,
    message: string,
    public data?: any
  ) {
    super(message)
    this.name = 'BusinessError'
  }
}

// 响应拦截器：处理 {code, message, data} 格式
request.interceptors.response.use(
  (response: AxiosResponse<APIResponse>) => {
    const { code, message, data } = response.data
    
    console.log('[DEBUG] response interceptor received:', {
      url: response.config.url,
      code,
      message,
      hasData: data !== undefined,
      dataType: data ? typeof data : 'undefined'
    })

    // 成功响应 (code === 0)
    if (code === 0) {
      console.log('[DEBUG] response interceptor returning data field')
      return data || response.data
    }

    // 业务错误 (code !== 0)
    // 后端返回 HTTP 200 + code 字段区分业务错误
    console.error(`[API Business Error] Code: ${code}, Message: ${message}`)
    return Promise.reject(new BusinessError(code, message, data))
  },
  (error) => {
    // 网络错误、超时等非 HTTP 200 的情况
    if (error instanceof BusinessError) {
      return Promise.reject(error)
    }

    const httpStatus = error.response?.status
    const errorData = error.response?.data

    // 尝试解析服务器返回的错误信息
    if (errorData?.code) {
      const { code, message } = errorData
      console.error(`[API Server Error] HTTP ${httpStatus}, Code: ${code}, Message: ${message}`)
      return Promise.reject(new BusinessError(code, message, errorData.data))
    }

    // 处理特定 HTTP 状态码
    if (httpStatus === 400) {
      console.error('[HTTP 400] 请求格式错误')
      return Promise.reject(new BusinessError(ErrorCode.INVALID_REQUEST, '请求格式错误'))
    }

    if (httpStatus === 401) {
      console.error('[HTTP 401] 未授权')
      // 清除 token 并跳转到登录
      clearAuthToken()
      window.location.href = '/login'
      return Promise.reject(new BusinessError(ErrorCode.UNAUTHORIZED, '未授权，请重新登录'))
    }

    if (httpStatus === 403) {
      console.error('[HTTP 403] 禁止访问')
      return Promise.reject(new BusinessError(ErrorCode.FORBIDDEN, '禁止访问'))
    }

    if (httpStatus === 404) {
      console.error('[HTTP 404] 未找到')
      return Promise.reject(new BusinessError(ErrorCode.NOT_FOUND, '请求的资源不存在'))
    }

    if (httpStatus === 500) {
      console.error('[HTTP 500] 服务器错误')
      return Promise.reject(new BusinessError(ErrorCode.INTERNAL_ERROR, '服务器内部错误'))
    }

    if (error.message === 'Network Error') {
      console.error('[Network Error] 无法连接到服务器')
      return Promise.reject(new BusinessError(ErrorCode.INTERNAL_ERROR, '网络连接失败，请检查后端服务'))
    }

    if (error.code === 'ECONNABORTED') {
      console.error('[Timeout Error] 请求超时')
      return Promise.reject(new BusinessError(ErrorCode.INTERNAL_ERROR, '请求超时'))
    }

    console.error('[Unknown Error]', error.message)
    return Promise.reject(error)
  }
)

// 生成 UUID
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

export default request
