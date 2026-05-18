<template>
  <div class="shell-command-execution-page">
    <div class="content-layout">
      <!-- Left Panel: Published Commands List -->
      <div class="list-panel">
        <div class="list-header">
          <div class="list-header-top">
            <h2>已发布命令</h2>
          </div>
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索命令..."
            @input="filterCommands"
          />
        </div>

        <div class="list-container">
          <!-- Commands by Server -->
          <div v-for="server in groupedCommandsByServer" :key="server.id" class="server-group">
            <div class="server-header">
              🖥️ {{ server.name }} <span class="cmd-count">({{ server.commands.length }})</span>
            </div>

            <div class="commands-items">
              <div
                v-for="cmd in server.commands"
                :key="cmd.id"
                class="list-item"
                :class="{ active: selectedCommandId === cmd.id }"
                @click="selectCommand(cmd, server)"
              >
                <div class="list-item-header">
                  <div class="list-item-name">{{ cmd.description || cmd.command }}</div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="filteredCommands.length === 0" class="empty-state">
            <p>暂无已发布的命令</p>
          </div>
        </div>
      </div>

      <!-- Right Panel: Command Details & Execution -->
      <div class="detail-panel">
        <div v-if="!selectedCommand" class="empty-detail">
          <p>请选择一个命令开始</p>
        </div>

        <div v-else class="detail-content">
          <!-- Command Info -->
          <div class="detail-section">
            <div class="section-header">
              <h3>命令详情</h3>
            </div>

            <div class="info-grid">
              <div class="info-item">
                <div class="info-item-label">服务器</div>
                <div class="info-item-value">{{ selectedServer?.name }}</div>
              </div>
              <div class="info-item">
                <div class="info-item-label">主机</div>
                <div class="info-item-code">{{ selectedServer?.host }}</div>
              </div>
              <div class="info-item">
                <div class="info-item-label">描述</div>
                <div class="info-item-value">{{ selectedCommand.description || '无' }}</div>
              </div>
              <div class="info-item">
                <div class="info-item-label">命令</div>
                <div class="info-item-code">{{ selectedCommand.command }}</div>
              </div>
            </div>
          </div>

          <!-- Execution Section -->
          <div class="detail-section">
            <div class="section-header">
              <h3>执行操作</h3>
            </div>

            <div class="execution-controls">
              <n-button
                type="primary"
                size="large"
                :loading="executing"
                @click="executeCommand"
              >
                执行命令
              </n-button>
            </div>
          </div>

          <!-- Execution History -->
          <div class="detail-section">
            <div class="section-header">
              <h3>执行历史</h3>
            </div>

            <div v-if="commandExecutions.length > 0" class="executions-list">
              <div
                v-for="exec in commandExecutions"
                :key="exec.id"
                class="execution-record"
                :class="`status-${exec.status}`"
              >
                <div class="exec-header">
                  <span class="exec-time">{{ formatTime(exec.created_at) }}</span>
                  <n-tag :type="getExecutionStatusType(exec.status)">
                    {{ getExecutionStatusLabel(exec.status) }}
                  </n-tag>
                </div>
                <div v-if="exec.error_message" class="exec-error">
                  ❌ {{ exec.error_message }}
                </div>
                <div v-if="exec.output" class="exec-output">
                  <code>{{ exec.output }}</code>
                </div>
              </div>
            </div>

            <div v-else class="empty-state">
              <p>暂无执行记录</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useShellStore } from '@/stores/shellStore'
import type { ShellCommand, ShellServer, ShellTaskExecution } from '@/types/api'
import { NButton, NTag, NInput, useMessage } from 'naive-ui'

// ============ Stores & Composables ============
const shellStore = useShellStore()
const message = useMessage()

// ============ State ============
const selectedCommandId = ref<number | null>(null)
const selectedCommand = ref<ShellCommand | null>(null)
const selectedServer = ref<ShellServer | null>(null)
const searchQuery = ref('')
const executing = ref(false)
const commandExecutions = ref<ShellTaskExecution[]>([])

// ============ Computed ============
const shellCommands = computed(() => shellStore.shellCommands)
const shellServers = computed(() => shellStore.shellServers)

// Filter published commands
const filteredCommands = computed(() => {
  return shellCommands.value.filter(cmd => {
    if (!cmd.is_published) return false
    if (!searchQuery.value) return true
    const lowerQuery = searchQuery.value.toLowerCase()
    return (
      cmd.command.toLowerCase().includes(lowerQuery) ||
      cmd.description?.toLowerCase().includes(lowerQuery)
    )
  })
})

// Group commands by server
const groupedCommandsByServer = computed(() => {
  const grouped: Record<
    number,
    { id: number; name: string; host: string; commands: ShellCommand[] }
  > = {}

  filteredCommands.value.forEach(cmd => {
    if (!grouped[cmd.server_id]) {
      const server = shellServers.value.find(s => s.id === cmd.server_id)
      grouped[cmd.server_id] = {
        id: cmd.server_id,
        name: server?.name || `Server ${cmd.server_id}`,
        host: server?.host || '',
        commands: []
      }
    }
    grouped[cmd.server_id].commands.push(cmd)
  })

  return Object.values(grouped)
})

// ============ Methods ============

/**
 * Load data on mount
 */
onMounted(async () => {
  await shellStore.fetchShellServers(1, 100)
  await shellStore.fetchShellCommands(1, 100)
})

/**
 * Filter commands when search changes
 */
const filterCommands = () => {
  // computed will reactively update
}

/**
 * Select a command
 */
const selectCommand = (cmd: ShellCommand, server: any) => {
  selectedCommandId.value = cmd.id
  selectedCommand.value = cmd
  selectedServer.value = server
  // Load execution history for this command
  loadExecutionHistory(cmd.id)
}

/**
 * Load execution history for a command
 */
const loadExecutionHistory = async (commandId: number) => {
  try {
    // Fetch the latest 5 executions
    const executions = await shellStore.fetchShellTaskExecutions(1, 5)
    commandExecutions.value = executions.filter((exec: ShellTaskExecution) => exec.command_id === commandId)
  } catch (err) {
    console.error('Failed to load execution history:', err)
  }
}

/**
 * Execute command
 */
const executeCommand = async () => {
  if (!selectedCommand.value || !selectedServer.value) return

  executing.value = true
  try {
    // Create a ShellTaskExecution record
    const execution: Omit<ShellTaskExecution, 'id' | 'created_at' | 'updated_at'> = {
      task_id: 0, // Virtual task ID for direct command execution
      server_id: selectedServer.value.id,
      command_id: selectedCommand.value.id,
      status: 'pending',
      output: '',
      error_message: '',
      exit_code: null,
      started_at: null,
      completed_at: null
    }

    // Call API to execute
    const result = await shellStore.executeShellCommand(execution)
    if (result) {
      message.success('命令已提交执行')
      // Reload execution history
      loadExecutionHistory(selectedCommand.value.id)
    }
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : '执行命令失败'
    message.error(errorMsg)
    console.error('Error executing command:', err)
  } finally {
    executing.value = false
  }
}

/**
 * Format time
 */
const formatTime = (dateStr: string | null | undefined): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

/**
 * Get execution status label
 */
const getExecutionStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    pending: '等待中',
    running: '执行中',
    success: '成功',
    failed: '失败'
  }
  return map[status] || status
}

/**
 * Get execution status tag type
 */
const getExecutionStatusType = (
  status: string
): 'default' | 'warning' | 'success' | 'error' => {
  const map: Record<string, 'default' | 'warning' | 'success' | 'error'> = {
    pending: 'default',
    running: 'warning',
    success: 'success',
    failed: 'error'
  }
  return map[status] || 'default'
}
</script>

<style scoped>
/* ============ Server Group Styles ============ */
.server-group {
  margin-bottom: 0;
  border-bottom: 1px solid var(--color-border);
}

.server-group:last-of-type {
  border-bottom: none;
}

.server-header {
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  background: var(--color-bg-dark);
  border-left: 3px solid transparent;
  letter-spacing: 0.5px;
}

.cmd-count {
  color: var(--color-primary);
  font-weight: 600;
}

.commands-items {
  padding: 0;
}

/* ============ Execution Controls ============ */
.execution-controls {
  display: flex;
  gap: 12px;
}

/* ============ Execution History ============ */
.executions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.execution-record {
  padding: 12px 16px;
  border-left: 3px solid var(--color-border-light);
  background: var(--color-bg-dark);
  border-radius: 0;
  border: 1px solid var(--color-border-light);
  border-left: 3px solid var(--color-border-light);
}

.execution-record.status-success {
  border-left-color: var(--color-success);
  background: rgba(45, 134, 89, 0.05);
}

.execution-record.status-failed {
  border-left-color: var(--color-error);
  background: rgba(230, 57, 70, 0.05);
}

.execution-record.status-running {
  border-left-color: var(--color-warning);
  background: rgba(247, 127, 0, 0.05);
}

.execution-record.status-pending {
  border-left-color: var(--color-info);
  background: rgba(45, 134, 89, 0.05);
}

.exec-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.exec-time {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.exec-error {
  color: var(--color-error);
  font-size: 12px;
  margin-bottom: 8px;
  padding: 8px;
  background: rgba(230, 57, 70, 0.05);
  border-radius: 0;
}

.exec-output {
  background: var(--color-bg-card);
  padding: 8px 12px;
  border-radius: 0;
  border: 1px solid var(--color-border);
  max-height: 200px;
  overflow-y: auto;
  font-size: 11px;
  font-family: 'Courier New', monospace;
}

.exec-output code {
  font-size: 11px;
  font-family: 'Courier New', monospace;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--color-text-primary);
}
</style>
