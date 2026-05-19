---
name: database-design
description: 发布控制系统 - SQLite数据库架构与模式设计
keywords: SQLite, 数据库设计, Schema, 关系模型, 数据完整性, 性能优化
---

# SQLite数据库架构设计指南

## 数据库概览

**系统**: SQLite3 + Schema Version Management (V3)
**特性**: 无需独立服务器、WAL mode高并发、完整事务支持

## Schema V3 - 完整数据模型

### 1. 应用管理模块

#### application (应用)
```sql
CREATE TABLE application (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  git_repo TEXT,
  build_type TEXT,
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

**用途**: 存储要部署的应用元信息
**关键字段**:
- `name` - 应用唯一标识
- `git_repo` - Git仓库地址
- `build_type` - 构建方式(docker/binary等)

#### environment (环境)
```sql
CREATE TABLE environment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  priority INTEGER,
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

**用途**: 定义逻辑环境分层
**例子**: dev → staging → production
**priority**: 用于排序展示

#### cluster (集群)
```sql
CREATE TABLE cluster (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL,  -- 仅支持 'kubernetes'，其他部署方式通过 shell_server 实现
  kubeconfig BLOB,     -- 加密存储
  k8s_connection_status TEXT,  -- 'connected', 'disconnected', 'unknown'
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

**用途**: 存储物理基础设施集群配置
**安全**: kubeconfig使用AES加密存储
**connection_status**: 前端显示连接状态而非kubeconfig内容

### 2. 发布管理模块

#### workload_target (★核心 - 应用到集群映射)
```sql
CREATE TABLE workload_target (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  app_id INTEGER NOT NULL,
  env_id INTEGER NOT NULL,
  cluster_id INTEGER NOT NULL,
  k8s_namespace TEXT,              -- K8s命名空间
  k8s_workload TEXT,               -- K8s Deployment/StatefulSet名称
  workload_type TEXT,              -- 'Deployment', 'StatefulSet', 'DaemonSet'
  container_name TEXT,             -- 特定容器(防止误更新sidecar)
  registry_domain TEXT,
  image_repo TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE(app_id, env_id, cluster_id),
  FOREIGN KEY(app_id) REFERENCES application(id),
  FOREIGN KEY(env_id) REFERENCES environment(id),
  FOREIGN KEY(cluster_id) REFERENCES cluster(id)
);
```

**用途**: **最核心表** - 定义"应用在某环境某集群上的部署配置"
**唯一性约束**: (app_id, env_id, cluster_id) 必须唯一
**示例**:
```
api-service + production + cluster-prod-1
  → k8s_namespace: production
  → k8s_workload: api-service
  → container_name: api-service
  → registry_domain: harbor.example.com
  → image_repo: company/api-service
```

**优化**: 
- 在三个外键上创建索引加速查询
- 一个应用可在多个环境的多个集群上部署

#### release_record (发布记录)
```sql
CREATE TABLE release_record (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  app_id INTEGER NOT NULL,
  env_id INTEGER NOT NULL,
  cluster_id INTEGER NOT NULL,
  image TEXT NOT NULL,
  status TEXT NOT NULL,  -- 'pending','validating','deploying','success','failed','rolled_back'
  previous_image TEXT,
  triggered_by TEXT,
  started_at DATETIME,
  completed_at DATETIME,
  error_message TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY(app_id) REFERENCES application(id),
  FOREIGN KEY(env_id) REFERENCES environment(id),
  FOREIGN KEY(cluster_id) REFERENCES cluster(id)
);
```

**用途**: 记录单次发布操作完整历史
**生命周期**: pending → validating → deploying → success/failed/rolled_back
**关键字段**:
- `previous_image` - 用于回滚
- `triggered_by` - 操作审计
- `error_message` - 失败原因

#### release_event (发布事件日志)
```sql
CREATE TABLE release_event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  release_id INTEGER NOT NULL,
  type TEXT NOT NULL,                    -- 'started', 'validating', 'deploying', 'pod_updated', 'success', 'failed', 'rolled_back'
  message TEXT,                          -- 事件描述信息
  created_at DATETIME NOT NULL,
  FOREIGN KEY(release_id) REFERENCES release_record(id)
);
```

**用途**: 记录发布全过程的细粒度事件，供前端实时展示
**事件流**: 启动 → 验证 → 部署 → Pod更新 → 完成/失败
**字段说明**:
- `type` - 事件类型标签
- `message` - 用户可读的事件描述

### 3. Shell执行模块

#### shell_server (SSH服务器)
```sql
CREATE TABLE shell_server (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  host TEXT NOT NULL,
  port INTEGER NOT NULL,
  username TEXT NOT NULL,
  auth_type TEXT NOT NULL,  -- 'password' or 'key'
  password BLOB,            -- 加密
  private_key BLOB,         -- 加密
  status TEXT,              -- 'active', 'inactive', 'error'
  last_connected DATETIME,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

**用途**: 管理SSH连接配置
**安全**: 密码和密钥都经过AES加密

#### shell_command (命令白名单)
```sql
CREATE TABLE shell_command (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  server_id INTEGER NOT NULL,
  command TEXT NOT NULL,
  requires_approval INTEGER DEFAULT 0,  -- 0=false, 1=true
  created_at DATETIME NOT NULL,
  FOREIGN KEY(server_id) REFERENCES shell_server(id)
);
```

**用途**: 定义服务器上允许执行的命令
**requires_approval**: 是否需要审批(敏感命令)

**用途**: 预定义可复用的Shell任务

#### shell_command_execution (执行记录)
```sql
CREATE TABLE shell_command_execution (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  status TEXT,  -- 'pending', 'running', 'success', 'failed'
  output TEXT,
  error_message TEXT,
  started_at DATETIME,
  completed_at DATETIME,
  created_at DATETIME NOT NULL,
  FOREIGN KEY(command_id) REFERENCES shell_command(id)
);
```

#### command_approval (审批流)
```sql
CREATE TABLE command_approval (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  command_id INTEGER NOT NULL,
  status TEXT,  -- 'pending', 'approved', 'rejected'
  approver TEXT,
  approved_at DATETIME,
  rejection_reason TEXT,
  created_at DATETIME NOT NULL,
  FOREIGN KEY(command_id) REFERENCES shell_command(id)
);
```

### 4. 系统管理

#### schema_version (版本管理)
```sql
CREATE TABLE schema_version (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  version INTEGER NOT NULL UNIQUE,
  description TEXT,
  applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**用途**: 追踪数据库Schema版本，支持未来迁移
**当前版本**: V3

## 索引策略

### 高频查询的索引

```sql
-- workload_target查询优化
CREATE INDEX idx_workload_target_app_id ON workload_target(app_id);
CREATE INDEX idx_workload_target_env_id ON workload_target(env_id);
CREATE INDEX idx_workload_target_cluster_id ON workload_target(cluster_id);
CREATE INDEX idx_workload_target_app_env_cluster ON workload_target(app_id, env_id, cluster_id);

-- release_record查询
CREATE INDEX idx_release_record_app_env_cluster ON release_record(app_id, env_id, cluster_id);
CREATE INDEX idx_release_record_status ON release_record(status);
CREATE INDEX idx_release_record_created_at ON release_record(created_at DESC);

-- release_event查询
CREATE INDEX idx_release_event_release_id ON release_event(release_id);
CREATE INDEX idx_release_event_created_at ON release_event(created_at DESC);

-- shell_command
CREATE INDEX idx_shell_command_server_id ON shell_command(server_id);
```

## 性能优化

### SQLite3 PRAGMA设置

```golang
// 在db.go中的初始化
PRAGMA journal_mode = WAL;        // 写入预见日志 - 提升并发性能
PRAGMA synchronous = NORMAL;      // 异步写入 - 提升速度
PRAGMA cache_size = -64000;       // 64MB缓存
PRAGMA foreign_keys = ON;         // 启用外键约束
PRAGMA busy_timeout = 5000;       // 5秒超时
```

### 查询优化建议

1. **批量插入** - 使用事务包装
2. **定期VACUUM** - 清理碎片
3. **避免SELECT *** - 只查需要的字段
4. **使用EXPLAIN QUERY PLAN** - 分析慢查询

## 数据一致性

### 外键约束
- 所有关联表都定义FOREIGN KEY
- 删除应用时自动级联删除发布记录

### 唯一性约束
- `application.name` - 应用名唯一
- `cluster.name` - 集群名唯一
- `(app_id, env_id, cluster_id)` - 工作负载映射唯一

### 敏感信息加密
- `cluster.kubeconfig` - AES加密存储
- `shell_server.password` - AES加密
- `shell_server.private_key` - AES加密
- Go模型层使用 `json:"-"` 隐藏这些字段

## 初始化流程

```
1. 启动应用 (main.go)
   ↓
2. database.Init(dbPath)
   ├─ 打开或创建SQLite文件
   ├─ 设置PRAGMA优化
   └─ 调用createTables()
   ↓
3. createTables()
   ├─ initSchemaVersion()
   ├─ applyMigrations()
   │  └─ 比较当前版本vs目标版本
   │     └─ 执行必要的迁移脚本
   └─ if INIT_DATA=true && version==1
      └─ InsertInitialData()
```

## 迁移策略

> **当前阶段**: 预开发阶段 (MVP)。数据库设计处于初稳定状态。未来可能需要迁移，通过问答流程确认。

### 迁移框架设计

本系统采用**版本化迁移**模式，支持未来扩展：

```golang
// internal/database/migration.go

// MigrationRegistry - 迁移注册表
var migrations = map[int]func(*sql.DB) error{
    1: migrateSchemaV1,  // 初始Schema
    2: migrateSchemaV2,  // 可选的扩展
    3: migrateSchemaV3,  // 当前版本
}

// Migrate - 执行从oldVersion到newVersion的迁移
func Migrate(db *sql.DB, oldVersion, newVersion int) error {
    for v := oldVersion + 1; v <= newVersion; v++ {
        migrator, ok := migrations[v]
        if !ok {
            return fmt.Errorf("migration v%d not found", v)
        }
        if err := migrator(db); err != nil {
            return fmt.Errorf("migration v%d failed: %w", v, err)
        }
    }
    return nil
}
```

### 版本升级示例

```golang
// migrateSchemaV3() - 添加workload_type字段等
func migrateSchemaV3(db *sql.DB) error {
    // 步骤1: 验证表存在性
    // 步骤2: 添加新列 (使用IF NOT EXISTS)
    alterSQL := `
    ALTER TABLE workload_target ADD COLUMN workload_type TEXT DEFAULT 'Deployment';
    ALTER TABLE release_record ADD COLUMN rollback_from_id INTEGER;
    `
    
    // 步骤3: 执行迁移 (事务包装)
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    if _, err := tx.Exec(alterSQL); err != nil {
        return err
    }
    
    // 步骤4: 数据迁移 (填充默认值、数据转换等)
    // 步骤5: 验证完整性 (检查没有NULL、约束符合等)
    // 步骤6: 更新schema_version表
    
    return tx.Commit().Error
}
```

### 向后兼容性策略

**当前立场**: 预开发MVP阶段，**暂不向后兼容**

| 场景 | 决策 | 理由 |
|------|------|------|
| 新增表 | 向前兼容 | 不影响现有代码 |
| 新增字段（非nullable） | 需确认 | 可能影响插入逻辑 |
| 删除表/字段 | 需确认 | 影响现有查询 |
| 字段类型变更 | 需确认 | 可能影响数据完整性 |
| 业务逻辑变更 | 需确认 | 评估遗留数据影响 |

**决定向后兼容的问答流程**:

当需要迁移时，提问以下问题：
1. 是否有生产环境数据需要保留？
2. 对应的业务流程有何影响？
3. 数据转换成本与风险评估？
4. 是否需要滚动升级（多版本并存）？
5. 回滚方案如何设计？

**示例答案决策**:
- ✅ "保留所有现有数据，新字段允许NULL" → 支持向后兼容
- ❌ "重建表结构，删除老字段" → 不支持，需新版本
- ✅ "添加新表，原表无改动" → 完全兼容，可直接升级

## 备份与恢复

```bash
# 备份
sqlite3 release_control.db ".dump" > backup.sql

# 恢复
sqlite3 release_control.db < backup.sql

# 导出CSV
sqlite3 -header -csv release_control.db "SELECT * FROM release_record;" > export.csv
```

## 监控指标

- 数据库文件大小
- 连接数/并发事务数
- 慢查询(>100ms)
- WAL文件大小
- 表行数统计

