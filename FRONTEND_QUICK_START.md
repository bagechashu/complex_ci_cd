# 前端快速开始指南

## 🎯 1 分钟快速开始

### 前置要求
- Node.js >= 16 （推荐 18+）
- npm >= 8

### 安装

```bash
cd /Users/op/Downloads/complex_ci_cd/frontend
npm install
```

### 开发

```bash
npm run dev
```

打开浏览器: **http://localhost:5173**

### 生产构建

```bash
npm run build
npm run preview
```

---

## 📋 项目结构速览

```
frontend/
├── src/
│   ├── api/              # HTTP 请求层（3个服务）
│   ├── stores/           # 状态管理（3个 Pinia stores）
│   ├── types/            # TypeScript 类型
│   ├── views/            # 3 个核心页面
│   ├── router/           # 路由配置
│   ├── styles/           # 全局样式
│   ├── utils/            # 工具函数
│   ├── App.vue           # 根组件
│   └── main.ts           # 入口
├── vite.config.ts        # Vite 配置（自动代理到后端）
└── package.json          # 依赖管理
```

---

## 🔌 对接后端 API

### 环境配置

1. 复制 `.env.example` 为 `.env`：

```bash
cp .env.example .env
```

2. 修改 `.env` 文件中的 API 地址：

```env
VITE_API_BASE_URL=http://localhost:8080
```

### 验证后端连接

启动后端服务后，前端会自动连接。可检查浏览器控制台是否有网络错误。

### API 端点验证

```bash
# 检查后端是否运行
curl http://localhost:8080/health

# 获取应用列表
curl http://localhost:8080/api/v1/applications

# 发起发布测试
curl -X POST http://localhost:8080/api/v1/releases \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "env_id": 1,
    "cluster_id": 1,
    "image": "myapp:v1.0.0",
    "user": "admin"
  }'
```

---

## 🌐 核心页面说明

### 1. 发布向导 (ReleaseFlow) `/release`

**功能**: 4 步向导式发布流程

```
Step 1: 选择应用
Step 2: 选择环境（根据可用集群动态加载）
Step 3: 选择集群
Step 4: 输入镜像标签
Step 5: 确认并发布
```

**发布后**:
- 自动启动轮询（2 秒间隔）
- 显示实时进度条
- 显示事件日志
- 完成时自动停止轮询

### 2. 发布历史 (ReleaseHistory) `/history`

**功能**: 管理所有发布记录

- 数据表格（应用、环境、集群、镜像、状态、时间）
- 筛选（按应用/环境/状态）
- 分页（每页 20 条）
- 快速操作：
  - 点击"详情"查看完整信息
  - 点击"回滚"快速回滚（仅成功的发布）

### 3. 发布详情 (ReleaseDetail) `/detail/:id`

**功能**: 详细查看发布过程

- 基本信息面板
- 实时进度条
- 完整事件日志
- 导出日志功能
- 回滚/重试按钮

---

## 🛠️ 常用命令

### 开发

```bash
# 启动开发服务器
npm run dev
make dev

# 代码检查和格式化
npm run lint
npm run format

# 构建
npm run build
make build

# 预览构建结果
npm run preview
make preview
```

### 清理

```bash
# 清理依赖和构建产物
make clean
```

---

## 🔍 调试技巧

### 1. 浏览器开发者工具

- **Network 标签**: 查看 API 请求和响应
- **Console 标签**: 查看日志和错误
- **Vue 调试工具**: 检查 Pinia stores 和组件状态

### 2. API 代理

Vite 自动代理 `/api/*` 请求到 `http://localhost:8080`:

```
浏览器请求 http://localhost:5173/api/v1/releases
  ↓ (由 Vite 代理)
后端服务 http://localhost:8080/api/v1/releases
```

### 3. 状态查看

使用 Vue DevTools 可实时查看：
- appStore （应用元数据）
- releaseStore （发布状态）
- uiStore （消息通知）

---

## ⚠️ 常见问题

### Q1: 启动时报错 "cannot find module"

**解决方案**:
```bash
npm install
npm run dev
```

### Q2: API 连接失败

**检查清单**:
- [ ] 后端服务是否启动 (`go run cmd/server/main.go`)
- [ ] 后端监听地址是否为 `localhost:8080`
- [ ] 防火墙是否阻止 8080 端口
- [ ] `.env` 文件中的 API 地址是否正确

### Q3: 发布进度不更新

**可能原因**:
- 后端返回的事件类型不匹配（检查事件类型）
- 轮询被中止（查看浏览器控制台）
- API 返回错误（查看 Network 标签）

**解决方案**:
- 打开浏览器 DevTools (F12)
- Network 标签查看请求是否成功
- Console 标签查看错误信息

### Q4: 如何修改轮询间隔

在 `.env` 中修改:
```env
VITE_POLLING_INTERVAL=2000  # 毫秒
```

### Q5: 表格显示为空或数据不更新

**检查**:
1. 后端是否返回了应用列表
2. 浏览器控制台是否有错误
3. 尝试刷新页面 (F5)

---

## 📦 依赖说明

| 包名 | 用途 | 版本 |
|------|------|------|
| vue | 前端框架 | 3.3.4 |
| vite | 构建工具 | 5.0.7 |
| vue-router | 路由管理 | 4.2.5 |
| pinia | 状态管理 | 2.1.6 |
| naive-ui | UI 组件库 | 2.34.6 |
| axios | HTTP 客户端 | 1.6.1 |
| typescript | 类型检查 | 5.2.2 |
| dayjs | 日期处理 | 1.11.10 |

---

## 🚀 生产部署

### 构建

```bash
npm run build
```

输出: `dist/` 目录（可直接部署到 nginx）

### Nginx 配置示例

```nginx
server {
    listen 80;
    server_name release.example.com;

    location / {
        root /path/to/dist;
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
    }
}
```

---

## 📚 更多资源

- **README.md** - 完整项目说明
- **ARCHITECTURE.md** - 架构设计和数据流
- **FRONTEND_SUMMARY.md** - 详细项目总结

---

## 🎓 下一步

1. **立即开始**
   ```bash
   npm install && npm run dev
   ```

2. **连接后端**
   - 启动后端服务
   - 修改 `.env` API 地址
   - 刷新浏览器验证

3. **测试完整流程**
   - 在 ReleaseFlow 页面走通 4 步流程
   - 查看发布进度和事件日志
   - 尝试回滚操作

4. **探索代码**
   - 查看 `src/stores/releaseStore.ts` 了解状态管理
   - 查看 `src/views/ReleaseFlow.vue` 了解页面结构
   - 查看 `src/api/release.ts` 了解 API 调用

---

**Happy Coding! 🎉**
