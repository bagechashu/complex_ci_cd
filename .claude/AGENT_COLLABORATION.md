# 🤝 前后端 Agent 协作约定

> 确保 BE Agent 和 FE Agent 高效协作，快速交付发布控制系统 MVP

---

## 📚 文档导航

**背景理解**（新人必读）：
- [核心问题和目标](../prompt/核心问题和目标.md) - 为什么要做这个系统？
- [架构设计](../prompt/架构设计.md) - 系统如何分层？分哪些阶段建设？

**协作执行**（本文件）：
- 前后端分工清晰
- 接口契约优先
- 沟通、Bug报告、质量检查规范

**代码实现**（Day 1-6 路线）：
- [BE Agent](./be.agent.md) - 后端 6 天实现计划
- [FE Agent](./fe.agent.md) - 前端 6 天实现计划

---

## 核心原则

### 分工清晰

| 领域 | BE Agent | FE Agent |
|------|----------|----------|
| **数据模型** | ✅ 设计 + 实现 | ❌ |
| **业务逻辑** | ✅ 实现 | ❌ |
| **API 设计** | ✅ 实现 | ✅ 消费 |
| **UI 组件** | ❌ | ✅ 实现 |
| **状态管理** | ❌ | ✅ 实现 (Pinia) |
| **HTTP 通讯** | ✅ 返回 JSON | ✅ 发送 JSON |
| **实时更新** | ✅ 提供事件日志 API | ✅ 轮询/WebSocket |

### 接口契约优先

在代码实现前，必须有**明确、文档化的后端 API 契约**（OpenAPI/Swagger），FE 基于契约进行开发。

完整的 API 文档、响应格式规范、数据类型定义详见：
- 📌 `skills/api-design/SKILL.md` - REST API 设计规范（所有端点定义）
- 📌 `be.agent.md` - 后端 API 实现指导（Day 5 API 接口层）
- 📌 `fe.agent.md` - 前端 API 集成指导（响应拦截器、类型定义）

---

## 沟通规范

### Slack/钉钉 快速反馈

| 情况 | 频道 | 内容 |
|------|------|------|
| 接口设计问题 | #release-system | 具体的 API 改进建议 |
| 测试发现的 Bug | #release-system | 完整的复现步骤 |
| 阻塞问题 | @mention | 紧急问题直接 @ |
| 日版本同步 | #release-system | 每日更新进度 (EOD) |

### 每日进度同步

**格式**:
```
✅ 完成: 
  - API 框架搭建
  - workload_target 表设计

🚀 今日进行中:
  - K8sDeployer 实现

🔴 阻塞:
  - 等待 Harbor registry 访问权限

📅 明日计划:
  - Deploy Service 完整实现
```

---

## Bug 报告规范

### 后端发现的前端 Bug

```
标题: [FE BUG] 发布列表分页不工作

详情:
- 环境: localhost:5173
- 步骤:
  1. 打开发布历史页面
  2. 设置 limit=5
  3. 点击下一页
- 预期: 显示第 2 页
- 实际: 还是显示第 1 页
- 网络: GET /api/v1/release?limit=5&offset=5 返回 200, 正确数据

代码位置: src/pages/ReleaseHistory.vue:123
```

### 前端发现的后端 Bug

```
标题: [BE BUG] Rollback API 返回错误响应格式

详情:
- 请求: POST /api/v1/release/123/rollback
- 期望返回: { code, data: { rollback_release_id, ... } }
- 实际返回: 直接返回 { release_id } (格式不一致)

影响: 前端 response.data.rollback_release_id 获取 undefined
```

---

## 部署前检查清单

### BE 部署前

- [ ] 所有接口都返回约定的 JSON 格式
- [ ] 错误时含有 code 和 request_id
- [ ] 202 Accepted 用于异步操作
- [ ] request_id 来自请求头 X-Request-ID
- [ ] 数据库事务保证一致性
- [ ] 日期字段都是 ISO 8601
- [ ] SQL 性能通过了压力测试

### FE 部署前

- [ ] 所有 API 调用都使用 TypeScript 类型
- [ ] 错误处理完善 (网络错误、业务错误)
- [ ] 轮询逻辑在页面卸载时清理
- [ ] 所有输入都经过前端验证
- [ ] 响应式设计适配移动端
- [ ] 加载态和禁用态显示正确
- [ ] 浏览器控制台无 TypeScript 错误

---

## 总结

✅ **API 契约是唯一的真理**  
✅ **并行开发，减少依赖**  
✅ **每日同步进度和问题**  
✅ **统一的数据格式和错误处理**  
✅ **充分的类型检查 (TypeScript)**  

🎯 **目标**: 6 天高质量交付发布控制系统 MVP
