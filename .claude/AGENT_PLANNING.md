# 🎯 Agent 协作规划

## 当前Agent配置

```
核心阵容 (当前采用):
  ✅ BE Agent     (后端高级开发)
  ✅ FE Agent     (前端高级开发)
```

---

## Agent 职责分工

### 后端Agent (be)

**核心工作**:
- Go语言开发、API设计、数据库设计
- K8s部署实现、SSH执行管理
- 发布流程、状态管理、错误处理
- **兼职运维**: 基础设施权限配置、部署测试

### 前端Agent (fe)

**核心工作**:
- Vue3/TypeScript UI开发
- Pinia状态管理、HTTP通信
- 发布流程UI、实时更新、历史管理
- 样式设计、组件复用

---

## 关键工作项的分工

| 工作项 | 负责方 | 备注 |
|--------|--------|------|
| 数据库设计与初始化 | BE | 包括索引、性能优化 |
| K8s部署实现 | BE | 集成client-go |
| Shell/SSH执行 | BE | 集成Salt/Ansible等 |
| REST API实现 | BE | go-chi框架 |
| Vue3页面开发 | FE | 所有UI组件 |
| Pinia状态管理 | FE | 缓存、轮询逻辑 |
| HTTP请求层 | FE | Axios、拦截器 |
| 发布流程集成测试 | BE/FE | 协作验证 |
| 环境配置 & kubeconfig | BE(兼) | 可向运维寻求协助 |
