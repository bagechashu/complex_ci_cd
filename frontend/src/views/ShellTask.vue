<template>
  <div class="shell-task-page">
    <div class="page-header">
      <h1>📋 Shell 任务</h1>
      <p class="description">创建并执行预定义的 Shell 任务</p>
      <n-button type="primary" @click="openCreateTaskModal">
        + 创建任务
      </n-button>
    </div>

    <div class="tasks-grid">
      <div v-if="tasks.length === 0" class="empty-state">
        <p>暂无任务</p>
      </div>

      <div v-for="task in tasks" :key="task.id" class="task-card">
        <div class="task-header">
          <div class="task-info">
            <h3>{{ task.name }}</h3>
            <p class="task-desc">{{ task.description || '（无描述）' }}</p>
          </div>
          <div class="task-actions">
            <n-button text size="small" @click="editTask(task)" title="编辑">
              ✏️
            </n-button>
            <n-button text size="small" @click="deleteTask(task.id)" title="删除">
              🗑️
            </n-button>
          </div>
        </div>

        <div class="task-details">
          <div class="detail-item">
            <span class="label">命令:</span>
            <span class="value code">{{ task.command || '-' }}</span>
          </div>
          <div class="detail-item">
            <span class="label">执行方式:</span>
            <span class="value">{{ task.execution_method === 'serial' ? '串行' : '并行' }}</span>
          </div>
          <div class="detail-item">
            <span class="label">目标服务器:</span>
            <span class="value">{{ task.server_ids?.length || 0 }} 台</span>
          </div>
          <div v-if="task.requires_approval" class="detail-item">
            <span class="label">⚠️ 需要审批</span>
          </div>
        </div>

        <div class="task-footer">
          <div class="execution-stats">
            <span v-if="task.executions">
              最近执行: {{ task.executions.length ? formatDate(task.executions[0].created_at) : '未执行' }}
            </span>
          </div>
          <n-button type="primary" @click="executeTask(task)">
            ▶ 执行
          </n-button>
        </div>
      </div>
    </div>

    <!-- Task Modal -->
    <div v-if="showTaskModal" class="modal-overlay" @click.self="closeTaskModal">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ editingTaskId ? '编辑任务' : '创建任务' }}</h2>
          <button class="close-btn" @click="closeTaskModal">✕</button>
        </div>

        <div class="modal-body">
          <form @submit.prevent="saveTask">
            <div class="form-group">
              <label>任务名称 *</label>
              <input
                v-model="taskForm.name"
                type="text"
                required
                class="form-input"
                placeholder="例如: 重启 Nginx"
              />
            </div>

            <div class="form-group">
              <label>描述</label>
              <textarea
                v-model="taskForm.description"
                class="form-textarea"
                placeholder="为此任务添加说明（可选）"
                rows="3"
              ></textarea>
            </div>

            <div class="form-group">
              <label>关联命令 *</label>
              <select v-model.number="taskForm.command_id" class="form-select" required>
                <option value="">-- 请选择命令 --</option>
                <optgroup v-for="(cmds, serverName) in groupedCommands" :key="serverName" :label="serverName">
                  <option v-for="cmd in cmds" :key="cmd.id" :value="cmd.id">
                    {{ cmd.description || cmd.command }}
                  </option>
                </optgroup>
              </select>
            </div>

            <div class="form-group">
              <label>目标服务器 *</label>
              <div class="servers-list">
                <label v-for="server in availableServers" :key="server.id" class="checkbox-item">
                  <input
                    type="checkbox"
                    :value="server.id"
                    :checked="taskForm.server_ids.includes(server.id)"
                    @change="toggleServer(server.id)"
                  />
                  {{ server.name }} ({{ server.host }})
                </label>
              </div>
            </div>

            <div class="form-group">
              <label>执行方式 *</label>
              <div class="radio-group">
                <label class="radio-item">
                  <input
                    type="radio"
                    value="serial"
                    v-model="taskForm.execution_method"
                    required
                  />
                  串行（逐个执行）
                </label>
                <label class="radio-item">
                  <input
                    type="radio"
                    value="parallel"
                    v-model="taskForm.execution_method"
                    required
                  />
                  并行（同时执行）
                </label>
              </div>
            </div>

            <div class="form-group">
              <label class="checkbox-item">
                <input
                  type="checkbox"
                  v-model="taskForm.requires_approval"
                />
                执行需要审批
              </label>
            </div>

            <div class="form-actions">
              <n-button type="default" @click="closeTaskModal">
                取消
              </n-button>
              <n-button type="primary" html-type="submit">
                {{ editingTaskId ? '更新' : '创建' }}
              </n-button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NButton } from 'naive-ui'
import type { ShellTask, ShellServer, ShellCommand } from '@/types/api'
import {
  listShellTasks,
  createShellTask,
  updateShellTask,
  deleteShellTask,
  listShellServers,
  listShellCommands
} from '@/api/shell'

const tasks = ref<ShellTask[]>([])
const servers = ref<ShellServer[]>([])
const commands = ref<ShellCommand[]>([])

const showTaskModal = ref(false)
const editingTaskId = ref<number | null>(null)
const taskForm = ref({
  name: '',
  description: '',
  command_id: 0,
  server_ids: [] as number[],
  execution_method: 'serial' as 'serial' | 'parallel',
  requires_approval: false
})

const availableServers = computed(() => servers.value)
const groupedCommands = computed(() => {
  const grouped: Record<string, ShellCommand[]> = {}
  commands.value.forEach(cmd => {
    const serverName = cmd.server_name || `Server ${cmd.server_id}`
    if (!grouped[serverName]) {
      grouped[serverName] = []
    }
    grouped[serverName].push(cmd)
  })
  return grouped
})

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    const [tasksRes, serversRes, commandsRes] = await Promise.all([
      listShellTasks(1, 100),
      listShellServers(1, 100),
      listShellCommands(1, 100)
    ])
    tasks.value = tasksRes.data
    servers.value = serversRes.data
    commands.value = commandsRes.data
  } catch (error) {
    console.error('Failed to load data:', error)
  }
}

function openCreateTaskModal() {
  editingTaskId.value = null
  taskForm.value = {
    name: '',
    description: '',
    command_id: 0,
    server_ids: [],
    execution_method: 'serial',
    requires_approval: false
  }
  showTaskModal.value = true
}

function closeTaskModal() {
  showTaskModal.value = false
  editingTaskId.value = null
}

function editTask(task: ShellTask) {
  editingTaskId.value = task.id
  taskForm.value = {
    name: task.name,
    description: task.description || '',
    command_id: task.command_id,
    server_ids: [...(task.server_ids || [])],
    execution_method: task.execution_method,
    requires_approval: task.requires_approval
  }
  showTaskModal.value = true
}

async function saveTask() {
  if (taskForm.value.server_ids.length === 0) {
    alert('请选择至少一个目标服务器')
    return
  }

  try {
    if (editingTaskId.value) {
      await updateShellTask(editingTaskId.value, {
        ...taskForm.value,
        command_id: taskForm.value.command_id
      } as any)
    } else {
      await createShellTask({
        ...taskForm.value,
        command_id: taskForm.value.command_id
      } as any)
    }
    await loadData()
    closeTaskModal()
  } catch (error) {
    alert('操作失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

async function deleteTask(id: number) {
  if (!confirm('确认删除该任务？')) return
  try {
    await deleteShellTask(id)
    await loadData()
  } catch (error) {
    alert('删除失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

function toggleServer(serverId: number) {
  const index = taskForm.value.server_ids.indexOf(serverId)
  if (index >= 0) {
    taskForm.value.server_ids.splice(index, 1)
  } else {
    taskForm.value.server_ids.push(serverId)
  }
}

async function executeTask(task: ShellTask) {
  // TODO: 实现任务执行逻辑
  alert(`执行任务: ${task.name}\n目前此功能正在开发中...`)
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleString('zh-CN')
}
</script>

<style scoped>
.shell-task-page {
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

.tasks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  flex: 1;
  overflow-y: auto;
  padding: 0 0 20px 0;
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #999;
}

.task-card {
  border: 1px solid #e0e0e0;
  border-radius: 0;
  background: white;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.2s;
}

.task-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #667eea;
}

.task-header {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.task-info h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.task-desc {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: #666;
}

.task-actions {
  display: flex;
  gap: 6px;
}



.task-details {
  padding: 16px;
  flex: 1;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.detail-item .label {
  color: #666;
  font-weight: 500;
}

.detail-item .value {
  color: #333;
}

.detail-item .code {
  font-family: 'Courier New', monospace;
  background-color: #f5f5f5;
  padding: 2px 6px;
  border-radius: 0;
}

.task-footer {
  padding: 12px 16px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #666;
}

.execution-stats {
  flex: 1;
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
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
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

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  font-size: 14px;
  color: #333;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 0;
  font-size: 14px;
  font-family: inherit;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 60px;
}

.servers-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 150px;
  overflow-y: auto;
  padding: 8px;
  border: 1px solid #e0e0e0;
  border-radius: 0;
  background-color: #fafafa;
}

.checkbox-item,
.radio-item {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
}

.checkbox-item input,
.radio-item input {
  cursor: pointer;
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 20px;
}


</style>
