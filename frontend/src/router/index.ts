/**
 * router/index.ts - 路由配置
 */

import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// 导入页面组件
import KubernetesRelease from '@/views/KubernetesRelease.vue'
import ClusterConfig from '@/views/ClusterConfig.vue'
import ShellTask from '@/views/ShellTask.vue'
import ServerConfig from '@/views/ServerConfig.vue'
import ReleaseHistory from '@/views/ReleaseHistory.vue'
import ExecutionHistory from '@/views/ExecutionHistory.vue'
import ReleaseDetail from '@/views/ReleaseDetail.vue'

// 路由定义
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/k8s-release'
  },
  {
    path: '/k8s-release',
    name: 'KubernetesRelease',
    component: KubernetesRelease,
    meta: {
      title: 'K8s 发布'
    }
  },
  {
    path: '/cluster-config',
    name: 'ClusterConfig',
    component: ClusterConfig,
    meta: {
      title: '集群配置'
    }
  },
  {
    path: '/shell-task',
    name: 'ShellTask',
    component: ShellTask,
    meta: {
      title: 'Shell 任务'
    }
  },
  {
    path: '/server-config',
    name: 'ServerConfig',
    component: ServerConfig,
    meta: {
      title: '服务器配置'
    }
  },
  {
    path: '/release-history',
    name: 'ReleaseHistory',
    component: ReleaseHistory,
    meta: {
      title: '发布历史'
    }
  },
  {
    path: '/execution-history',
    name: 'ExecutionHistory',
    component: ExecutionHistory,
    meta: {
      title: '执行历史'
    }
  },
  {
    path: '/detail/:id',
    name: 'ReleaseDetail',
    component: ReleaseDetail,
    meta: {
      title: '发布详情'
    }
  }
]

// 创建路由器
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title || '页面'} - 发布控制系统`
  next()
})

export default router
