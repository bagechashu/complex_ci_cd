<template>
  <div class="execution-history-page">
    <div class="page-header">
      <h1>📊 执行历史</h1>
      <p class="description">查看 Shell 执行记录和输出结果</p>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>任务筛选:</label>
        <select v-model.number="selectedTaskFilter" class="filter-select">
          <option value="">-- 全部任务 --</option>
          <option v-for="task in tasks" :key="task.id" :value="task.id">
            {{ task.name }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label>状态筛选:</label>
        <select v-model="selectedStatusFilter" class="filter-select">
          <option value="">-- 全部状态 --</option>
          <option value="pending">等待中</option>
          <option value="running">执行中</option>
          <option value="success">成功</option>
          <option value="failed">失败</option>
        </select>
      </div>

      <div class="filter-group">
        <label>服务器筛选:</label>
        <select v-model.number="selectedServerFilter" class="filter-select">
          <option value="">-- 全部服务器 --</option>
          <option v-for="server in servers" :key="server.id" :value="server.id">
            {{ server.name }}
          </option>
        </select>
      </div>

      <n-button type="default" size="small" @click="loadExecutions" title="刷新">
        🔄 刷新
      </n-button>
    </div>

    <div class="executions-container">
      <div v-if="filteredExecutions.length === 0" class="empty-state">
        <p>暂无执行记录</p>
      </div>

      <div v-else class="executions-table">
        <div class="table-header">
          <div class="col-id">ID</div>
          <div class="col-task">任务名称</div>
          <div class="col-server">服务器</div>
          <div class="col-status">状态</div>
          <div class="col-time">执行时间</div>
          <div class="col-actions">操作</div>
        </div>

        <div v-for="execution in filteredExecutions" :key="execution.id" class="table-row">
          <div class="col-id">#{{ execution.id }}</div>
          <div class="col-task">{{ truncateString(execution.task_name, 40) }}</div>
          <div class="col-server">{{ truncateString(execution.server_name, 40) }}</div>
          <div class="col-status">
            <span class="status-badge" :class="`status-${execution.status}`">
              {{ statusText(execution.status) }}
            </span>
          </div>
          <div class="col-time">
            {{ execution.started_at ? formatDate(execution.started_at) : '-' }}
          </div>
          <div class="col-actions">
            <n-button type="default" size="small" @click="viewExecution(execution)">
              查看详情
            </n-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Execution Detail Modal -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="closeDetailModal">
      <div class="modal modal-large">
        <div class="modal-header">
          <div>
            <h2>执行详情 #{{ selectedExecution?.id }}</h2>
            <p class="detail-subtitle">{{ selectedExecution?.task_name }}</p>
          </div>
          <button class="close-btn" @click="closeDetailModal">✕</button>
        </div>

        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-item">
              <label>任务名称:</label>
              <span>{{ selectedExecution?.task_name || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>服务器:</label>
              <span>{{ selectedExecution?.server_name || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>命令:</label>
              <span class="code">{{ selectedExecution?.command || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>状态:</label>
              <span class="status-badge" :class="`status-${selectedExecution?.status}`">
                {{ statusText(selectedExecution?.status) }}
              </span>
            </div>
          </div>

          <div class="detail-section">
            <h4>执行时间</h4>
            <div class="time-info">
              <div>
                <span class="time-label">开始时间:</span>
                <span>{{ selectedExecution?.started_at ? formatDateTime(selectedExecution.started_at) : '-' }}</span>
              </div>
              <div>
                <span class="time-label">结束时间:</span>
                <span>{{ selectedExecution?.completed_at ? formatDateTime(selectedExecution.completed_at) : '未完成' }}</span>
              </div>
              <div v-if="selectedExecution?.started_at && selectedExecution?.completed_at">
                <span class="time-label">耗时:</span>
                <span>{{ formatDuration(calcMs(selectedExecution.started_at, selectedExecution.completed_at)) }}</span>
              </div>
            </div>
          </div>

          <div v-if="selectedExecution?.exit_code !== null && selectedExecution?.exit_code !== undefined" class="detail-section">
            <h4>退出码</h4>
            <div class="exit-code" :class="selectedExecution.exit_code === 0 ? 'success' : 'error'">
              {{ selectedExecution.exit_code }}
            </div>
          </div>

          <div v-if="selectedExecution?.output" class="detail-section">
            <h4>执行输出</h4>
            <pre class="output-box"><code>{{ selectedExecution.output }}</code></pre>
          </div>

          <div v-if="selectedExecution?.error_message" class="detail-section">
            <h4>错误信息</h4>
            <div class="error-box">
              <pre><code>{{ selectedExecution.error_message }}</code></pre>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <n-button type="default" @click="closeDetailModal">
            关闭
          </n-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { NButton } from 'naive-ui'
import type { ShellTaskExecution, ShellTask, ShellServer } from '@/types/api'
import { formatDate, formatDateTime, truncateString, calculateDuration as calcMs, formatDuration } from '@/utils/format'
import {
  listShellTaskExecutions,
  getShellTaskExecution,
  listShellTasks,
  listShellServers
} from '@/api/shell'

const executions = ref<ShellTaskExecution[]>([])
const tasks = ref<ShellTask[]>([])
const servers = ref<ShellServer[]>([])

const selectedTaskFilter = ref<number | ''>('')
const selectedStatusFilter = ref('')
const selectedServerFilter = ref<number | ''>('')
const showDetailModal = ref(false)
const selectedExecution = ref<ShellTaskExecution | null>(null)

let refreshInterval: ReturnType<typeof setInterval> | null = null

const filteredExecutions = computed(() => {
  return executions.value.filter(exec => {
    if (selectedTaskFilter.value && exec.task_id !== selectedTaskFilter.value) {
      return false
    }
    if (selectedStatusFilter.value && exec.status !== selectedStatusFilter.value) {
      return false
    }
    if (selectedServerFilter.value && exec.server_id !== selectedServerFilter.value) {
      return false
    }
    return true
  })
})

onMounted(async () => {
  await loadAllData()
  // 自动刷新，每 5 秒
  refreshInterval = setInterval(async () => {
    await loadExecutions()
  }, 5000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

async function loadAllData() {
  try {
    const [execRes, tasksRes, serversRes] = await Promise.all([
      listShellTaskExecutions(1, 100),
      listShellTasks(1, 100),
      listShellServers(1, 100)
    ])
    executions.value = execRes.data
    tasks.value = tasksRes.data
    servers.value = serversRes.data
  } catch (error) {
    console.error('Failed to load data:', error)
  }
}

async function loadExecutions() {
  try {
    const res = await listShellTaskExecutions(1, 100)
    executions.value = res.data
  } catch (error) {
    console.error('Failed to load executions:', error)
  }
}

async function viewExecution(execution: ShellTaskExecution) {
  try {
    selectedExecution.value = await getShellTaskExecution(execution.id)
    showDetailModal.value = true
  } catch (error) {
    console.error('Failed to load execution details:', error)
    alert('加载详情失败')
  }
}

function closeDetailModal() {
  showDetailModal.value = false
  selectedExecution.value = null
}

function statusText(status?: string): string {
  const statusMap: Record<string, string> = {
    pending: '等待中',
    running: '执行中',
    success: '成功',
    failed: '失败'
  }
  return statusMap[status || ''] || status || '-'
}

</script>

<style scoped>
.execution-history-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 20px;
}

.page-header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  color: #1a1a1a;
}

.page-header .description {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.filters-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
  padding: 16px;
  background: white;
  border-radius: 0;
  border: 1px solid #e0e0e0;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-group label {
  font-weight: 500;
  color: #666;
  font-size: 14px;
}

.filter-select {
  padding: 6px 10px;
  border: 1px solid #e0e0e0;
  border-radius: 0;
  font-size: 14px;
  min-width: 150px;
}



.executions-container {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 0;
  border: 1px solid #e0e0e0;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #999;
  font-size: 16px;
}

.executions-table {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 60px 1fr 1fr 100px 150px 100px;
  gap: 12px;
  padding: 12px 16px;
  background-color: #f5f5f5;
  border-bottom: 2px solid #e0e0e0;
  font-weight: 600;
  font-size: 13px;
  color: #666;
  flex-shrink: 0;
}

.table-row {
  display: grid;
  grid-template-columns: 60px 1fr 1fr 100px 150px 100px;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  align-items: center;
  font-size: 13px;
  transition: background-color 0.2s;
}

.table-row:hover {
  background-color: #fafafa;
}

.col-id {
  color: #999;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.col-task,
.col-server {
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-status {
  display: flex;
  justify-content: center;
}

.status-badge {
  padding: 3px 8px;
  border-radius: 0;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.status-pending {
  background-color: #e0e0e0;
  color: #666;
}

.status-running {
  background-color: #fff3cd;
  color: #856404;
}

.status-success {
  background-color: #d4edda;
  color: #155724;
}

.status-failed {
  background-color: #f8d7da;
  color: #721c24;
}

.col-time {
  color: #666;
  font-size: 12px;
}

.col-actions {
  display: flex;
  justify-content: center;
}



.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 0;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-large {
  max-width: 800px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
}

.detail-subtitle {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: #666;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #999;
}

.modal-body {
  padding: 20px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  flex-direction: column;
}

.detail-item label {
  font-weight: 600;
  color: #666;
  font-size: 12px;
  margin-bottom: 4px;
}

.detail-item span {
  color: #333;
  font-size: 14px;
}

.detail-item .code {
  font-family: 'Courier New', monospace;
  background-color: #f5f5f5;
  padding: 4px 8px;
  border-radius: 0;
  font-size: 12px;
}

.detail-section {
  margin-bottom: 20px;
  padding: 16px;
  background-color: #fafafa;
  border-radius: 0;
  border: 1px solid #f0f0f0;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #333;
}

.time-info {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  font-size: 13px;
}

.time-label {
  color: #666;
  font-weight: 500;
  display: block;
  margin-bottom: 4px;
}

.exit-code {
  padding: 8px 12px;
  border-radius: 0;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  font-weight: 600;
  background-color: white;
  border: 1px solid #e0e0e0;
  display: inline-block;
}

.exit-code.success {
  background-color: #d4edda;
  color: #155724;
  border-color: #c3e6cb;
}

.exit-code.error {
  background-color: #f8d7da;
  color: #721c24;
  border-color: #f5c6cb;
}

.output-box,
.error-box {
  background-color: white;
  border: 1px solid #e0e0e0;
  border-radius: 0;
  padding: 12px;
  overflow-x: auto;
  max-height: 300px;
  font-size: 12px;
}

.output-box code,
.error-box code {
  font-family: 'Courier New', monospace;
  color: #333;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.error-box {
  background-color: #fff5f5;
}

.error-box code {
  color: #d32f2f;
}

.modal-footer {
  padding: 16px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  justify-content: flex-end;
}
</style>
