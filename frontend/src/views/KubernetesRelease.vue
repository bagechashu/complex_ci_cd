<template>
  <div class="k8s-release-page">
    <div class="page-header">
      <h1>🚀 K8s 发布</h1>
      <p class="description">管理应用与集群的部署映射，发布新版本</p>
      <button class="btn-primary" @click="openCreateApplicationModal">
        + 添加应用
      </button>
    </div>

    <!-- Error Message -->
    <div v-if="errorMessage" class="error-banner">
      <span>⚠️ {{ errorMessage }}</span>
      <button class="close-banner-btn" @click="errorMessage = null">×</button>
    </div>

    <!-- Loading Indicator -->
    <div v-if="isLoading" class="loading-overlay">
      <div class="loading-spinner">
        <div class="spinner"></div>
        <p>加载数据中...</p>
      </div>
    </div>

    <div class="content-layout">
      <!-- Left Panel: Applications List -->
      <div class="list-panel">
        <div class="list-header">
          <h2>应用列表</h2>
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索应用..."
          />
          <div class="sort-controls">
            <button
              :class="{ active: sortBy === 'name' }"
              @click="sortBy = 'name'"
              title="按名称排序"
            >
              名称
            </button>
            <button
              :class="{ active: sortBy === 'config-count' }"
              @click="sortBy = 'config-count'"
              title="按集群数排序"
            >
              集群数
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
          <div v-if="applications.length === 0" class="empty-state">
            <p>暂无应用，点击上方"添加应用"创建</p>
          </div>

          <div
            v-for="app in filteredApplications"
            :key="app.id"
            class="list-item"
            :class="{ active: selectedApplicationId === app.id }"
            @click="selectApplication(app)"
          >
            <div class="list-item-header">
              <div class="app-name">{{ app.name }}</div>
              <div class="app-actions">
                <button
                  class="icon-btn"
                  @click.stop="deleteApplication(app.id)"
                  title="删除"
                >
                  🗑️
                </button>
              </div>
            </div>
            <div class="app-info">
              <span class="badge">{{ app.image_name }}</span>
              <span class="config-count">
                {{ (allMappingsByApp[app.id] || []).length }} 集群
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Panel: Application Details -->
      <div class="detail-panel">
        <div v-if="!selectedApplication" class="empty-detail">
          <p>请选择一个应用开始配置</p>
        </div>

        <div v-else class="detail-content">
          <!-- Application Info -->
          <div class="detail-section">
            <div class="section-header">
              <h3>应用信息</h3>
              <button class="btn-secondary" @click="openEditApplicationModal">
                编辑应用
              </button>
            </div>

            <div class="info-grid">
              <div class="info-item">
                <label>应用名称:</label>
                <span>{{ selectedApplication.name }}</span>
              </div>
              <div class="info-item">
                <label>镜像名称:</label>
                <span class="code">{{ selectedApplication.image_name }}</span>
              </div>
              <div class="info-item">
                <label>仓库地址:</label>
                <span class="code">{{ selectedApplication.git_repo }}</span>
              </div>
              <div class="info-item">
                <label>构建类型:</label>
                <span>{{ selectedApplication.build_type }}</span>
              </div>
            </div>
          </div>

          <!-- Cluster Mappings -->
          <div class="detail-section">
            <div class="section-header">
              <h3>集群部署配置</h3>
              <button class="btn-primary-small" @click="openManageClusterModal">
                ⚙️ 管理集群关联
              </button>
            </div>

            <!-- Summary -->
            <div v-if="clusterMappings.length > 0" class="cluster-summary">
              <p>已关联 <strong>{{ clusterMappings.length }}</strong> 个集群环境</p>
              <div class="env-tags">
                <span v-for="mapping in clusterMappings" :key="mapping.id" class="env-tag">
                  {{ mapping.cluster_name }} / {{ mapping.environment }}
                </span>
              </div>
            </div>

            <div v-if="clusterMappings.length === 0" class="empty-state">
              <p>暂无集群配置，点击"管理集群关联"添加</p>
            </div>

            <div class="mappings-list">
              <div
                v-for="mapping in clusterMappings"
                :key="mapping.id"
                class="mapping-item"
              >
                <div class="mapping-header">
                  <div class="cluster-info">
                    <strong>{{ mapping.cluster_name }}</strong>
                    <span class="env-badge">{{ mapping.environment }}</span>
                  </div>
                  <div class="mapping-actions">
                    <button
                      class="btn-small"
                      @click="openReleaseModal(mapping)"
                    >
                      📤 发布
                    </button>
                    <button
                      class="btn-small"
                      @click="openEditClusterMappingModal(mapping)"
                    >
                      ✏️ 编辑
                    </button>
                    <button
                      class="btn-small btn-danger"
                      @click="deleteClusterMapping(mapping.id)"
                    >
                      🗑️
                    </button>
                  </div>
                </div>

                <div class="mapping-details">
                  <div class="detail-line">
                    <strong>Namespace:</strong> <code>{{ mapping.k8s_namespace }}</code>
                  </div>
                  <div class="detail-line">
                    <strong>工作负载类型:</strong> {{ mapping.workload_type }}
                  </div>
                  <div class="detail-line">
                    <strong>工作负载名称:</strong> <code>{{ mapping.workload_name }}</code>
                  </div>
                  <div class="detail-line">
                    <strong>当前镜像:</strong> <code>{{ mapping.current_image || '未配置' }}</code>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Release Actions -->
          <div class="detail-section">
            <div class="section-header">
              <h3>快速发布</h3>
            </div>
            <p class="help-text">
              选择上方集群配置中的"发布"按钮，或点击编辑配置
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <!-- Create/Edit Application Modal -->
    <div v-if="showApplicationModal" class="modal-overlay" @click="closeApplicationModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>{{ editingApplicationId ? '编辑应用' : '新建应用' }}</h2>
          <button class="close-btn" @click="closeApplicationModal">×</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>应用名称 *</label>
            <input
              v-model="applicationForm.name"
              type="text"
              class="form-input"
              placeholder="例如: api-service"
            />
          </div>

          <div class="form-group">
            <label>镜像名称 *</label>
            <input
              v-model="applicationForm.image_name"
              type="text"
              class="form-input"
              placeholder="例如: api-service (不含registry前缀)"
            />
          </div>

          <div class="form-group">
            <label>仓库地址</label>
            <input
              v-model="applicationForm.git_repo"
              type="text"
              class="form-input"
              placeholder="例如: https://github.com/company/api-service.git"
            />
          </div>

          <div class="form-group">
            <label>构建类型 *</label>
            <select v-model="applicationForm.build_type" class="form-input">
              <option value="docker">Docker</option>
              <option value="maven">Maven</option>
              <option value="npm">NPM</option>
              <option value="go">Go</option>
              <option value="other">其他</option>
            </select>
          </div>

          <div class="form-group">
            <label>描述</label>
            <textarea
              v-model="applicationForm.description"
              class="form-input form-textarea"
              rows="3"
            ></textarea>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-secondary" @click="closeApplicationModal">取消</button>
          <button class="btn-primary" @click="saveApplication">保存</button>
        </div>
      </div>
    </div>

    <!-- Edit Cluster Mapping Modal -->
    <div v-if="showClusterMappingModal" class="modal-overlay" @click="closeClusterMappingModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>{{ editingMappingId ? '编辑集群配置' : '添加集群配置' }}</h2>
          <button class="close-btn" @click="closeClusterMappingModal">×</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>集群 *</label>
            <select v-model="clusterMappingForm.cluster_id" class="form-input">
              <option value="">-- 请选择集群 --</option>
              <option v-for="cluster in sortedClusters" :key="cluster.id" :value="cluster.id">
                {{ cluster.name }} ({{ cluster.environment }})
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>Namespace *</label>
            <input
              v-model="clusterMappingForm.k8s_namespace"
              type="text"
              class="form-input"
              placeholder="例如: default 或 production"
            />
          </div>

          <div class="form-group">
            <label>工作负载类型 *</label>
            <select v-model="clusterMappingForm.workload_type" class="form-input">
              <option value="Deployment">Deployment</option>
              <option value="StatefulSet">StatefulSet</option>
              <option value="DaemonSet">DaemonSet</option>
            </select>
          </div>

          <div class="form-group">
            <label>工作负载名称 *</label>
            <input
              v-model="clusterMappingForm.workload_name"
              type="text"
              class="form-input"
              placeholder="例如: api-service-deployment"
            />
          </div>

          <div class="form-group">
            <label>容器名称</label>
            <input
              v-model="clusterMappingForm.container_name"
              type="text"
              class="form-input"
              placeholder="例如: api-service (如不填，默认为工作负载名称)"
            />
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-secondary" @click="closeClusterMappingModal">取消</button>
          <button class="btn-primary" @click="saveClusterMapping">保存</button>
        </div>
      </div>
    </div>

    <!-- Release Modal -->
    <div v-if="showReleaseModal" class="modal-overlay" @click="closeReleaseModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>发布新版本</h2>
          <button class="close-btn" @click="closeReleaseModal">×</button>
        </div>

        <div class="modal-body">
          <div class="info-box">
            <p>
              <strong>应用:</strong> {{ selectedApplication?.name }}
            </p>
            <p>
              <strong>集群:</strong> {{ releaseInfo.cluster_name }}
              <span class="env-badge">{{ releaseInfo.environment }}</span>
            </p>
            <p>
              <strong>Namespace:</strong> <code>{{ releaseInfo.k8s_namespace }}</code>
            </p>
            <p>
              <strong>工作负载:</strong>
              <code>{{ releaseInfo.workload_name }}</code> ({{ releaseInfo.workload_type }})
            </p>
          </div>

          <div class="form-group">
            <label>镜像标签 *</label>
            <input
              v-model="releaseForm.image_tag"
              type="text"
              class="form-input"
              placeholder="例如: v1.0.0 或 main-abc123"
            />
            <p class="help-text">
              完整镜像地址:
              <code>{{ releaseInfo.registry_prefix }}/{{ selectedApplication?.image_name }}:{{ releaseForm.image_tag }}</code>
            </p>
          </div>

          <div v-if="false" class="form-group">
            <label>
              <input type="checkbox" v-model="releaseForm.dryRun" />
              <span>试运行（不实际执行）</span>
            </label>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-secondary" @click="closeReleaseModal">取消</button>
          <button class="btn-success" @click="confirmRelease">确认发布</button>
        </div>
      </div>
    </div>

    <!-- Manage Cluster Associations Modal -->
    <div v-if="showManageClusterModal" class="modal-overlay" @click="closeManageClusterModal">
      <div class="modal modal-large" @click.stop>
        <div class="modal-header">
          <h2 v-if="!showEditFormInManageModal">管理集群关联 - {{ selectedApplication?.name }}</h2>
          <h2 v-else>编辑集群配置</h2>
          <button class="close-btn" @click="closeManageClusterModal">×</button>
        </div>

        <div class="modal-body">
          <!-- Cluster List View -->
          <div v-if="!showEditFormInManageModal">
            <!-- Available Clusters -->
            <div class="cluster-management">
              <h3>可用集群</h3>
              <p class="help-text">点击集群进行配置，或点击"添加"快速关联</p>

              <div class="clusters-grid">
                <div v-for="cluster in availableClustersForManage" :key="cluster.id" class="cluster-card">
                  <div class="cluster-card-header">
                    <strong>{{ cluster.name }}</strong>
                    <span class="env-badge">{{ cluster.environment }}</span>
                  </div>

                  <div class="cluster-card-body">
                    <p class="label">Kubernetes 集群</p>
                    <p class="registry">
                      <small>{{ cluster.registry_prefix }}</small>
                    </p>
                  </div>

                  <div class="cluster-card-footer">
                    <button
                      v-if="!isClusterAlreadyMapped(cluster.id)"
                      class="btn-link"
                      @click="quickAddClusterInManageModal(cluster)"
                    >
                      + 添加
                    </button>
                    <button
                      v-else
                      class="btn-link btn-success"
                      @click="startEditInManageModal(getMappingForCluster(cluster.id)!)"
                    >
                      ✓ 已配置
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <hr class="divider" />

            <!-- Current Mappings -->
            <div class="current-mappings">
              <h3>已关联集群 ({{ clusterMappings.length }})</h3>

              <div v-if="clusterMappings.length === 0" class="empty-state">
                <p>暂无关联集群</p>
              </div>

              <div class="mappings-management-list">
                <div
                  v-for="mapping in clusterMappings"
                  :key="mapping.id"
                  class="mapping-manage-item"
                >
                  <div class="mapping-manage-header">
                    <div class="cluster-badge">
                      <strong>{{ mapping.cluster_name }}</strong>
                      <span class="env-badge">{{ mapping.environment }}</span>
                    </div>
                  </div>

                  <div class="mapping-manage-body">
                    <div class="config-row">
                      <span class="config-label">Namespace:</span>
                      <code>{{ mapping.k8s_namespace }}</code>
                    </div>
                    <div class="config-row">
                      <span class="config-label">工作负载:</span>
                      <code>{{ mapping.workload_type }}/{{ mapping.workload_name }}</code>
                    </div>
                    <div v-if="mapping.container_name" class="config-row">
                      <span class="config-label">容器:</span>
                      <code>{{ mapping.container_name }}</code>
                    </div>
                  </div>

                  <div class="mapping-manage-footer">
                    <button
                      class="btn-link btn-edit"
                      @click="startEditInManageModal(mapping)"
                    >
                      编辑
                    </button>
                    <button
                      class="btn-link btn-danger"
                      @click="deleteClusterMapping(mapping.id)"
                    >
                      删除
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Edit Form View -->
          <div v-else>
            <button class="btn-back" @click="cancelEditInManageModal">
              ← 返回集群列表
            </button>

            <div class="edit-form-container">
              <div class="form-group">
                <label>集群 *</label>
                <select v-model="clusterMappingForm.cluster_id" class="form-input" :disabled="!!editingMappingInManageModal?.id">
                  <option value="">-- 请选择集群 --</option>
                  <option v-for="cluster in sortedClusters" :key="cluster.id" :value="cluster.id">
                    {{ cluster.name }} ({{ cluster.environment }})
                  </option>
                </select>
                <p v-if="editingMappingInManageModal?.id" class="help-text">创建后不可更改集群</p>
              </div>

              <div class="form-group">
                <label>Namespace *</label>
                <input
                  v-model="clusterMappingForm.k8s_namespace"
                  type="text"
                  class="form-input"
                  placeholder="例如: default 或 production"
                />
              </div>

              <div class="form-group">
                <label>工作负载类型 *</label>
                <select v-model="clusterMappingForm.workload_type" class="form-input">
                  <option value="Deployment">Deployment</option>
                  <option value="StatefulSet">StatefulSet</option>
                  <option value="DaemonSet">DaemonSet</option>
                </select>
              </div>

              <div class="form-group">
                <label>工作负载名称 *</label>
                <input
                  v-model="clusterMappingForm.workload_name"
                  type="text"
                  class="form-input"
                  placeholder="例如: api-service-deployment"
                />
              </div>

              <div class="form-group">
                <label>容器名称</label>
                <input
                  v-model="clusterMappingForm.container_name"
                  type="text"
                  class="form-input"
                  placeholder="例如: api-service (如不填，默认为工作负载名称)"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button v-if="!showEditFormInManageModal" class="btn-secondary" @click="closeManageClusterModal">关闭</button>
          <button v-else class="btn-secondary" @click="cancelEditInManageModal">取消</button>
          <button v-if="showEditFormInManageModal" class="btn-primary" @click="saveClusterMappingInManageModal">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getApplications, getClusters, createRelease, getEnvironments } from '@/api/metadata'
import {
  getClusterMappingsByApp,
  createClusterMapping,
  updateClusterMapping,
  deleteClusterMapping as apiDeleteClusterMapping
} from '@/api/cluster-mapping'
import type { Application, Cluster, ClusterMapping, Environment } from '@/types/api'

// State
const searchQuery = ref('')
const selectedApplicationId = ref<number | null>(null)

const applications = ref<Application[]>([])
const clusters = ref<Cluster[]>([])
const environments = ref<Environment[]>([])
const clusterMappings = ref<ClusterMapping[]>([])
// Store all mappings by app_id for list display
const allMappingsByApp = ref<{ [appId: number]: ClusterMapping[] }>({})

// Build environment name to id mapping
const environmentMap = computed(() => {
  const map: { [key: string]: number } = {}
  environments.value.forEach(env => {
    map[env.name.toLowerCase()] = env.id
  })
  return map
})

// Application Modal
const showApplicationModal = ref(false)
const editingApplicationId = ref<number | null>(null)
const applicationForm = ref<Partial<Application>>({
  build_type: 'docker'
})

// Cluster Mapping Modal
const showClusterMappingModal = ref(false)
const editingMappingId = ref<number | null>(null)
const clusterMappingForm = ref<Partial<ClusterMapping>>({
  workload_type: 'Deployment'
})

// Manage Cluster Modal
const showManageClusterModal = ref(false)
const availableClustersForManage = ref<Cluster[]>([])
const showEditFormInManageModal = ref(false)
const editingMappingInManageModal = ref<ClusterMapping | null>(null)

// Release Modal
const showReleaseModal = ref(false)
const releaseInfo = ref<Partial<ClusterMapping & { registry_prefix?: string }>>({})
const releaseForm = ref({
  image_tag: '',
  dryRun: false
})

// Loading and Error States
const isLoading = ref(true)
const loadingApps = ref(false)
const loadingMappings = ref(false)
const errorMessage = ref<string | null>(null)

// Sorting
const sortBy = ref<'name' | 'config-count'>('name')
const sortOrder = ref<'asc' | 'desc'>('asc')

// Computed
const filteredApplications = computed(() => {
  // Filter
  let filtered = applications.value.filter(app =>
    app.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    app.image_name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )

  // Sort
  filtered.sort((a, b) => {
    let compareVal = 0
    
    if (sortBy.value === 'name') {
      compareVal = a.name.localeCompare(b.name)
    } else if (sortBy.value === 'config-count') {
      // Get count of mapped clusters for each app from pre-loaded data
      const aCount = (allMappingsByApp.value[a.id] || []).length
      const bCount = (allMappingsByApp.value[b.id] || []).length
      compareVal = aCount - bCount
    }

    return sortOrder.value === 'asc' ? compareVal : -compareVal
  })

  return filtered
})

const selectedApplication = computed(() => {
  return applications.value.find(app => app.id === selectedApplicationId.value)
})

const sortedClusters = computed(() => {
  // Sort clusters by environment, then by name
  return clusters.value.slice().sort((a, b) => {
    const aEnv = a.environment || '未分类'
    const bEnv = b.environment || '未分类'
    const envCompare = aEnv.localeCompare(bEnv)
    if (envCompare !== 0) return envCompare
    return (a.name || '').localeCompare(b.name || '')
  })
})

// Functions
const selectApplication = async (app: Application) => {
  selectedApplicationId.value = app.id
  // Use pre-loaded mappings if available, otherwise fetch
  if (allMappingsByApp.value[app.id]) {
    clusterMappings.value = allMappingsByApp.value[app.id]
  } else {
    try {
      const mappings = await getClusterMappingsByApp(app.id)
      clusterMappings.value = mappings
      allMappingsByApp.value[app.id] = mappings
    } catch (error) {
      console.error('Failed to load cluster mappings:', error)
      clusterMappings.value = []
    }
  }
}

const openCreateApplicationModal = () => {
  editingApplicationId.value = null
  applicationForm.value = { build_type: 'docker' }
  showApplicationModal.value = true
}

const openEditApplicationModal = () => {
  if (selectedApplication.value) {
    editingApplicationId.value = selectedApplication.value.id
    applicationForm.value = { ...selectedApplication.value }
    showApplicationModal.value = true
  }
}

const closeApplicationModal = () => {
  showApplicationModal.value = false
  editingApplicationId.value = null
}

const saveApplication = async () => {
  try {
    // Applications might need special handling for create vs update
    // For now, just log the action
    alert('应用信息已保存')
    closeApplicationModal()
  } catch (error) {
    console.error('Failed to save application:', error)
    alert('保存应用失败')
  }
}

const deleteApplication = async (appId: number) => {
  if (confirm('确定删除此应用吗？')) {
    try {
      // Applications deletion would be implemented in the API
      alert('应用已删除')
      selectedApplicationId.value = null
      if (selectedApplication.value) {
        applications.value = applications.value.filter(a => a.id !== appId)
      }
    } catch (error) {
      console.error('Failed to delete application:', error)
      alert('删除应用失败')
    }
  }
}

const openEditClusterMappingModal = (mapping: ClusterMapping | null) => {
  if (mapping) {
    editingMappingId.value = mapping.id
    clusterMappingForm.value = { ...mapping }
  } else {
    editingMappingId.value = null
    clusterMappingForm.value = { workload_type: 'Deployment' }
  }
  showClusterMappingModal.value = true
}

const closeClusterMappingModal = () => {
  showClusterMappingModal.value = false
  editingMappingId.value = null
}

const saveClusterMapping = async () => {
  try {
    if (!clusterMappingForm.value.cluster_id) {
      alert('请选择集群')
      return
    }
    if (!clusterMappingForm.value.k8s_namespace) {
      alert('请输入 Namespace')
      return
    }
    if (!clusterMappingForm.value.workload_name) {
      alert('请输入工作负载名称')
      return
    }

    const envId = getEnvIdForCluster(clusterMappingForm.value.cluster_id)
    if (!envId) {
      alert('无法确定环境，请确保集群和环境已正确配置')
      return
    }

    const data = {
      ...clusterMappingForm.value,
      app_id: selectedApplication.value?.id,
      env_id: envId
    }

    if (editingMappingId.value) {
      await updateClusterMapping(editingMappingId.value, data)
      alert('集群配置已更新')
    } else {
      await createClusterMapping(data)
      alert('集群配置已创建')
    }

    // Reload mappings
    if (selectedApplication.value) {
      await selectApplication(selectedApplication.value)
    }
    closeClusterMappingModal()
  } catch (error) {
    console.error('Failed to save cluster mapping:', error)
    alert(`保存失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

const deleteClusterMapping = async (mappingId: number) => {
  if (confirm('确定删除此集群配置吗？')) {
    try {
      await apiDeleteClusterMapping(mappingId)
      alert('集群配置已删除')
      // Reload mappings
      if (selectedApplication.value) {
        await selectApplication(selectedApplication.value)
      }
    } catch (error) {
      console.error('Failed to delete cluster mapping:', error)
      alert(`删除失败: ${error instanceof Error ? error.message : '未知错误'}`)
    }
  }
}


const openReleaseModal = (mapping: ClusterMapping) => {
  releaseInfo.value = mapping
  releaseForm.value = { image_tag: '', dryRun: false }
  showReleaseModal.value = true
}

const closeReleaseModal = () => {
  showReleaseModal.value = false
}

const confirmRelease = async () => {
  if (!releaseForm.value.image_tag.trim()) {
    alert('请输入镜像标签')
    return
  }
  
  try {
    if (!selectedApplication.value || !releaseInfo.value || !releaseInfo.value.cluster_id) {
      alert('缺少发布信息')
      return
    }

    const envId = releaseInfo.value.env_id || getEnvIdForCluster(releaseInfo.value.cluster_id as string | number)
    if (!envId) {
      alert('无法确定环境')
      return
    }
    
    await createRelease({
      app_id: selectedApplication.value.id,
      cluster_id: releaseInfo.value.cluster_id,
      env_id: envId,
      image_tag: releaseForm.value.image_tag,
      dryRun: releaseForm.value.dryRun
    })
    
    alert('发布已提交')
    closeReleaseModal()
  } catch (error) {
    console.error('Failed to create release:', error)
    alert('提交发布失败')
  }
}

const openManageClusterModal = async () => {
  availableClustersForManage.value = clusters.value
  showEditFormInManageModal.value = false
  editingMappingInManageModal.value = null
  showManageClusterModal.value = true
}

const closeManageClusterModal = () => {
  showManageClusterModal.value = false
  showEditFormInManageModal.value = false
  editingMappingInManageModal.value = null
}

const startEditInManageModal = (mapping: ClusterMapping) => {
  editingMappingInManageModal.value = mapping
  clusterMappingForm.value = { ...mapping }
  showEditFormInManageModal.value = true
}

const cancelEditInManageModal = () => {
  showEditFormInManageModal.value = false
  editingMappingInManageModal.value = null
}

const saveClusterMappingInManageModal = async () => {
  try {
    if (!clusterMappingForm.value.cluster_id) {
      alert('请选择集群')
      return
    }
    if (!clusterMappingForm.value.k8s_namespace) {
      alert('请输入 Namespace')
      return
    }
    if (!clusterMappingForm.value.workload_name) {
      alert('请输入工作负载名称')
      return
    }

    const envId = getEnvIdForCluster(clusterMappingForm.value.cluster_id)
    if (!envId) {
      alert('无法确定环境，请确保集群和环境已正确配置')
      return
    }

    const data = {
      ...clusterMappingForm.value,
      app_id: selectedApplication.value?.id,
      env_id: envId
    }

    if (editingMappingInManageModal.value?.id) {
      await updateClusterMapping(editingMappingInManageModal.value.id, data)
      alert('集群配置已更新')
    } else {
      await createClusterMapping(data)
      alert('集群配置已创建')
    }

    // Reload mappings
    if (selectedApplication.value) {
      await selectApplication(selectedApplication.value)
    }
    cancelEditInManageModal()
  } catch (error) {
    console.error('Failed to save cluster mapping:', error)
    alert(`保存失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

const isClusterAlreadyMapped = (clusterId: number | string): boolean => {
  return clusterMappings.value.some(m => m.cluster_id === clusterId)
}

const getMappingForCluster = (clusterId: number | string): ClusterMapping | null => {
  return clusterMappings.value.find(m => m.cluster_id === clusterId) || null
}

const getEnvIdForCluster = (clusterId: number | string): number | null => {
  // For existing mappings, use the env_id from the mapping
  const existingMapping = getMappingForCluster(clusterId)
  if (existingMapping?.env_id) {
    return existingMapping.env_id
  }

  // For new clusters, find env_id based on cluster environment
  const cluster = clusters.value.find(c => c.id === clusterId)
  if (cluster?.environment) {
    const envName = cluster.environment.toLowerCase()
    // Try exact match first
    if (environmentMap.value[envName]) {
      return environmentMap.value[envName]
    }
    // Try prefix match (e.g., 'dev' -> 'development')
    const prefixMatch = Object.entries(environmentMap.value).find(([env]) =>
      env.startsWith(envName)
    )
    if (prefixMatch) {
      return prefixMatch[1]
    }
  }

  // Default to first environment if found
  return environments.value.length > 0 ? environments.value[0].id : 1
}

const quickAddCluster = (cluster: Cluster) => {
  // Open edit modal with pre-filled cluster info (used from detail panel)
  clusterMappingForm.value = {
    cluster_id: Number(cluster.id),
    workload_type: 'Deployment'
  }
  showClusterMappingModal.value = true
}

const quickAddClusterInManageModal = (cluster: Cluster) => {
  // Add cluster quickly in manage modal
  clusterMappingForm.value = {
    cluster_id: Number(cluster.id),
    workload_type: 'Deployment'
  }
  editingMappingInManageModal.value = null
  showEditFormInManageModal.value = true
}

const loadApplications = async () => {
  try {
    const data = await getApplications()
    applications.value = data
    // Pre-load cluster mappings for all applications
    for (const app of data) {
      try {
        const mappings = await getClusterMappingsByApp(app.id)
        allMappingsByApp.value[app.id] = mappings
      } catch (error) {
        console.error(`Failed to load mappings for app ${app.id}:`, error)
        allMappingsByApp.value[app.id] = []
      }
    }
  } catch (error) {
    console.error('Failed to load applications:', error)
    alert('加载应用列表失败')
  }
}

const loadClusters = async () => {
  try {
    const data = await getClusters()
    clusters.value = data
  } catch (error) {
    console.error('Failed to load clusters:', error)
    alert('加载集群列表失败')
  }
}

const loadEnvironments = async () => {
  try {
    const data = await getEnvironments()
    environments.value = data
  } catch (error) {
    console.error('Failed to load environments:', error)
    // Environments are optional, so we don't show an error
  }
}

// Lifecycle
onMounted(async () => {
  try {
    isLoading.value = true
    errorMessage.value = null
    await loadApplications()
    await loadClusters()
    await loadEnvironments()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载数据失败'
  } finally {
    isLoading.value = false
  }
})
</script>

<style scoped>
.k8s-release-page {
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
  transition: background 0.2s;
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
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-header h2 {
  margin: 0;
  font-size: 16px;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.sort-controls {
  display: flex;
  gap: 8px;
  align-items: center;
}

.sort-controls button {
  padding: 6px 12px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.sort-controls button:hover {
  border-color: #1890ff;
  color: #1890ff;
}

.sort-controls button.active {
  background: #e6f7ff;
  border-color: #1890ff;
  color: #1890ff;
  font-weight: 600;
}

.sort-controls button.order-btn {
  min-width: 32px;
  font-size: 16px;
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

.app-name {
  font-weight: 600;
  color: #1a1a1a;
}

.app-actions {
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

.app-info {
  display: flex;
  gap: 8px;
  font-size: 12px;
}

.badge {
  background: #f0f0f0;
  padding: 2px 8px;
  border-radius: 3px;
  color: #666;
}

.config-count {
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

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
}

.btn-secondary {
  padding: 6px 12px;
  background: white;
  color: #1a1a1a;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: background 0.2s;
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

.mappings-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mapping-item {
  padding: 12px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
}

.mapping-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.cluster-info {
  display: flex;
  gap: 8px;
  align-items: center;
}

.env-badge {
  background: #e6f7ff;
  color: #1890ff;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 600;
}

.mapping-actions {
  display: flex;
  gap: 6px;
}

.btn-small {
  padding: 4px 8px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
}

.btn-small:hover {
  background: #40a9ff;
}

.btn-small.btn-danger {
  background: #ff4d4f;
}

.btn-small.btn-danger:hover {
  background: #ff7875;
}

.mapping-details {
  font-size: 12px;
}

.detail-line {
  margin-bottom: 6px;
  padding: 6px;
  background: white;
  border-radius: 3px;
}

.detail-line strong {
  color: #666;
}

.detail-line code {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  color: #d32f2f;
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
  background: #85ce61;
}

.btn-primary-small {
  padding: 6px 12px;
  background: #1890ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  transition: background 0.2s;
}

.btn-primary-small:hover {
  background: #40a9ff;
}

.btn-link {
  background: none;
  border: none;
  color: #1890ff;
  cursor: pointer;
  padding: 0;
  font-size: 12px;
  text-decoration: none;
  transition: color 0.2s;
}

.btn-link:hover {
  color: #40a9ff;
}

.btn-link.btn-success {
  color: #52c41a;
  background: none;
  padding: 0;
}

.btn-link.btn-success:hover {
  color: #85ce61;
}

.btn-link.btn-edit {
  color: #1890ff;
}

.btn-link.btn-danger {
  color: #ff4d4f;
}

.btn-link.btn-danger:hover {
  color: #ff7875;
}

.modal-large {
  max-width: 800px;
}

.cluster-summary {
  background: #f0f8ff;
  border: 1px solid #b3e5fc;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 16px;
}

.cluster-summary p {
  margin: 0 0 8px 0;
  font-size: 13px;
}

.env-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.env-tag {
  background: #e6f7ff;
  color: #1890ff;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.cluster-management {
  margin-bottom: 24px;
}

.cluster-management h3 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
}

.clusters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.cluster-card {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 12px;
  background: white;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.cluster-card:hover {
  border-color: #1890ff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.1);
}

.cluster-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  gap: 8px;
}

.cluster-card-header strong {
  font-size: 13px;
  margin: 0;
}

.cluster-card-body {
  margin-bottom: 8px;
}

.cluster-card-body .label {
  font-size: 12px;
  color: #666;
  margin: 0 0 4px 0;
}

.cluster-card-body .registry {
  font-size: 11px;
  color: #999;
  margin: 0;
  word-break: break-all;
}

.cluster-card-footer {
  display: flex;
  gap: 8px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.divider {
  margin: 24px 0;
  border: none;
  border-top: 1px solid #eee;
}

.current-mappings h3 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
}

.mappings-management-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mapping-manage-item {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 12px;
  background: white;
}

.mapping-manage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.cluster-badge {
  display: flex;
  gap: 8px;
  align-items: center;
}

.cluster-badge strong {
  font-size: 13px;
  margin: 0;
}

.mapping-manage-body {
  margin-bottom: 10px;
  padding: 8px;
  background: #f9f9f9;
  border-radius: 3px;
  font-size: 12px;
}

.config-row {
  margin-bottom: 6px;
  display: flex;
  gap: 8px;
}

.config-row:last-child {
  margin-bottom: 0;
}

.config-label {
  font-weight: 600;
  color: #666;
  min-width: 70px;
}

.config-row code {
  background: white;
  padding: 2px 6px;
  border-radius: 2px;
  color: #d32f2f;
  font-family: 'Courier New', monospace;
  word-break: break-all;
}

.mapping-manage-footer {
  display: flex;
  gap: 12px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.empty-state {
  text-align: center;
  padding: 24px;
  color: #999;
  font-size: 14px;
  background: #f5f5f5;
  border-radius: 4px;
}

.empty-state p {
  margin: 0;
}

.btn-success:hover {
  background: #73d13d;
}

/* Error and Loading States */
.error-banner {
  position: fixed;
  top: 60px;
  left: 0;
  right: 0;
  background: #fff2e8;
  border-bottom: 2px solid #ff7875;
  padding: 12px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  z-index: 100;
}

.error-banner span {
  color: #d32f2f;
  font-weight: 600;
}

.close-banner-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #ff7875;
}

.close-banner-btn:hover {
  color: #d32f2f;
}

.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 99;
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  background: white;
  padding: 32px;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f0f0f0;
  border-top-color: #1890ff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-spinner p {
  margin: 0;
  color: #666;
  font-weight: 600;
}

@media (max-width: 1200px) {
  .content-layout {
    grid-template-columns: 1fr;
  }
}

/* Manage Modal Edit Form */
.btn-back {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  margin-bottom: 16px;
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #1890ff;
  transition: all 0.2s;
}

.btn-back:hover {
  background: #e6f7ff;
  border-color: #1890ff;
}

.edit-form-container {
  padding: 8px 0;
}
</style>
