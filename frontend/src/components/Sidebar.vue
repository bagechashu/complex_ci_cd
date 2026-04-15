<template>
  <aside class="sidebar" :class="{ 'sidebar-collapsed': isCollapsed }">
    <!-- Logo Section -->
    <div class="sidebar-header">
      <div class="logo-wrapper">
        <div class="logo-icon">📦</div>
        <span v-if="!isCollapsed" class="logo-text">发布系统</span>
      </div>
      <button class="collapse-btn" @click="toggleCollapse" :title="isCollapsed ? '展开' : '折叠'">
        <span class="collapse-icon">{{ isCollapsed ? '→' : '←' }}</span>
      </button>
    </div>

    <!-- Menu Items -->
    <nav class="sidebar-menu">
      <div
        v-for="item in menuItems"
        :key="item.name"
        class="menu-item"
        :class="{ active: isActive(item.name) }"
        @click="navigateTo(item.path)"
      >
        <div class="menu-icon">{{ item.icon }}</div>
        <span v-if="!isCollapsed" class="menu-label">{{ item.label }}</span>
        <div v-if="!isCollapsed" class="menu-tooltip">{{ item.label }}</div>
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

const router = useRouter()
const route = useRoute()
const isCollapsed = ref(false)
const showUserMenu = ref(false)

// Menu items configuration
const menuItems: MenuItem[] = [
  {
    name: 'ReleaseFlow',
    path: '/release',
    label: '发布向导',
    icon: '🚀',
    meta: { title: '发布向导' }
  },
  {
    name: 'ReleaseHistory',
    path: '/history',
    label: '发布历史',
    icon: '📋',
    meta: { title: '发布历史' }
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
  width: 80px;
}

/* ============ Sidebar Header ============ */
.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid rgba(45, 134, 89, 0.15);
}

.logo-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  flex: 1;
  overflow: hidden;
}

.logo-icon {
  font-size: 24px;
  flex-shrink: 0;
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
  padding: 4px 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background-color 0.2s;
  font-size: 16px;
}

.collapse-btn:hover {
  background-color: rgba(45, 134, 89, 0.1);
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

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  color: #4a4a4a;
  user-select: none;
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
  height: 24px;
  background-color: #2d8659;
  border-radius: 0 2px 2px 0;
}

.menu-icon {
  font-size: 18px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.menu-label {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.menu-tooltip {
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  background-color: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s ease;
  margin-left: 8px;
  z-index: 10;
}

.sidebar-collapsed .menu-item:hover .menu-tooltip {
  opacity: 1;
}

/* ============ Sidebar Footer ============ */
.sidebar-footer {
  padding: 16px 8px;
  border-top: 1px solid rgba(45, 134, 89, 0.15);
  overflow-x: hidden;
}

.footer-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #4a4a4a;
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
}

.footer-label {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
