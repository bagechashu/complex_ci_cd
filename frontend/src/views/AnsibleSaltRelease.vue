<template>
  <div class="ansible-salt-page">
    <div class="page-header">
      <h1>🔧 Ansible/Salt 发布</h1>
      <p class="description">执行 Ansible/Salt 命令进行应用部署（需要审批）</p>
    </div>

    <div class="content-grid">
      <!-- Left Panel: Command Execution -->
      <div class="panel command-panel">
        <div class="panel-header">
          <h2>执行命令</h2>
        </div>

        <div class="panel-body">
          <!-- Method Selection -->
          <div class="form-group">
            <label>执行方式 *</label>
            <div class="radio-group">
              <label class="radio-item">
                <input type="radio" v-model="form.method" value="ansible" />
                <span>Ansible</span>
              </label>
              <label class="radio-item">
                <input type="radio" v-model="form.method" value="salt" />
                <span>Salt</span>
              </label>
            </div>
          </div>

          <!-- Target Servers/Hosts -->
          <!-- TODO: Load available hosts based on method -->
          <div class="form-group">
            <label>目标主机/Minion *</label>
            <textarea
              v-model="form.targetHosts"
              class="form-input form-textarea"
              placeholder="输入主机名或 IP，多个用逗号或换行分隔"
              rows="4"
            ></textarea>
          </div>

          <!-- Command Template -->
          <div class="form-group">
            <label>
              <span>命令模板</span>
              <span class="optional">(可选)</span>
            </label>
            <select v-model="form.commandTemplate" class="form-input">
              <option value="">-- 自定义命令 --</option>
              <option value="deploy">部署应用</option>
              <option value="restart">重启服务</option>
              <option value="health-check">健康检查</option>
              <option value="rollback">回滚部署</option>
            </select>
          </div>

          <!-- Command Input -->
          <div class="form-group">
            <label>执行命令 *</label>
            <textarea
              v-model="form.command"
              class="form-input form-textarea code-input"
              placeholder="例如: systemctl restart myapp"
              rows="6"
            ></textarea>
            <div class="help-text">
              示例: 
              <br/>Ansible: <code>ansible-playbook deploy.yml -i inventory.ini</code>
              <br/>Salt: <code>salt 'minion-id' cmd.run 'systemctl restart myapp'</code>
            </div>
          </div>

          <!-- Execution Options -->
          <div class="form-group">
            <label class="checkbox-item">
              <input type="checkbox" v-model="form.dryRun" />
              <span>试运行（不实际执行）</span>
            </label>
          </div>

          <!-- Submit Approval Request -->
          <button class="btn-primary btn-large" @click="submitForApproval">
            提交审批
          </button>
        </div>
      </div>

      <!-- Right Panel: Approval & History -->
      <div class="panel status-panel">
        <!-- Pending Approvals -->
        <div class="section">
          <div class="section-header">
            <h3>待审批命令</h3>
            <span class="badge">{{ pendingApprovals.length }}</span>
          </div>

          <!-- TODO: Load pending approvals from API -->
          <div v-if="pendingApprovals.length > 0" class="approvals-list">
            <div
              v-for="approval in pendingApprovals"
              :key="approval.id"
              class="approval-item"
            >
              <div class="approval-header">
                <div class="approval-info">
                  <strong>{{ approval.method }}</strong>
                  <span class="status-badge pending">等待中</span>
                </div>
                <div class="approval-time">{{ formatTime(approval.created_at) }}</div>
              </div>

              <div class="approval-command">
                <code>{{ approval.command }}</code>
              </div>

              <div class="approval-actions">
                <button class="btn-small btn-success" @click="approveCommand(approval.id)">
                  ✓ 批准
                </button>
                <button class="btn-small btn-danger" @click="rejectCommand(approval.id)">
                  ✗ 拒绝
                </button>
              </div>
            </div>
          </div>

          <div v-else class="empty-state">
            <p>没有等待审批的命令</p>
          </div>
        </div>

        <!-- Execution History -->
        <div class="section">
          <div class="section-header">
            <h3>执行历史</h3>
          </div>

          <!-- TODO: Load execution history from API -->
          <div v-if="executionHistory.length > 0" class="history-list">
            <div
              v-for="history in executionHistory"
              :key="history.id"
              class="history-item"
              :class="{ [history.status]: true }"
            >
              <div class="history-header">
                <strong>{{ history.method }}</strong>
                <span class="status-badge" :class="history.status">
                  {{ statusText(history.status) }}
                </span>
              </div>

              <div class="history-command">
                <code>{{ history.command }}</code>
              </div>

              <div class="history-time">{{ formatTime(history.executed_at) }}</div>

              <!-- TODO: Show execution result -->
              <div v-if="history.result" class="history-result">
                <details>
                  <summary>执行结果</summary>
                  <pre>{{ history.result }}</pre>
                </details>
              </div>
            </div>
          </div>

          <div v-else class="empty-state">
            <p>没有执行历史</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface CommandApproval {
  id: number
  method: 'ansible' | 'salt'
  targetHosts: string
  command: string
  approvalStatus: 'pending' | 'approved' | 'rejected'
  created_at: string
}

interface ExecutionHistory {
  id: number
  method: 'ansible' | 'salt'
  command: string
  status: 'success' | 'failed' | 'pending'
  executed_at: string
  result?: string
}

// State
const form = ref({
  method: 'ansible' as 'ansible' | 'salt',
  targetHosts: '',
  commandTemplate: '',
  command: '',
  dryRun: false
})

const pendingApprovals = ref<CommandApproval[]>([])
const executionHistory = ref<ExecutionHistory[]>([])

// Functions
const submitForApproval = async () => {
  if (!form.value.command.trim()) {
    alert('请输入执行命令')
    return
  }

  if (!form.value.targetHosts.trim()) {
    alert('请输入目标主机')
    return
  }

  // TODO: Submit for approval via API POST /api/v1/command-approvals
  console.log('TODO: Submit for approval', form.value)

  // Reset form
  form.value = {
    method: 'ansible',
    targetHosts: '',
    commandTemplate: '',
    command: '',
    dryRun: false
  }
}

const approveCommand = async (approvalId: number) => {
  // TODO: Approve command via API PUT /api/v1/command-approvals/:id/approve
  console.log('TODO: Approve command', approvalId)
}

const rejectCommand = async (approvalId: number) => {
  // TODO: Reject command via API PUT /api/v1/command-approvals/:id/reject
  console.log('TODO: Reject command', approvalId)
}

const formatTime = (dateStr: string): string => {
  // TODO: Format timestamp
  return dateStr
}

const statusText = (status: string): string => {
  const statusMap: Record<string, string> = {
    success: '成功',
    failed: '失败',
    pending: '执行中'
  }
  return statusMap[status] || status
}

const loadPendingApprovals = async () => {
  // TODO: Fetch from API GET /api/v1/command-approvals?status=pending
  console.log('TODO: Load pending approvals')
}

const loadExecutionHistory = async () => {
  // TODO: Fetch from API GET /api/v1/command-approvals?status=approved,rejected
  console.log('TODO: Load execution history')
}

// Lifecycle
onMounted(async () => {
  await loadPendingApprovals()
  await loadExecutionHistory()
})
</script>

<style scoped>
.ansible-salt-page {
  padding: 24px;
}

.page-header {
  margin-bottom: 32px;
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

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.panel {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.panel-header h2 {
  margin: 0;
  font-size: 16px;
}

.panel-body {
  padding: 24px;
  max-height: 800px;
  overflow-y: auto;
}

.command-panel {
  display: flex;
  flex-direction: column;
}

.status-panel {
  display: flex;
  flex-direction: column;
}

.section {
  margin-bottom: 32px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  color: #1a1a1a;
}

.badge {
  background: #ff4d4f;
  color: white;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 600;
}

/* Form Groups */
.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #333;
}

.optional {
  font-weight: normal;
  color: #999;
  font-size: 12px;
}

.form-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: #1890ff;
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.1);
}

.form-textarea {
  resize: vertical;
  font-family: 'Courier New', monospace;
}

.code-input {
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
}

.help-text {
  margin-top: 8px;
  font-size: 12px;
  color: #666;
  line-height: 1.5;
}

.help-text code {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

.radio-group,
.checkbox-group {
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

/* Buttons */
.btn-primary,
.btn-success,
.btn-danger,
.btn-small {
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 600;
}

.btn-primary {
  background: #1890ff;
  color: white;
  padding: 10px 16px;
}

.btn-primary:hover {
  background: #40a9ff;
}

.btn-large {
  width: 100%;
  padding: 12px 24px;
  font-size: 16px;
}

.btn-small {
  padding: 6px 12px;
  font-size: 12px;
  margin-right: 8px;
}

.btn-success {
  background: #52c41a;
  color: white;
}

.btn-success:hover {
  background: #73d13d;
}

.btn-danger {
  background: #ff4d4f;
  color: white;
}

.btn-danger:hover {
  background: #ff7875;
}

/* Approval Items */
.approvals-list,
.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.approval-item,
.history-item {
  padding: 12px;
  background: #f9f9f9;
  border: 1px solid #eee;
  border-radius: 6px;
  border-left: 4px solid #faad14;
}

.approval-header,
.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.approval-info {
  display: flex;
  gap: 8px;
  align-items: center;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.pending {
  background: #fff1f0;
  color: #ff4d4f;
}

.status-badge.success {
  background: #f6ffed;
  color: #52c41a;
}

.status-badge.failed {
  background: #fff1f0;
  color: #ff4d4f;
}

.approval-time,
.history-time {
  font-size: 12px;
  color: #999;
}

.approval-command,
.history-command {
  background: white;
  padding: 8px 12px;
  border-radius: 4px;
  margin-bottom: 12px;
  font-size: 12px;
  overflow-x: auto;
}

.approval-command code,
.history-command code {
  font-family: 'Courier New', monospace;
  color: #d32f2f;
  word-break: break-all;
}

.approval-actions {
  display: flex;
  gap: 8px;
}

.history-item.success {
  border-left-color: #52c41a;
}

.history-item.failed {
  border-left-color: #ff4d4f;
}

.history-result {
  margin-top: 12px;
}

.history-result details {
  cursor: pointer;
}

.history-result summary {
  font-size: 12px;
  color: #1890ff;
  padding: 8px;
  background: #f0f8ff;
  border-radius: 3px;
}

.history-result pre {
  margin: 8px 0 0 0;
  padding: 8px;
  background: #f5f5f5;
  border-radius: 3px;
  font-size: 11px;
  overflow-x: auto;
  max-height: 200px;
}

.empty-state {
  text-align: center;
  padding: 32px 16px;
  color: #999;
}

@media (max-width: 1200px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}
</style>
