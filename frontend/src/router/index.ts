/**
 * router/index.ts - 路由配置
 */

import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// 导入页面组件
import ReleaseFlow from '@/views/ReleaseFlow.vue'
import ReleaseHistory from '@/views/ReleaseHistory.vue'
import ReleaseDetail from '@/views/ReleaseDetail.vue'

// 路由定义
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/release'
  },
  {
    path: '/release',
    name: 'ReleaseFlow',
    component: ReleaseFlow,
    meta: {
      title: '发布向导'
    }
  },
  {
    path: '/history',
    name: 'ReleaseHistory',
    component: ReleaseHistory,
    meta: {
      title: '发布历史'
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
