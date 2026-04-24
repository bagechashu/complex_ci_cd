---
name: frontend-css-architecture
description: 发布控制系统 - 前端CSS架构、样式统一化、主题管理、响应式设计
keywords: CSS, 样式管理, 主题定制, CSS变量, 响应式, 样式统一化, Naive UI
---

# 前端 CSS 架构与样式管理指南

## 概述

本指南规范发布控制系统前端的 CSS 架构，包括：
- 样式统一化（避免重复代码）
- CSS 变量主题管理
- 响应式设计规范
- 样式组织结构
- 常见组件样式模式

---

## CSS 架构概览

### 样式文件组织

```
frontend/src/styles/
├─ main.css              # 全局样式入口（导入其他所有样式）
├─ views.css             # 所有视图页面的共用样式（中央样式库）
├─ theme.ts              # 主题配置（CSS变量）
└─ [页面特定样式]        # 各Vue文件中的scoped样式
```

### CSS 导入链路

```
main.css
  ├─ @import views.css (所有页面共用样式 400+ 行)
  ├─ @import 全局重置
  └─ @import 通用工具类

views.css (中央样式库)
  ├─ CSS Variables (颜色、间距、尺寸)
  ├─ 页面布局 (.page-container, .content-layout)
  ├─ 列表样式 (.list-panel, .list-item, .list-header)
  ├─ 表单样式 (.form-group, .form-input, .form-row)
  ├─ 详情页样式 (.detail-panel, .detail-section, .info-grid)
  ├─ 分页样式 (.pagination-controls)
  ├─ 徽章样式 (.badge, .status-badge)
  ├─ 模态框样式 (.modal-overlay, .modal)
  └─ 按钮样式 (.header-actions)

Individual .vue files (scoped样式)
  └─ 仅包含页面特定的自定义样式
```

---

## CSS 变量系统（来自 theme.ts）

### 颜色变量

```css
:root {
  /* 主题色 */
  --color-primary: #2d8659;              /* 森林绿 - 主操作按钮 */
  
  /* 文字颜色 */
  --color-text-primary: #1a1a1a;         /* 主文本 */
  --color-text-secondary: #4a4a4a;       /* 次级文本 */
  --color-text-muted: #999999;           /* 弱化文本 */
  
  /* 背景颜色 */
  --color-bg-card: #ffffff;              /* 卡片背景 */
  --color-bg-dark: #f5f5f5;              /* 深灰背景 */
  --color-bg-light: #e5e5e5;             /* 浅灰背景 */
  
  /* 边框颜色 */
  --color-border: #eeeeee;               /* 标准边框 */
  --color-border-light: #f0f0f0;         /* 浅色边框 */
  
  /* 状态颜色 */
  --color-success: #2d8659;              /* 成功 - 同主色 */
  --color-error: #e63946;                /* 错误 - 红色 */
  --color-warning: #f77f00;              /* 警告 - 橙色 */
  --color-info: #0066cc;                 /* 信息 - 蓝色 */
  
  /* 列表活跃背景 */
  --color-bg-list-active: #e8f5f0;       /* 列表项活跃背景 */
}
```

### 间距与尺寸变量

```css
:root {
  /* 间距 (gap/padding/margin) */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 12px;
  --spacing-lg: 16px;
  --spacing-xl: 20px;
  --spacing-xxl: 24px;
  --spacing-3xl: 30px;
  
  /* 字体 */
  --font-size-xs: 12px;
  --font-size-sm: 13px;
  --font-size-base: 14px;
  --font-size-lg: 16px;
  --font-size-xl: 18px;
  
  /* 圆角 */
  --border-radius: 0px;                 /* 发布控制系统采用直角风格 */
  
  /* 阴影 */
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 2px 8px rgba(0,0,0,0.1);
}
```

---

## views.css 统一样式库（重点内容）

### 1. 页面布局样式

#### `.page-container` - 页面根容器

```css
.page-container {
  padding: 24px;
  width: 100%;
  height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-card);
}
```

**用法**: 所有页面的根div必须使用此class

#### `.content-layout` - 列表+详情布局

```css
.content-layout {
  display: flex;
  gap: 20px;
  flex: 1;
  overflow: hidden;
}
```

**用法**: 列表-详情两列布局的容器

#### `.list-panel` - 左侧列表面板

```css
.list-panel {
  width: 300px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--color-border);
  overflow: hidden;
}
```

**用法**: 左侧固定宽度列表

#### `.detail-panel` - 右侧详情面板

```css
.detail-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
```

**用法**: 右侧可伸缩的详情区域

---

### 2. 列表样式

#### `.list-header` - 列表头部

```css
.list-header {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  border-bottom: 1px solid var(--color-border);
}

.list-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
}
```

**用法**: 列表顶部区域（标题、搜索、排序）

#### `.list-item` - 列表项

```css
.list-item {
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.list-item:hover {
  background-color: var(--color-bg-light);
}

.list-item.active {
  background-color: var(--color-bg-list-active);
  border-left: 3px solid var(--color-primary);
  padding-left: 13px;  /* 调整padding补偿border占用的空间 */
}
```

**用法**: 列表中的每一项

**规范**:
- 活跃项必须添加 `.active` class
- 要改变活跃项样式时，只改 views.css 中的 `.list-item.active`，不要在页面中重复定义

---

### 3. 表单样式

#### `.form-group` - 表单组

```css
.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.form-group label {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 12px;
}
```

#### `.form-input`, `.form-select`, `.form-textarea` - 表单元素

```css
.form-input,
.form-select,
.form-textarea {
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  background: var(--color-bg-card);
  color: var(--color-text-primary);
  font-size: 14px;
  border-radius: 0;
  transition: border-color 0.2s ease;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: inset 0 0 0 1px var(--color-primary);
}
```

#### `.form-row` - 表单两列布局

```css
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

@media (max-width: 600px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}
```

---

### 4. 详情页样式

#### `.detail-content` - 详情内容区

```css
.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
```

#### `.detail-section` - 详情区块

```css
.detail-section {
  padding: 16px;
  margin-bottom: 30px;
  background: var(--color-bg-dark);
  border: 1px solid var(--color-border-light);
  border-radius: 0;
}

.detail-section:last-child {
  margin-bottom: 0;
}
```

**规范**: 
- 所有详情内容都应该用 `.detail-section` 包裹
- 不应在页面中定义 `.detail-section` 的样式，使用 views.css 中的统一样式
- 如果需要自定义背景色或边框，应该在页面中定义新的 class，而不是覆盖 `.detail-section`

#### `.section-header` - 区块标题（标题 + 操作按钮）

```css
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 0;
  margin: 0 0 16px 0;
  border: none;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
  flex: 0 1 auto;
}
```

**用法示例**:

```vue
<div class="section-header">
  <h3>集群信息</h3>
  <div class="header-actions">
    <button class="n-button">编辑集群</button>
  </div>
</div>
```

**规范**:
- h3 自动左对齐（flex: 0 1 auto）
- 操作按钮自动右对齐（justify-content: space-between）
- h3 和按钮在同一行

#### `.header-actions` - 区块操作按钮组

```css
.header-actions {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}
```

#### `.info-grid` - 信息网格（2列展示）

```css
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

@media (max-width: 600px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}
```

#### `.info-item` - 信息项

```css
.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item label {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 12px;
}

.info-item span {
  color: var(--color-text-primary);
  font-size: 14px;
  word-break: break-all;
}
```

---

### 5. 分页样式

#### `.pagination-controls` - 分页容器

```css
.pagination-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 12px 0;
  margin-top: auto;
  border-top: 1px solid var(--color-border);
}
```

#### `.pagination-btn` - 分页按钮

```css
.pagination-btn {
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  background: var(--color-bg-card);
  color: var(--color-text-primary);
  cursor: pointer;
  font-size: 12px;
  border-radius: 0;
  transition: all 0.2s ease;
}

.pagination-btn:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination-btn.active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}
```

---

### 6. 徽章与状态样式

#### `.badge` - 通用徽章

```css
.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px 12px;
  background: var(--color-bg-light);
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 600;
  border-radius: 0;
  white-space: nowrap;
}
```

#### `.status-badge` - 状态徽章

```css
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 0;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.connected {
  background: rgba(45, 134, 89, 0.12);
  color: var(--color-primary);
}

.status-badge.disconnected {
  background: rgba(230, 57, 70, 0.12);
  color: var(--color-error);
}

.status-badge.unknown {
  background: rgba(249, 165, 0, 0.12);
  color: var(--color-warning);
}
```

---

### 7. 模态框样式

#### `.modal-overlay` - 模态框背景

```css
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
```

#### `.modal` - 模态框

```css
.modal {
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: 0;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  max-width: 90vw;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
```

---

## 响应式设计规范

### 断点

```css
/* 桌面优先设计 (Desktop-first) */

/* 标准桌面: 1200px 及以上 */
@media (min-width: 1200px) {
  /* ... */
}

/* 平板: 768px - 1199px */
@media (max-width: 1199px) {
  /* ... */
}

/* 手机: 600px - 767px */
@media (max-width: 767px) {
  /* ... */
}

/* 小屏手机: 小于 600px */
@media (max-width: 599px) {
  /* ... */
}
```

### 推荐响应式调整

```css
/* 两列网格 → 单列 */
.info-grid {
  grid-template-columns: 1fr 1fr;
}

@media (max-width: 600px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}

/* 两列布局 → 堆叠 */
.content-layout {
  display: flex;
  gap: 20px;
}

@media (max-width: 600px) {
  .content-layout {
    flex-direction: column;
  }
  
  .list-panel {
    width: 100%;
    max-height: 300px;
  }
}
```

---

## Vue 组件中的样式使用规范

### ✅ DO - 正确做法

```vue
<template>
  <div class="page-container">
    <div class="content-layout">
      <div class="list-panel">
        <div class="list-header">
          <h2>集群列表</h2>
        </div>
        <div class="list-container">
          <div 
            v-for="cluster in clusters"
            :key="cluster.id"
            class="list-item"
            :class="{ active: cluster.id === selectedId }"
            @click="selectCluster(cluster)"
          >
            {{ cluster.name }}
          </div>
        </div>
      </div>
      
      <div class="detail-panel" v-if="selectedCluster">
        <div class="detail-content">
          <div class="detail-section">
            <div class="section-header">
              <h3>集群信息</h3>
              <div class="header-actions">
                <button class="n-button">编辑</button>
              </div>
            </div>
            <div class="info-grid">
              <div class="info-item">
                <label>名称</label>
                <span>{{ selectedCluster.name }}</span>
              </div>
              <div class="info-item">
                <label>环境</label>
                <span>{{ selectedCluster.env }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 仅定义页面特定的自定义样式 */

/* 示例1: 特殊的颜色 */
.status-connected {
  color: #2d8659;
  font-weight: 600;
}

/* 示例2: 页面特定的布局微调 */
.cluster-card {
  background: linear-gradient(135deg, #2d8659 0%, #1a5c3a 100%);
  padding: 20px;
  border-radius: 4px;
}

/* 示例3: 动画效果 */
.fade-enter-active {
  transition: opacity 0.3s ease;
}
</style>
```

### ❌ DON'T - 错误做法

```vue
<template>
  <div class="page-container">
    <!-- ❌ 不要使用内联样式 -->
    <div style="display: flex; gap: 20px; flex: 1;">
      
      <!-- ❌ 不要定义已在 views.css 中的样式 -->
      <div class="list-panel" style="width: 300px;">
      
      <!-- ❌ 不要重复定义 .section-header -->
      <div class="section-header" style="justify-content: space-between;">
    </div>
  </div>
</template>

<style scoped>
/* ❌ 不要重复定义通用样式 */
.list-item {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  cursor: pointer;
}

.list-item.active {
  background-color: #e8f5f0;
  border-left: 3px solid #2d8659;
}

/* ❌ 不要创建与 views.css 中同名的 class */
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
}
</style>
```

---

## 样式统一化检查清单

在创建或修改页面时，使用此清单：

### 页面初始化

- [ ] 使用 `.page-container` 作为根元素
- [ ] 两列布局时使用 `.content-layout`
- [ ] 列表使用 `.list-panel`, `.list-header`, `.list-item`
- [ ] 详情使用 `.detail-panel`, `.detail-content`, `.detail-section`

### 样式定义

- [ ] 仅在页面 scoped 中定义页面特定的样式
- [ ] 不重复定义 views.css 中已有的样式（`.list-item`, `.section-header` 等）
- [ ] 所有颜色使用 CSS 变量（`var(--color-primary)` 等）
- [ ] 所有间距使用预定义的间距变量

### 交互状态

- [ ] 列表项选中时自动添加 `.active` class
- [ ] 使用 `.status-badge` 显示连接状态
- [ ] 使用 `.header-actions` 包裹区块操作按钮

### 响应式设计

- [ ] `.info-grid` 在小屏上切换为单列
- [ ] `.content-layout` 在手机上切换为竖向堆叠
- [ ] 表单在小屏上适当调整

---

## 常见问题

### Q: 我需要为某个列表项添加特殊样式，应该在哪里定义？

**A**: 在页面的 `<style scoped>` 中定义新的 class 名称，不要覆盖 `.list-item`。例如：

```vue
<style scoped>
.my-special-item {
  background: linear-gradient(...);
}
</style>

<template>
  <div class="list-item my-special-item">
    ...
  </div>
</template>
```

### Q: 我想改变活跃列表项的样式，应该怎么做？

**A**: 如果是系统全局变化，更新 views.css 中的 `.list-item.active`。如果只是某个页面需要不同样式，在页面中定义新的 class：

```vue
<style scoped>
.my-list-item.active {
  background: custom-color;
  border-left: 5px solid custom-color;
}
</style>
```

### Q: CSS 变量如何使用？

**A**: 直接在样式中使用 `var()` 函数：

```css
.my-element {
  color: var(--color-primary);           /* #2d8659 */
  background: var(--color-bg-dark);      /* #f5f5f5 */
  padding: var(--spacing-lg);            /* 16px */
  margin-bottom: var(--spacing-3xl);     /* 30px */
}
```

---

## 总结

| 原则 | 说明 |
|------|------|
| **中央样式库** | 所有通用样式在 views.css 中定义，避免重复 |
| **主题一致性** | 使用 CSS 变量管理颜色和间距，便于切换主题 |
| **单一责任** | 页面 scoped 样式仅定义该页面特定的内容 |
| **响应式优先** | 设计时考虑 600px 断点的响应式调整 |
| **命名规范** | 使用 BEM 或其他约定，保持命名一致性 |
| **避免嵌套** | CSS 中避免过深的嵌套（最多 2-3 层） |

