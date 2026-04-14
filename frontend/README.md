# 发布控制系统 - 前端

Vue3 + TypeScript + Naive UI 开发的现代化发布控制系统前端。

## 项目特性

- ✅ **Vite** - 极速开发和构建
- ✅ **Vue3 + TypeScript** - 类型安全和最新语法
- ✅ **Naive UI** - 企业级UI组件库
- ✅ **Pinia** - 简洁高效的状态管理
- ✅ **Axios** - 带拦截器的HTTP客户端
- ✅ **Vue Router** - 完整的路由系统
- ✅ **实时轮询** - 发布进度实时更新

## 快速开始

### 安装依赖

```bash
cd frontend
npm install
```

### 开发

```bash
npm run dev
```

访问 http://localhost:5173

### 生产构建

```bash
npm run build
```

### 代码检查和格式化

```bash
npm run lint
npm run format
```

## 项目结构

```
frontend/
├── src/
│   ├── api/                 # HTTP API 服务层
│   │   ├── request.ts      # Axios 实例和拦截器
│   │   ├── release.ts      # 发布 API
│   │   └── metadata.ts     # 元数据 API
│   ├── stores/             # Pinia 状态管理
│   │   ├── appStore.ts     # 应用/环境/集群数据
│   │   ├── releaseStore.ts # 发布流程状态
│   │   └── uiStore.ts      # 全局 UI 状态
│   ├── types/              # TypeScript 类型定义
│   │   ├── api.ts          # API 返回类型
│   │   └── models.ts       # 业务模型类型
│   ├── views/              # 页面组件
│   │   ├── ReleaseFlow.vue     # 发布向导（P0）
│   │   ├── ReleaseHistory.vue  # 发布历史（P0）
│   │   └── ReleaseDetail.vue   # 发布详情（P1）
│   ├── router/             # 路由配置
│   ├── styles/             # 全局样式
│   ├── utils/              # 工具函数
│   ├── App.vue             # 根组件
│   └── main.ts             # 应用入口
├── index.html              # HTML 模板
├── vite.config.ts          # Vite 配置
├── tsconfig.json           # TypeScript 配置
├── package.json            # 项目依赖
└── README.md               # 本文件
```

## 核心功能模块

### 1. 发布向导 (ReleaseFlow) - P0

四步向导式发布流程：

1. **选择应用** - 从元数据列表中选择
2. **选择环境** - 选择 dev/staging/production 等
3. **选择集群** - 根据应用和环境动态加载可用集群
4. **选择镜像** - 输入镜像标签（格式: registry/app:vX.X.X）
5. **确认发布** - 显示发布详情，点击发布

**特性**:
- ✅ 表单验证
- ✅ 实时进度显示
- ✅ 事件日志流
- ✅ 发布后可查看详情

### 2. 发布历史 (ReleaseHistory) - P0

完整的发布记录管理：

- 表格展示所有发布记录
- 按应用/环境/状态筛选
- 支持分页查看
- 每条记录显示应用、环境、镜像、状态、发起人、时间等
- 点击"详情"查看完整事件日志
- 成功的发布可快速回滚

**特性**:
- ✅ 动态筛选和搜索
- ✅ 分页加载
- ✅ 状态标签彩色展示
- ✅ 回滚确认对话框

### 3. 发布详情 (ReleaseDetail) - P1

发布全流程详细信息：

- 基本信息（应用、环境、集群、镜像等）
- 实时进度条
- 完整的事件日志
- 错误详情展示
- 日志导出功能
- 回滚/重试操作

**特性**:
- ✅ 实时轮询更新
- ✅ 事件日志详情
- ✅ 错误日志展示
- ✅ 日志导出下载

## 状态管理架构

### appStore

管理应用元数据（读取后端接口）：

```typescript
- applications: Application[]
- environments: Environment[]
- clusters: Cluster[]
- deploymentTargets: DeploymentTarget[]

Actions:
- fetchApplications()
- fetchEnvironments()
- fetchClusters()
- fetchDeploymentTargets()
- initializeMetadata() // 初始化所有元数据
```

### releaseStore

管理发布流程和历史：

```typescript
- currentRelease: ReleaseResponse | null
- releaseHistory: ReleaseResponse[]
- releaseEvents: ReleaseEvent[]
- isPolling: boolean

Actions:
- createRelease() // 发起发布
- fetchReleaseStatus() // 查询状态
- fetchReleaseEvents() // 获取事件
- fetchReleaseHistory() // 获取历史
- startPolling() // 启动轮询
- stopPolling() // 停止轮询
- rollback() // 回滚发布
```

### uiStore

全局 UI 状态：

```typescript
- messages: Message[]
- loading: boolean

Actions:
- showMessage()
- success() / error() / warning() / info()
- setLoading()
```

## API 集成

### 后端 API 地址

默认连接 `http://localhost:8080` (可通过 `.env` 配置)

### 核心端点

```
POST   /api/v1/releases
GET    /api/v1/releases
GET    /api/v1/releases/{id}
GET    /api/v1/releases/{id}/events
POST   /api/v1/releases/{id}/rollback

GET    /api/v1/applications
GET    /api/v1/environments
GET    /api/v1/clusters
GET    /api/v1/deployment-targets
GET    /api/v1/deployment-targets/app/{appId}/env/{envId}
```

### 请求/响应拦截

所有请求自动添加 `X-Request-ID` 头用于链路追踪。

错误自动转换为 Error 对象，UI 层统一处理。

## 配置

### 环境变量

复制 `.env.example` 为 `.env` 并配置：

```bash
# Backend API
VITE_API_BASE_URL=http://localhost:8080

# App Settings
VITE_APP_NAME=Release Control System
VITE_APP_VERSION=0.1.0

# UI Settings
VITE_POLLING_INTERVAL=2000  # 轮询间隔（毫秒）
VITE_MAX_LOG_ITEMS=50       # 最多显示事件数
```

### Vite 配置

```typescript
// vite.config.ts
- 开发服务器端口: 5173
- API 代理: /api -> http://localhost:8080
- 生产构建输出: dist/
```

## 开发指南

### 添加新页面

1. 在 `src/views/` 创建新的 `.vue` 文件
2. 在 `src/router/index.ts` 注册路由
3. 在 `App.vue` 添加导航链接

### 添加新的 API 调用

1. 在 `src/api/` 中创建新的 service 文件
2. 定义 TypeScript 类型
3. 通过 Pinia store 封装调用

### 添加新的状态

1. 在 `src/stores/` 创建新的 store
2. 使用 Pinia 的 `defineStore`
3. 在组件中通过 `useXxxStore()` 使用

### 添加全局样式

编辑 `src/styles/main.css`，会自动应用到所有组件。

### 类型定义

所有 API 类型在 `src/types/api.ts`，业务模型在 `src/types/models.ts`。

## 常用命令

```bash
# 开发
npm run dev

# 构建
npm run build

# 预览构建结果
npm run preview

# 代码检查和格式化
npm run lint
npm run format
```

## 浏览器支持

- Chrome/Edge 最新版
- Firefox 最新版
- Safari 最新版

## 性能优化

- ✅ 代码分割（按路由）
- ✅ Lazy Loading（路由和组件）
- ✅ CSS 压缩和 Tree Shaking
- ✅ 虚拟滚动（事件日志）
- ✅ 轮询去重（2秒一次）

## 常见问题

**Q: 如何连接到不同的后端地址?**
A: 编辑 `.env` 文件，修改 `VITE_API_BASE_URL`。

**Q: 为什么发布进度不更新?**
A: 检查浏览器控制台是否有网络错误，确认后端服务已启动。

**Q: 如何修改 UI 主题?**
A: Naive UI 主题配置在 App.vue，可通过 `<n-config-provider>` 自定义。

**Q: 事件日志如何导出?**
A: 在发布详情页点击"导出日志"按钮，会下载 `release-{id}-logs.txt` 文件。

## 依赖说明

- **vue@3** - 核心框架
- **vue-router@4** - 路由管理
- **pinia@2** - 状态管理
- **naive-ui@2** - UI 组件库
- **axios@1** - HTTP 客户端
- **dayjs@1** - 日期处理
- **vite@5** - 构建工具
- **typescript@5** - 类型检查

## 下一步优化

### 短期
- [ ] 添加登录和权限管理
- [ ] 支持导入/导出发布配置
- [ ] 实现灰度发布预览

### 中期
- [ ] WebSocket 实时推送（替代轮询）
- [ ] 发布日志搜索和过滤
- [ ] 发布对比（版本 A vs 版本 B）
- [ ] 通知集成（Slack/钉钉）

### 长期
- [ ] 多语言支持
- [ ] 暗黑主题
- [ ] 移动端适配
- [ ] 离线支持

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue 或 PR。
