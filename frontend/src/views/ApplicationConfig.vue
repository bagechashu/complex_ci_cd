<template>
  <div class="app-config-page">
    <div class="page-header">
      <h1>🔧 应用配置</h1>
      <p class="description">配置应用与 K8s 集群的部署映射关系（应用是整个系统的核心）</p>
    </div>

    <div class="content">
      <!-- Left Panel: Applications List -->
      <div class="panel applications-panel">
        <div class="panel-header">
          <h2>应用列表</h2>
          <button class="btn-primary" @click="showCreateApp = true">+ 新建应用</button>
        </div>

        <!-- Search Box -->
        <div class="search-box">
          <input
            v-model="searchKeyword"
            type="text"
            placeholder="搜索应用名称或镜像..."
            @input="handleSearch"
          />
        </div>

        <div class="applications-list">
          <div
            v-for="app in applications"
            :key="app.id"
            class="app-item"
            :class="{ active: selectedAppId === app.id }"
            @click="selectApp(app.id)"
          >
            <div class="app-name">{{ app.name }}</div>
            <div class="app-image">{{ app.image_name }}</div>
          </div>
          
          <div v-if="applications.length === 0" class="empty-list">
            <p>暂无应用</p>
          </div>
        </div>

        <!-- Pagination -->
        <div class="pagination">
          <button
            :disabled="currentPage === 1"
            @click="goToPreviousPage"
            class="pagination-btn"
          >
            ← 上一页
          </button>
          <span class="pagination-info">
            第 {{ currentPage }} / {{ totalPages }} 页
          </span>
          <button
            :disabled="currentPage === totalPages"
            @click="goToNextPage"
            class="pagination-btn"
          >
            下一页 →
          </button>
        </div>
      </div>

      <!-- Right Panel: App Details & Cluster Mappings -->
      <div class="panel details-panel">
        <div v-if="selectedApp" class="details-content">
          <!-- Application Info -->
          <div class="section">
            <h3>应用信息</h3>
            <div class="info-item">
              <label>应用名称:</label>
              <span>{{ selectedApp.name }}</span>
            </div>
            <div class="info-item">
              <label>镜像名(无前缀):</label>
              <span>{{ selectedApp.image_name }}</span>
            </div>
            <div class="info-item">
              <label>Git 仓库:</label>
              <span>{{ selectedApp.git_repo || '未配置' }}</span>
            </div>
            <div class="info-item">
              <label>构建类型:</label>
              <span>{{ selectedApp.build_type || 'docker' }}</span>
            </div>
            <button class="btn-secondary" @click="editSelectedApp">编辑应用</button>
          </div>

          <!-- K8s Cluster Mappings -->
          <div class="section">
            <h3>K8s 集群部署配置</h3>
            <p class="section-desc">指定该应用在各个 K8s 集群中部署的 Namespace 和 Workload</p>

            <!-- TODO: Load and display cluster mappings -->
            <div v-if="clusterMappings.length > 0" class="mappings-list">
              <div
                v-for="mapping in clusterMappings"
                :key="mapping.cluster_id"
                class="mapping-item"
              >
                <div class="mapping-header">
                  <h4>{{ mapping.cluster_name }}</h4>
                  <button class="btn-small" @click="editMapping(mapping)">编辑</button>
                  <button class="btn-small btn-danger" @click="deleteMapping(mapping.id)">删除</button>
                </div>
                <div class="mapping-details">
                  <div>
                    <strong>Namespace:</strong> {{ mapping.k8s_namespace || '未配置' }}
                  </div>
                  <div>
                    <strong>Workload:</strong> {{ mapping.k8s_workload || '未配置' }}
                  </div>
                  <div>
                    <strong>Container:</strong> {{ mapping.container_name || '未配置' }}
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="empty-state">
              <p>暂无集群配置，请添加</p>
            </div>

            <button class="btn-primary" @click="showAddMapping = true">+ 添加集群配置</button>
          </div>
        </div>

        <div v-else class="empty-state">
          <p>请选择一个应用查看详情</p>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <!-- TODO: Create Application Modal -->
    <!-- TODO: Edit Application Modal -->
    <!-- TODO: Add/Edit Cluster Mapping Modal -->
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import type { Application, ClusterMapping } from '@/types/api'
import { getApplications } from '@/api/metadata'

// State
const applications = ref<Application[]>([])
const selectedAppId = ref<number | null>(null)
const clusterMappings = ref<ClusterMapping[]>([])
const showCreateApp = ref(false)
const showAddMapping = ref(false)

// Pagination & Search
const currentPage = ref(1)
const totalPages = ref(1)
const totalCount = ref(0)
const pageSize = 10
const searchKeyword = ref('')
let searchTimeout: ReturnType<typeof setTimeout>

// Computed
const selectedApp = computed(() => {
  return applications.value.find(app => app.id === selectedAppId.value)
})

// Functions
const loadApplications = async () => {
  try {
    const response = await getApplications(currentPage.value, pageSize, searchKeyword.value)
    applications.value = response.data
    currentPage.value = response.page
    totalPages.value = response.totalPages
    totalCount.value = response.total
  } catch (error) {
    console.error('Failed to load applications:', error)
  }
}

const handleSearch = () => {
  // Debounce search input
  clearTimeout(searchTimeout)
  currentPage.value = 1
  searchTimeout = setTimeout(() => {
    loadApplications()
  }, 300)
}

const goToPreviousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loadApplications()
  }
}

const goToNextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadApplications()
  }
}

const selectApp = async (appId: number) => {
  selectedAppId.value = appId
  try {
    const response = await fetch(`/api/v1/workload-targets/app/${appId}`)
    const result = await response.json()
    clusterMappings.value = result.data || []
  } catch (error) {
    console.error('Failed to load cluster mappings:', error)
  }
}

const editSelectedApp = () => {
  if (selectedAppId.value) {
    showCreateApp.value = true
  }
}

const editMapping = (mapping: ClusterMapping) => {
  showAddMapping.value = true
}

const deleteMapping = async (mappingId: number) => {
  if (confirm('确认删除此映射吗？')) {
    try {
      const response = await fetch(`/api/v1/workload-targets/${mappingId}`, {
        method: 'DELETE'
      })
      if (response.ok) {
        if (selectedAppId.value) {
          await selectApp(selectedAppId.value)
        }
      }
    } catch (error) {
      console.error('Failed to delete mapping:', error)
    }
  }
}

// Lifecycle
onMounted(async () => {
  await loadApplications()
})
</script>

<style scoped>
.app-config-page {
  padding: 24px;
  height: 100%;
  overflow-y: auto;
}

.page-header {
  margin-bottom: 24px;
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

.content {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 24px;
  min-height: 600px;
}

.panel {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow-y: auto;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.panel-header h2 {
  margin: 0;
  font-size: 16px;
}

.applications-list {
  max-height: 600px;
  overflow-y: auto;
}

.app-item {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  cursor: pointer;
  transition: background 0.2s;
}

.app-item:hover {
  background: #f9f9f9;
}

.app-item.active {
  background: #e6f7ff;
  border-left: 3px solid #1890ff;
}

.app-name {
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 4px;
}

.app-image {
  font-size: 12px;
  color: #999;
}

.details-content {
  padding: 24px;
}

.section {
  margin-bottom: 32px;
}

.section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #1a1a1a;
}

.section-desc {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: #666;
}

.info-item {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 12px;
  margin-bottom: 8px;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.info-item label {
  font-weight: 600;
  color: #666;
}

.mappings-list {
  margin-bottom: 16px;
}

.mapping-item {
  padding: 12px;
  margin-bottom: 12px;
  background: #f9f9f9;
  border-radius: 6px;
  border: 1px solid #eee;
}

.mapping-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.mapping-header h4 {
  margin: 0;
  font-size: 14px;
}

.mapping-details {
  font-size: 13px;
  color: #666;
  line-height: 1.6;
}

.mapping-details div {
  margin: 4px 0;
}

.empty-state {
  text-align: center;
  padding: 32px 16px;
  color: #999;
}

.empty-list {
  text-align: center;
  padding: 32px 16px;
  color: #999;
}

/* Search Box */
.search-box {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.search-box input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.search-box input:focus {
  border-color: #1890ff;
  box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-top: 1px solid #eee;
  background: #fafafa;
}

.pagination-btn {
  padding: 6px 12px;
  border: 1px solid #d9d9d9;
  background: white;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.pagination-btn:hover:not(:disabled) {
  border-color: #1890ff;
  color: #1890ff;
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination-info {
  font-size: 12px;
  color: #666;
}

/* Buttons */
.btn-primary,
.btn-secondary,
.btn-small,
.btn-danger {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #1890ff;
  color: white;
}

.btn-primary:hover {
  background: #40a9ff;
}

.btn-secondary {
  background: #f0f0f0;
  color: #1a1a1a;
}

.btn-secondary:hover {
  background: #e0e0e0;
}

.btn-small {
  padding: 4px 12px;
  font-size: 12px;
  margin-right: 8px;
}

.btn-danger {
  background: #faad14;
  color: white;
}

.btn-danger:hover {
  background: #ffc53d;
}

@media (max-width: 1200px) {
  .content {
    grid-template-columns: 1fr;
  }
}
</style>
