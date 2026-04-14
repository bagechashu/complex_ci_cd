/**
 * request.ts - Axios 请求拦截器和实例
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

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

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // 202 Accepted - 异步处理的响应
    if (response.status === 202 || response.status === 200) {
      return response.data
    }
    return response.data
  },
  (error) => {
    // 错误处理
    const errorResponse = error.response?.data

    if (errorResponse?.error) {
      const { code, message } = errorResponse.error
      console.error(`[API Error] ${code}: ${message}`)
      return Promise.reject(new Error(message))
    }

    if (error.message === 'Network Error') {
      console.error('[Network Error] 无法连接到服务器')
      return Promise.reject(new Error('网络连接失败，请检查后端服务'))
    }

    if (error.code === 'ECONNABORTED') {
      console.error('[Timeout Error] 请求超时')
      return Promise.reject(new Error('请求超时'))
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
