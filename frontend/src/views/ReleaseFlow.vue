<template>
  <div class="release-flow">
    <n-card title="发布向导" size="large">
      <!-- Steps -->
      <n-steps
        v-model:current="currentStep"
        :status="stepStatus"
        :process-dot="true"
      >
        <n-step title="选择应用" />
        <n-step title="选择环境" />
        <n-step title="选择集群" />
        <n-step title="选择镜像" />
        <n-step title="确认并发布" />
      </n-steps>

      <!-- Content -->
      <div class="step-content">
        <!-- Step 1: Select Application -->
        <div v-if="currentStep === 1" class="step-item">
          <n-form label-placement="left">
            <n-form-item label="应用" required>
              <n-select
                v-model:value="form.app_id"
                :options="appOptions"
                placeholder="请选择应用"
                clearable
              />
            </n-form-item>
          </n-form>
        </div>

        <!-- Step 2: Select Environment -->
        <div v-if="currentStep === 2" class="step-item">
          <n-form label-placement="left">
            <n-form-item label="环境" required>
              <n-select
                v-model:value="form.env_id"
                :options="envOptions"
                placeholder="请选择环境"
                clearable
              />
            </n-form-item>
          </n-form>
        </div>

        <!-- Step 3: Select Cluster -->
        <div v-if="currentStep === 3" class="step-item">
          <n-form label-placement="left">
            <n-form-item label="目标集群" required>
              <n-select
                v-model:value="form.cluster_id"
                :options="clusterOptions"
                placeholder="请选择集群"
                clearable
              />
            </n-form-item>
          </n-form>
        </div>

        <!-- Step 4: Select Image Tag -->
        <div v-if="currentStep === 4" class="step-item">
          <n-form label-placement="left">
            <n-form-item label="镜像标签" required>
              <n-input
                v-model:value="form.image"
                placeholder="例如: myapp:v1.0.0"
                clearable
              />
            </n-form-item>
          </n-form>
        </div>

        <!-- Step 5: Confirm and Release -->
        <div v-if="currentStep === 5" class="step-item">
          <div class="confirm-info">
            <n-row :gutter="[20, 20]">
              <n-col :span="12">
                <div class="info-item">
                  <span class="label">应用:</span>
                  <span class="value">{{ selectedAppName }}</span>
                </div>
              </n-col>
              <n-col :span="12">
                <div class="info-item">
                  <span class="label">环境:</span>
                  <span class="value">{{ selectedEnvName }}</span>
                </div>
              </n-col>
              <n-col :span="12">
                <div class="info-item">
                  <span class="label">集群:</span>
                  <span class="value">{{ selectedClusterName }}</span>
                </div>
              </n-col>
              <n-col :span="12">
                <div class="info-item">
                  <span class="label">镜像:</span>
                  <span class="value">{{ form.image }}</span>
                </div>
              </n-col>
            </n-row>
          </div>
        </div>
      </div>

      <!-- Release Progress (after submission) -->
      <div v-if="releaseStore.currentRelease" class="release-progress">
        <n-divider />
        <h3>发布进度</h3>
        <div class="progress-item">
          <span class="status-text">状态: </span>
          <n-tag
            :type="getStatusType(releaseStore.currentRelease.status)"
            round
          >
            {{ getStatusLabel(releaseStore.currentRelease.status) }}
          </n-tag>
        </div>
        <div class="progress-item">
          <span class="status-text">进度:</span>
          <n-progress
            :percentage="releaseStore.progressPercentage"
            :height="24"
            :show-indicator="true"
          />
        </div>
        <div class="event-log">
          <h4>事件日志</h4>
          <div class="event-list">
            <div
              v-for="(event, i) in releaseStore.releaseEvents.slice(-5)"
              :key="i"
              class="event-item"
            >
              <span class="time">{{ formatDateTime(event.created_at, 'HH:mm:ss') }}</span>
              <span class="message">{{ event.message }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="actions">
        <n-button
          v-if="currentStep > 1 && !releaseStore.currentRelease"
          @click="previousStep"
        >
          上一步
        </n-button>
        <n-button
          v-if="currentStep < 5 && !releaseStore.currentRelease"
          type="primary"
          @click="nextStep"
        >
          下一步
        </n-button>
        <n-button
          v-if="currentStep === 5 && !releaseStore.currentRelease"
          type="success"
          size="large"
          :loading="releaseStore.isCreatingRelease"
          @click="submitRelease"
        >
          确认发布
        </n-button>
        <n-button
          v-if="releaseStore.currentRelease"
          @click="resetForm"
        >
          返回
        </n-button>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NCard, NSteps, NStep, NForm, NFormItem, NSelect, NInput, NTag, NProgress, NDivider, NRow, NCol, NButton } from 'naive-ui'
import { useReleaseStore } from '@/stores/releaseStore'
import { useAppStore } from '@/stores/appStore'
import { useUiStore } from '@/stores/uiStore'
import { formatDateTime, getStatusLabel } from '@/utils/format'
import type { ReleaseFlowForm } from '@/types/models'

const releaseStore = useReleaseStore()
const appStore = useAppStore()
const uiStore = useUiStore()

// ============ State ============
const currentStep = ref(1)
const form = ref<ReleaseFlowForm>({
  app_id: null,
  env_id: null,
  cluster_id: null,
  image: ''
})

// ============ Computed ============
const stepStatus = computed(() => {
  if (releaseStore.currentRelease?.status === 'failed') return 'error'
  if (releaseStore.currentRelease?.status === 'success') return 'finish'
  return 'process'
})

const appOptions = computed(() =>
  appStore.applications.map(app => ({
    label: app.name,
    value: app.id
  }))
)

const envOptions = computed(() =>
  appStore.environments.map(env => ({
    label: env.name,
    value: env.id
  }))
)

const clusterOptions = computed(() => {
  if (!form.value.app_id || !form.value.env_id) return []
  const clusters = appStore.getAvailableClusters(form.value.app_id, form.value.env_id)
  return clusters.map(c => ({
    label: c.name,
    value: c.id
  }))
})

const selectedAppName = computed(
  () => appStore.getApplicationById(form.value.app_id || 0)?.name || '-'
)

const selectedEnvName = computed(
  () => appStore.getEnvironmentById(form.value.env_id || 0)?.name || '-'
)

const selectedClusterName = computed(
  () => appStore.getClusterById(form.value.cluster_id || 0)?.name || '-'
)

// ============ Methods ============
const validateStep = (): boolean => {
  switch (currentStep.value) {
    case 1:
      if (!form.value.app_id) {
        uiStore.warning('请选择应用')
        return false
      }
      break
    case 2:
      if (!form.value.env_id) {
        uiStore.warning('请选择环境')
        return false
      }
      break
    case 3:
      if (!form.value.cluster_id) {
        uiStore.warning('请选择集群')
        return false
      }
      break
    case 4:
      if (!form.value.image) {
        uiStore.warning('请输入镜像标签')
        return false
      }
      break
  }
  return true
}

const nextStep = () => {
  if (validateStep()) {
    currentStep.value++
  }
}

const previousStep = () => {
  currentStep.value--
}

const submitRelease = async () => {
  try {
    await releaseStore.createRelease({
      app_id: form.value.app_id!,
      env_id: form.value.env_id!,
      cluster_id: form.value.cluster_id!,
      image: form.value.image
    })
    
    uiStore.success('发布已启动')
    
    // 启动轮询
    if (releaseStore.currentRelease?.id) {
      await releaseStore.startPolling(releaseStore.currentRelease.id)
    }
  } catch (error) {
    uiStore.error(error instanceof Error ? error.message : '发布失败')
  }
}

const resetForm = () => {
  releaseStore.clearCurrent()
  currentStep.value = 1
  form.value = {
    app_id: null,
    env_id: null,
    cluster_id: null,
    image: ''
  }
}

const getStatusType = (status: string) => {
  const typeMap: Record<string, 'success' | 'error' | 'warning' | 'info'> = {
    pending: 'warning',
    validating: 'info',
    deploying: 'info',
    success: 'success',
    failed: 'error',
    rolled_back: 'warning'
  }
  return typeMap[status] || 'info'
}

// 当集群选项变化时，重置集群选择
watch(() => [form.value.app_id, form.value.env_id], () => {
  form.value.cluster_id = null
})
</script>

<style scoped>
.release-flow {
  max-width: 800px;
  margin: 0 auto;
}

.step-content {
  margin: 30px 0;
  min-height: 200px;
}

.step-item {
  padding: 20px 0;
}

.confirm-info {
  background-color: #ffffff;
  padding: 20px;
  border-radius: 4px;
}

.info-item {
  display: flex;
  gap: 12px;
}

.info-item .label {
  font-weight: 600;
  min-width: 80px;
}

.info-item .value {
  flex: 1;
  word-break: break-all;
}

.release-progress {
  margin-top: 20px;
}

.progress-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 16px 0;
}

.status-text {
  font-weight: 600;
  min-width: 80px;
}

.event-log {
  margin-top: 20px;
}

.event-log h4 {
  margin-bottom: 12px;
}

.event-list {
  background-color: #ffffff;
  border-radius: 4px;
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.event-item {
  display: flex;
  gap: 12px;
  padding: 8px 0;
  font-size: 13px;
  border-bottom: 1px solid #e8e8e8;
}

.event-item:last-child {
  border-bottom: none;
}

.event-item .time {
  color: #999;
  min-width: 70px;
  font-family: monospace;
}

.event-item .message {
  flex: 1;
  color: #333;
}

.actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-top: 30px;
}
</style>
