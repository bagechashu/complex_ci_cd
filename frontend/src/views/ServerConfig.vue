<template>
  <div class="server-config-page">
    <div class="content-layout">
      <!-- Left Panel: Servers List -->
      <div class="list-panel">
        <div class="list-header">
          <div class="list-header-top">
            <h2>服务器列表</h2>
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
import { NButton, NDropdown } from 'naive-ui'
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

const headerMenuOptions = computed(() => [
  {
    label: '+ 添加服务器',
    key: 'add-server'
  }
])

const handleHeaderMenuSelect = (key: string) => {
  if (key === 'add-server') {
    openCreateServerModal()
  }
}

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
    // 后端返回 {data: [...], page, pageSize, total, totalPages}
    servers.value = Array.isArray(res.data) ? res.data : []
  } catch (error) {
    console.error('Failed to load servers:', error)
    servers.value = []
  }
}

async function loadCommands(serverId: number) {
  try {
    const res = await listShellCommands(1, 100, serverId)
    // 后端返回 {data: [...], page, pageSize, total, totalPages}
    commands.value = Array.isArray(res.data) ? res.data : []
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
/* ============ Server-Specific Styles ============ */

/* 服务器名称 */
.server-name {
  font-weight: 500;
  color: var(--color-text-primary);
}

/* 状态徽章 */
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

/* 服务器信息 */
.server-info {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.auth-type {
  background-color: var(--color-bg-light);
  padding: 2px 6px;
  border-radius: 0;
}

/* 空命令提示 */
.empty-commands {
  padding: 20px;
  text-align: center;
  color: var(--color-text-muted);
}

/* 命令列表 */
.commands-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.command-item {
  border: 1px solid var(--color-border);
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
  color: var(--color-text-primary);
  background-color: white;
  padding: 6px 8px;
  border-radius: 0;
  border: 1px solid var(--color-border);
  margin-bottom: 4px;
}

.command-desc {
  font-size: 12px;
  color: var(--color-text-secondary);
}

/* 命令操作按钮 */
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

/* 命令状态 */
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
</style>
