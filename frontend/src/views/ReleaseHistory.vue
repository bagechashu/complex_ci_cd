<template>
  <div class="release-history">
    <n-card title="发布历史" size="large">
      <!-- Filters -->
      <div class="filters">
        <n-space>
          <n-select
            v-model:value="filterAppId"
            :options="appFilterOptions"
            placeholder="筛选应用"
            clearable
            style="width: 150px"
          />
          <n-select
            v-model:value="filterEnvId"
            :options="envFilterOptions"
            placeholder="筛选环境"
            clearable
            style="width: 150px"
          />
          <n-select
            v-model:value="filterStatus"
            :options="statusFilterOptions"
            placeholder="筛选状态"
            clearable
            style="width: 150px"
          />
          <n-button @click="fetchHistory">刷新</n-button>
        </n-space>
      </div>

      <!-- Table -->
      <n-data-table
        :columns="columns"
        :data="filteredHistory"
        :loading="releaseStore.isLoadingHistory"
        :pagination="false"
        :bordered="false"
        :single-line="false"
      />

      <!-- Pagination -->
      <div class="pagination">
        <n-pagination
          v-model:page="currentPage"
          :page-size="pageSize"
          :item-count="totalCount"
          @update:page="handlePageChange"
        />
      </div>
    </n-card>

    <!-- Detail Modal -->
    <n-modal
      v-model:show="showDetailModal"
      title="发布详情"
      preset="dialog"
      size="large"
      :show-icon="false"
    >
      <div v-if="selectedRelease" class="detail-content">
        <n-descriptions :columns="2" size="small">
          <n-descriptions-item label="应用">
            {{ appStore.getApplicationById(selectedRelease.app_id)?.name }}
          </n-descriptions-item>
          <n-descriptions-item label="环境">
            {{ appStore.getEnvironmentById(selectedRelease.env_id)?.name }}
          </n-descriptions-item>
          <n-descriptions-item label="集群">
            {{ appStore.getClusterById(selectedRelease.cluster_id)?.name }}
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-tag :type="getStatusType(selectedRelease.status)" round>
              {{ getStatusLabel(selectedRelease.status) }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="镜像">
            {{ selectedRelease.image }}
          </n-descriptions-item>
          <n-descriptions-item label="发起人">
            {{ selectedRelease.triggered_by }}
          </n-descriptions-item>
          <n-descriptions-item label="发起时间">
            {{ formatDateTime(selectedRelease.created_at) }}
          </n-descriptions-item>
          <n-descriptions-item label="完成时间">
            {{ formatDateTime(selectedRelease.completed_at) }}
          </n-descriptions-item>
        </n-descriptions>

        <!-- Events -->
        <n-divider />
        <h4>事件日志</h4>
        <div v-if="selectedEvents && selectedEvents.length > 0" class="events">
          <div v-for="event in selectedEvents" :key="event.id" class="event">
            <span class="time">{{ formatDateTime(event.created_at, 'HH:mm:ss') }}</span>
            <span class="type">[{{ event.type }}]</span>
            <span class="message">{{ event.message }}</span>
          </div>
        </div>
        <div v-else class="no-events">暂无事件日志</div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { h } from 'vue'
import {
  NCard,
  NSpace,
  NSelect,
  NButton,
  NDataTable,
  NPagination,
  NTag,
  NModal,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NPopconfirm
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useReleaseStore } from '@/stores/releaseStore'
import { useAppStore } from '@/stores/appStore'
import { useUiStore } from '@/stores/uiStore'
import { formatDateTime, getStatusLabel } from '@/utils/format'
import type { ReleaseResponse, ReleaseEvent } from '@/types/api'

const releaseStore = useReleaseStore()
const appStore = useAppStore()
const uiStore = useUiStore()

// ============ State ============
const currentPage = ref(1)
const pageSize = 20
const filterAppId = ref<number | null>(null)
const filterEnvId = ref<number | null>(null)
const filterStatus = ref<string | null>(null)
const showDetailModal = ref(false)
const selectedRelease = ref<ReleaseResponse | null>(null)
const selectedEvents = ref<ReleaseEvent[]>([])

// ============ Computed ============
const appFilterOptions = computed(() =>
  appStore.applications.map(app => ({
    label: app.name,
    value: app.id
  }))
)

const envFilterOptions = computed(() =>
  appStore.environments.map(env => ({
    label: env.name,
    value: env.id
  }))
)

const statusFilterOptions = computed(() => [
  { label: '待发布', value: 'pending' },
  { label: '验证中', value: 'validating' },
  { label: '部署中', value: 'deploying' },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' },
  { label: '已回滚', value: 'rolled_back' }
])

const filteredHistory = computed(() => {
  let result = releaseStore.releaseHistoryFormatted

  if (filterAppId.value) {
    result = result.filter(r => r.app_id === filterAppId.value)
  }
  if (filterEnvId.value) {
    result = result.filter(r => r.env_id === filterEnvId.value)
  }
  if (filterStatus.value) {
    result = result.filter(r => r.status === filterStatus.value)
  }

  return result
})

const totalCount = computed(() => filteredHistory.value.length)

const columns = computed<DataTableColumns<ReleaseResponse>>(() => [
  {
    title: '应用',
    key: 'app_name',
    width: 120
  },
  {
    title: '环境',
    key: 'env_name',
    width: 100
  },
  {
    title: '集群',
    key: 'cluster_name',
    width: 120
  },
  {
    title: '镜像',
    key: 'image',
    width: 200,
    ellipsis: true
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) =>
      h(NTag, { type: getStatusType(row.status) }, () =>
        getStatusLabel(row.status)
      )
  },
  {
    title: '发起人',
    key: 'triggered_by',
    width: 100
  },
  {
    title: '发起时间',
    key: 'created_at',
    width: 160,
    render: (row) => formatDateTime(row.created_at, 'MM-DD HH:mm')
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    align: 'center',
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          {
            type: 'info',
            size: 'small',
            onClick: () => showDetail(row)
          },
          () => '详情'
        ),
        row.status === 'success'
          ? h(
              NPopconfirm,
              {
                onPositiveClick: () => handleRollback(row.id)
              },
              {
                default: () => '确认回滚到此版本？',
                trigger: () =>
                  h(NButton, { type: 'warning', size: 'small' }, () => '回滚')
              }
            )
          : null
      ])
  }
])

// ============ Methods ============
const fetchHistory = async () => {
  try {
    const offset = (currentPage.value - 1) * pageSize
    await releaseStore.fetchReleaseHistory(pageSize, offset)
  } catch (error) {
    uiStore.error('获取发布历史失败')
  }
}

const handlePageChange = () => {
  fetchHistory()
}

const showDetail = async (release: ReleaseResponse) => {
  selectedRelease.value = release
  try {
    const events = await releaseStore.fetchReleaseEvents(release.id)
    selectedEvents.value = events || []
  } catch {
    selectedEvents.value = []
  }
  showDetailModal.value = true
}

const handleRollback = async (releaseId: number) => {
  try {
    await releaseStore.rollback(releaseId)
    uiStore.success('回滚已启动')
    showDetailModal.value = false
    // 刷新列表
    setTimeout(() => fetchHistory(), 2000)
  } catch (error) {
    uiStore.error('回滚失败')
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

// ============ Lifecycle ============
onMounted(() => {
  fetchHistory()
})
</script>

<style scoped>
.release-history {
  max-width: 1200px;
  margin: 0 auto;
}

.filters {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #e8e8e8;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

.detail-content {
  padding: 20px 0;
}

.events {
  background-color: #f5f7fa;
  border-radius: 4px;
  padding: 12px;
  margin-top: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.event {
  display: flex;
  gap: 12px;
  padding: 8px 0;
  font-size: 13px;
  border-bottom: 1px solid #e8e8e8;
}

.event:last-child {
  border-bottom: none;
}

.event .time {
  color: #999;
  min-width: 70px;
  font-family: monospace;
}

.event .type {
  color: #667eea;
  font-weight: 600;
  min-width: 100px;
}

.event .message {
  flex: 1;
  color: #333;
}

.no-events {
  text-align: center;
  color: #999;
  padding: 20px;
}
</style>
