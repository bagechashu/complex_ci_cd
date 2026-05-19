---
description: "Design system architecture, analyze component interactions, and generate implementation skeletons for release control system features"
name: "System Architecture Design"
argument-hint: "Feature/component to design or analyze"
agent: "agent"
tools: ["semantic-search", "file-access", "workspace-files"]
---

# System Architecture Design

You are helping design or improve system architecture for a release control system (CI/CD control platform) with a Go backend and Vue3 frontend.

## 🎯 何时使用此 Prompt？

✅ **使用场景** - 需要架构级设计时：
- 添加**全新的业务域**（如权限审批系统、通知中心、监控告警等）
- 大规模**重构现有模块**（如 Release 流程演进、权限体系升级）
- 设计**新的数据结构**或**服务层方法**
- 新员工**学习系统架构设计思路**

❌ **不需要使用** - 日常开发场景：
- 日常 Bug 修复或小功能增强 → 用 [be.agent.md](../agents/be.agent.md) / [fe.agent.md](../agents/fe.agent.md)
- 现有 API 的简单调用 → 用相关 **Skill** 文件（api-design、pinia-stores 等）
- 数据库查询优化 → 用 [database-design/SKILL.md](../skills/database-design/SKILL.md)
- 编码实现细节 → 用相应的 **Golang**、**Naive UI** 等 Skill

## 使用方式

1. **在 Chat 中直接输入需求**（告诉我你要设计什么）
2. **参考此 Prompt**（我会按照下面的流程来帮你）
3. **我会给你完整的架构提案**（包括分层、代码骨架、集成方案）

---

## Task Options

### Option 1: Design New Component/Feature
When given a feature requirement or user story:

1. **Understand Context**: Ask about integration points, deployment targets, and existing patterns
2. **Propose Architecture**: 
   - Map to existing domain models and services
   - Identify new data structures needed
   - Show how it integrates with current layers (handlers → services → repository → database)
3. **Generate Skeleton Code**:
   - Go: models, service interfaces, repository patterns
   - Vue3: components, stores (Pinia), API service layer
   - Database: table definitions or migrations
4. **Document Decisions**: Why this structure? Trade-offs vs alternatives?

### Option 2: Analyze Existing Component
When analyzing an existing feature:

1. **Map Current Structure**: Trace request flow from frontend UI through API handler to database
2. **Identify Patterns**: What design patterns are used? (dependency injection, factory, etc.)
3. **Assess Quality**: 
   - Is separation of concerns maintained?
   - Are there coupling issues?
   - Does it match project conventions?
4. **Suggest Improvements**: What could be refactored or optimized?

## Project Context

- **Backend**: Go + SQLite, organized in `internal/` with handlers → services → repository pattern
- **Frontend**: Vue3 + Naive UI + Pinia stores, organized by feature/domain
- **Domains**: Applications, Clusters, Environments, Shell Servers, Shell Commands, Workloads, Release Events
- **Deployment**: Kubernetes deployers, shell command executors, release workflows

## Key Constraints

✅ DO:
- Reference [ARCHITECTURE.md](../../backend/ARCHITECTURE.md) and [ARCHITECTURE.md](../../frontend/ARCHITECTURE.md) 
- Follow existing patterns (DDD, service layer, Pinia stores)
- Use Go concurrency patterns (goroutines, channels, context)
- Align frontend with Naive UI and CSS architecture
- Consider database normalization and indexing

❌ DON'T:
- Invent new patterns; adapt existing ones
- Add dependencies without justification
- Create monolithic services; keep services focused
- Mix business logic in handlers or components

## Deliverables

When designing, provide:

```
## Architecture Proposal

### Component Overview
[High-level description + interaction diagram if helpful]

### Data Model
- [New tables/fields needed]
- [Relationships to existing models]

### Backend Implementation
- Services (new methods, responsibilities)
- Repository (query patterns, transactions)
- Handlers (endpoint signatures, validation)

### Frontend Implementation
- Components (hierarchy, props, events)
- Pinia store (state, actions, getters)
- API layer (request/response types)

### Database Schema
[CREATE TABLE or migration script if applicable]

### Integration Points
- Existing APIs it consumes
- New APIs it exposes
- Permission/approval workflows

### Trade-offs & Rationale
- Why this design over alternatives?
- Scalability considerations
- Security implications
```

## Example Invocations

- "Design a notification system that alerts users when releases fail"
- "Analyze the Shell Command execution flow and suggest refactoring opportunities"
- "Create the data model and service layer for multi-environment promotion workflows"
- "Generate handlers and Pinia stores for the new environment approval dashboard"
