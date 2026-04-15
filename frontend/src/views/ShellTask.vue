<template>
  <div class="shell-release-page">
    <div class="page-header">
      <h1>🔧 Shell 任务</h1>
      <p class="description">创建和执行 Ansible/Salt 命令任务</p>
      <button class="btn-primary" @click="openCreateTaskModal">
        + 添加任务
      </button>
    </div>

    <div class="content-layout">
      <!-- Left Panel: Tasks List -->
      <div class="list-panel">
        <div class="list-header">
          <h2>任务列表</h2>
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索任务..."
          />
        </div>

        <div class="list-container">
          <div v-if="tasks.length === 0" class="empty-state">
            <p>暂无任务</p>
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
              <div class="task-actions">
                <button
                  class="icon-btn"
                  @click.stop="deleteTask(task.id)"
                  title="删除"
                >
                  🗑️
                </button>
              </div>
            </div>
            <div class="task-info">
              <span class="method-badge">{{ task.method }}</span>
              <!-- TODO: Show server count -->
              <span class="server-count">0 服务器</span>
            </div>
          </div>
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
              <button class="btn-secondary" @click="openEditTaskModal">
                编辑任务
              </button>
            </div>

            <div class="info-grid">
              <div class="info-item">
                <label>任务名称:</label>
                <span>{{ selectedTask.name }}</span>
              </div>
              <div class="info-item">
                <label>执行方式:</label>
                <span class="method-badge">{{ selectedTask.method }}</span>
              </div>
              <div class="info-item">
                <label>工作目录:</label>
                <span class="code">{{ selectedTask.working_directory }}</span>
              </div>
            </div>

            <div class="command-box">
              <label>执行命令:</label>
              <pre><code>{{ selectedTask.command }}</code></pre>
            </div>
          </div>

          <!-- Server Targets -->
          <div class="detail-section">
            <div class="section-header">
              <h3>目标服务器</h3>
            </div>

            <!-- TODO: Show configured servers for this task -->
            <div v-if="!selectedTask.servers || selectedTask.servers.length === 0" class="empty-state">
              <p>暂无配置服务器</p>
            </div>

            <div class="servers-list">
              <div
                v-for="server in selectedTask.servers"
                :key="server"
                class="server-item"
              >
                <span>{{ server }}</span>
              </div>
            </div>
          </div>

          <!-- Quick Execute -->
          <div class="detail-section">
            <div class="section-header">
              <h3>执行</h3>
            </div>
            <button class="btn-success btn-large" @click="openExecuteModal">
              ▶ 执行任务
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Task Modal -->
    <div v-if="showTaskModal" class="modal-overlay" @click="closeTaskModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>{{ editingTaskId ? '编辑任务' : '新建任务' }}</h2>
          <button class="close-btn" @click="closeTaskModal">×</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>任务名称 *</label>
            <input
              v-model="taskForm.name"
              type="text"
              class="form-input"
              placeholder="例如: 部署应用、重启服务"
            />
          </div>

          <div class="form-group">
            <label>执行方式 *</label>
            <div class="radio-group">
              <label class="radio-item">
                <input type="radio" v-model="taskForm.method" value="ansible" />
                <span>Ansible</span>
              </label>
              <label class="radio-item">
                <input type="radio" v-model="taskForm.method" value="salt" />
                <span>Salt</span>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label>执行目录 *</label>
            <input
              v-model="taskForm.working_directory"
              type="text"
              class="form-input"
              placeholder="例如: /opt/apps/myapp"
            />
          </div>

          <div class="form-group">
            <label>执行命令 *</label>
            <textarea
              v-model="taskForm.command"
              class="form-input form-textarea code-input"
              rows="8"
              placeholder="示例:&#10;Ansible: ansible-playbook deploy.yml -i inventory.ini&#10;Salt: salt 'minion-id' cmd.run 'systemctl restart myapp'"
            ></textarea>
          </div>

          <div class="form-group">
            <label>目标服务器/Minion</label>
            <textarea
              v-model="taskForm.servers_text"
              class="form-input form-textarea"
              rows="4"
              placeholder="输入服务器名称或 Minion ID，多个用逗号或换行分隔"
            ></textarea>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-secondary" @click="closeTaskModal">取消</button>
          <button class="btn-primary" @click="saveTask">保存</button>
        </div>
      </div>
    </div>

    <!-- Execute Task Modal -->
    <div v-if="showExecuteModal" class="modal-overlay" @click="closeExecuteModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>执行任务确认</h2>
          <button class="close-btn" @click="closeExecuteModal">×</button>
        </div>

        <div class="modal-body">
          <div class="info-box">
            <p>
              <strong>任务名称:</strong> {{ selectedTask?.name }}
            </p>
            <p>
              <strong>执行方式:</strong> {{ selectedTask?.method }}
            </p>
            <p>
              <strong>工作目录:</strong> <code>{{ selectedTask?.working_directory }}</code>
            </p>
            <p>
              <strong>执行命令:</strong>
            </p>
            <pre><code>{{ selectedTask?.command }}</code></pre>
            <p>
              <strong>目标服务器:</strong>
              {{ selectedTask?.servers?.join(', ') || '无' }}
            </p>
          </div>

          <div class="form-group">
            <label>执行理由/备注</label>
            <textarea
              v-model="executeForm.reason"
              class="form-input form-textarea"
              rows="3"
              placeholder="选项: 输入执行原因"
            ></textarea>
          </div>

          <div class="form-group">
            <label class="checkbox-item">
              <input type="checkbox" v-model="executeForm.dryRun" />
              <span>试运行（不实际执行）</span>
            </label>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-secondary" @click="closeExecuteModal">取消</button>
          <button class="btn-success" @click="confirmExecute">确认执行</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface Task {
  id: number
  name: string
  method: 'ansible' | 'salt'
  working_directory: string
  command: string
  servers?: string[]
}

// State
const searchQuery = ref('')
const selectedTaskId = ref<number | null>(null)

const tasks = ref<Task[]>([])

// Task Modal
const showTaskModal = ref(false)
const editingTaskId = ref<number | null>(null)
const taskForm = ref<any>({
  method: 'ansible'
})

// Execute Modal
const showExecuteModal = ref(false)
const executeForm = ref({
  reason: '',
  dryRun: false
})

// Computed
const filteredTasks = computed(() => {
  return tasks.value.filter(task =>
    task.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const selectedTask = computed(() => {
  return tasks.value.find(task => task.id === selectedTaskId.value)
})

// Functions
const selectTask = (task: Task) => {
  selectedTaskId.value = task.id
}

const openCreateTaskModal = () => {
  editingTaskId.value = null
  taskForm.value = {
    method: 'ansible',
    servers_text: ''
  }
  showTaskModal.value = true
}

const openEditTaskModal = () => {
  if (selectedTask.value) {
    editingTaskId.value = selectedTask.value.id
    taskForm.value = {
      ...selectedTask.value,
      servers_text: (selectedTask.value.servers || []).join('\n')
    }
    showTaskModal.value = true
  }
}

const closeTaskModal = () => {
  showTaskModal.value = false
  editingTaskId.value = null
}

const saveTask = async () => {
  if (!taskForm.value.name?.trim()) {
    alert('请输入任务名称')
    return
  }
  if (!taskForm.value.working_directory?.trim()) {
    alert('请输入执行目录')
    return
  }
  if (!taskForm.value.command?.trim()) {
    alert('请输入执行命令')
    return
  }

  // Parse servers from textarea
  const servers = taskForm.value.servers_text
    ?.split(/[\n,]/)
    .map((s: string) => s.trim())
    .filter((s: string) => s) || []

  // TODO: Save task via API
  console.log('TODO: Save task', { ...taskForm.value, servers })
  closeTaskModal()
}

const deleteTask = async (taskId: number) => {
  if (confirm('确定删除此任务吗？')) {
    // TODO: Delete task via API
    console.log('TODO: Delete task', taskId)
  }
}

const openExecuteModal = () => {
  executeForm.value = { reason: '', dryRun: false }
  showExecuteModal.value = true
}

const closeExecuteModal = () => {
  showExecuteModal.value = false
}

const confirmExecute = async () => {
  // TODO: Execute task via API
  console.log('TODO: Execute task', {
    taskId: selectedTaskId.value,
    reason: executeForm.value.reason,
    dryRun: executeForm.value.dryRun
  })
  closeExecuteModal()
}

const loadTasks = async () => {
  // TODO: Fetch from API GET /api/v1/shell-tasks
  console.log('TODO: Load shell tasks')
}

// Lifecycle
onMounted(async () => {
  await loadTasks()
})
</script>

<style scoped>
.shell-release-page {
  padding: 24px;
  height: 100vh;
  display: flex;
  flex-direction: column;
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

.description {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.btn-primary {
  padding: 10px 16px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-primary:hover {
  background: #40a9ff;
}

.content-layout {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 24px;
  flex: 1;
  overflow: hidden;
}

/* List Panel */
.list-panel {
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.list-header {
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.list-header h2 {
  margin: 0 0 12px 0;
  font-size: 16px;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.list-container {
  flex: 1;
  overflow-y: auto;
}

.empty-state {
  padding: 32px 16px;
  text-align: center;
  color: #999;
}

.list-item {
  padding: 12px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.2s;
}

.list-item:hover {
  background: #f9f9f9;
}

.list-item.active {
  background: #e6f7ff;
  border-left: 3px solid #1890ff;
}

.list-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.task-name {
  font-weight: 600;
  color: #1a1a1a;
}

.task-actions {
  display: flex;
  gap: 4px;
}

.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 4px;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.icon-btn:hover {
  opacity: 1;
}

.task-info {
  display: flex;
  gap: 8px;
  font-size: 12px;
}

.method-badge {
  background: #f0f0f0;
  padding: 2px 8px;
  border-radius: 3px;
  color: #666;
  font-weight: 600;
}

.server-count {
  color: #999;
}

/* Detail Panel */
.detail-panel {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow-y: auto;
  padding: 24px;
}

.empty-detail {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #999;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.detail-section {
  padding: 16px;
  background: #f9f9f9;
  border-radius: 6px;
  border: 1px solid #f0f0f0;
}

.detail-section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
}

.btn-secondary {
  padding: 6px 12px;
  background: white;
  color: #1a1a1a;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.btn-secondary:hover {
  background: #f5f5f5;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.info-item label {
  font-weight: 600;
  color: #666;
  font-size: 12px;
}

.info-item span {
  margin-top: 4px;
  color: #1a1a1a;
}

.code {
  font-family: 'Courier New', monospace;
  background: white;
  padding: 2px 6px;
  border-radius: 3px;
}

.command-box {
  margin-top: 12px;
}

.command-box label {
  font-weight: 600;
  color: #666;
  font-size: 12px;
  display: block;
  margin-bottom: 8px;
}

.command-box pre {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
  margin: 0;
  overflow-x: auto;
}

.command-box code {
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.servers-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.server-item {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 12px;
}

.btn-success {
  padding: 10px 24px;
  background: #52c41a;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-success:hover {
  background: #73d13d;
}

.btn-large {
  width: 100%;
  padding: 12px 24px;
  font-size: 16px;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  padding: 16px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  padding: 24px;
}

.modal-footer {
  padding: 16px;
  border-top: 1px solid #eee;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #1890ff;
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.1);
}

.form-textarea {
  resize: vertical;
}

.code-input {
  font-family: 'Courier New', monospace;
  background: #f5f5f5;
}

.radio-group {
  display: flex;
  gap: 16px;
}

.radio-item,
.checkbox-item {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.radio-item input,
.checkbox-item input {
  margin-right: 8px;
  cursor: pointer;
}

.radio-item span,
.checkbox-item span {
  font-size: 14px;
}

.info-box {
  background: #f0f8ff;
  border: 1px solid #b3e5fc;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 16px;
  font-size: 14px;
}

.info-box p {
  margin: 6px 0;
}

.info-box code {
  background: white;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  color: #d32f2f;
}

.info-box pre {
  background: white;
  border: 1px solid #b3e5fc;
  border-radius: 3px;
  padding: 8px;
  margin: 8px 0;
  overflow-x: auto;
}

.info-box code {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
}

@media (max-width: 1200px) {
  .content-layout {
    grid-template-columns: 1fr;
  }
}
</style>
