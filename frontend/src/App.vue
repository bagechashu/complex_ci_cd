<template>
  <div class="layout">
    <!-- Header -->
    <div class="layout-header">
      <h1>📦 发布控制系统</h1>
      <div class="layout-nav">
        <router-link
          to="/release"
          :class="{ active: $route.name === 'ReleaseFlow' }"
        >
          发布向导
        </router-link>
        <router-link
          to="/history"
          :class="{ active: $route.name === 'ReleaseHistory' }"
        >
          发布历史
        </router-link>
      </div>
    </div>

    <!-- Content Area -->
    <div class="layout-content">
      <div class="container">
        <!-- Messages Notification -->
        <div v-if="uiStore.messages.length > 0" class="message-container">
          <div
            v-for="msg in uiStore.messages"
            :key="msg.id"
            :class="['message', `message-${msg.type}`]"
          >
            {{ msg.content }}
          </div>
        </div>

        <!-- Main Router View -->
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUiStore } from '@/stores/uiStore'
import { useAppStore } from '@/stores/appStore'

const uiStore = useUiStore()
const appStore = useAppStore()

// 初始化应用数据
onMounted(async () => {
  try {
    await appStore.initializeMetadata()
  } catch (error) {
    uiStore.error('初始化应用数据失败')
    console.error('Failed to initialize metadata:', error)
  }
})
</script>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #f5f7fa;
}

.layout-header {
  height: 60px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.layout-header h1 {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.layout-nav {
  display: flex;
  gap: 20px;
  margin-left: auto;
}

.layout-nav a {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  font-size: 14px;
  transition: color 0.3s;
  padding: 4px 12px;
  border-bottom: 2px solid transparent;
}

.layout-nav a:hover {
  color: white;
}

.layout-nav a.active {
  color: white;
  border-bottom-color: white;
}

.layout-content {
  flex: 1;
  overflow: auto;
  padding: 24px;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
}

.message-container {
  position: fixed;
  top: 80px;
  right: 24px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message {
  padding: 12px 16px;
  border-radius: 4px;
  font-size: 14px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  animation: slideIn 0.3s ease-out;
}

.message-success {
  background-color: #55efc4;
  color: #00b894;
}

.message-error {
  background-color: #ff7675;
  color: #d63031;
}

.message-warning {
  background-color: #ffeaa7;
  color: #d97706;
}

.message-info {
  background-color: #74b9ff;
  color: #0984e3;
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .layout-header {
    padding: 0 16px;
  }

  .layout-content {
    padding: 16px;
  }

  .layout-nav {
    gap: 12px;
  }

  .message-container {
    right: 12px;
  }
}
</style>
