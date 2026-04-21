<template>
  <div class="shell-tasks-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-title">
        <h1>🐚 Shell 任务管理</h1>
        <p class="subtitle">创建和管理预定义的 Shell 执行任务（DDD 架构）</p>
      </div>
      <div class="header-actions">
        <n-button type="primary" @click="openCreateModal" :loading="createLoading">
          + 创建任务
        </n-button>
      </div>
    </div>

    <!-- Error Message -->
    <n-alert v-if="error" type="error" closable @close="clearError" class="mb-4">
      {{ error }}
    </n-alert>

    <!-- Toolbar -->
    <div class="toolbar">
      <div class="toolbar-left">
        <n-checkbox v-model:checked="selectAll" @update:checked="handleSelectAllChange">
          全选
        </n-checkbox>
        <span v-if="selectedTaskIds.length > 0" class="selection-info">
          已选择 {{ selectedTaskIds.length }} 项
        </span>
      </div>
      <div class="toolbar-right">
        <n-button
          v-if="selectedTaskIds.length > 0"
          type="error"
          @click="deleteBatch"
          :loading="deleteLoading"
        >
          批量删除
        </n-button>
      </div>
    </div>

    <!-- Tasks Table -->
    <n-card class="tasks-table">
      <n-spin :show="tasksLoading">
        <n-data-table
          :columns="columns"
          :data="shellTasks"
          :bordered="false"
          :single-line="false"
          :loading="tasksLoading"
          striped
          class="shell-tasks-table"
        />
      </n-spin>
    </n-card>

    <!-- Pagination -->
    <div class="pagination-container">
      <n-pagination
        v-model:page="currentPage"
        :page-count="pagination.totalPages"
        :page-size="pagination.pageSize"
        show-size-picker
        :page-sizes="[5, 10, 20, 50]"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
        style="text-align: right"
      />
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
      title="确认删除"
      preset="dialog"
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
import { ref, computed, onMounted, h } from 'vue'
import { useShellStore } from '@/stores/shellStore'
import type { DataTableColumn } from 'naive-ui'
import type { ShellTask } from '@/types/api'
import { formatDateTime } from '@/utils/format'
import {
  NButton,
  NTag,
  NSpace,
  NPopconfirm,
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
const currentPage = ref(1)
const selectAll = ref(false)

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
    required: true,
    message: '请选择至少一个目标服务器',
    trigger: 'change'
  }
}

// ============ Computed ============
const pagination = computed(() => shellStore.pagination)
const tasksLoading = computed(() => shellStore.tasksLoading)
const createLoading = computed(() => shellStore.createLoading)
const updateLoading = computed(() => shellStore.updateLoading)
const deleteLoading = computed(() => shellStore.deleteLoading)
const error = computed(() => shellStore.error)
const shellTasks = computed(() => shellStore.shellTasks)
const selectedTaskIds = computed(() => shellStore.selectedTaskIds)
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

// ============ Table Columns ============
const columns: DataTableColumn<ShellTask>[] = [
  {
    type: 'selection',
    width: 40,
    align: 'center'
  },
  {
    title: '任务 ID',
    key: 'id',
    width: 80,
    align: 'center',
    render: (row) => row.id
  },
  {
    title: '任务名称',
    key: 'name',
    width: 200,
    ellipsis: true,
    render: (row) => h('span', { class: 'font-semibold' }, row.name)
  },
  {
    title: '描述',
    key: 'description',
    width: 220,
    ellipsis: true,
    render: (row) =>
      row.description ? h('span', { class: 'text-gray-600' }, row.description) : h('span', { class: 'text-gray-400' }, '无')
  },
  {
    title: '执行方式',
    key: 'execution_method',
    width: 120,
    align: 'center',
    render: (row) => {
      const isSerial = row.execution_method === 'serial'
      return h(
        NTag,
        { type: isSerial ? 'warning' : 'success', round: true },
        () => isSerial ? '🔄 串行' : '⚡ 并行'
      )
    }
  },
  {
    title: '目标服务器',
    key: 'server_ids',
    width: 150,
    align: 'center',
    render: (row) =>
      h('span', { class: 'font-mono' }, `${row.server_ids?.length || 0} 台`)
  },
  {
    title: '需要审批',
    key: 'requires_approval',
    width: 100,
    align: 'center',
    render: (row) => {
      if (row.requires_approval) {
        return h(NTag, { type: 'error', round: true }, () => '⚠️ 是')
      } else {
        return h(NTag, { type: 'default', round: true }, () => '否')
      }
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    align: 'center',
    render: (row) => formatDateTime(row.created_at)
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    align: 'center',
    fixed: 'right',
    render: (row) =>
      h(
        NSpace,
        { size: 'small', justify: 'center' },
        () => [
          h(
            NButton,
            {
              text: true,
              type: 'primary',
              size: 'small',
              onClick: () => openEditModal(row)
            },
            () => '✏️ 编辑'
          ),
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleDeleteTask(row)
            },
            {
              default: () => `确定删除 "${row.name}" 吗？`,
              trigger: () =>
                h(
                  NButton,
                  {
                    text: true,
                    type: 'error',
                    size: 'small'
                  },
                  () => '🗑️ 删除'
                )
            }
          )
        ]
      )
  }
]

// ============ Methods ============

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
      message.success('任务更新成功')
      showCreateModal.value = false
    }
  } else {
    // Create mode
    const result = await shellStore.createShellTaskAction({
      name: taskForm.value.name || '',
      description: taskForm.value.description || '',
      command_id: taskForm.value.command_id || 0,
      execution_method: taskForm.value.execution_method || 'serial',
      server_ids: taskForm.value.server_ids || [],
      requires_approval: taskForm.value.requires_approval || false
    })
    if (result) {
      message.success('任务创建成功')
      showCreateModal.value = false
      // Refresh list if on first page
      if (currentPage.value === 1) {
        await shellStore.fetchShellTasks(1, pagination.value.pageSize)
      }
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
 * Delete single task
 */
const handleDeleteTask = async (task: ShellTask) => {
  const result = await shellStore.deleteShellTaskAction(task.id)
  if (result) {
    message.success(`任务 "${task.name}" 已删除`)
    await shellStore.fetchShellTasks(currentPage.value, pagination.value.pageSize)
  }
}

/**
 * Confirm delete from modal
 */
const confirmDelete = async () => {
  if (taskToDelete.value) {
    await handleDeleteTask(taskToDelete.value)
    showDeleteConfirm.value = false
    taskToDelete.value = null
  }
}

/**
 * Batch delete
 */
const deleteBatch = async () => {
  if (selectedTaskIds.value.length === 0) {
    message.warning('请先选择要删除的任务')
    return
  }

  const count = await shellStore.deleteMultipleShellTasks(selectedTaskIds.value)
  message.success(`已删除 ${count} 个任务`)
  shellStore.clearSelection()
  selectAll.value = false
  await shellStore.fetchShellTasks(currentPage.value, pagination.value.pageSize)
}

/**
 * Handle page change
 */
const handlePageChange = async (page: number) => {
  currentPage.value = page
  await shellStore.fetchShellTasks(page, pagination.value.pageSize)
}

/**
 * Handle page size change
 */
const handlePageSizeChange = async (pageSize: number) => {
  currentPage.value = 1
  await shellStore.fetchShellTasks(1, pageSize)
}

/**
 * Handle select all
 */
const handleSelectAllChange = (checked: boolean) => {
  if (checked) {
    shellStore.selectAllTasks()
  } else {
    shellStore.clearSelection()
  }
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
.shell-tasks-page {
  padding: 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.header-title h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.subtitle {
  margin: 8px 0 0 0;
  font-size: 14px;
  color: #666;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.mb-4 {
  margin-bottom: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: white;
  border-radius: 6px;
  border: 1px solid #e0e6ed;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.selection-info {
  font-size: 14px;
  color: #666;
  margin-left: 8px;
}

.tasks-table {
  margin-bottom: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
  padding: 16px;
  background: white;
  border-radius: 6px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.font-semibold {
  font-weight: 600;
}

.font-mono {
  font-family: 'Courier New', monospace;
}

.text-gray-600 {
  color: #666;
}

.text-gray-400 {
  color: #999;
}
</style>
