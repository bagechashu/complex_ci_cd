---
name: shell-service
description: 发布控制系统 - Shell/SSH命令执行服务
keywords: SSH执行, Shell命令, Salt, Ansible, 远程执行, 命令白名单
---

# Shell/SSH命令执行服务指南

## 概述

ShellService 提供通过SSH执行远程Shell命令的能力，用于处理Salt、Ansible等不通过标准DeployStrategy执行的部署方式。

**设计原则**:
- 只执行预定义的**白名单命令**（shell_command表）
- 支持密钥和密码两种认证方式（均使用AES加密存储）
- **单服务器单命令执行模型**（前端直接选择已发布命令执行）
- 所有执行结果完整记录到shell_command_execution表
- SSH连接支持缓存和复用

---

## 核心概念

### 数据模型关系

```
shell_server (SSH服务器配置)
  ├─ host, port, username
  ├─ auth_type: 'password' or 'key'
  ├─ password/private_key (AES加密)
  └─ status: 'active', 'inactive', 'error'
       ↓ (1:N)
shell_command (命令白名单，按服务器归组)
  ├─ server_id: 该命令绑定的服务器
  ├─ command: 实际执行的命令
  ├─ description: 命令说明
  └─ is_published: 是否已发布（未发布不能在UI显示）
       ↓ (1:N)
shell_command_execution (执行记录)
  ├─ command_id, server_id: 关联的命令和服务器
  ├─ status: 'pending', 'running', 'success', 'failed'
  ├─ output: 命令输出
  ├─ exit_code: Unix退出码
  ├─ error_message: 执行错误信息
  └─ command_params: 执行参数（如镜像版本等）
```

---

## ShellService API

### 1. 单命令执行

```go
// ExecuteCommand 在单个服务器上执行命令
func (s *ShellService) ExecuteCommand(
    ctx context.Context,
    commandID int,      // shell_command表的ID
    serverID int,       // shell_server表的ID
) (exitCode int, output string, err error)
```

**执行流程**:
1. 验证命令ID存在且已发布(is_published=true)
2. 验证服务器ID存在且状态为active
3. 记录执行开始时间
4. 建立SSH连接（支持缓存）
5. 执行命令并捕获输出
6. 记录执行结果和退出码
7. 更新服务器最后连接时间

**示例**:
```go
exitCode, output, err := shellService.ExecuteCommand(ctx, commandID=5, serverID=2, nil)
if err != nil {
    log.Printf("Execution failed: %v", err)
}
log.Printf("Exit code: %d\nOutput: %s", exitCode, output)
```

### 2. 执行历史查询

```go
// GetCommandExecutions 查询执行历史
func (r *ShellCommandExecutionRepository) GetExecutionsByPage(
    ctx context.Context,
    page, pageSize int,
) ([]ShellCommandExecution, int, error)
```

**特性**:
- 按时间倒序获取执行记录
- 包含完整的输出和错误信息
- 支持前端分页展示

**示例**:
```go
executions, total, err := execRepo.GetExecutionsByPage(ctx, page=1, pageSize=20)
for _, exec := range executions {
    log.Printf("[%s] Command %d: %s (exit=%d)", 
        exec.UpdatedAt, exec.ID, exec.Status, exec.ExitCode)
}
```

---

## 使用场景

### 场景1: 直接执行已发布命令

**流程**:
1. 前端显示已发布的Shell命令列表（按服务器分组）
2. 用户点击"执行"按钮
3. 创建shell_command_execution记录，状态=pending
4. 后端异步执行命令，更新状态
5. 前端轮询查询执行结果

**代码示例**:
```go
// 1. 数据库中预定义命令
// INSERT INTO shell_command (server_id, command, is_published)
// VALUES (1, 'salt "prod-*" state.apply webserver', true)

// 2. 前端调用执行端点
// POST /v1/shell-commands/execute
// { "command_id": 1, "server_id": 1 }

// 3. 后端创建执行记录
exec := &models.ShellCommandExecution{
    CommandID: 1,
    ServerID: 1,
    Status: "pending",
}
execRepo.Create(ctx, exec)  // 返回带ID的执行记录

// 4. 异步执行（后台任务）
exitCode, output, err := shellService.ExecuteCommand(ctx, 1, 1, &exec.ID)
```

### 认证方式

#### 1. 密码认证

```go
server := &models.ShellServer{
    Name:     "prod-server-1",
    Host:     "192.168.1.10",
    Port:     22,
    Username: "deploy",
    AuthType: "password",
    Password: encryptedPassword,  // 使用AES加密存储
    Status:   "active",
}
```

### 2. SSH密钥认证

```go
server := &models.ShellServer{
    Name:       "prod-server-2",
    Host:       "192.168.1.11",
    Port:       22,
    Username:   "deploy",
    AuthType:   "key",
    PrivateKey: encryptedPrivateKey,  // 使用AES加密存储
    Status:     "active",
}
```

---

## 连接管理

### SSH连接缓存

```go
// 连接自动缓存 (host:port → *ssh.Client)
// 默认最多缓存50个连接

// 再次连接到同一服务器时，会复用缓存
// 如果连接已断开，自动清理并重新建立

exitCode1, _, _ := shellService.ExecuteCommand(ctx, cmd1, srv1, nil)  // 新建连接
exitCode2, _, _ := shellService.ExecuteCommand(ctx, cmd2, srv1, nil)  // 复用连接
```

### 连接验证

```go
// 每次获取缓存连接时，先验证连接有效性
// 如果连接失败，自动清理缓存并创建新连接
```

### 关闭所有连接

```go
// 应用关闭时调用
defer shellService.CloseConnections()
```

---

## 命令白名单机制

### 设计目的

- **安全性**: 防止执行任意命令
- **可追踪性**: 只能执行预定义的命令
- **审计**: 所有执行都有记录

### 命令状态

| 状态 | is_published | 说明 |
|------|--------------|------|
| 草稿 | false | 定义但未发布，不能执行 |
| 已发布 | true | 可以执行 |

### 命令定义示例

```sql
-- Salt命令
INSERT INTO shell_command (server_id, command, description, is_published)
VALUES (1, 'salt "prod-*" state.apply webserver', 'Apply webserver state on prod', true);

-- Ansible命令
INSERT INTO shell_command (server_id, command, description, is_published)
VALUES (2, 'ansible-playbook /opt/deploy.yml -i /opt/inventory', 'Deploy application', true);

-- 自定义脚本
INSERT INTO shell_command (server_id, command, description, is_published)
VALUES (3, 'bash /opt/scripts/deploy.sh --env prod', 'Production deployment', true);

-- 健康检查
INSERT INTO shell_command (server_id, command, description, is_published)
VALUES (4, 'curl -s http://localhost:8080/health | jq .', 'Health check', true);
```

---

## 执行记录

### shell_command_execution表

```go
type ShellCommandExecution struct {
    ID            int        // 执行记录ID
    ServerID      int        // 执行的服务器
    CommandID     int        // 执行的命令
    Status        string     // running, success, failed
    Output        string     // 完整的stdout+stderr
    ErrorMessage  string     // 执行错误信息
    ExitCode      *int       // Unix退出码
    StartedAt     *time.Time // 开始时间
    CompletedAt   *time.Time // 完成时间
    CreatedAt     time.Time  // 记录创建时间
}
```

### 查询执行历史

```go
// 获取特定命令在特定服务器上的最新执行记录
execution, err := execRepo.GetLatestByCommandAndServer(ctx, commandID=1, serverID=2)

// 获取特定服务器的所有执行记录
executions, total, err := execRepo.ListByServer(ctx, serverID=2, offset=0, limit=20)

// 查询执行时长
duration := execution.GetDuration()  // 返回秒数
```

---

## 超时设置

### 默认超时

- **SSH连接建立**: 10秒
- **单个命令执行**: 5分钟
- **Context超时**: 由调用者控制

### 自定义超时

```go
// 设置整体超时为2分钟
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

results, err := shellService.ExecuteCommandParallel(ctx, commandID, serverIDs, nil)
```

---

## 错误处理

### 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| "command not found" | 命令ID不存在 | 检查command_id |
| "command not published" | 命令未发布 | 在管理界面发布命令 |
| "server not found" | 服务器ID不存在 | 检查server_id |
| "server not active" | 服务器状态非active | 检查服务器状态 |
| "SSH connection failed" | SSH连接失败 | 检查网络、认证信息 |
| "command execution timeout" | 命令执行超时 | 增加超时时间或优化命令 |

### 错误恢复

```go
exitCode, output, err := shellService.ExecuteCommand(ctx, cmdID, srvID, nil)
if err != nil {
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        log.Printf("Timeout executing command")
    case strings.Contains(err.Error(), "SSH"):
        log.Printf("Connection error, retry later")
    case strings.Contains(err.Error(), "not published"):
        log.Printf("Command not published")
    default:
        log.Printf("Execution error: %v", err)
    }
}
```

---

## 集成示例

### 在Release流程中使用

```go
// internal/services/release_service.go

func (s *ReleaseService) DeployUsingSaltStack(
    ctx context.Context,
    release *models.ReleaseRecord,
    saltServerID int,
) error {
    // 1. 获取执行命令ID（预定义在数据库中）
    saltCommand, err := s.commandRepo.GetByID(ctx, release.SaltCommandID)
    if err != nil {
        return err
    }

    // 2. 执行Salt部署
    exitCode, output, err := s.shellService.ExecuteCommand(
        ctx,
        saltCommand.ID,
        saltServerID,
        nil,
    )

    if err != nil || exitCode != 0 {
        return fmt.Errorf("salt deployment failed: exit_code=%d, error=%v", exitCode, err)
    }

    // 3. 记录部署成功
    s.logger.Info("Salt deployment succeeded", "output", output)
    return nil
}
```

---

## 最佳实践

### 1. 命令设计

```go
// ✅ 好的例子：命令清晰、幂等、有返回值
"salt 'prod-web-*' state.apply webserver --out=json"

// ❌ 不好的例子：命令复杂、副作用多、难追踪
"cd /opt/app && git pull && make build && make deploy"
```

### 2. 错误处理

```go
// ✅ 记录详细信息
if err != nil {
    s.log.Error("Shell execution failed",
        "command_id", cmdID,
        "server_id", srvID,
        "exit_code", exitCode,
        "error", err,
        "output", output[:min(len(output), 500)],  // 截断长输出
    )
}

// ❌ 忽略错误
_, _, _ = shellService.ExecuteCommand(ctx, cmdID, srvID, nil)
```

### 3. 资源管理

```go
// ✅ 应用关闭时清理连接
defer func() {
    if err := shellService.CloseConnections(); err != nil {
        log.Printf("Failed to close SSH connections: %v", err)
    }
}()

// ❌ 连接泄漏
```

### 4. 日志记录

```go
// ✅ 记录所有执行
func (s *ShellService) ExecuteCommand(...) (..., err error) {
    defer func() {
        s.log.Info("Command execution completed",
            "command_id", commandID,
            "server_id", serverID,
            "exit_code", exitCode,
            "duration", time.Since(startTime),
            "error", err,
        )
    }()
    // ...
}
```

