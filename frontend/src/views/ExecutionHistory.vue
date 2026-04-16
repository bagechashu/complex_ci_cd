<template>
  <div class="execution-history-page">
    <div class="page-header">
      <h1>📋 执行历史</h1>
      <p class="description">查看 Shell 任务的执行历史和结果</p>
    </div>

    <div class="content-wrapper">
      <!-- Filters -->
      <div class="filters">
        <input
          v-model="searchQuery"
          type="text"
          class="search-input"
          placeholder="搜索任务名称或服务器..."
        />

        <select v-model="filterStatus" class="filter-select">
          <option value="">所有状态</option>
          <option value="success">成功</option>
          <option value="failed">失败</option>
          <option value="running">执行中</option>
        </select>

        <select v-model="filterTimeRange" class="filter-select">
          <option value="">全部时间</option>
          <option value="1h">最近 1 小时</option>
          <option value="24h">最近 24 小时</option>
          <option value="7d">最近 7 天</option>
          <option value="30d">最近 30 天</option>
        </select>
      </div>

      <!-- History List -->
      <div class="history-container">
        <!-- TODO: Load execution history from API -->
        <div v-if="histories.length === 0" class="empty-state">
          <p>暂无执行历史</p>
        </div>

        <div class="history-grid">
          <div
            v-for="history in filteredHistories"
            :key="history.id"
            class="history-card"
            :class="{ [history.status]: true }"
          >
            <!-- Card Header -->
            <div class="card-header">
              <div class="header-left">
                <div class="status-icon">
                  <span v-if="history.status === 'success'">✓</span>
                  <span v-else-if="history.status === 'failed'">✕</span>
                  <span v-else>◐</span>
                </div>
                <div class="header-info">
                  <div class="task-name">{{ history.task_name }}</div>
                  <div class="status-text">{{ statusText(history.status) }}</div>
                </div>
              </div>
              <div class="header-time">
                {{ formatTime(history.executed_at) }}
              </div>
            </div>

            <!-- Card Body -->
            <div class="card-body">
              <div class="info-row">
                <span class="label">执行方式:</span>
                <span class="method-badge">{{ history.method }}</span>
              </div>

              <div class="info-row">
                <span class="label">服务器:</span>
                <span class="servers">{{ history.servers?.join(', ') || '无' }}</span>
              </div>

              <div class="info-row">
                <span class="label">执行人:</span>
                <span>{{ history.executed_by || '自动' }}</span>
              </div>

              <div v-if="history.reason" class="info-row">
                <span class="label">执行理由:</span>
                <span>{{ history.reason }}</span>
              </div>

              <div class="info-row">
                <span class="label">耗时:</span>
                <span>{{ formatDuration(history.duration) }}</span>
              </div>

              <!-- Command -->
              <div class="command-section">
                <label>执行命令:</label>
                <div class="command-box">
                  <pre><code>{{ history.command }}</code></pre>
                </div>
              </div>

              <!-- Result -->
              <div v-if="history.result" class="result-section">
                <details>
                  <summary>执行结果 ({{ history.status }})</summary>
                  <div class="result-box" :class="{ [history.status]: true }">
                    <pre><code>{{ history.result }}</code></pre>
                  </div>
                </details>
              </div>
            </div>

            <!-- Card Footer -->
            <div class="card-footer">
              <button class="btn-small" @click="retryExecution(history.id)">
                🔄 重新执行
              </button>
              <button class="btn-small" @click="viewDetails(history.id)">
                👁 查看详情
              </button>
              <button class="btn-small btn-danger" @click="deleteHistory(history.id)">
                🗑 删除
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="pagination">
        <button
          class="page-btn"
          @click="previousPage"
          :disabled="currentPage === 1"
        >
          ← 上一页
        </button>

        <span class="page-info">
          第 {{ currentPage }} / {{ totalPages }} 页
        </span>

        <button
          class="page-btn"
          @click="nextPage"
          :disabled="currentPage === totalPages"
        >
          下一页 →
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getExecutionHistories } from '@/api/metadata'

interface ExecutionHistory {
  id: number
  task_name: string
  method: 'ansible' | 'salt'
  servers?: string[]
  command: string
  status: 'success' | 'failed' | 'running'
  executed_at: string
  executed_by?: string
  reason?: string
  duration?: number
  result?: string
}

// State
const searchQuery = ref('')
const filterStatus = ref('')
const filterTimeRange = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

const histories = ref<ExecutionHistory[]>([])

// Computed
const filteredHistories = computed(() => {
  let filtered = histories.value

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(h =>
      h.task_name.toLowerCase().includes(query) ||
      h.servers?.some(s => s.toLowerCase().includes(query))
    )
  }

  // Status filter
  if (filterStatus.value) {
    filtered = filtered.filter(h => h.status === filterStatus.value)
  }

  // Time range filter
  if (filterTimeRange.value) {
    const now = new Date()
    let cutoffTime = new Date()

    switch (filterTimeRange.value) {
      case '1h':
        cutoffTime.setHours(cutoffTime.getHours() - 1)
        break
      case '24h':
        cutoffTime.setDate(cutoffTime.getDate() - 1)
        break
      case '7d':
        cutoffTime.setDate(cutoffTime.getDate() - 7)
        break
      case '30d':
        cutoffTime.setDate(cutoffTime.getDate() - 30)
        break
    }

    filtered = filtered.filter(h => new Date(h.executed_at) > cutoffTime)
  }

  return filtered
})

const paginatedHistories = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredHistories.value.slice(start, end)
})

const totalPages = computed(() => {
  return Math.ceil(filteredHistories.value.length / pageSize.value)
})

// Functions
const statusText = (status: string): string => {
  const map: Record<string, string> = {
    success: '成功',
    failed: '失败',
    running: '执行中'
  }
  return map[status] || status
}

const formatTime = (dateStr: string): string => {
  try {
    const date = new Date(dateStr)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch {
    return dateStr
  }
}

const formatDuration = (seconds?: number): string => {
  if (!seconds) return '--'
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`
  return `${Math.floor(seconds / 3600)}小时${Math.floor((seconds % 3600) / 60)}分`
}

const retryExecution = async (historyId: number) => {
  try {
    // Retry execution would call the execute endpoint with the same task
    console.log('Retrying execution:', historyId)
    alert('重试逻辑需要根据任务类型实现')
  } catch (error) {
    console.error('Failed to retry execution:', error)
    alert('重试执行失败')
  }
}

const viewDetails = (historyId: number) => {
  // TODO: Navigate to detailed view
  console.log('TODO: View execution details', historyId)
}

const deleteHistory = async (historyId: number) => {
  if (confirm('确定删除此执行记录吗？')) {
    try {
      // Would require a delete API endpoint
      console.log('Deleting history:', historyId)
      const index = histories.value.findIndex(h => h.id === historyId)
      if (index > -1) {
        histories.value.splice(index, 1)
      }
    } catch (error) {
      console.error('Failed to delete history:', error)
      alert('删除失败')
    }
  }
}

const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

const loadExecutionHistories = async () => {
  try {
    const data = await getExecutionHistories()
    histories.value = data
  } catch (error) {
    console.error('Failed to load execution histories:', error)
  }
}

// Lifecycle
onMounted(async () => {
  await loadExecutionHistories()
})
</script>

<style scoped>
.execution-history-page {
  padding: 24px;
  min-height: 100vh;
  background: #f5f5f5;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  color: #1a1a1a;
}

.description {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.content-wrapper {
  max-width: 1400px;
  margin: 0 auto;
}

/* Filters */
.filters {
  display: grid;
  grid-template-columns: 1fr 200px 200px;
  gap: 12px;
  margin-bottom: 24px;
}

.search-input,
.filter-select {
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  background: white;
}

.search-input:focus,
.filter-select:focus {
  outline: none;
  border-color: #1890ff;
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.1);
}

/* History Container */
.history-container {
  background: white;
  border-radius: 8px;
  padding: 24px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.history-grid {
  display: grid;
  gap: 16px;
}

.history-card {
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  background: white;
  border-left: 4px solid #999;
  transition: all 0.2s;
}

.history-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.history-card.success {
  border-left-color: #52c41a;
}

.history-card.failed {
  border-left-color: #ff4d4f;
}

.history-card.running {
  border-left-color: #1890ff;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.header-left {
  display: flex;
  gap: 12px;
  align-items: center;
  flex: 1;
}

.status-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 18px;
  color: white;
}

.history-card.success .status-icon {
  background: #52c41a;
}

.history-card.failed .status-icon {
  background: #ff4d4f;
}

.history-card.running .status-icon {
  background: #1890ff;
}

.header-info {
  flex: 1;
}

.task-name {
  font-weight: 600;
  font-size: 14px;
  color: #1a1a1a;
}

.status-text {
  font-size: 12px;
  color: #999;
}

.header-time {
  text-align: right;
  font-size: 12px;
  color: #999;
}

.card-body {
  padding: 16px;
}

.info-row {
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: 12px;
  margin-bottom: 12px;
  font-size: 13px;
}

.label {
  font-weight: 600;
  color: #666;
}

.method-badge {
  background: #f0f0f0;
  padding: 2px 8px;
  border-radius: 3px;
  display: inline-block;
  width: fit-content;
}

.servers {
  color: #1a1a1a;
  word-break: break-all;
}

.command-section {
  margin: 16px 0;
}

.command-section label {
  font-weight: 600;
  color: #666;
  font-size: 12px;
  display: block;
  margin-bottom: 8px;
}

.command-box {
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
  overflow-x: auto;
}

.command-box pre {
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
}

.command-box code {
  font-family: 'Courier New', monospace;
  color: #d32f2f;
}

.result-section {
  margin: 16px 0;
}

.result-section summary {
  font-weight: 600;
  color: #666;
  cursor: pointer;
  font-size: 12px;
  padding: 8px;
  background: #f9f9f9;
  border-radius: 3px;
}

.result-section summary:hover {
  background: #f0f0f0;
}

.result-box {
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
  margin-top: 8px;
  max-height: 300px;
  overflow-y: auto;
}

.result-box.success {
  background: #f6ffed;
  border-color: #b7eb8f;
}

.result-box.failed {
  background: #fff1f0;
  border-color: #ffccc7;
}

.result-box pre {
  margin: 0;
  font-size: 11px;
  line-height: 1.4;
  color: #1a1a1a;
}

.result-box code {
  font-family: 'Courier New', monospace;
}

.card-footer {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #f0f0f0;
  background: #fafafa;
  border-bottom-left-radius: 6px;
  border-bottom-right-radius: 6px;
}

.btn-small {
  padding: 4px 8px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
}

.btn-small:hover {
  background: #40a9ff;
}

.btn-small.btn-danger {
  background: #ff4d4f;
}

.btn-small.btn-danger:hover {
  background: #ff7875;
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
  padding: 16px;
  background: white;
  border-radius: 8px;
}

.page-btn {
  padding: 8px 16px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.page-btn:hover:not(:disabled) {
  background: #40a9ff;
}

.page-btn:disabled {
  background: #d9d9d9;
  cursor: not-allowed;
}

.page-info {
  font-size: 14px;
  color: #666;
  min-width: 100px;
  text-align: center;
}

@media (max-width: 1200px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
