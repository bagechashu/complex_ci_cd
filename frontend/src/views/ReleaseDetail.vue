<template>
  <div class="release-detail">
    <n-card size="large">
      <template #header>
        <div class="card-header">
          <span>发布详情 #{{ releaseId }}</span>
          <n-button text-color="#999" @click="goBack">返回</n-button>
        </div>
      </template>

      <!-- Loading State -->
      <n-spin v-if="loading" />

      <!-- Error State -->
      <n-alert
        v-if="error && !loading"
        type="error"
        closable
        @close="error = null"
      >
        {{ error }}
      </n-alert>

      <!-- Content -->
      <div v-if="release && !loading" class="content">
        <!-- Basic Info -->
        <div class="section">
          <h3>基本信息</h3>
          <n-descriptions :columns="2" size="small">
            <n-descriptions-item label="应用">
              {{ appStore.getApplicationById(release.app_id)?.name }}
            </n-descriptions-item>
            <n-descriptions-item label="环境">
              {{ appStore.getEnvironmentById(release.env_id)?.name }}
            </n-descriptions-item>
            <n-descriptions-item label="集群">
              {{ appStore.getClusterById(release.cluster_id)?.name }}
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="getStatusType(release.status)" round>
                {{ getStatusLabel(release.status) }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="当前镜像">
              {{ release.image }}
            </n-descriptions-item>
            <n-descriptions-item label="上一版本">
              {{ release.previous_image || '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="发起人">
              {{ release.triggered_by }}
            </n-descriptions-item>
            <n-descriptions-item label="用时">
              {{ formatDuration(duration) }}
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <!-- Progress -->
        <div class="section">
          <h3>进度</h3>
          <div class="progress-item">
            <n-progress :percentage="progressPercentage" :height="24" />
          </div>
        </div>

        <!-- Events -->
        <div class="section">
          <div class="section-header">
            <h3>事件日志</h3>
            <n-space>
              <n-button
                text
                type="info"
                size="small"
                @click="exportLogs"
              >
                导出日志
              </n-button>
              <n-button text type="info" size="small" @click="refreshEvents">
                刷新
              </n-button>
            </n-space>
          </div>

          <div v-if="events && events.length > 0" class="events">
            <div v-for="event in events" :key="event.id" class="event">
              <div class="event-header">
                <span class="time">{{ formatDateTime(event.created_at, 'YYYY-MM-DD HH:mm:ss') }}</span>
                <n-tag type="info" size="small">{{ event.type }}</n-tag>
              </div>
              <div class="event-message">{{ event.message }}</div>
              <div v-if="event.details" class="event-details">
                {{ event.details }}
              </div>
            </div>
          </div>
          <div v-else class="no-events">暂无事件日志</div>
        </div>

        <!-- Error Details -->
        <div v-if="release.error_msg" class="section">
          <n-alert type="error" :title="`错误: ${release.status}`">
            {{ release.error_msg }}
          </n-alert>
        </div>

        <!-- Actions -->
        <div class="actions">
          <n-button
            v-if="release.status === 'success'"
            type="warning"
            @click="handleRollback"
          >
            回滚
          </n-button>
          <n-button
            v-if="['failed', 'pending'].includes(release.status)"
            type="primary"
            @click="handleRetry"
          >
            重新发布
          </n-button>
        </div>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NCard,
  NSpin,
  NAlert,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NProgress,
  NSpace,
  NButton,
  NPopconfirm
} from 'naive-ui'
import { useReleaseStore } from '@/stores/releaseStore'
import { useAppStore } from '@/stores/appStore'
import { useUiStore } from '@/stores/uiStore'
import { formatDateTime, getStatusLabel, formatDuration } from '@/utils/format'
import type { ReleaseResponse, ReleaseEvent } from '@/types/api'

const router = useRouter()
const route = useRoute()

const releaseStore = useReleaseStore()
const appStore = useAppStore()
const uiStore = useUiStore()

// ============ State ============
const releaseId = computed(() => parseInt(route.params.id as string))
const loading = ref(true)
const error = ref<string | null>(null)
const release = ref<ReleaseResponse | null>(null)
const events = ref<ReleaseEvent[]>([])
let pollingInterval: NodeJS.Timeout | null = null

// ============ Computed ============
const progressPercentage = computed(() => {
  if (!events.value || events.value.length === 0) return 0

  const eventTypes = events.value.map(e => e.type)
  const hasStarted = eventTypes.includes('started')
  const hasValidating = eventTypes.includes('validating')
  const hasDeploying = eventTypes.includes('deploying')
  const hasSuccess = eventTypes.includes('success')
  const hasFailed = eventTypes.includes('failed')

  if (hasFailed) return 0
  if (hasSuccess) return 100
  if (hasDeploying) return 75
  if (hasValidating) return 50
  if (hasStarted) return 25
  return 10
})

const duration = computed(() => {
  if (!release.value?.started_at || !release.value?.completed_at) return null

  const start = new Date(release.value.started_at).getTime()
  const end = new Date(release.value.completed_at).getTime()
  return end - start
})

// ============ Methods ============
const loadRelease = async () => {
  loading.value = true
  error.value = null

  try {
    const data = await releaseStore.fetchReleaseStatus(releaseId.value)
    release.value = data
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

const refreshEvents = async () => {
  try {
    const eventList = await releaseStore.fetchReleaseEvents(releaseId.value)
    events.value = eventList || []
  } catch (err) {
    uiStore.error('刷新事件失败')
  }
}

const startPolling = () => {
  if (pollingInterval) clearInterval(pollingInterval)

  pollingInterval = setInterval(async () => {
    try {
      await loadRelease()
      if (release.value && ['success', 'failed'].includes(release.value.status)) {
        if (pollingInterval) clearInterval(pollingInterval)
      }
    } catch (err) {
      // 忽略轮询错误
    }
  }, 2000)
}

const exportLogs = async () => {
  try {
    const logs = events.value
      .map(
        e =>
          `[${e.created_at}] ${e.type}: ${e.message}`
      )
      .join('\n')

    const element = document.createElement('a')
    element.setAttribute(
      'href',
      'data:text/plain;charset=utf-8,' + encodeURIComponent(logs)
    )
    element.setAttribute('download', `release-${releaseId.value}-logs.txt`)
    element.style.display = 'none'
    document.body.appendChild(element)
    element.click()
    document.body.removeChild(element)

    uiStore.success('日志已导出')
  } catch (err) {
    uiStore.error('导出失败')
  }
}

const handleRollback = async () => {
  try {
    await releaseStore.rollback(releaseId.value)
    uiStore.success('回滚已启动')
    startPolling()
  } catch (err) {
    uiStore.error('回滚失败')
  }
}

const handleRetry = () => {
  router.push({
    name: 'ReleaseFlow',
    query: { retry: releaseId.value }
  })
}

const goBack = () => {
  router.push({ name: 'ReleaseHistory' })
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

// ============ Lifecycle ============
onMounted(async () => {
  await loadRelease()
  await refreshEvents()

  if (release.value && !['success', 'failed', 'rolled_back'].includes(release.value.status)) {
    startPolling()
  }
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})
</script>

<style scoped>
.release-detail {
  max-width: 1000px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.content {
  padding: 20px 0;
}

.section {
  margin-bottom: 30px;
}

.section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  border-bottom: 2px solid #667eea;
  padding-bottom: 8px;
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

.progress-item {
  padding: 12px 0;
}

.events {
  background-color: #ffffff;
  border-radius: 4px;
  padding: 12px;
}

.event {
  padding: 12px;
  border-bottom: 1px solid #e8e8e8;
  margin-bottom: 12px;
}

.event:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.event-header {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 8px;
}

.event-header .time {
  font-size: 12px;
  color: #999;
  font-family: monospace;
  min-width: 180px;
}

.event-message {
  font-size: 14px;
  color: #333;
  margin-bottom: 4px;
}

.event-details {
  font-size: 12px;
  color: #666;
  background-color: white;
  padding: 8px;
  border-left: 2px solid #667eea;
  margin-top: 4px;
  overflow-x: auto;
}

.no-events {
  text-align: center;
  color: #999;
  padding: 30px;
}

.actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e8e8e8;
}
</style>
