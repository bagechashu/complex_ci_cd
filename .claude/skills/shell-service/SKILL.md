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
- 支持单服务器和多服务器执行（串行/并行）
- 所有执行结果完整记录到shell_exec_task表
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
shell_command (命令白名单)
  ├─ command: 实际执行的命令
  ├─ description: 命令说明
  └─ is_published: 是否已发布（未发布不能执行）
       ↓ (1:N)
shell_exec_task (执行记录)
  ├─ status: 'running', 'success', 'failed'
  ├─ output: 命令输出
  ├─ exit_code: Unix退出码
  └─ error_message: 执行错误信息
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
    taskID *int,        // 可选：关联的shell_exec ID
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

### 2. 并行执行（多服务器）

```go
// ExecuteTaskParallel 在多个服务器上并行执行同一命令
func (s *ShellService) ExecuteTaskParallel(
    ctx context.Context,
    commandID int,
    serverIDs []int,    // 服务器ID列表
    taskID *int,        // 可选：关联的shell_exec ID
) (results map[int]ExecutionResult, err error)
```

**特性**:
- 为每个服务器启动独立的Goroutine
- 所有执行并行进行（受系统资源限制）
- 单个服务器失败不影响其他服务器
- 返回所有服务器的执行结果

**示例**:
```go
results, err := shellService.ExecuteTaskParallel(
    ctx,
    commandID=10,
    serverIDs=[]int{1, 2, 3, 4},
    nil,
)

for serverID, result := range results {
    if result.Error != nil {
        log.Printf("Server %d failed: %v", serverID, result.Error)
    } else {
        log.Printf("Server %d: exit_code=%d, output_size=%d", 
            serverID, result.ExitCode, len(result.Output))
    }
}
```

### 3. 串行执行（多服务器）

```go
// ExecuteTaskSerial 在多个服务器上串行执行同一命令
func (s *ShellService) ExecuteTaskSerial(
    ctx context.Context,
    commandID int,
    serverIDs []int,    // 服务器ID列表
    taskID *int,        // 可选：关联的shell_exec ID
) (results map[int]ExecutionResult, err error)
```

**特性**:
- 按顺序在服务器上执行命令
- 单个服务器失败后继续执行下一个
- 适合对执行顺序有依赖的场景
- Context取消时立即停止

**示例**:
```go
// 先在dev环境验证，再在prod环境部署
results, err := shellService.ExecuteTaskSerial(
    ctx,
    commandID=15,
    serverIDs=[]int{devServerID, prodServerID},
    nil,
)
```

---

## 使用场景

### 场景1: 执行Salt命令

```go
// 1. 数据库中预定义命令
// INSERT INTO shell_command (server_id, command, is_published)
// VALUES (1, 'salt "prod-*" state.apply webserver', true)

// 2. 在代码中执行
exitCode, output, err := shellService.ExecuteCommand(ctx, commandID=1, serverID=1, nil)

// 示例输出:
// exit_code: 0
// output: "summary: {...}, duration: 45.123s"
```

### 场景2: 执行Ansible Playbook

```go
// 1. 服务器上已准备好playbook
// /opt/playbooks/deploy.yml
// /opt/inventory/production

// 2. 数据库中预定义命令
// INSERT INTO shell_command (server_id, command, is_published)
// VALUES (2, 'ansible-playbook /opt/playbooks/deploy.yml -i /opt/inventory/production -e image=myapp:v1.2.3', true)

// 3. 在代码中执行
exitCode, output, err := shellService.ExecuteCommand(ctx, commandID=3, serverID=2, nil)
```

### 场景3: 在多个集群并行部署

```go
// 在4个集群上并行执行相同的部署脚本
results, err := shellService.ExecuteTaskParallel(
    ctx,
    commandID=20,  // 部署脚本命令
    serverIDs=[]int{clusterA, clusterB, clusterC, clusterD},
    nil,
)

// 检查所有结果
successCount := 0
for _, result := range results {
    if result.IsSuccess() {
        successCount++
    }
}

if successCount == len(results) {
    log.Printf("All clusters deployed successfully")
} else {
    log.Printf("Deployment failed on some clusters: %d/%d", successCount, len(results))
}
```

---

## 认证方式

### 1. 密码认证

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

### shell_exec_task表

```go
type ShellExecTask struct {
    ID            int        // 执行记录ID
    TaskID        int        // 关联的task ID（可选）
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
// 获取特定任务在特定服务器上的最新执行记录
execution, err := execRepo.GetLatestByTaskAndServer(ctx, taskID=1, serverID=2)

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

results, err := shellService.ExecuteTaskParallel(ctx, commandID, serverIDs, nil)
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

