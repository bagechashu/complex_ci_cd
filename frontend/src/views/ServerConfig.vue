<template>
  <div class="server-config-page">
    <div class="page-header">
      <h1>🖥️ 服务器配置</h1>
      <p class="description">管理 Shell 任务执行的目标服务器和可执行命令</p>
      <button class="btn-primary" @click="openCreateServerModal">
        + 添加服务器
      </button>
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
              <div class="server-actions">
                <button
                  class="icon-btn"
                  @click.stop="deleteServer(server.id)"
                  title="删除"
                >
                  🗑️
                </button>
              </div>
            </div>
            <div class="server-info">
              <span class="ip-badge">{{ server.ip_address }}</span>
              <span class="method-badge">{{ server.connection_method }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Panel: Server Details -->
      <div class="detail-panel">
        <div v-if="!selectedServer" class="empty-detail">
          <p>请选择一个服务器查看详情</p>
        </div>

        <div v-else class="detail-content">
          <!-- Server Info -->
          <div class="detail-section">
            <div class="section-header">
              <h3>服务器信息</h3>
              <button class="btn-secondary" @click="openEditServerModal">
                编辑服务器
              </button>
            </div>

            <div class="info-grid">
              <div class="info-item">
                <label>服务器名称:</label>
                <span>{{ selectedServer.name }}</span>
              </div>
              <div class="info-item">
                <label>IP 地址:</label>
                <span class="code">{{ selectedServer.ip_address }}</span>
              </div>
              <div class="info-item">
                <label>连接方式:</label>
                <span>{{ selectedServer.connection_method }}</span>
              </div>
              <div class="info-item">
                <label>SSH 端口:</label>
                <span>{{ selectedServer.ssh_port || 22 }}</span>
              </div>
            </div>
          </div>

          <!-- Credentials -->
          <div class="detail-section">
            <h3>连接凭证</h3>
            <div class="credential-box">
              <div class="info-line">
                <strong>用户名:</strong> {{ selectedServer.username }}
              </div>
              <div v-if="selectedServer.password" class="info-line">
                <strong>密码:</strong> ••••••••••••
              </div>
              <div v-if="selectedServer.private_key" class="info-line">
                <strong>私钥:</strong> (已配置)
              </div>
              <div v-if="selectedServer.connection_params" class="info-line">
                <strong>其他参数:</strong>
                <code>{{ selectedServer.connection_params }}</code>
              </div>
            </div>
          </div>

          <!-- Allowed Commands -->
          <div class="detail-section">
            <h3>允许执行的命令</h3>

            <div v-if="!selectedServer.allowed_commands || selectedServer.allowed_commands.length === 0" class="empty-state">
              <p>暂无限制（允许所有命令）</p>
            </div>

            <div class="commands-list">
              <div
                v-for="(cmd, index) in selectedServer.allowed_commands"
                :key="index"
                class="command-item"
              >
                <span>{{ cmd }}</span>
              </div>
            </div>
          </div>

          <!-- Usage Info -->
          <div class="detail-section">
            <h3>服务器状态</h3>
            <p class="help-text">
              最后连接: TODO
              <br/>
              关联任务: 0
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Server Modal -->
    <div v-if="showServerModal" class="modal-overlay" @click="closeServerModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>{{ editingServerId ? '编辑服务器' : '新建服务器' }}</h2>
          <button class="close-btn" @click="closeServerModal">×</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>服务器名称 *</label>
            <input
              v-model="serverForm.name"
              type="text"
              class="form-input"
              placeholder="例如: prod-server-1 或 应用服务器-A"
            />
          </div>

          <div class="form-group">
            <label>IP 地址 *</label>
            <input
              v-model="serverForm.ip_address"
              type="text"
              class="form-input"
              placeholder="例如: 192.168.1.10 或 server.example.com"
            />
          </div>

          <div class="form-group">
            <label>SSH 端口</label>
            <input
              v-model.number="serverForm.ssh_port"
              type="number"
              class="form-input"
              placeholder="默认: 22"
            />
          </div>

          <div class="form-group">
            <label>连接方式 *</label>
            <select v-model="serverForm.connection_method" class="form-input">
              <option value="ssh">SSH</option>
              <option value="ansible-inventory">Ansible Inventory</option>
              <option value="salt-minion">Salt Minion</option>
            </select>
          </div>

          <div class="form-group">
            <label>用户名 *</label>
            <input
              v-model="serverForm.username"
              type="text"
              class="form-input"
              placeholder="例如: root 或 ubuntu"
            />
          </div>

          <div class="form-group">
            <label>认证方式</label>
            <div class="radio-group">
              <label class="radio-item">
                <input type="radio" v-model="serverForm.auth_type" value="password" />
                <span>密码</span>
              </label>
              <label class="radio-item">
                <input type="radio" v-model="serverForm.auth_type" value="key" />
                <span>密钥</span>
              </label>
            </div>
          </div>

          <div v-if="serverForm.auth_type === 'password'" class="form-group">
            <label>密码 *</label>
            <input
              v-model="serverForm.password"
              type="password"
              class="form-input"
              placeholder="输入服务器密码"
            />
          </div>

          <div v-if="serverForm.auth_type === 'key'" class="form-group">
            <label>私钥内容 *</label>
            <textarea
              v-model="serverForm.private_key"
              class="form-input form-textarea"
              rows="6"
              placeholder="粘贴 SSH 私钥内容（BEGIN/END markers）"
            ></textarea>
          </div>

          <div class="form-group">
            <label>允许的命令（可选）</label>
            <textarea
              v-model="serverForm.allowed_commands_text"
              class="form-input form-textarea"
              rows="4"
              placeholder="输入允许执行的命令，每行一个。留空表示允许所有命令。&#10;例如:&#10;systemctl restart.*&#10;/opt/apps/.*/deploy.sh"
            ></textarea>
          </div>

          <div class="form-group">
            <label>其他连接参数</label>
            <textarea
              v-model="serverForm.connection_params"
              class="form-input form-textarea"
              rows="2"
              placeholder="例如: timeout=30,retries=3"
            ></textarea>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-secondary" @click="closeServerModal">取消</button>
          <button class="btn-primary" @click="saveServer">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface Server {
  id: number | string
  name: string
  ip_address: string
  ssh_port?: number
  connection_method: string
  username: string
  auth_type: 'password' | 'key'
  password?: string
  private_key?: string
  connection_params?: string
  allowed_commands?: string[]
}

// State
const searchQuery = ref('')
const selectedServerId = ref<number | string | null>(null)

const servers = ref<Server[]>([])

// Server Modal
const showServerModal = ref(false)
const editingServerId = ref<number | string | null>(null)
const serverForm = ref<any>({
  ssh_port: 22,
  connection_method: 'ssh',
  auth_type: 'password'
})

// Computed
const filteredServers = computed(() => {
  return servers.value.filter(server =>
    server.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    server.ip_address.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const selectedServer = computed(() => {
  return servers.value.find(s => s.id === selectedServerId.value)
})

// Functions
const selectServer = (server: Server) => {
  selectedServerId.value = server.id
}

const openCreateServerModal = () => {
  editingServerId.value = null
  serverForm.value = {
    ssh_port: 22,
    connection_method: 'ssh',
    auth_type: 'password',
    allowed_commands_text: ''
  }
  showServerModal.value = true
}

const openEditServerModal = () => {
  if (selectedServer.value) {
    editingServerId.value = selectedServer.value.id
    serverForm.value = {
      ...selectedServer.value,
      allowed_commands_text: (selectedServer.value.allowed_commands || []).join('\n')
    }
    showServerModal.value = true
  }
}

const closeServerModal = () => {
  showServerModal.value = false
  editingServerId.value = null
}

const saveServer = async () => {
  if (!serverForm.value.name?.trim()) {
    alert('请输入服务器名称')
    return
  }
  if (!serverForm.value.ip_address?.trim()) {
    alert('请输入 IP 地址')
    return
  }
  if (!serverForm.value.username?.trim()) {
    alert('请输入用户名')
    return
  }

  if (serverForm.value.auth_type === 'password') {
    if (!serverForm.value.password?.trim()) {
      alert('请输入密码')
      return
    }
  } else if (serverForm.value.auth_type === 'key') {
    if (!serverForm.value.private_key?.trim()) {
      alert('请输入私钥内容')
      return
    }
  }

  // Parse allowed commands
  const allowedCommands = serverForm.value.allowed_commands_text
    ?.split('\n')
    .map((cmd: string) => cmd.trim())
    .filter((cmd: string) => cmd) || []

  // TODO: Save server via API
  console.log('TODO: Save server', { ...serverForm.value, allowed_commands: allowedCommands })
  closeServerModal()
}

const deleteServer = async (serverId: number | string) => {
  if (confirm('确定删除此服务器吗？')) {
    // TODO: Delete server via API
    console.log('TODO: Delete server', serverId)
  }
}

const loadServers = async () => {
  // TODO: Fetch from API GET /api/v1/shell-servers
  console.log('TODO: Load shell servers')
}

// Lifecycle
onMounted(async () => {
  await loadServers()
})
</script>

<style scoped>
.server-config-page {
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

.server-name {
  font-weight: 600;
  color: #1a1a1a;
}

.server-actions {
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

.server-info {
  display: flex;
  gap: 8px;
  font-size: 12px;
}

.ip-badge {
  background: #f0f0f0;
  padding: 2px 8px;
  border-radius: 3px;
  color: #666;
  font-family: 'Courier New', monospace;
}

.method-badge {
  background: #e6f7ff;
  color: #1890ff;
  padding: 2px 8px;
  border-radius: 3px;
  font-weight: 600;
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

.credential-box {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
}

.info-line {
  margin-bottom: 8px;
  font-size: 12px;
}

.info-line strong {
  color: #666;
  min-width: 100px;
  display: inline-block;
}

.info-line code {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  color: #d32f2f;
}

.commands-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.command-item {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 12px;
  font-family: 'Courier New', monospace;
}

.help-text {
  font-size: 12px;
  color: #999;
  margin: 0;
  line-height: 1.6;
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
  font-family: 'Courier New', monospace;
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

@media (max-width: 1200px) {
  .content-layout {
    grid-template-columns: 1fr;
  }
}
</style>
