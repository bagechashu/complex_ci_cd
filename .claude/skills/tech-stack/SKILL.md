---
name: tech-stack
description: 发布控制系统 - 前后端完整技术栈指南
keywords: Go, Vue3, TypeScript, SQLite, K8s, Pinia, Axios, Naive UI
---

# 发布控制系统 - 完整技术栈

## 后端技术栈

### 核心框架与库

| 技术 | 版本 | 用途 | 关键特性 |
|------|------|------|---------|
| Go | 1.26+ | 编程语言 | 高性能、内存高效、并发强大 |
| go-chi | v5.2+ | Web框架 | 轻量级、路由、中间件、无依赖 |
| SQLite3 | 最新 | 数据库 | 本地部署、无需服务器、WAL mode |
| mattn/go-sqlite3 | v1.14+ | SQLite驱动 | 原生C扩展、高性能 |

### 功能库

| 库 | 用途 | 例子 |
|----|------|------|
| go-chi/cors | CORS中间件 | 跨域资源共享 |
| crypto/aes | 加密 | kubeconfig加密存储 |
| encoding/json | JSON序列化 | API请求/响应 |
| net/http | HTTP服务 | 基础HTTP功能 |
| os/signal | 信号处理 | 优雅关闭 |

### 测试框架

| 库 | 版本 | 用途 |
|----|------|------|
| testify | v1.11+ | 断言库、Mock工具、测试套件 |

### 项目结构

```
backend/
├── cmd/server/main.go           # 应用入口
├── internal/
│   ├── handlers/                # HTTP处理层
│   │   ├── api_handlers.go       # 应用/集群/发布处理
│   │   ├── shell_handlers.go     # Shell命令处理
│   │   ├── dtos.go               # 请求/响应DTO ⭐ 新增
│   │   ├── errors.go             # 统一错误处理 ⭐ 新增
│   │   └── router.go             # 路由定义
│   ├── services/                 # 业务服务层 ⭐ 核心升级
│   │   ├── release_service.go    # 发布生命周期管理
│   │   ├── application_service.go # 应用管理
│   │   ├── cluster_service.go     # 集群管理
│   │   ├── workload_service.go    # 部署目标管理
│   │   └── shell_service.go       # SSH命令执行
│   ├── models/                   # 数据模型structs
│   │   ├── application.go
│   │   ├── cluster.go
│   │   ├── workload_target.go
│   │   ├── release_record.go
│   │   ├── shell_server.go
│   │   ├── shell_command.go
│   │   └── ...
│   ├── repository/               # 数据访问层
│   │   ├── application_repo.go
│   │   ├── cluster_repo.go
│   │   ├── workload_target_repo.go
│   │   ├── release_record_repo.go
│   │   └── ...
│   ├── deployers/                # 部署策略
│   │   ├── deployer.go           # DeployStrategy接口
│   │   ├── k8s_deployer.go       # K8s实现
│   │   └── factory.go            # 工厂模式
│   ├── database/                 # 数据库初始化
│   │   ├── sqlite.go             # Schema定义(V1-V3)
│   │   ├── db.go                 # 初始化、连接
│   │   └── migration.go          # 初始数据
│   ├── config/                   # 配置管理
│   │   └── config.go
├── pkg/
│   ├── logger/                   # 结构化日志
│   ├── middleware/               # 中间件
│   │   └── validator.go          # 请求验证 ⭐ 新增
│   └── utils/                    # 工具函数
├── db/
│   └── init_data.sql             # 初始SQL数据
├── go.mod
├── go.sum
└── Makefile
```

### 数据库设计 (Schema V3)

**核心表**:
- `application` - 应用信息
- `environment` - 逻辑环境(dev/staging/prod)
- `cluster` - K8s集群(仅支持Kubernetes类型)
- `workload_target` - **应用→环境→集群映射(唯一键)**
- `release_record` - 发布记录及生命周期
- `release_event` - 发布事件日志

**Shell执行表** (SSH命令执行，用于Salt/Ansible等):
- `shell_server` - SSH服务器配置(密钥加密)
- `shell_command` - 允许执行的命令白名单(is_published安全标记)
- `shell_command_execution` - 命令执行记录
- `command_approval` - 命令审批流程

**版本管理表**:
- `schema_version` - 数据库版本追踪

### 关键特性

1. **多部署方式**: 支持K8s/Salt/Ansible策略模式
2. **敏感信息保护**: kubeconfig/密钥/密码都经过AES加密，使用json:"-"隐藏
3. **异步发布**: 使用Goroutines处理长时间部署，支持发布进度实时查询
4. **数据库并发控制**: SQLite WAL mode, PRAGMA优化
5. **完整事件日志**: 记录发布全生命周期事件供前端实时展示

---

## 前端技术栈

### 核心框架与库

| 技术 | 版本 | 用途 | 关键特性 |
|------|------|------|---------|
| Vue | 3.3+ | 前端框架 | Composition API、TypeScript支持、响应式 |
| TypeScript | 5.2+ | 编程语言 | 类型安全、IDE支持、编译检查 |
| Vite | 5.0+ | 构建工具 | 快速开发、HMR、最小化产物 |
| Pinia | 2.1+ | 状态管理 | 轻量级、类型安全、DevTools |
| Vue Router | 4.2+ | SPA路由 | 嵌套路由、懒加载、动态路由 |
| Axios | 1.6+ | HTTP客户端 | 拦截器、错误处理、超时控制 |
| Naive UI | 2.34+ | UI组件库 | 企业级、丰富组件、主题定制 |
| Day.js | 1.11+ | 日期处理 | 轻量级、格式化、时区支持 |

### 开发工具

| 工具 | 版本 | 用途 |
|------|------|------|
| ESLint | 8.53+ | 代码检查 |
| Prettier | 3.0+ | 代码格式化 |
| vue-tsc | 1.8+ | Vue类型检查 |
| @vitejs/plugin-vue | 5.0+ | Vite Vue插件 |

### 项目结构

```
frontend/
├── src/
│   ├── main.ts                   # 应用入口
│   ├── App.vue                   # 根组件
│   ├── theme.ts                  # 主题配置
│   ├── router/
│   │   └── index.ts              # 路由配置
│   ├── views/                    # 页面组件
│   │   ├── ReleaseFlow.vue       # 发布工作流(核心)
│   │   ├── ReleaseHistory.vue    # 发布历史
│   │   ├── ReleaseDetail.vue     # 发布详情
│   │   ├── KubernetesRelease.vue # K8s发布
│   │   ├── ClusterConfig.vue     # 集群配置
│   │   ├── ServerConfig.vue      # 服务器配置
│   │   ├── ShellExec.vue         # Shell执行
│   │   └── ExecutionHistory.vue  # 执行历史
│   ├── components/               # 可复用组件
│   │   └── Sidebar.vue           # 侧边栏
│   ├── stores/                   # Pinia状态管理
│   │   ├── appStore.ts           # 应用/环境/集群元数据
│   │   ├── releaseStore.ts       # 发布流程和历史
│   │   └── uiStore.ts            # 全局UI状态
│   ├── api/                      # API服务层
│   │   ├── request.ts            # Axios实例、拦截器
│   │   ├── release.ts            # 发布API
│   │   ├── metadata.ts           # 元数据API
│   │   ├── cluster-mapping.ts    # 集群映射API
│   │   ├── shell.ts              # Shell执行API
│   │   └── index.ts              # 统一导出
│   ├── types/                    # TypeScript类型
│   │   ├── api.ts                # API请求/响应类型
│   │   └── models.ts             # 业务模型类型
│   ├── utils/                    # 工具函数
│   │   ├── auth.ts               # 认证工具
│   │   ├── format.ts             # 格式化工具
│   │   └── ...
│   ├── styles/
│   │   └── main.css              # 全局样式
│   └── assets/
├── index.html
├── vite.config.ts                # Vite配置
├── tsconfig.json                 # TypeScript配置
├── tsconfig.node.json            # TS配置(Vite)
├── package.json
└── .eslintrc.cjs / prettier.config.js
```

### Pinia Store设计

```typescript
// appStore - 应用元数据
- applications: Application[]
- environments: Environment[]
- clusters: Cluster[]
- workloadTargets: WorkloadTarget[]
- allMappingsByApp: Map<appId, mappings[]> // 缓存优化

// releaseStore - 发布流程
- releases: ReleaseRecord[]
- currentRelease: ReleaseRecord | null
- releaseEvents: ReleaseEvent[]
- isPolling: boolean

// uiStore - UI全局状态
- sidebarCollapsed: boolean
- currentTheme: 'light' | 'dark'
- notification: Notification | null
```

### API与后端契约

**关键端点**:
- POST /api/v1/releases - 创建发布
- GET /api/v1/releases - 发布列表
- GET /api/v1/releases/{id} - 发布详情
- GET /api/v1/releases/{id}/events - 发布事件
- POST /api/v1/releases/{id}/rollback - 回滚
- GET /api/v1/applications - 应用列表
- GET /api/v1/environments - 环境列表
- GET /api/v1/clusters - 集群列表
- GET /api/v1/app-cluster-configs - 工作负载配置

### 关键特性

1. **类型安全**: 完整的TypeScript类型定义，编译时检查
2. **状态管理**: Pinia store实现响应式状态，DevTools调试
3. **HTTP拦截**: Axios拦截器处理认证、错误、超时
4. **实时更新**: 轮询获取发布进度(2秒间隔)
5. **企业级UI**: Naive UI组件库，支持亮/暗主题
6. **响应式设计**: CSS Grid/Flexbox自适应布局
7. **开发体验**: Vite HMR、TypeScript strict mode

---

## 开发流程

### 后端开发

```bash
# 依赖管理
go mod tidy
go mod download

# 开发运行(使用air热重载)
air

# 编译构建
go build -o server ./cmd/server

# 测试
go test ./...
```

### 前端开发

```bash
# 安装依赖
npm install

# 开发服务器(HMR)
npm run dev

# 类型检查
vue-tsc

# 代码检查/格式化
npm run lint
npm run format

# 生产构建
npm run build

# 预览产物
npm run preview
```

### 全栈开发

```bash
# 根目录运行
npm run dev  # 前后端同时启动

# 后端: http://localhost:8080
# 前端: http://localhost:5173
```

---

## 性能优化

### 后端
- SQLite WAL mode减少写入阻塞
- 数据库连接缓存
- Goroutine池管理异步任务
- 定期清理过期数据

### 前端
- Vite代码分割、Tree shaking
- Lazy loading路由组件
- Pinia store缓存减少API调用
- Naive UI按需导入
- 图片优化、CDN部署

---

## 安全考虑

### 后端
- 敏感信息AES加密(kubeconfig、密钥)
- JSON序列化隐藏字段(json:"-")
- CORS配置限制来源
- 请求日志记录链路追踪ID

### 前端
- TypeScript strict mode预防类型错误
- Axios请求超时控制
- 敏感信息不保存到localStorage
- 错误信息不泄露内部细节

---

## 扩展建议

1. **认证**: 集成OAuth2/OIDC、JWT令牌刷新
2. **WebSocket**: 实时发布推送，替代轮询
3. **监控**: Prometheus指标、链路追踪、日志聚合
4. **缓存**: Redis分布式缓存、数据库查询缓存
5. **容器化**: Docker/K8s部署、CI/CD流水线
6. **国际化**: i18n多语言支持
7. **性能**: 数据库查询优化、索引调整
