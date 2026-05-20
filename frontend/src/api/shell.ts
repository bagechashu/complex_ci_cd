/**
 * shell.ts - Shell 服务器、命令和执行相关的 API 调用
 */

import type {
  ShellServer,
  ShellCommand,
  ShellCommandExecution,
  PaginatedResponse
} from '@/types/api'
import request from './request'

// ============ Shell Server API ============

export const listShellServers = (
  page: number = 1,
  pageSize: number = 10
): Promise<PaginatedResponse<ShellServer>> => {
  return request.get('/v1/shell-servers', {
    params: { page, pageSize }
  })
}

export const getShellServer = (id: number): Promise<ShellServer> => {
  return request.get(`/v1/shell-servers/${id}`)
}

export const createShellServer = (data: Omit<ShellServer, 'id' | 'created_at' | 'updated_at'>): Promise<ShellServer> => {
  return request.post('/v1/shell-servers', data)
}

export const updateShellServer = (id: number, data: Partial<ShellServer>): Promise<ShellServer> => {
  return request.put(`/v1/shell-servers/${id}`, data)
}

export const deleteShellServer = (id: number): Promise<void> => {
  return request.delete(`/v1/shell-servers/${id}`)
}

// ============ Shell Command API ============

export const listShellCommands = (
  page: number = 1,
  pageSize: number = 10,
  serverID?: number
): Promise<PaginatedResponse<ShellCommand>> => {
  const params: any = { page, pageSize }
  if (serverID) params.serverID = serverID
  return request.get('/v1/shell-commands', { params })
}

export const createShellCommand = (data: Omit<ShellCommand, 'id' | 'created_at' | 'updated_at'>): Promise<ShellCommand> => {
  return request.post('/v1/shell-commands', data)
}

export const publishShellCommand = (id: number): Promise<void> => {
  return request.post(`/v1/shell-commands/${id}/publish`, {})
}

export const unpublishShellCommand = (id: number): Promise<void> => {
  return request.post(`/v1/shell-commands/${id}/unpublish`, {})
}

export const deleteShellCommand = (id: number): Promise<void> => {
  return request.delete(`/v1/shell-commands/${id}`)
}

// ============ Shell Command Execution API ============

export const listShellCommandExecutions = (
  page: number = 1,
  pageSize: number = 10,
  commandID?: number
): Promise<PaginatedResponse<ShellCommandExecution>> => {
  const params: any = { page, pageSize }
  if (commandID) params.commandID = commandID
  return request.get('/v1/shell-command-executions', { params })
}

export const getShellCommandExecution = (id: number): Promise<ShellCommandExecution> => {
  return request.get(`/v1/shell-command-executions/${id}`)
}

export const executeShellCommand = (
  commandId: number,
  data: {
    server_id: number
    command_params?: string
  }
): Promise<ShellCommandExecution> => {
  return request.post(`/v1/shell-commands/${commandId}/execute`, data)
}

