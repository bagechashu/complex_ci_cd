<template>
  <div class="shell-tasks-page">
    <div class="content-layout">
      <!-- Left Panel: Shell Tasks List -->
      <div class="list-panel">
        <div class="list-header">
          <div class="list-header-top">
            <h2>任务列表</h2>
            <div class="header-menu">
              <n-dropdown trigger="click" :options="headerMenuOptions" @select="handleHeaderMenuSelect">
                <n-button text type="primary" class="menu-btn">⋮</n-button>
              </n-dropdown>
            </div>
          </div>
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索任务..."
          />
          <div class="sort-controls">
            <button
              :class="{ active: true }"
              disabled
              title="按名称排序"
            >
              名称
            </button>
            <button
              :class="{ 'order-btn': true, active: true }"
              @click="sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'"
              :title="`${sortOrder === 'asc' ? '升序' : '降序'}`"
            >
              {{ sortOrder === 'asc' ? '↑' : '↓' }}
            </button>
          </div>
        </div>

        <div class="list-container">
          <div v-if="filteredTasks.length === 0" class="empty-state">
            <p>暂无任务，点击上方"+"创建</p>
          </div>

          <div
            v-for="task in filteredTasks"
            :key="task.id"
            class="list-item"
            :class="{ active: selectedTaskId === task.id }"
            @click="selectTask(task)"
          >
            <div class="list-item-header">
              <div class="task-name">{{ task.name }}</div>
            </div>
            <div class="task-info">
              <span class="badge">{{ getExecutionMethodLabel(task.execution_method) }}</span>
              <span class="badge-server">{{ task.server_ids?.length || 0 }} 台服务器</span>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="pagination-controls">
          <button
            :disabled="currentPage === 1"
            class="pagination-btn"
            @click="previousPage"
          >
            ← 上一页
          </button>
          <span class="pagination-info">
            第 {{ currentPage }} / {{ totalPages }} 页 (共 {{ totalTasks }} 个任务)
          </span>
          <button
            :disabled="currentPage === totalPages"
            class="pagination-btn"
            @click="nextPage"
          >
            下一页 →
          </button>
        </div>
      </div>

      <!-- Right Panel: Task Details -->
      <div class="detail-panel">
        <div v-if="!selectedTask" class="empty-detail">
          <p>请选择一个任务开始配置</p>
        </div>

        <div v-else class="detail-content">
          <!-- Task Info -->
          <div class="detail-section">
            <div class="section-header">
              <h3>任务信息</h3>
              <div class="header-actions">
                <n-button @click="openEditModal(selectedTask)">
                  编辑
                </n-button>
                <n-button type="error" @click="deleteTask(selectedTask.id)">
                  删除
                </n-button>
              </div>
            </div>

            <div class="info-grid">
              <div class="info-item">
                <label>任务名称:</label>
                <span>{{ selectedTask.name }}</span>
              </div>
              <div class="info-item">
                <label>任务描述:</label>
                <span>{{ selectedTask.description || '无' }}</span>
              </div>
              <div class="info-item">
                <label>执行方式:</label>
                <span>{{ getExecutionMethodLabel(selectedTask.execution_method) }}</span>
              </div>
              <div class="info-item">
                <label>关联命令:</label>
                <span class="code">{{ getCommandName(selectedTask.command_id) }}</span>
              </div>
            </div>
          </div>

          <!-- Servers -->
          <div class="detail-section">
            <div class="section-header">
              <h3>目标服务器</h3>
            </div>

            <div v-if="selectedTask.server_ids && selectedTask.server_ids.length > 0" class="servers-list">
              <p class="servers-summary">已配置 <strong>{{ selectedTask.server_ids.length }}</strong> 台服务器</p>
              <div class="servers-grid">
                <div v-for="serverId in selectedTask.server_ids" :key="serverId" class="server-item">
                  <span class="server-badge">🖥️ {{ getServerName(serverId) }}</span>
                </div>
              </div>
            </div>

            <div v-else class="empty-state">
              <p>暂无服务器配置</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <n-modal
      v-model:show="showCreateModal"
      :title="editingTaskId ? '编辑任务' : '创建任务'"
      preset="dialog"
      @positive-click="handleSaveTask"
      @negative-click="handleCancelModal"
      :positive-text="editingTaskId ? '更新' : '创建'"
      negative-text="取消"
      :loading="createLoading || updateLoading"
    >
      <n-form
        ref="formRef"
        :model="taskForm"
        :rules="formRules"
        label-placement="left"
        label-width="120px"
      >
        <!-- Task Name -->
        <n-form-item label="任务名称" path="name">
          <n-input
            v-model:value="taskForm.name"
            placeholder="例如：重启 Nginx 服务"
            clearable
          />
        </n-form-item>

        <!-- Task Description -->
        <n-form-item label="任务描述" path="description">
          <n-input
            v-model:value="taskForm.description"
            type="textarea"
            placeholder="为此任务添加说明（可选）"
            :rows="3"
          />
        </n-form-item>

        <!-- Command Selection -->
        <n-form-item label="关联命令" path="command_id">
          <n-select
            v-model:value="taskForm.command_id"
            :options="commandOptions"
            placeholder="请选择要执行的命令"
            clearable
            filterable
          />
        </n-form-item>

        <!-- Execution Method -->
        <n-form-item label="执行方式" path="execution_method">
          <n-radio-group v-model:value="taskForm.execution_method">
            <n-space>
              <n-radio value="serial">
                <span>🔄 串行（逐个执行）</span>
              </n-radio>
              <n-radio value="parallel">
                <span>⚡ 并行（同时执行）</span>
              </n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <!-- Server Selection -->
        <n-form-item label="目标服务器" path="server_ids">
          <n-tree-select
            v-model:value="taskForm.server_ids"
            :options="serverTreeOptions"
            multiple
            cascade
            placeholder="请选择目标服务器"
            filterable
            default-expand-all
            show-path
          />
        </n-form-item>

        <!-- Approval Requirement -->
        <n-form-item label="是否需要审批" path="requires_approval">
          <n-checkbox v-model:checked="taskForm.requires_approval">
            ⚠️ 此任务执行前需要审批
          </n-checkbox>
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- Delete Confirmation -->
    <n-modal
      v-model:show="showDeleteConfirm"
      title="删除任务"
      positive-text="删除"
      negative-text="取消"
      type="error"
      @positive-click="confirmDelete"
      @negative-click="showDeleteConfirm = false"
    >
      确定要删除任务 <strong>{{ taskToDelete?.name }}</strong> 吗？此操作不可撤销。
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useShellStore } from '@/stores/shellStore'
import type { ShellTask } from '@/types/api'
import {
  NButton,
  NTag,
  NSpace,
  useMessage
} from 'naive-ui'

// ============ Stores & Composables ============
const shellStore = useShellStore()
const message = useMessage()

// ============ State ============
const showCreateModal = ref(false)
const showDeleteConfirm = ref(false)
const editingTaskId = ref<number | null>(null)
const taskToDelete = ref<ShellTask | null>(null)
const selectedTaskId = ref<number | null>(null)
const currentPage = ref(1)
const pageSize = 10
const searchQuery = ref('')
const sortOrder = ref<'asc' | 'desc'>('asc')

const taskForm = ref<Partial<ShellTask>>({
  name: '',
  description: '',
  command_id: 0,
  execution_method: 'serial',
  server_ids: [],
  requires_approval: false
})

const formRules = {
  name: {
    required: true,
    message: '请输入任务名称',
    trigger: 'blur'
  },
  command_id: {
    required: true,
    message: '请选择关联的命令',
    type: 'number',
    trigger: 'change'
  },
  server_ids: {
    validator: (rule, value) => {
      if (!Array.isArray(value) || value.length === 0) {
        return new Error('请选择至少一个目标服务器')
      }
      return true
    },
    trigger: 'change'
  }
}

// ============ Computed ============
const tasksLoading = computed(() => shellStore.tasksLoading)
const createLoading = computed(() => shellStore.createLoading)
const updateLoading = computed(() => shellStore.updateLoading)
const deleteLoading = computed(() => shellStore.deleteLoading)
const error = computed(() => shellStore.error)
const shellTasks = computed(() => shellStore.shellTasks)
const shellServers = computed(() => shellStore.shellServers)
const shellCommands = computed(() => shellStore.shellCommands)

// Command options for select
const commandOptions = computed(() => {
  return shellCommands.value.map(cmd => ({
    label: `${cmd.description || cmd.command}`,
    value: cmd.id
  }))
})

// Server tree options
const serverTreeOptions = computed(() => {
  return shellServers.value.map(server => ({
    label: `${server.name} (${server.host})`,
    key: server.id,
    value: server.id,
    children: []
  }))
})

// Header Menu Options
const headerMenuOptions = computed(() => [
  {
    label: '+ 创建任务',
    key: 'create-task'
  }
])

const handleHeaderMenuSelect = (key: string) => {
  if (key === 'create-task') {
    openCreateModal()
  }
}

// Filtered tasks
const filteredTasks = computed(() => {
  let filtered = shellTasks.value.filter(task =>
    task.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    task.description?.toLowerCase().includes(searchQuery.value.toLowerCase())
  )

  // Sort by name
  filtered.sort((a, b) => {
    const compareVal = a.name.localeCompare(b.name)
    return sortOrder.value === 'asc' ? compareVal : -compareVal
  })

  return filtered
})

// Pagination
const totalTasks = computed(() => filteredTasks.value.length)
const totalPages = computed(() => Math.ceil(totalTasks.value / pageSize))
const paginatedTasks = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return filteredTasks.value.slice(start, end)
})

// Selected task
const selectedTask = computed(() => {
  return shellTasks.value.find(task => task.id === selectedTaskId.value)
})

// ============ Methods ============

/**
 * Select task
 */
const selectTask = (task: ShellTask) => {
  selectedTaskId.value = task.id
}

/**
 * Open create modal
 */
const openCreateModal = () => {
  editingTaskId.value = null
  taskForm.value = {
    name: '',
    description: '',
    command_id: 0,
    execution_method: 'serial',
    server_ids: [],
    requires_approval: false
  }
  showCreateModal.value = true
}

/**
 * Open edit modal
 */
const openEditModal = (task: ShellTask) => {
  editingTaskId.value = task.id
  taskForm.value = {
    name: task.name,
    description: task.description,
    command_id: task.command_id,
    execution_method: task.execution_method,
    server_ids: task.server_ids,
    requires_approval: task.requires_approval
  }
  showCreateModal.value = true
}

/**
 * Save task
 */
const handleSaveTask = async () => {
  if (editingTaskId.value) {
    // Update mode
    const result = await shellStore.updateShellTaskAction(editingTaskId.value, taskForm.value)
    if (result) {
      message.success('任务已更新')
      showCreateModal.value = false
    }
  } else {
    // Create mode
    const result = await shellStore.createShellTaskAction(taskForm.value)
    if (result) {
      message.success('任务已创建')
      showCreateModal.value = false
      currentPage.value = 1
    }
  }
}

/**
 * Cancel modal
 */
const handleCancelModal = () => {
  showCreateModal.value = false
  editingTaskId.value = null
}

/**
 * Delete task
 */
const deleteTask = async (taskId: number) => {
  const task = shellTasks.value.find(t => t.id === taskId)
  if (!task) return
  taskToDelete.value = task
  showDeleteConfirm.value = true
}

/**
 * Confirm delete
 */
const confirmDelete = async () => {
  if (taskToDelete.value) {
    const result = await shellStore.deleteShellTaskAction(taskToDelete.value.id)
    if (result) {
      message.success('任务已删除')
      if (selectedTaskId.value === taskToDelete.value.id) {
        selectedTaskId.value = null
      }
      showDeleteConfirm.value = false
      taskToDelete.value = null
      // Reload list
      await shellStore.fetchShellTasks(currentPage.value, pageSize)
    }
  }
}

/**
 * Handle page change
 */
const previousPage = async () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

const nextPage = async () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

/**
 * Helper functions
 */
const getExecutionMethodLabel = (method: string): string => {
  const map: Record<string, string> = {
    serial: '🔄 串行执行',
    parallel: '⚡ 并行执行'
  }
  return map[method] || method
}

const getCommandName = (commandId: number): string => {
  const cmd = shellCommands.value.find(c => c.id === commandId)
  return cmd ? (cmd.description || cmd.command) : '未知命令'
}

const getServerName = (serverId: number): string => {
  const server = shellServers.value.find(s => s.id === serverId)
  return server ? `${server.name} (${server.host})` : `服务器 ${serverId}`
}

/**
 * Clear error
 */
const clearError = () => {
  shellStore.clearError()
}

// ============ Lifecycle ============
onMounted(async () => {
  await shellStore.initializeData()
  await shellStore.fetchShellTasks(1, 10)
})
</script>

<style scoped>
/* ============ Task-Specific Styles ============ */

/* 任务名称 */
.task-name {
  font-weight: 500;
  color: var(--color-text-primary);
}

/* 任务操作按钮 */
.task-actions {
  display: flex;
  gap: 4px;
}

/* 图标按钮 */
.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 0;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.icon-btn:hover {
  opacity: 1;
}

/* 任务信息标签 */
.task-info {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
}

/* 徽章样式 */
.badge {
  background-color: var(--color-bg-light);
  padding: 3px 8px;
  border-radius: 0;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.badge-server {
  background-color: var(--color-bg-list-active);
  color: var(--color-primary);
  padding: 3px 8px;
  border-radius: 0;
}

/* 服务器列表 */
.servers-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.servers-summary {
  font-size: 14px;
  color: var(--color-text-primary);
  margin: 0;
}

.servers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 8px;
}

.server-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background-color: var(--color-bg-light);
  border-radius: 0;
  border: 1px solid var(--color-border);
}

.server-badge {
  font-size: 12px;
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 审批信息 */
.approval-info {
  display: flex;
  gap: 12px;
}

.approval-badge {
  padding: 8px 16px;
  border-radius: 0;
  font-size: 14px;
  font-weight: 500;
}

.approval-required {
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeaa7;
}

.approval-not-required {
  background-color: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 覆盖 views.css 中的 sort-controls 按钮样式 */
.sort-controls button {
  flex: 1;
  padding: 6px 8px;
  border: 1px solid var(--color-border);
  border-radius: 0;
  background: var(--color-bg-card);
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.sort-controls button:hover {
  background-color: var(--color-bg-list-hover);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.sort-controls button.active {
  background-color: var(--color-bg-list-active);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.sort-controls button.order-btn {
  flex: 0 1 auto;
  width: 40px;
}
</style>
