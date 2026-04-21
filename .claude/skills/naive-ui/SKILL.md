---
name: naive-ui
description: Naive UI - 企业级Vue3组件库，用于发布控制系统的前端UI开发
keywords: 组件库, Vue3, UI, 企业级, 主题定制, 国际化
---

# Naive UI 组件库使用指南

## 概述

Naive UI (v2.34+) 是一个为 Vue 3 设计的企业级UI组件库，提供丰富的组件和完整的主题定制能力。

## 常用组件

### 布局组件
- `<n-layout>` - 页面布局容器
- `<n-layout-sider>` - 侧边栏
- `<n-layout-content>` - 内容区域
- `<n-layout-header>` - 页眉

### 表单组件
- `<n-input>` - 文本输入框
- `<n-select>` - 下拉选择框
- `<n-form>` - 表单容器
- `<n-form-item>` - 表单项
- `<n-input-number>` - 数字输入
- `<n-checkbox>` - 复选框
- `<n-radio>` - 单选框
- `<n-switch>` - 开关

### 按钮组件
- `<n-button>` - 普通按钮
- `<n-button-group>` - 按钮组

### 数据展示
- `<n-data-table>` - 数据表格
- `<n-list>` - 列表
- `<n-card>` - 卡片
- `<n-statistic>` - 统计数据
- `<n-tag>` - 标签/徽章
- `<n-badge>` - 徽章
- `<n-progress>` - 进度条
- `<n-timeline>` - 时间线

### 反馈组件
- `<n-modal>` - 模态框/对话框
- `<n-popconfirm>` - 气泡确认框
- `<n-drawer>` - 抽屉
- `<n-message>` - 消息提示 (使用useMessage)
- `<n-notification>` - 通知 (使用useNotification)
- `<n-alert>` - 警告框
- `<n-result>` - 结果页

### 导航组件
- `<n-menu>` - 菜单
- `<n-breadcrumb>` - 面包屑
- `<n-tabs>` - 标签页

## 常见使用模式

### 1. 基础按钮
```vue
<n-button type="primary" @click="handleClick">
  创建发布
</n-button>
```

### 2. 表单
```vue
<n-form :model="form" :rules="rules">
  <n-form-item label="应用" path="app_id">
    <n-select v-model:value="form.app_id" :options="appOptions" />
  </n-form-item>
</n-form>
```

### 3. 数据表格
```vue
<n-data-table :columns="columns" :data="releases" />
```

### 4. 模态框
```vue
<n-modal v-model:show="showModal" title="创建发布">
  <!-- 内容 -->
</n-modal>
```

### 5. 消息提示
```typescript
import { useMessage } from 'naive-ui'
const message = useMessage()
message.success('操作成功')
message.error('操作失败')
```

## 主题定制

在 `src/theme.ts` 中定义主题配置：

```typescript
import { darkTheme, zhCN } from 'naive-ui'

export const naiveTheme = {
  theme: darkTheme,
  locale: zhCN,
  dateLocale: zhCN,
}
```

## 在应用中集成

```typescript
// main.ts
import naive from 'naive-ui'
app.use(naive)
```

## 注意事项

- Naive UI 组件使用 `v-model:value` 绑定，不是 `v-model`
- 颜色使用 RGB 格式
- 所有 Naive UI 插槽都使用作用域插槽
- 响应式设计需要使用 CSS Grid 或 Flexbox