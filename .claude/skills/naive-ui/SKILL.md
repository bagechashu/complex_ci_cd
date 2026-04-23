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

#### 基础表格
```vue
<n-data-table :columns="columns" :data="releases" />
```

#### 带水平滚动的表格（列数多时）
Naive UI data-table 默认会自动计算宽度，当列太多时需要处理水平滚动。方式如下：

**方案1：使用 scroll-x 属性（推荐）**
```vue
<n-data-table 
  :columns="columns" 
  :data="releases"
  :scroll-x="1200"
/>
```
- `scroll-x` 属性指定表格内容的最小宽度，当内容宽度超过容器时会出现水平滚动条
- 值为数字（像素）或 'auto'

**方案2：CSS 容器滚动**
```vue
<div style="overflow-x: auto">
  <n-data-table 
    :columns="columns" 
    :data="releases"
    class="data-table"
  />
</div>
```
```css
.data-table {
  min-width: max-content;
}
```

#### 表格常用属性
```vue
<n-data-table
  :columns="columns"
  :data="tableData"
  :bordered="false"
  :single-line="false"
  :striped="true"
  :loading="loading"
  :row-key="(row) => row.id"
  v-model:checked-row-keys="selectedRowKeys"
  :scroll-x="1200"
  :flex-height="true"
/>
```

**关键属性说明：**
- `columns` - 列配置数组，每列需要 `key`, `title`, `render` (可选)
- `data` - 表格数据数组
- `bordered` - 是否显示边框
- `single-line` - 是否单行模式（不换行）
- `striped` - 是否显示斑马纹
- `loading` - 加载状态
- `row-key` - 行唯一标识，可以是字符串属性名或返回值的函数
- `v-model:checked-row-keys` - 勾选的行键数组（用于选择功能）
- `scroll-x` - 水平滚动阈值（当表格宽度超过此值时显示滚动条）
- `flex-height` - 是否自动计算高度以填充容器
- `pagination` - 分页配置对象
- `on-update:page` - 页码变化回调
- `on-update:page-size` - 每页大小变化回调

#### 列配置示例
```typescript
const columns = [
  {
    type: 'selection', // 选择列
    fixed: 'left',     // 固定在左侧
  },
  {
    key: 'id',
    title: '任务 ID',
    width: 80,
    fixed: 'left',  // 固定列（不会被滚动隐藏）
  },
  {
    key: 'name',
    title: '任务名称',
    width: 200,
  },
  {
    key: 'status',
    title: '状态',
    render: (row) => h(NTag, { type: row.status === 'success' ? 'success' : 'warning' }, () => row.status)
  },
  {
    key: 'actions',
    title: '操作',
    width: 150,
    fixed: 'right',  // 固定在右侧（"操作"列常用这个设置）
    render: (row) => h('div', [
      h(NButton, { text: true, size: 'small' }, () => '编辑'),
      h(NButton, { text: true, size: 'small' }, () => '删除'),
    ])
  }
]
```

**列配置关键字段：**
- `type` - 特殊列类型：'selection' (复选框)、'expand' (展开)
- `key` - 对应数据的键名
- `title` - 列标题
- `width` - 列宽（像素或百分比）
- `fixed` - 固定位置：'left' 或 'right'（用于锁定重要列）
- `render` - 自定义渲染函数 `(row, rowIndex) => VNode`
- `align` - 对齐方式：'left'、'center'、'right'
- `ellipsis` - 文本溢出时是否显示省略号

#### 水平滚动最佳实践

当表格列数多或列宽较大时，需要启用水平滚动。关键点：

1. **计算总宽度**
   - 将所有列的宽度相加，得到表格总宽度
   - 例如：8列 × 平均150px = 1200px

2. **设置 scroll-x 阈值**
   ```vue
   <n-data-table :scroll-x="1200" />
   ```
   - 当表格宽度超过此值时，会显示水平滚动条
   - 推荐值为所有列总宽度或容器宽度，选择较小的那个

3. **固定重要列**
   ```typescript
   // 固定左侧的选择列和ID列
   { type: 'selection', fixed: 'left', width: 40 }
   { key: 'id', title: 'ID', width: 80, fixed: 'left' }
   
   // 固定右侧的操作列
   { key: 'actions', title: '操作', width: 150, fixed: 'right' }
   ```
   - `fixed: 'left'` - 列会固定在左侧，用户滚动时不会隐藏
   - `fixed: 'right'` - 列会固定在右侧，通常用于"操作"列

4. **为所有列设置宽度**
   - 必须为列设置固定宽度（width）
   - 不设置宽度的列在滚动时可能出现排版问题
   - 重要列可用 `ellipsis: true` 处理文本溢出

5. **示例配置**
   ```typescript
   const columns = [
     { type: 'selection', width: 40, fixed: 'left' },        // 选择列 - 固定左侧
     { key: 'id', title: 'ID', width: 80, fixed: 'left' },   // ID列 - 固定左侧
     { key: 'name', title: '名称', width: 200, ellipsis: true },
     { key: 'desc', title: '描述', width: 250, ellipsis: true },
     { key: 'status', title: '状态', width: 120 },
     { key: 'date', title: '日期', width: 150 },
     { key: 'actions', title: '操作', width: 150, fixed: 'right' }  // 操作列 - 固定右侧
   ]
   // 总宽度 = 40 + 80 + 200 + 250 + 120 + 150 + 150 = 990px
   // 设置 scroll-x="990" 或根据容器宽度调整
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

### 通用注意事项
- Naive UI 组件使用 `v-model:value` 绑定，不是 `v-model`
- 颜色使用 RGB 格式
- 所有 Naive UI 插槽都使用作用域插槽
- 响应式设计需要使用 CSS Grid 或 Flexbox

### 数据表格常见问题

**问题1：表格列太多，超出容器宽度无法看到所有列**
- 解决：使用 `:scroll-x` 属性启用水平滚动
- 确保每列都有 `width` 定义
- 用 `fixed: 'left'` 或 `fixed: 'right'` 锁定重要列

**问题2：行高不一致**
- 解决：设置 `:single-line="false"` 允许文本换行
- 或使用 `ellipsis: true` 截断长文本

**问题3：选择行功能不工作**
- 检查：是否为列设置了 `type: 'selection'`
- 检查：是否使用了 `v-model:checked-row-keys` 双向绑定
- 检查：`:row-key` 是否正确指向行的唯一标识

**问题4：固定列不显示**
- 检查：固定列是否有 `width` 属性
- 检查：是否使用了 `:scroll-x` 属性（fixed列需要scroll-x才能生效）
- 确保 `fixed: 'left'` 或 `fixed: 'right'` 设置正确

**问题5：表格排版错乱**
- 原因：列没有指定宽度
- 解决：为所有列添加 `width` 属性
- 动态内容列可用 `ellipsis: true` 处理溢出