<template>
  <div class="execution-history-page">
    <n-card size="large">
      <!-- Filters -->
      <div class="filters">
        <n-space>
          <n-select
            v-model:value="selectedTaskFilter"
            :options="taskFilterOptions"
            placeholder="筛选任务"
            clearable
            style="width: 150px"
          />
          <n-select
            v-model:value="selectedStatusFilter"
            :options="statusFilterOptions"
            placeholder="筛选状态"
            clearable
            style="width: 150px"
          />
          <n-select
            v-model:value="selectedServerFilter"
            :options="serverFilterOptions"
            placeholder="筛选服务器"
            clearable
            style="width: 150px"
          />
          <n-button @click="loadExecutions">刷新</n-button>
        </n-space>
      </div>

      <!-- Table -->
      <n-data-table
        :columns="columns"
        :data="paginatedExecutions"
        :loading="false"
        :pagination="false"
        :bordered="false"
        :single-line="false"
      />

      <!-- Pagination -->
      <div class="pagination">
        <n-pagination
          v-model:page="currentPage"
          :page-size="pageSize"
          :item-count="totalCount"
          @update:page="handlePageChange"
        />
      </div>
    </n-card>

    <!-- Execution Detail Modal -->
    <n-modal
      v-model:show="showDetailModal"
      title="执行详情"
      preset="dialog"
      size="large"
      :show-icon="false"
    >
      <div v-if="selectedExecution" class="detail-content">
        <n-descriptions :columns="2" size="small">
          <n-descriptions-item label="任务名称">
            {{ selectedExecution.task_name }}
          </n-descriptions-item>
          <n-descriptions-item label="服务器">
            {{ selectedExecution.server_name }}
          </n-descriptions-item>
          <n-descriptions-item label="命令">
            <code>{{ selectedExecution.command }}</code>
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-tag :type="getStatusType(selectedExecution.status)">
              {{ statusText(selectedExecution.status) }}
            </n-tag>
          </n-descriptions-item>
        </n-descriptions>

        <n-divider />
        <h4>执行时间</h4>
        <div class="time-info">
          <div>
            <span class="time-label">开始时间:</span>
            <span>{{ selectedExecution.started_at ? formatDateTime(selectedExecution.started_at) : '-' }}</span>
          </div>
          <div>
            <span class="time-label">结束时间:</span>
            <span>{{ selectedExecution.completed_at ? formatDateTime(selectedExecution.completed_at) : '未完成' }}</span>
          </div>
          <div v-if="selectedExecution.started_at && selectedExecution.completed_at">
            <span class="time-label">耗时:</span>
            <span>{{ formatDuration(calcMs(selectedExecution.started_at, selectedExecution.completed_at)) }}</span>
          </div>
        </div>

        <n-divider v-if="selectedExecution.exit_code !== null && selectedExecution.exit_code !== undefined" />
        <div v-if="selectedExecution.exit_code !== null && selectedExecution.exit_code !== undefined">
          <h4>退出码</h4>
          <div class="exit-code" :class="selectedExecution.exit_code === 0 ? 'success' : 'error'">
            {{ selectedExecution.exit_code }}
          </div>
        </div>

        <n-divider v-if="selectedExecution.output" />
        <div v-if="selectedExecution.output">
          <h4>执行输出</h4>
          <pre class="output-box"><code>{{ selectedExecution.output }}</code></pre>
        </div>

        <n-divider v-if="selectedExecution.error_message" />
        <div v-if="selectedExecution.error_message">
          <h4>错误信息</h4>
          <div class="error-box">
            <pre><code>{{ selectedExecution.error_message }}</code></pre>
          </div>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { h } from 'vue'
import {
  NCard,
  NSpace,
  NSelect,
  NButton,
  NDataTable,
  NPagination,
  NTag,
  NModal,
  NDescriptions,
  NDescriptionsItem,
  NDivider
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
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

// ============ Pagination State ============
const currentPage = ref(1)
const pageSize = 20

// ============ Filter State ============
const selectedTaskFilter = ref<number | null>(null)
const selectedStatusFilter = ref<string | null>(null)
const selectedServerFilter = ref<number | null>(null)
const showDetailModal = ref(false)
const selectedExecution = ref<ShellTaskExecution | null>(null)

let refreshInterval: ReturnType<typeof setInterval> | null = null

// ============ Filter Options ============
const taskFilterOptions = computed(() =>
  tasks.value.map(task => ({
    label: task.name,
    value: task.id
  }))
)

const statusFilterOptions = computed(() => [
  { label: '等待中', value: 'pending' },
  { label: '执行中', value: 'running' },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' }
])

const serverFilterOptions = computed(() =>
  servers.value.map(server => ({
    label: server.name,
    value: server.id
  }))
)

// ============ Filtered Data ============
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

// ============ Paginated Data ============
const totalCount = computed(() => filteredExecutions.value.length)

const paginatedExecutions = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return filteredExecutions.value.slice(start, end)
})

// ============ Table Columns ============
const columns = computed<DataTableColumns<ShellTaskExecution>>(() => [
  {
    title: 'ID',
    key: 'id',
    width: 80,
    render: (row) => `#${row.id}`
  },
  {
    title: '任务名称',
    key: 'task_name',
    width: 150,
    ellipsis: true
  },
  {
    title: '服务器',
    key: 'server_name',
    width: 150,
    ellipsis: true
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) =>
      h(NTag, { type: getStatusType(row.status) }, () =>
        statusText(row.status)
      )
  },
  {
    title: '执行时间',
    key: 'started_at',
    width: 160,
    render: (row) => formatDateTime(row.started_at, 'MM-DD HH:mm')
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    align: 'center',
    render: (row) =>
      h(NButton, {
        type: 'info',
        size: 'small',
        onClick: () => viewExecution(row)
      }, () => '详情')
  }
])

// ============ Lifecycle ============
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

// ============ Methods ============
async function loadAllData() {
  try {
    const [execRes, tasksRes, serversRes] = await Promise.all([
      listShellTaskExecutions(1, 100),
      listShellTasks(1, 100),
      listShellServers(1, 100)
    ])
    // 响应拦截器已提取 data 字段，返回的是 PaginatedResponse<T>
    executions.value = (execRes as any)?.data || []
    tasks.value = (tasksRes as any)?.data || []
    servers.value = (serversRes as any)?.data || []
  } catch (error) {
    console.error('Failed to load data:', error)
  }
}

async function loadExecutions() {
  try {
    const offset = (currentPage.value - 1) * pageSize
    const res = await listShellTaskExecutions(currentPage.value, pageSize)
    // 响应拦截器已提取 data 字段，返回的是 PaginatedResponse<ShellTaskExecution>
    executions.value = (res as any)?.data || []
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
  }
}

function handlePageChange() {
  // 分页变化时保持当前过滤状态
}

function getStatusType(status?: string): 'default' | 'warning' | 'success' | 'error' {
  const statusTypeMap: Record<string, 'default' | 'warning' | 'success' | 'error'> = {
    pending: 'default',
    running: 'warning',
    success: 'success',
    failed: 'error'
  }
  return statusTypeMap[status || ''] || 'default'
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
/* ============ Execution History-Specific Styles ============ */

/* 时间信息网格 */
.time-info {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  font-size: 13px;
}

.time-label {
  color: var(--color-text-secondary);
  font-weight: 500;
  display: block;
  margin-bottom: 4px;
}

/* 退出码徽章 */
.exit-code {
  padding: 8px 12px;
  border-radius: 0;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  font-weight: 600;
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border);
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

/* 输出框 */
.output-box,
.error-box {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: 0;
  padding: 12px;
  overflow-x: auto;
  max-height: 300px;
  font-size: 12px;
}

.output-box code,
.error-box code {
  font-family: 'Courier New', monospace;
  color: var(--color-text-primary);
  white-space: pre-wrap;
  word-wrap: break-word;
}

.error-box {
  background-color: #fff5f5;
}

.error-box code {
  color: #d32f2f;
}
</style>
