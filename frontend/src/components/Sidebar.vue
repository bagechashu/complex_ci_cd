<template>
  <aside class="sidebar" :class="{ 'sidebar-collapsed': isCollapsed }">
    <!-- Logo Section -->
    <div class="sidebar-header">
      <div class="logo-wrapper" @click="toggleCollapse">
        <div class="logo-icon">📦</div>
        <span v-if="!isCollapsed" class="logo-text">发布系统</span>
      </div>
      <button class="collapse-btn" @click="toggleCollapse" :title="isCollapsed ? '展开' : '折叠'">
        <span class="collapse-icon">{{ isCollapsed ? '→' : '←' }}</span>
      </button>
    </div>

    <!-- Menu Items with Groups -->
    <nav class="sidebar-menu">
      <div v-for="group in menuGroups" :key="group.id" class="menu-group">
        <!-- Group Header -->
        <div
          class="group-header"
          @click="toggleGroup(group.id)"
          :class="{ 'group-collapsed': !isGroupExpanded(group.id) }"
          :title="isCollapsed ? group.label : ''"
        >
          <div class="group-icon">{{ group.icon }}</div>
          <span v-if="!isCollapsed" class="group-label">{{ group.label }}</span>
          <div v-if="!isCollapsed" class="group-toggle">
            <span class="toggle-icon">{{ isGroupExpanded(group.id) ? '▼' : '▶' }}</span>
          </div>
        </div>

        <!-- Group Items -->
        <transition name="group-expand">
          <div v-if="isGroupExpanded(group.id) || isCollapsed" class="group-items">
            <div
              v-for="item in group.items"
              :key="item.name"
              class="menu-item"
              :class="{ active: isActive(item.name) }"
              @click="navigateTo(item.path)"
            >
              <div class="menu-icon">{{ item.icon }}</div>
              <span v-if="!isCollapsed" class="menu-label">{{ item.label }}</span>
              <div v-if="!isCollapsed" class="menu-tooltip">{{ item.label }}</div>
            </div>
          </div>
        </transition>
      </div>
    </nav>

    <!-- Footer Section -->
    <div class="sidebar-footer">
      <div class="footer-item" :title="isCollapsed ? '用户菜单' : ''"
        @click="showUserMenu = !showUserMenu">
        <div class="footer-icon">👤</div>
        <span v-if="!isCollapsed" class="footer-label">用户</span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

interface MenuItem {
  name: string
  path: string
  label: string
  icon: string
  meta?: {
    title?: string
  }
}

interface MenuGroup {
  id: string
  label: string
  icon: string
  items: MenuItem[]
}

const router = useRouter()
const route = useRoute()
const isCollapsed = ref(false)
const showUserMenu = ref(false)
const expandedGroups = ref<Set<string>>(new Set(['k8s-release', 'shell-command-execution']))

// Menu groups configuration
const menuGroups: MenuGroup[] = [
  {
    id: 'k8s-release',
    label: 'K8s 发布管理',
    icon: '🚀',
    items: [
      {
        name: 'KubernetesRelease',
        path: '/k8s-release',
        label: 'K8s 发布',
        icon: '🚀',
        meta: { title: 'K8s 发布' }
      },
      {
        name: 'ClusterConfig',
        path: '/cluster-config',
        label: '集群配置',
        icon: '⚙️',
        meta: { title: '集群配置' }
      },
      {
        name: 'ReleaseHistory',
        path: '/release-history',
        label: '发布历史',
        icon: '📋',
        meta: { title: '发布历史' }
      }
    ]
  },
  {
    id: 'shell-command-execution',
    label: 'Shell 命令管理',
    icon: '🐚',
    items: [
      {
        name: 'ShellCommandExecution',
        path: '/shell-command-execution',
        label: '命令执行',
        icon: '🐚',
        meta: { title: 'Shell 命令执行' }
      },
      {
        name: 'ServerConfig',
        path: '/server-config',
        label: '服务器配置',
        icon: '🖥️',
        meta: { title: '服务器配置' }
      },
      {
        name: 'ExecutionHistory',
        path: '/execution-history',
        label: '执行历史',
        icon: '📊',
        meta: { title: '执行历史' }
      }
    ]
  }
]

// Check if menu item is active
const isActive = (itemName: string): boolean => {
  return route.name === itemName
}

// Navigate to path
const navigateTo = (path: string) => {
  router.push(path)
}

// Toggle sidebar collapse state
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

// Toggle group expand/collapse
const toggleGroup = (groupId: string) => {
  if (expandedGroups.value.has(groupId)) {
    expandedGroups.value.delete(groupId)
  } else {
    expandedGroups.value.add(groupId)
  }
}

// Check if group is expanded
const isGroupExpanded = (groupId: string): boolean => {
  return expandedGroups.value.has(groupId)
}
</script>

<style scoped>
.sidebar {
  width: 250px;
  background: #f5f5f5;
  color: #1a1a1a;
  display: flex;
  flex-direction: column;
  height: 100vh;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.08);
  transition: width 0.3s ease;
  overflow-y: auto;
  overflow-x: hidden;
}

.sidebar-collapsed {
  width: 70px;
}

/* ============ Sidebar Header ============ */
.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid rgba(45, 134, 89, 0.15);
  transition: all 0.3s ease;
}

.sidebar-collapsed .sidebar-header {
  padding: 0 10px;
  justify-content: center;
}

.logo-wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  cursor: pointer;
  flex: 1;
  overflow: visible;
  transition: all 0.3s ease;
  border-radius: 0;
  padding: 0;
  height: 60px;
}

.logo-wrapper:hover {
  background-color: rgba(45, 134, 89, 0.08);
}

.sidebar-collapsed .logo-wrapper {
  flex: 1;
  justify-content: center;
  gap: 0;
  padding: 0;
}

.logo-icon {
  font-size: 24px;
  flex-shrink: 0;
  transition: font-size 0.3s ease;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  width: 32px;
}

.sidebar-collapsed .logo-icon {
  font-size: 28px;
  height: 40px;
  width: 40px;
}

.logo-text {
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.collapse-btn {
  background: transparent;
  border: none;
  color: #2d8659;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
  font-size: 14px;
  flex-shrink: 0;
  margin-left: 4px;
  height: 32px;
  width: 32px;
}

.collapse-btn:hover {
  background-color: rgba(45, 134, 89, 0.1);
}

.sidebar-collapsed .collapse-btn {
  display: none;
}

.collapse-btn:hover {
  background-color: rgba(45, 134, 89, 0.1);
}

.collapse-btn:active {
  background-color: rgba(45, 134, 89, 0.15);
}

.collapse-icon {
  display: inline-block;
  transition: transform 0.3s ease;
}

/* ============ Sidebar Menu ============ */
.sidebar-menu {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px 8px;
  gap: 8px;
  overflow-y: auto;
  overflow-x: hidden;
}

/* ============ Menu Group ============ */
.menu-group {
  margin-bottom: 0;
  transition: all 0.3s ease;
}

.menu-group:not(:first-child) {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(45, 134, 89, 0.1);
}

.sidebar-collapsed .menu-group:not(:first-child) {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 2px solid rgba(45, 134, 89, 0.12);
}

.group-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  color: #2d8659;
  user-select: none;
  font-weight: 500;
  font-size: 13px;
}

.sidebar-collapsed .group-header {
  display: none;
}

.group-header:hover {
  background-color: rgba(45, 134, 89, 0.08);
}

.group-header.group-collapsed {
  color: #666;
}

.group-icon {
  font-size: 16px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: font-size 0.3s ease;
}

.sidebar-collapsed .group-icon {
  font-size: 20px;
}

.group-label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: opacity 0.3s ease;
}

.sidebar-collapsed .group-label {
  display: none;
}

.group-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #999;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.sidebar-collapsed .group-toggle {
  display: none;
}

.toggle-icon {
  display: inline-block;
  transition: transform 0.3s ease;
}

/* ============ Group Items ============ */
.group-items {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-left: 8px;
  transition: all 0.3s ease;
}

.sidebar-collapsed .group-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 0 6px;
  align-items: center;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  color: #4a4a4a;
  user-select: none;
  font-size: 14px;
  margin-left: 8px;
}

.sidebar-collapsed .menu-item {
  justify-content: center;
  align-items: center;
  padding: 8px;
  margin: 0;
  gap: 0;
  border-radius: 6px;
  height: 40px;
  width: 40px;
  flex-shrink: 0;
  grid-column: auto;
}

.menu-item:hover {
  background-color: rgba(45, 134, 89, 0.1);
  color: #2d8659;
}

.menu-item.active {
  background-color: rgba(45, 134, 89, 0.12);
  color: #2d8659;
  font-weight: 500;
}

.menu-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 4px;
  height: 20px;
  background-color: #2d8659;
  border-radius: 0 2px 2px 0;
}

.sidebar-collapsed .menu-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  height: 100%;
  background-color: #2d8659;
  border-radius: 0 3px 3px 0;
  transform: none;
}

.menu-icon {
  font-size: 16px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: font-size 0.3s ease;
}

.sidebar-collapsed .menu-icon {
  font-size: 18px;
}

.menu-label {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: opacity 0.3s ease;
}

.sidebar-collapsed .menu-label {
  display: none;
}

.menu-tooltip {
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  background-color: rgba(0, 0, 0, 0.85);
  color: white;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s ease;
  margin-left: 12px;
  z-index: 1000;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.sidebar-collapsed .menu-item:hover .menu-tooltip,
.sidebar-collapsed .group-header:hover::after {
  opacity: 1;
}

/* ============ Sidebar Footer ============ */
.sidebar-footer {
  padding: 12px 8px;
  border-top: 1px solid rgba(45, 134, 89, 0.15);
  overflow-x: hidden;
  transition: all 0.3s ease;
}

.sidebar-collapsed .sidebar-footer {
  padding: 12px;
}

.footer-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #4a4a4a;
  position: relative;
}

.sidebar-collapsed .footer-item {
  justify-content: center;
  padding: 10px 12px;
  gap: 0;
}

.footer-item:hover {
  background-color: rgba(45, 134, 89, 0.1);
  color: #2d8659;
}

.footer-icon {
  font-size: 18px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: font-size 0.3s ease;
}

.sidebar-collapsed .footer-icon {
  font-size: 20px;
}

.footer-label {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: opacity 0.3s ease;
}

.sidebar-collapsed .footer-label {
  display: none;
}

/* ============ Scrollbar ============ */
.sidebar::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 0;
}

.sidebar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* ============ Transitions ============ */
.group-expand-enter-active,
.group-expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.group-expand-enter-from,
.group-expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.group-expand-enter-to,
.group-expand-leave-from {
  opacity: 1;
  max-height: 500px;
}

/* ============ Responsive Design ============ */
@media (max-width: 600px) {
  .sidebar {
    display: none !important;
  }
}
</style>
