/**
 * uiStore.ts - UI 状态管理（加载状态、通知等）
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Message {
  id: string
  type: 'info' | 'success' | 'warning' | 'error'
  content: string
  duration?: number
}

export const useUiStore = defineStore('ui', () => {
  // ============ State ============
  const messages = ref<Message[]>([])
  const loading = ref(false)

  // ============ Actions ============

  /**
   * 显示消息
   */
  const showMessage = (
    content: string,
    type: Message['type'] = 'info',
    duration = 3000
  ) => {
    const id = Date.now().toString()
    const message: Message = { id, content, type, duration }
    messages.value.push(message)

    if (duration > 0) {
      setTimeout(() => {
        removeMessage(id)
      }, duration)
    }

    return id
  }

  /**
   * 移除消息
   */
  const removeMessage = (id: string) => {
    messages.value = messages.value.filter(m => m.id !== id)
  }

  /**
   * 快捷方法
   */
  const success = (content: string, duration?: number) =>
    showMessage(content, 'success', duration)
  const error = (content: string, duration?: number) =>
    showMessage(content, 'error', duration)
  const warning = (content: string, duration?: number) =>
    showMessage(content, 'warning', duration)
  const info = (content: string, duration?: number) =>
    showMessage(content, 'info', duration)

  /**
   * 设置全局加载状态
   */
  const setLoading = (value: boolean) => {
    loading.value = value
  }

  return {
    // State
    messages,
    loading,

    // Actions
    showMessage,
    removeMessage,
    success,
    error,
    warning,
    info,
    setLoading
  }
})
