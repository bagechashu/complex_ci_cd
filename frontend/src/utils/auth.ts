/**
 * auth.ts - 认证相关工具函数
 */

/**
 * 获取当前登录用户
 */
export const getCurrentUser = (): string => {
  // 尝试从 localStorage 获取用户信息
  const userStr = localStorage.getItem('current_user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      return user.username || user.name || 'unknown'
    } catch {
      return 'current-user'
    }
  }
  
  // 尝试从 sessionStorage 获取
  const sessionUser = sessionStorage.getItem('current_user')
  if (sessionUser) {
    try {
      const user = JSON.parse(sessionUser)
      return user.username || user.name || 'unknown'
    } catch {
      return 'current-user'
    }
  }
  
  // 从 Cookie 获取用户信息
  const cookies = document.cookie.split(';')
  for (const cookie of cookies) {
    const [key, value] = cookie.trim().split('=')
    if (key === 'current_user' && value) {
      try {
        const user = JSON.parse(decodeURIComponent(value))
        return user.username || user.name || 'unknown'
      } catch {
        return decodeURIComponent(value)
      }
    }
  }
  
  // 默认返回当前用户
  return 'current-user'
}

/**
 * 获取认证令牌
 */
export const getAuthToken = (): string | null => {
  return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token') || null
}

/**
 * 设置认证令牌
 */
export const setAuthToken = (token: string): void => {
  localStorage.setItem('auth_token', token)
}

/**
 * 清除认证信息
 */
export const clearAuth = (): void => {
  localStorage.removeItem('auth_token')
  localStorage.removeItem('current_user')
  sessionStorage.removeItem('auth_token')
  sessionStorage.removeItem('current_user')
}

/**
 * 检查是否已认证
 */
export const isAuthenticated = (): boolean => {
  return !!getAuthToken()
}
