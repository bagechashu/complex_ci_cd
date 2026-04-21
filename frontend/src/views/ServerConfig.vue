<template>
  <div class="server-config-page">
    <div class="page-header">
      <h1>🖥️ 服务器配置</h1>
      <p class="description">管理 Shell 执行的目标服务器和可执行命令</p>
      <n-button type="primary" @click="openCreateServerModal">
        + 添加服务器
      </n-button>
    </div>

    <div class="content-layout">
      <!-- Left Panel: Servers List -->
      <div class="list-panel">
        <div class="list-header">
          <h2>服务器列表</h2>
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索服务器..."
          />
        </div>

        <div class="list-container">
          <div v-if="servers.length === 0" class="empty-state">
            <p>暂无服务器</p>
          </div>

          <div
            v-for="server in filteredServers"
            :key="server.id"
            class="list-item"
            :class="{ active: selectedServerId === server.id }"
            @click="selectServer(server)"
          >
            <div class="list-item-header">
              <div class="server-name">{{ server.name }}</div>
              <div class="status-badge" :class="`status-${server.status}`">
                {{ server.status }}
              </div>
            </div>
            <div class="server-info">
              <span>{{ server.host }}:{{ server.port }}</span>
              <span class="auth-type">{{ server.auth_type }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Panel: Server Details and Commands -->
      <div class="detail-panel">
        <div v-if="!selectedServer" class="empty-detail">
          <p>请选择一个服务器查看详情和管理命令</p>
        </div>

        <div v-else class="detail-content">
          <!-- Server Info -->
          <div class="detail-section">
            <div class="section-header">
              <h3>服务器信息</h3>
              <div class="header-actions">
                <n-button @click="openEditServerModal">
                  编辑
                </n-button>
                <n-button type="error" @click="deleteServer(selectedServer.id)">
                  删除
                </n-button>
              </div>
            </div>

            <div class="info-grid">
              <div class="info-item">
                <label>服务器名称:</label>
                <span>{{ selectedServer.name }}</span>
              </div>
              <div class="info-item">
                <label>主机地址:</label>
                <span class="code">{{ selectedServer.host }}:{{ selectedServer.port }}</span>
              </div>
              <div class="info-item">
                <label>用户名:</label>
                <span>{{ selectedServer.username }}</span>
              </div>
              <div class="info-item">
                <label>认证方式:</label>
                <span>{{ selectedServer.auth_type === 'password' ? '密码' : '私钥' }}</span>
              </div>
              <div class="info-item">
                <label>状态:</label>
                <span class="status-badge" :class="`status-${selectedServer.status}`">
                  {{ selectedServer.status }}
                </span>
              </div>
              <div class="info-item">
                <label>最后连接:</label>
                <span>{{ selectedServer.last_connected ? formatDate(selectedServer.last_connected) : '未连接' }}</span>
              </div>
            </div>
          </div>

          <!-- Commands Management -->
          <div class="detail-section">
            <div class="section-header">
              <h3>允许执行的命令</h3>
              <n-button type="primary" @click="openCreateCommandModal">
                + 添加命令
              </n-button>
            </div>

            <div v-if="commands.length === 0" class="empty-commands">
              <p>该服务器暂无命令配置</p>
            </div>

            <div v-else class="commands-list">
              <div v-for="cmd in commands" :key="cmd.id" class="command-item">
                <div class="command-header">
                  <div class="command-info">
                    <div class="command-text">{{ cmd.command }}</div>
                    <div class="command-desc">{{ cmd.description || '（无描述）' }}</div>
                  </div>
                  <div class="command-actions">
                    <n-button
                      v-if="!cmd.is_published"
                      type="primary"
                      @click="publishCommand(cmd.id)"
                    >
                      发布
                    </n-button>
                    <n-button
                      v-else
                      type="warning"
                      @click="unpublishCommand(cmd.id)"
                    >
                      收回
                    </n-button>
                    <n-button
                      type="error"
                      @click="deleteCommand(cmd.id)"
                    >
                      删除
                    </n-button>
                  </div>
                </div>
                <div class="command-status">
                  <span v-if="cmd.is_published" class="published-badge">
                    ✓ 已发布
                  </span>
                  <span v-else class="unpublished-badge">
                    ○ 未发布
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Server Modal -->
    <div v-if="showServerModal" class="modal-overlay" @click.self="closeServerModal">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ editingServerId ? '编辑服务器' : '添加服务器' }}</h2>
          <n-button text type="error" @click="closeServerModal">✕</n-button>
        </div>

        <div class="modal-body">
          <form @submit.prevent="saveServer">
            <div class="form-group">
              <label>服务器名称 *</label>
              <input
                v-model="serverForm.name"
                type="text"
                required
                class="form-input"
                placeholder="例如: prod-server-01"
              />
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>主机地址 *</label>
                <input
                  v-model="serverForm.host"
                  type="text"
                  required
                  class="form-input"
                  placeholder="例如: 192.168.1.10"
                />
              </div>
              <div class="form-group">
                <label>SSH 端口 *</label>
                <input
                  v-model.number="serverForm.port"
                  type="number"
                  required
                  class="form-input"
                  placeholder="22"
                />
              </div>
            </div>

            <div class="form-group">
              <label>用户名 *</label>
              <input
                v-model="serverForm.username"
                type="text"
                required
                class="form-input"
                placeholder="例如: root"
              />
            </div>

            <div class="form-group">
              <label>认证方式 *</label>
              <select v-model="serverForm.auth_type" class="form-select" required>
                <option value="password">密码</option>
                <option value="key">私钥</option>
              </select>
            </div>

            <div class="form-group">
              <label>
                {{ serverForm.auth_type === 'password' ? '密码' : '私钥文本' }}
                {{ editingServerId ? '（留空保持不变）' : '' }} *
              </label>
              <textarea
                v-model="serverForm.password"
                :placeholder="
                  serverForm.auth_type === 'password'
                    ? '输入 SSH 密码'
                    : '输入私钥内容（包括 -----BEGIN PRIVATE KEY----- 等行）'
                "
                class="form-textarea"
                :required="!editingServerId"
              ></textarea>
            </div>

            <div class="form-actions">
              <n-button type="default" @click="closeServerModal">
                取消
              </n-button>
              <n-button type="primary" html-type="submit">
                {{ editingServerId ? '更新' : '创建' }}
              </n-button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Command Modal -->
    <div v-if="showCommandModal" class="modal-overlay" @click.self="closeCommandModal">
      <div class="modal">
        <div class="modal-header">
          <h2>添加命令</h2>
          <n-button text type="error" @click="closeCommandModal">✕</n-button>
        </div>

        <div class="modal-body">
          <form @submit.prevent="saveCommand">
            <div class="form-group">
              <label>命令 *</label>
              <input
                v-model="commandForm.command"
                type="text"
                required
                class="form-input"
                placeholder="例如: systemctl restart nginx"
              />
            </div>

            <div class="form-group">
              <label>描述</label>
              <textarea
                v-model="commandForm.description"
                class="form-textarea"
                placeholder="为此命令添加说明（可选）"
                rows="4"
              ></textarea>
            </div>

            <p class="form-tip">
              💡 提示：命令创建后初始状态为"未发布"，需要点击"发布"按钮后才允许在 Shell 执行中执行。
            </p>

            <div class="form-actions">
              <n-button type="default" @click="closeCommandModal">
                取消
              </n-button>
              <n-button type="primary" html-type="submit">
                创建命令
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
import type { ShellServer, ShellCommand } from '@/types/api'
import { NButton } from 'naive-ui'
import {
  listShellServers,
  createShellServer,
  updateShellServer,
  deleteShellServer,
  listShellCommands,
  createShellCommand,
  publishShellCommand,
  unpublishShellCommand,
  deleteShellCommand
} from '@/api/shell'

const servers = ref<ShellServer[]>([])
const commands = ref<ShellCommand[]>([])
const selectedServerId = ref<number | null>(null)
const selectedServer = computed(() =>
  servers.value.find(s => s.id === selectedServerId.value)
)

const searchQuery = ref('')
const filteredServers = computed(() =>
  servers.value.filter(s =>
    s.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    s.host.includes(searchQuery.value)
  )
)

const showServerModal = ref(false)
const showCommandModal = ref(false)
const editingServerId = ref<number | null>(null)
const serverForm = ref({
  name: '',
  host: '',
  port: 22,
  username: '',
  auth_type: 'password' as 'password' | 'key',
  password: ''
})
const commandForm = ref({
  command: '',
  description: ''
})

onMounted(async () => {
  await loadServers()
})

async function loadServers() {
  try {
    const res = await listShellServers(1, 100)
    servers.value = res.data
  } catch (error) {
    console.error('Failed to load servers:', error)
  }
}

async function loadCommands(serverId: number) {
  try {
    const res = await listShellCommands(1, 100, serverId)
    commands.value = res.data
  } catch (error) {
    console.error('Failed to load commands:', error)
    commands.value = []
  }
}

async function selectServer(server: ShellServer) {
  selectedServerId.value = server.id
  await loadCommands(server.id)
}

function openCreateServerModal() {
  editingServerId.value = null
  serverForm.value = {
    name: '',
    host: '',
    port: 22,
    username: '',
    auth_type: 'password',
    password: ''
  }
  showServerModal.value = true
}

function openEditServerModal() {
  if (!selectedServer.value) return
  editingServerId.value = selectedServer.value.id
  serverForm.value = {
    name: selectedServer.value.name,
    host: selectedServer.value.host,
    port: selectedServer.value.port,
    username: selectedServer.value.username,
    auth_type: selectedServer.value.auth_type,
    password: ''
  }
  showServerModal.value = true
}

function closeServerModal() {
  showServerModal.value = false
  editingServerId.value = null
}

async function saveServer() {
  try {
    if (editingServerId.value) {
      await updateShellServer(editingServerId.value, serverForm.value)
    } else {
      await createShellServer({
        ...serverForm.value,
        status: 'inactive' as const,
        last_connected: null
      })
    }
    await loadServers()
    closeServerModal()
  } catch (error) {
    alert('操作失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

async function deleteServer(id: number) {
  if (!confirm('确认删除该服务器？')) return
  try {
    await deleteShellServer(id)
    await loadServers()
    selectedServerId.value = null
  } catch (error) {
    alert('删除失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

function openCreateCommandModal() {
  commandForm.value = { command: '', description: '' }
  showCommandModal.value = true
}

function closeCommandModal() {
  showCommandModal.value = false
}

async function saveCommand() {
  if (!selectedServerId.value) return
  try {
    await createShellCommand({
      server_id: selectedServerId.value,
      ...commandForm.value,
      is_published: false
    })
    await loadCommands(selectedServerId.value)
    closeCommandModal()
  } catch (error) {
    alert('操作失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

async function publishCommand(id: number) {
  try {
    await publishShellCommand(id)
    if (selectedServerId.value) {
      await loadCommands(selectedServerId.value)
    }
  } catch (error) {
    alert('操作失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

async function unpublishCommand(id: number) {
  try {
    await unpublishShellCommand(id)
    if (selectedServerId.value) {
      await loadCommands(selectedServerId.value)
    }
  } catch (error) {
    alert('操作失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

async function deleteCommand(id: number) {
  if (!confirm('确认删除该命令？')) return
  try {
    await deleteShellCommand(id)
    if (selectedServerId.value) {
      await loadCommands(selectedServerId.value)
    }
  } catch (error) {
    alert('删除失败: ' + (error instanceof Error ? error.message : String(error)))
  }
}

function formatDate(dateString: string | null): string {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}
</script>

<style scoped>
.server-config-page {
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

.content-layout {
  display: flex;
  gap: 20px;
  flex: 1;
  overflow: hidden;
}

.list-panel,
.detail-panel {
  border: 1px solid #e0e0e0;
  border-radius: 0;
  background: white;
  display: flex;
  flex-direction: column;
}

.list-panel {
  width: 300px;
  overflow: hidden;
}

.detail-panel {
  flex: 1;
  overflow: hidden;
}

.list-header,
.section-header {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
}

.list-header h2,
.section-header h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 0;
  font-size: 14px;
}

.list-container {
  flex: 1;
  overflow-y: auto;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: #999;
}

.list-item {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.2s;
}

.list-item:hover,
.list-item.active {
  background-color: #f5f5f5;
}

.list-item.active {
  background-color: #e8f4f8;
  border-left: 3px solid #667eea;
  padding-left: 13px;
}

.list-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.server-name {
  font-weight: 500;
  color: #333;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 0;
  font-size: 12px;
  text-transform: uppercase;
}

.status-active {
  background-color: #d4edda;
  color: #155724;
}

.status-inactive {
  background-color: #f8d7da;
  color: #721c24;
}

.status-error {
  background-color: #f8d7da;
  color: #721c24;
}

.server-info {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #666;
}

.auth-type {
  background-color: #f0f0f0;
  padding: 2px 6px;
  border-radius: 0;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.empty-detail {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #999;
}

.detail-section {
  margin-bottom: 30px;
  padding-bottom: 30px;
  border-bottom: 1px solid #e0e0e0;
}

.detail-section:last-child {
  border-bottom: none;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0;
  margin: 0 0 16px 0;
  border: none;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.info-item label {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.info-item span {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.info-item .code {
  font-family: 'Courier New', monospace;
  background-color: #f5f5f5;
  padding: 4px 8px;
  border-radius: 0;
}

.empty-commands {
  padding: 20px;
  text-align: center;
  color: #999;
}

.commands-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.command-item {
  border: 1px solid #e0e0e0;
  border-radius: 0;
  padding: 12px;
  background-color: #fafafa;
}

.command-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 8px;
}

.command-info {
  flex: 1;
}

.command-text {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #333;
  background-color: white;
  padding: 6px 8px;
  border-radius: 0;
  border: 1px solid #e0e0e0;
  margin-bottom: 4px;
}

.command-desc {
  font-size: 12px;
  color: #666;
}

.command-actions {
  display: flex;
  gap: 6px;
}

.btn-publish,
.btn-unpublish,
.btn-delete {
  padding: 4px 8px;
  font-size: 12px;
  border: none;
  border-radius: 0;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-publish {
  background-color: #d4edda;
  color: #155724;
}

.btn-publish:hover {
  background-color: #c3e6cb;
}

.btn-unpublish {
  background-color: #fff3cd;
  color: #856404;
}

.btn-unpublish:hover {
  background-color: #ffeaa7;
}

.btn-delete {
  background-color: #f8d7da;
  color: #721c24;
}

.btn-delete:hover {
  background-color: #f5c6cb;
}

.command-status {
  font-size: 12px;
}

.published-badge {
  color: #155724;
  font-weight: 500;
}

.unpublished-badge {
  color: #856404;
  font-weight: 500;
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
  min-height: 80px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.form-tip {
  background-color: #e7f3ff;
  border-left: 3px solid #667eea;
  padding: 12px;
  border-radius: 0;
  font-size: 13px;
  color: #666;
  margin: 16px 0;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 20px;
}


</style>
