<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <div class="app-wrapper">
      <!-- Sidebar -->
      <Sidebar />

      <!-- Main Content -->
      <div class="layout">
        <!-- Header -->
        <div class="layout-header">
          <h1>{{ pageTitle }}</h1>
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
    </div>
  </n-config-provider>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUiStore } from '@/stores/uiStore'
import { useAppStore } from '@/stores/appStore'
import Sidebar from '@/components/Sidebar.vue'
import { themeOverrides } from '@/theme'

const route = useRoute()
const uiStore = useUiStore()
const appStore = useAppStore()

const pageTitle = computed(() => {
  return route.meta.title ? `${route.meta.title}` : '发布控制系统'
})

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
.app-wrapper {
  display: flex;
  height: 100vh;
  background-color: #f5f5f5;
}

.layout {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.layout-header {
  height: 60px;
  background: #f0f0f0;
  color: #2d8659;
  display: flex;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  flex-shrink: 0;
  border-bottom: 1px solid rgba(45, 134, 89, 0.15);
}

.layout-header h1 {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.layout-content {
  flex: 1;
  overflow: auto;
  background-color: #e5e5e5;
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
  border-radius: 0;
  font-size: 14px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  animation: slideIn 0.3s ease-out;
}

.message-success {
  background-color: #2d8659;
  color: #ffffff;
}

.message-error {
  background-color: #e63946;
  color: #ffffff;
}

.message-warning {
  background-color: #f77f00;
  color: #ffffff;
}

.message-info {
  background-color: #2d8659;
  color: #ffffff;
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
  .app-wrapper {
    flex-direction: column;
  }

  .layout-header {
    padding: 0 16px;
  }

  .layout-content {
    padding: 16px;
  }

  .message-container {
    right: 12px;
  }
}
</style>
