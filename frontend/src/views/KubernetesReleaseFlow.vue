<template>
  <div class="k8s-release-page">
    <div class="page-header">
      <h1>🚀 K8s 发布向导</h1>
      <p class="description">为 K8s 应用发布新的镜像版本</p>
    </div>

    <div class="wizard-container">
      <!-- Steps Indicator -->
      <div class="steps-indicator">
        <div v-for="(step, index) in steps" :key="index" class="step" :class="{ active: currentStep === index + 1 }">
          <div class="step-number">{{ index + 1 }}</div>
          <div class="step-title">{{ step }}</div>
        </div>
      </div>

      <!-- Step Content -->
      <div class="step-content">
        <!-- Step 1: Select Application -->
        <div v-if="currentStep === 1" class="step-pane">
          <h3>选择应用</h3>
          <p>选择要发布的应用</p>
          
          <!-- Search Box -->
          <div class="form-group">
            <label>搜索应用</label>
            <input
              v-model="appSearchKeyword"
              type="text"
              class="form-input"
              placeholder="搜索应用名称或镜像..."
              @input="loadApplications"
            />
          </div>

          <!-- Application Selection -->
          <div class="form-group">
            <label>应用 *</label>
            <select v-model="form.applicationId" class="form-input">
              <option value="">-- 请选择应用 --</option>
              <option v-for="app in filteredApplications" :key="app.id" :value="app.id">
                {{ app.name }} ({{ app.image_name }})
              </option>
            </select>
            <div v-if="filteredApplications.length === 0" class="no-result">
              未找到匹配的应用
            </div>
          </div>
        </div>

        <!-- Step 2: Select Cluster & Namespace -->
        <div v-if="currentStep === 2" class="step-pane">
          <h3>选择 K8s 集群</h3>
          <p>选择目标集群和 Namespace</p>

          <!-- TODO: Load available clusters for selected app -->
          <div class="form-group">
            <label>集群 *</label>
            <select v-model="form.clusterId" class="form-input" @change="onClusterChange">
              <option value="">-- 请选择集群 --</option>
              <option v-for="cluster in availableClusters" :key="cluster.id" :value="cluster.id">
                {{ cluster.name }} ({{ cluster.labels }})
              </option>
            </select>
          </div>

          <div v-if="selectedClusterMapping" class="info-box">
            <h4>部署信息</h4>
            <div class="info-item">
              <strong>Namespace:</strong> {{ selectedClusterMapping.k8s_namespace }}
            </div>
            <div class="info-item">
              <strong>Workload:</strong> {{ selectedClusterMapping.k8s_workload }}
            </div>
            <div class="info-item">
              <strong>Container:</strong> {{ selectedClusterMapping.container_name }}
            </div>
            <div class="info-item">
              <strong>镜像前缀:</strong> {{ selectedCluster?.registry_prefix }}
            </div>
          </div>
        </div>

        <!-- Step 3: Select/Input Image Tag -->
        <div v-if="currentStep === 3" class="step-pane">
          <h3>选择镜像标签</h3>
          <p>输入要发布的镜像标签</p>

          <div class="form-group">
            <label>镜像标签 (如: v1.0.0, main-abc123) *</label>
            <input
              v-model="form.imageTag"
              type="text"
              class="form-input"
              placeholder="例如: v1.0.0 或 main-abc123"
            />
          </div>

          <!-- TODO: Show recent tags from container registry -->
          <div class="recent-tags">
            <h4>最近使用的标签</h4>
            <p>TODO: Load recent tags</p>
          </div>
        </div>

        <!-- Step 4: Review & Confirm -->
        <div v-if="currentStep === 4" class="step-pane">
          <h3>确认发布</h3>
          <p>检查以下信息，确认无误后点击发布</p>

          <div class="review-info">
            <div class="review-item">
              <strong>应用:</strong> {{ selectedApp?.name }}
            </div>
            <div class="review-item">
              <strong>集群:</strong> {{ selectedCluster?.name }}
            </div>
            <div class="review-item">
              <strong>Namespace:</strong> {{ selectedClusterMapping?.k8s_namespace }}
            </div>
            <div class="review-item">
              <strong>Workload:</strong> {{ selectedClusterMapping?.k8s_workload }}
            </div>
            <div class="review-item">
              <strong>完整镜像 URI:</strong>
              <span class="full-image">
                {{ completeImageUri }}
              </span>
            </div>
          </div>

          <!-- TODO: Add approval flow info if needed -->
        </div>

        <!-- Step 5: Release Result -->
        <div v-if="currentStep === 5" class="step-pane">
          <h3>发布结果</h3>
          <!-- TODO: Show release result -->
          <p>TODO: Display release result</p>
        </div>
      </div>

      <!-- Navigation Buttons -->
      <div class="wizard-footer">
        <button
          v-if="currentStep > 1"
          class="btn-secondary"
          @click="previousStep"
        >
          ← 上一步
        </button>

        <button
          v-if="currentStep < 4"
          class="btn-primary"
          @click="nextStep"
          :disabled="!canProceedToNextStep()"
        >
          下一步 →
        </button>

        <button
          v-if="currentStep === 4"
          class="btn-success"
          @click="submitRelease"
        >
          🚀 立即发布
        </button>

        <div v-if="currentStep === 5" class="completion-message">
          <p>✅ 发布已完成，请查看发布历史</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { ClusterMapping, Application, Cluster } from '@/types/api'
import { getApplications, getClusters, createRelease } from '@/api/metadata'
import { getClusterMappingsByApp } from '@/api/cluster-mapping'

// Use ClusterMapping from api types instead of defining locally

// State
const steps = ['选择应用', '选择集群', '选择镜像', '确认发布', '发布结果']
const currentStep = ref(1)

const form = ref({
  applicationId: null as number | null,
  clusterId: null as string | number | null,
  imageTag: ''
})

const applications = ref<Application[]>([])
const clusters = ref<Cluster[]>([])
const availableClusters = ref<Cluster[]>([])
const clusterMappings = ref<ClusterMapping[]>([])
const appSearchKeyword = ref('')

// Computed
const filteredApplications = computed(() => {
  if (!appSearchKeyword.value) {
    return applications.value
  }
  const keyword = appSearchKeyword.value.toLowerCase()
  return applications.value.filter(app =>
    app.name.toLowerCase().includes(keyword) ||
    app.image_name.toLowerCase().includes(keyword)
  )
})

const selectedApp = computed(() => {
  return applications.value.find(app => app.id === form.value.applicationId)
})

const selectedCluster = computed(() => {
  return availableClusters.value.find(c => c.id === form.value.clusterId)
})

const selectedClusterMapping = computed(() => {
  return clusterMappings.value.find(m => m.cluster_id === form.value.clusterId)
})

const completeImageUri = computed(() => {
  if (!selectedCluster.value || !selectedApp.value || !form.value.imageTag) {
    return '(未完成)'
  }
  const prefix = selectedCluster.value.registry_prefix || 'registry.example.com'
  return `${prefix}/${selectedApp.value.image_name}:${form.value.imageTag}`
})

// Functions
const loadApplications = async () => {
  try {
    // Load applications with search keyword (use a large pageSize to get all matching apps)
    const data = await getApplications(1, 50, appSearchKeyword.value)
    applications.value = data.data
    
    // Also load clusters
    const clustersData = await getClusters()
    clusters.value = clustersData
  } catch (error) {
    console.error('Failed to load applications:', error)
  }
}

const onClusterChange = async () => {
  if (form.value.applicationId && form.value.clusterId) {
    try {
      const mappings = await getClusterMappingsByApp(form.value.applicationId)
      clusterMappings.value = mappings
    } catch (error) {
      console.error('Failed to load cluster mapping:', error)
    }
  }
}

const canProceedToNextStep = (): boolean => {
  switch (currentStep.value) {
    case 1:
      return form.value.applicationId !== null
    case 2:
      return form.value.clusterId !== null
    case 3:
      return form.value.imageTag.trim() !== ''
    default:
      return true
  }
}

const nextStep = () => {
  if (canProceedToNextStep()) {
    currentStep.value++
  }
}

const previousStep = () => {
  if (currentStep.value > 1) {
    currentStep.value--
  }
}

const submitRelease = async () => {
  try {
    if (!form.value.applicationId || !form.value.clusterId || !form.value.imageTag.trim()) {
      alert('请完成所有步骤')
      return
    }
    
    await createRelease({
      app_id: form.value.applicationId,
      cluster_id: form.value.clusterId,
      image_tag: form.value.imageTag
    })
    
    currentStep.value = 5
  } catch (error) {
    console.error('Failed to submit release:', error)
    alert('发布提交失败')
  }
}

// Lifecycle
onMounted(async () => {
  await loadApplications()
})
</script>

<style scoped>
.k8s-release-page {
  padding: 24px;
  max-width: 1000px;
  margin: 0 auto;
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

.wizard-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.steps-indicator {
  display: flex;
  background: #f5f5f5;
  border-bottom: 1px solid #eee;
  overflow-x: auto;
}

.step {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 16px;
  cursor: pointer;
  transition: background 0.2s;
  min-width: 150px;
}

.step:hover {
  background: #efefef;
}

.step.active {
  background: #e6f7ff;
}

.step-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #ddd;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  margin-right: 12px;
}

.step.active .step-number {
  background: #1890ff;
}

.step-title {
  font-size: 14px;
  color: #666;
}

.step.active .step-title {
  color: #1890ff;
  font-weight: 600;
}

.step-content {
  padding: 32px;
  min-height: 400px;
}

.step-pane h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #1a1a1a;
}

.step-pane p {
  margin: 0 0 24px 0;
  color: #666;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #333;
}

.form-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-input:focus {
  outline: none;
  border-color: #1890ff;
  box-shadow: 0 0 0 3px rgba(24, 144, 255, 0.1);
}

.form-input:disabled {
  background: #f5f5f5;
  cursor: not-allowed;
}

.no-result {
  margin-top: 8px;
  padding: 8px 12px;
  color: #999;
  font-size: 13px;
  text-align: center;
}

.info-box {
  background: #f0f8ff;
  border: 1px solid #b3e5fc;
  border-radius: 4px;
  padding: 16px;
  margin-top: 16px;
}

.info-box h4 {
  margin: 0 0 12px 0;
  color: #1890ff;
}

.info-item {
  margin-bottom: 8px;
  font-size: 14px;
}

.recent-tags {
  margin-top: 24px;
  padding: 16px;
  background: #fafafa;
  border-radius: 4px;
}

.recent-tags h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
}

.review-info {
  background: #f0f8ff;
  border: 1px solid #b3e5fc;
  border-radius: 4px;
  padding: 20px;
}

.review-item {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #d6ecf5;
}

.review-item:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.full-image {
  font-family: 'Courier New', monospace;
  background: white;
  padding: 8px 12px;
  border-radius: 4px;
  display: block;
  margin-top: 4px;
  word-break: break-all;
  color: #d32f2f;
  font-weight: 600;
}

.wizard-footer {
  display: flex;
  gap: 12px;
  padding: 24px 32px;
  background: #f5f5f5;
  border-top: 1px solid #eee;
  justify-content: flex-end;
  align-items: center;
}

.btn-primary,
.btn-secondary,
.btn-success {
  padding: 10px 24px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 600;
}

.btn-primary {
  background: #1890ff;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #40a9ff;
}

.btn-primary:disabled {
  background: #d9d9d9;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  color: #1a1a1a;
  border: 1px solid #ddd;
}

.btn-secondary:hover {
  background: #fafafa;
}

.btn-success {
  background: #52c41a;
  color: white;
}

.btn-success:hover {
  background: #73d13d;
}

.completion-message {
  color: #52c41a;
  font-size: 16px;
  font-weight: 600;
}
</style>
