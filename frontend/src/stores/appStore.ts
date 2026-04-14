/**
 * appStore.ts - 应用/环境/集群元数据状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Application,
  Environment,
  Cluster,
  DeploymentTarget
} from '@/types/api'
import { metadataAPI } from '@/api/metadata'

export const useAppStore = defineStore('app', () => {
  // ============ State ============
  const applications = ref<Application[]>([])
  const environments = ref<Environment[]>([])
  const clusters = ref<Cluster[]>([])
  const deploymentTargets = ref<DeploymentTarget[]>([])

  const applicationsLoading = ref(false)
  const environmentsLoading = ref(false)
  const clustersLoading = ref(false)
  const deploymentTargetsLoading = ref(false)

  const error = ref<string | null>(null)

  // ============ Getters ============
  const getApplicationById = computed(() => (id: number) => {
    return applications.value.find(app => app.id === id)
  })

  const getEnvironmentById = computed(() => (id: number) => {
    return environments.value.find(env => env.id === id)
  })

  const getClusterById = computed(() => (id: number) => {
    return clusters.value.find(cluster => cluster.id === id)
  })

  /**
   * 获取指定应用在指定环境下可用的集群
   */
  const getAvailableClusters = computed(
    () => (appId: number, envId: number) => {
      return deploymentTargets.value
        .filter(dt => dt.app_id === appId && dt.env_id === envId)
        .map(dt => getClusterById.value(dt.cluster_id))
        .filter((c): c is Cluster => c !== undefined)
    }
  )

  // ============ Actions ============
  const fetchApplications = async () => {
    applicationsLoading.value = true
    error.value = null
    try {
      applications.value = await metadataAPI.getApplications()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取应用列表失败'
    } finally {
      applicationsLoading.value = false
    }
  }

  const fetchEnvironments = async () => {
    environmentsLoading.value = true
    error.value = null
    try {
      environments.value = await metadataAPI.getEnvironments()
      // 按 rank 排序
      environments.value.sort((a, b) => a.rank - b.rank)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取环境列表失败'
    } finally {
      environmentsLoading.value = false
    }
  }

  const fetchClusters = async () => {
    clustersLoading.value = true
    error.value = null
    try {
      clusters.value = await metadataAPI.getClusters()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取集群列表失败'
    } finally {
      clustersLoading.value = false
    }
  }

  const fetchDeploymentTargets = async () => {
    deploymentTargetsLoading.value = true
    error.value = null
    try {
      deploymentTargets.value = await metadataAPI.getDeploymentTargets()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取部署目标失败'
    } finally {
      deploymentTargetsLoading.value = false
    }
  }

  /**
   * 初始化所有元数据
   */
  const initializeMetadata = async () => {
    try {
      await Promise.all([
        fetchApplications(),
        fetchEnvironments(),
        fetchClusters(),
        fetchDeploymentTargets()
      ])
    } catch (err) {
      error.value = err instanceof Error ? err.message : '初始化元数据失败'
    }
  }

  return {
    // State
    applications,
    environments,
    clusters,
    deploymentTargets,
    applicationsLoading,
    environmentsLoading,
    clustersLoading,
    deploymentTargetsLoading,
    error,

    // Getters
    getApplicationById,
    getEnvironmentById,
    getClusterById,
    getAvailableClusters,

    // Actions
    fetchApplications,
    fetchEnvironments,
    fetchClusters,
    fetchDeploymentTargets,
    initializeMetadata
  }
})
