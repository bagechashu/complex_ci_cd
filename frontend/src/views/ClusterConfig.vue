<template>
  <div class="cluster-config-page">
    <div class="content-layout">
      <!-- Left Panel: Clusters List -->
      <div class="list-panel">
        <div class="list-header">
          <div class="list-header-top">
            <h2>集群列表</h2>
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
            placeholder="搜索集群..."
          />
        </div>

        <div class="list-container">
          <div v-if="clusters.length === 0" class="empty-state">
            <p>暂无集群</p>
          </div>

          <div
            v-for="cluster in filteredClusters"
            :key="cluster.id"
            class="list-item"
            :class="{ active: selectedClusterId === cluster.id }"
            @click="selectCluster(cluster)"
          >
            <div class="list-item-header">
              <div class="cluster-name">{{ cluster.name }}</div>
              <div class="cluster-actions">
                <button
                  class="icon-btn"
                  @click.stop="deleteCluster(cluster.id)"
                  title="删除"
                >
                  🗑️
                </button>
              </div>
            </div>
            <div class="cluster-info">
              <span class="env-badge">{{ cluster.environment }}</span>
              <span class="type-badge">{{ cluster.type }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Panel: Cluster Details -->
      <div class="detail-panel">
        <div v-if="!selectedCluster" class="empty-detail">
          <p>请选择一个集群查看详情</p>
        </div>

        <div v-else class="detail-content">
          <!-- Cluster Info -->
          <div class="detail-section">
            <div class="section-header">
              <h3>集群信息</h3>
              <n-button @click="openEditClusterModal">
                编辑集群
              </n-button>
            </div>

            <div class="info-grid">
              <div class="info-item">
                <label>集群名称:</label>
                <span>{{ selectedCluster.name }}</span>
              </div>
              <div class="info-item">
                <label>环境:</label>
                <span class="env-badge">{{ selectedCluster.environment }}</span>
              </div>
              <div class="info-item">
                <label>集群类型:</label>
                <span>{{ selectedCluster.type }}</span>
              </div>
              <div class="info-item">
                <label>镜像仓库前缀:</label>
                <span class="code">{{ selectedCluster.registry_prefix }}</span>
              </div>
              <div class="info-item">
                <label>Kubernetes 连接状态:</label>
                <span :class="['status-badge', selectedCluster.k8s_connection_status]">
                  {{ getConnectionStatusLabel(selectedCluster.k8s_connection_status) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Security Notice -->
          <div class="detail-section">
            <h3>🔐 安全信息</h3>
            <div class="security-notice">
              <p><strong>Kubeconfig 是敏感信息</strong>，系统已加密存储，不显示具体内容。</p>
              <p v-if="selectedCluster.k8s_connection_status === 'connected'" class="status-ok">
                ✓ Kubernetes 集群连接正常
              </p>
              <p v-else-if="selectedCluster.k8s_connection_status === 'disconnected'" class="status-error">
                ✗ 无法连接 Kubernetes 集群，请检查 Kubeconfig 配置
              </p>
              <p v-else class="status-unknown">
                ⚠ 连接状态未知
              </p>
              <p class="help-text">要更新 Kubeconfig，请点击"编辑集群"按钮。</p>
            </div>
          </div>

          <!-- Connected Applications -->
          <div class="detail-section">
            <h3>关联应用</h3>
            <div v-if="loadingApplications" class="loading">
              <p>加载中...</p>
            </div>
            <div v-else-if="clusterApplications.length === 0" class="empty-apps">
              <p class="help-text">该集群暂无关联应用</p>
            </div>
            <div v-else class="apps-list">
              <div v-for="app in clusterApplications" :key="app.id" class="app-item">
                <div class="app-name">{{ app.name }}</div>
                <div class="app-image">{{ app.image_name }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Cluster Modal -->
    <div v-if="showClusterModal" class="modal-overlay" @click="closeClusterModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>{{ editingClusterId ? '编辑集群' : '新建集群' }}</h2>
          <button class="close-btn" @click="closeClusterModal">×</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>集群名称 *</label>
            <input
              v-model="clusterForm.name"
              type="text"
              class="form-input"
              placeholder="例如: k8s-prod 或 k8s-staging"
            />
          </div>

          <div class="form-group">
            <label>环境 *</label>
            <select v-model="clusterForm.environment" class="form-input">
              <option value="">-- 请选择环境 --</option>
              <option value="dev">开发环境 (dev)</option>
              <option value="staging">预发布环境 (staging)</option>
              <option value="production">生产环境 (production)</option>
              <option value="testing">测试环境 (testing)</option>
            </select>
          </div>

          <div class="form-group">
            <label>集群类型 *</label>
            <select v-model="clusterForm.type" class="form-input">
              <option value="kubernetes">Kubernetes</option>
              <option value="k3s">K3s</option>
              <option value="openshift">OpenShift</option>
            </select>
          </div>

          <div class="form-group">
            <label>镜像仓库前缀 *</label>
            <input
              v-model="clusterForm.registry_prefix"
              type="text"
              class="form-input"
              placeholder="例如: docker.io/company 或 harbor.example.com/company"
            />
            <p class="help-text">
              完整镜像地址 = 前缀 + 应用镜像名 + 标签
              <br/>
              例如:
              <code>{{ clusterForm.registry_prefix }}/api-service:v1.0.0</code>
            </p>
          </div>

          <div class="form-group">
            <label>Kubeconfig 文件内容 *</label>
            <textarea
              v-model="clusterForm.kubeconfig"
              class="form-input form-textarea"
              rows="10"
              :placeholder="editingClusterId ? '输入新的 kubeconfig 内容来更新（当前值已隐藏）' : '将 kubeconfig 文件内容粘贴到这里'"
            ></textarea>
            <p class="help-text">
              <strong>⚠️ Kubeconfig 是敏感信息</strong>，系统加密存储。编辑时必须提供完整的新 kubeconfig 内容，当前值不显示。
              <br/>
              通常可以从 ~/.kube/config 获取，或从集群管理员获得。
              <br/>
              <span v-if="editingClusterId && clusterForm.kubeconfig === undefined" style="color: #ff7d00;">
                💡 提示：保存时将验证 Kubeconfig 连接性
              </span>
            </p>
          </div>
        </div>

        <div class="modal-footer">
          <n-button type="default" @click="closeClusterModal">取消</n-button>
          <n-button type="primary" @click="saveCluster">保存</n-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NButton, NDropdown } from 'naive-ui'
import type { Application, Cluster } from '@/types/api'
import { getClusters, createCluster, updateCluster, deleteCluster as apiDeleteCluster, getApplicationsByCluster } from '@/api/metadata'

// State
const searchQuery = ref('')
const selectedClusterId = ref<number | string | null>(null)
const clusterApplications = ref<Application[]>([])
const loadingApplications = ref(false)

const clusters = ref<Cluster[]>([])

// Cluster Modal
const showClusterModal = ref(false)
const editingClusterId = ref<number | string | null>(null)
const clusterForm = ref<Partial<Cluster>>({
  type: 'kubernetes'
})

// Computed
const filteredClusters = computed(() => {
  return clusters.value.filter(cluster =>
    cluster.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    (cluster.environment?.toLowerCase() ?? '').includes(searchQuery.value.toLowerCase())
  )
})

const selectedCluster = computed(() => {
  return clusters.value.find(c => c.id === selectedClusterId.value)
})

// Header Menu Options
const headerMenuOptions = computed(() => [
  {
    label: '+ 添加集群',
    key: 'add-cluster'
  }
])

// Functions
const selectCluster = (cluster: Cluster) => {
  selectedClusterId.value = cluster.id
  loadApplicationsForCluster(cluster.id)
}

const loadApplicationsForCluster = async (clusterId: number | string) => {
  loadingApplications.value = true
  try {
    clusterApplications.value = await getApplicationsByCluster(clusterId)
  } catch (error) {
    console.error('Failed to load applications for cluster:', error)
    clusterApplications.value = []
  } finally {
    loadingApplications.value = false
  }
}

const openCreateClusterModal = () => {
  editingClusterId.value = null
  clusterForm.value = { type: 'kubernetes' }
  showClusterModal.value = true
}

const handleHeaderMenuSelect = (key: string) => {
  if (key === 'add-cluster') {
    openCreateClusterModal()
  }
}

const openEditClusterModal = () => {
  if (selectedCluster.value) {
    editingClusterId.value = selectedCluster.value.id
    // Don't include kubeconfig in form to prevent showing sensitive data
    const { kubeconfig, ...rest } = selectedCluster.value as any
    clusterForm.value = { 
      ...rest,
      kubeconfig: undefined // Force empty, secure by default
    }
    showClusterModal.value = true
  }
}

const closeClusterModal = () => {
  showClusterModal.value = false
  editingClusterId.value = null
}

const saveCluster = async () => {
  if (!clusterForm.value.name?.trim()) {
    alert('请输入集群名称')
    return
  }
  if (!clusterForm.value.environment) {
    alert('请选择环境')
    return
  }
  if (!clusterForm.value.registry_prefix?.trim()) {
    alert('请输入镜像仓库前缀')
    return
  }
  if (!clusterForm.value.kubeconfig?.trim()) {
    alert('请输入 Kubeconfig 内容')
    return
  }

  try {
    if (editingClusterId.value) {
      await updateCluster(editingClusterId.value as number, clusterForm.value)
    } else {
      await createCluster(clusterForm.value)
    }
    await loadClusters()
    closeClusterModal()
  } catch (error) {
    console.error('Failed to save cluster:', error)
    alert('保存集群失败')
  }
}

const deleteCluster = async (clusterId: number | string) => {
  if (confirm('确定删除此集群吗？关联的应用配置也会被清除。')) {
    try {
      await apiDeleteCluster(clusterId as number)
      await loadClusters()
      selectedClusterId.value = null
    } catch (error) {
      console.error('Failed to delete cluster:', error)
      alert('删除集群失败')
    }
  }
}

const loadClusters = async () => {
  try {
    const data = await getClusters()
    clusters.value = data
  } catch (error) {
    console.error('Failed to load clusters:', error)
  }
}

const getConnectionStatusLabel = (status?: string): string => {
  switch (status) {
    case 'connected':
      return '✓ 已连接'
    case 'disconnected':
      return '✗ 未连接'
    default:
      return '⚠ 未知'
  }
}

// Lifecycle
onMounted(async () => {
  await loadClusters()
})
</script>

<style scoped>
.cluster-config-page {
  padding: 24px;
  height: 100vh;
  display: flex;
  flex-direction: column;
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
  border-radius: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.list-header {
  padding: 16px;
  border-bottom: 1px solid #eee;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.list-header h2 {
  margin: 0;
  font-size: 16px;
}

.header-menu {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.menu-btn {
  font-size: 20px !important;
  padding: 4px 8px !important;
  min-width: auto !important;
  font-weight: bold !important;
  letter-spacing: 2px !important;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 0;
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

.cluster-name {
  font-weight: 600;
  color: #1a1a1a;
}

.cluster-actions {
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

.cluster-info {
  display: flex;
  gap: 8px;
  font-size: 12px;
}

.env-badge {
  background: #e6f7ff;
  color: #1890ff;
  padding: 2px 8px;
  border-radius: 0;
  font-weight: 600;
}

.type-badge {
  background: #f0f0f0;
  color: #666;
  padding: 2px 8px;
  border-radius: 0;
}

/* Detail Panel */
.detail-panel {
  background: white;
  border-radius: 0;
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
  border-radius: 0;
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
  border-radius: 0;
}

.kubeconfig-display {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 0;
  padding: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.kubeconfig-display pre {
  margin: 0;
  font-size: 11px;
  line-height: 1.5;
}

.help-text {
  font-size: 12px;
  color: #999;
  margin: 0;
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
  border-radius: 0;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  max-width: 600px;
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
  border-radius: 0;
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

/* Status Badge */
.status-badge {
  padding: 4px 12px;
  border-radius: 0;
  font-weight: 600;
  display: inline-block;
  font-size: 14px;
}

.status-badge.connected {
  background: #f0f9ff;
  color: #22863a;
  border: 1px solid #22863a;
}

.status-badge.disconnected {
  background: #fff5f5;
  color: #cb2431;
  border: 1px solid #cb2431;
}

.status-badge.unknown {
  background: #fffbea;
  color: #d79a3a;
  border: 1px solid #d79a3a;
}

/* Security Notice */
.security-notice {
  background: #fef5e7;
  border: 1px solid #f9e79f;
  border-radius: 0;
  padding: 12px;
}

.security-notice p {
  margin: 8px 0;
  font-size: 14px;
  line-height: 1.5;
}

.security-notice .status-ok {
  color: #22863a;
}

.security-notice .status-error {
  color: #cb2431;
}

.security-notice .status-unknown {
  color: #d79a3a;
}

.security-notice .help-text {
  margin-top: 12px;
}

/* Applications List */
.loading {
  padding: 12px;
  color: #666;
  font-size: 14px;
  text-align: center;
}

.empty-apps {
  padding: 12px 0;
}

.apps-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.app-item {
  padding: 12px;
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 0;
  transition: all 0.2s ease;
}

.app-item:hover {
  background: #f9f9f9;
  border-color: #1890ff;
  box-shadow: 0 1px 4px rgba(24, 144, 255, 0.1);
}

.app-name {
  font-weight: 600;
  color: #1a1a1a;
  font-size: 14px;
  margin-bottom: 4px;
}

.app-image {
  font-size: 12px;
  color: #666;
  font-family: 'Courier New', monospace;
  word-break: break-all;
}

@media (max-width: 1200px) {
  .content-layout {
    grid-template-columns: 1fr;
  }
}
</style>
