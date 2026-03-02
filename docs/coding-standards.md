# AgentTeams 编码规范

本文档定义了 AgentTeams 项目的编码规范，所有贡献者必须遵循。

## 目录

- [通用规范](#通用规范)
- [Go 编码规范 (Server)](#go-编码规范-server)
- [C++ 编码规范 (Agent)](#c-编码规范-agent)
- [数据库规范](#数据库规范)
- [API 设计规范](#api-设计规范)
- [Git 提交规范](#git-提交规范)

---

## 通用规范

### 文件命名

| 类型 | 规范 | 示例 |
|------|------|------|
| Go 文件 | 小写蛇形命名 | `agent_handler.go`, `task_service.go` |
| C++ 文件 | 小写蛇形命名 | `websocket_client.cpp`, `task_worker.hpp` |
| 配置文件 | 小写蛇形命名 | `config.yaml`, `docker-compose.yaml` |
| 文档文件 | 大写蛇形命名 | `README.md`, `CONTRIBUTING.md` |

### 目录结构

```
AgentTeams/
├── agent/                     # C++ Agent 代码
│   ├── src/
│   │   ├── main/              # 主进程
│   │   ├── heartbeat/         # 心跳工作进程
│   │   └── task/              # 任务工作进程
│   ├── include/               # 公共头文件
│   ├── tests/                 # 单元测试
│   ├── CMakeLists.txt
│   └── conanfile.txt
├── server/                    # Go Server 代码
│   ├── cmd/                   # 入口程序
│   │   └── server/
│   │       └── main.go
│   ├── internal/              # 私有代码
│   │   ├── modules/           # 业务模块
│   │   │   ├── auth/
│   │   │   ├── agent/
│   │   │   ├── task/
│   │   │   ├── monitor/
│   │   │   └── user/
│   │   └── pkg/               # 内部公共包
│   │       ├── config/
│   │       ├── database/
│   │       ├── logger/
│   │       └── middleware/
│   ├── api/                   # API 定义
│   │   └── openapi.yaml
│   ├── configs/               # 配置文件
│   ├── scripts/               # 脚本
│   └── go.mod
├── docs/                      # 文档
├── openspec/                  # OpenSpec 变更管理
└── deployments/               # 部署配置
    ├── docker/
    └── kubernetes/
```

### 字符编码

- 所有源文件使用 **UTF-8** 编码
- 源文件末尾保留一个空行
- 不使用 TAB 字符，使用空格缩进

---

## Go 编码规范 (Server)

### 代码格式

```go
// 使用 gofmt 和 goimports 格式化代码
// 安装: go install golang.org/x/tools/cmd/goimports@latest

// 运行: goimports -w .
```

### 命名规范

#### 包命名

```go
// 包名使用小写单词，不使用下划线或驼峰
package auth
package agent
package task

// 不推荐
package authModule
package task_service
```

#### 导出函数/方法

```go
// 导出函数使用 PascalCase
func CreateAgent(ctx context.Context, req *CreateAgentRequest) (*Agent, error) {
    // ...
}

// 私有函数使用 camelCase
func validateToken(token string) error {
    // ...
}
```

#### 接口命名

```go
// 接口名使用动词或名词 + er 后缀
type AgentRepository interface {
    Create(ctx context.Context, agent *Agent) error
    GetByID(ctx context.Context, id string) (*Agent, error)
    Update(ctx context.Context, agent *Agent) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter *AgentFilter) ([]*Agent, error)
}

// 单方法接口使用 -er 后缀
type TaskExecutor interface {
    Execute(ctx context.Context, task *Task) error
}
```

#### 常量命名

```go
// 导出常量使用 PascalCase
const (
    StatusPending   = "pending"
    StatusRunning   = "running"
    StatusSuccess   = "success"
    StatusFailed    = "failed"
    StatusTimeout   = "timeout"
    StatusCancelled = "cancelled"
)

// 私有常量使用 camelCase 或全大写下划线分隔
const (
    defaultTimeout     = 30 * time.Second
    maxRetryAttempts   = 3
    HEARTBEAT_INTERVAL = 1 * time.Second
)
```

### 错误处理

```go
// 错误处理规范

// 1. 不要忽略错误
// 错误
result, _ := someFunction()

// 正确
result, err := someFunction()
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}

// 2. 使用自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// 3. 错误包装使用 %w
if err != nil {
    return fmt.Errorf("failed to create agent: %w", err)
}

// 4. 使用 errors.Is 和 errors.As
if errors.Is(err, ErrNotFound) {
    // handle not found
}

var valErr *ValidationError
if errors.As(err, &valErr) {
    // handle validation error
}
```

### 结构体定义

```go
// 结构体定义顺序：导出字段在前，私有字段在后
type Agent struct {
    // 导出字段
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"uniqueIndex;size:100"`
    Status    string    `json:"status" gorm:"size:20;default:'offline'"`
    Token     string    `json:"-" gorm:"size:64"` // 敏感字段不序列化
    Metadata  JSONB     `json:"metadata" gorm:"type:jsonb"`
    Version   string    `json:"version" gorm:"size:20"`
    LastSeen  time.Time `json:"last_seen_at"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

    // 私有字段
    mu sync.RWMutex
}

// 使用构造函数创建实例
func NewAgent(name string) *Agent {
    return &Agent{
        ID:     uuid.New().String(),
        Name:   name,
        Status: StatusOffline,
        Token:  generateToken(),
    }
}
```

### HTTP Handler 规范

```go
// Handler 结构体模式
type AgentHandler struct {
    service   AgentService
    validator *validator.Validate
}

func NewAgentHandler(service AgentService) *AgentHandler {
    return &AgentHandler{
        service:   service,
        validator: validator.New(),
    }
}

// 请求/响应使用独立的结构体
type CreateAgentRequest struct {
    Name     string          `json:"name" validate:"required,min=1,max=100"`
    Metadata json.RawMessage `json:"metadata" validate:"omitempty"`
}

type AgentResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Status    string    `json:"status"`
    Version   string    `json:"version"`
    LastSeen  time.Time `json:"last_seen_at"`
    CreatedAt time.Time `json:"created_at"`
}

// Handler 方法签名
func (h *AgentHandler) Create(c *gin.Context) {
    var req CreateAgentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse(err))
        return
    }

    if err := h.validator.Struct(req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse(err))
        return
    }

    agent, err := h.service.Create(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse(err))
        return
    }

    c.JSON(http.StatusCreated, AgentResponse{
        ID:        agent.ID,
        Name:      agent.Name,
        Status:    agent.Status,
        Version:   agent.Version,
        LastSeen:  agent.LastSeen,
        CreatedAt: agent.CreatedAt,
    })
}
```

### 并发规范

```go
// 使用 context 控制超时和取消
func (s *Service) DoWork(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // 使用 select 处理超时
    select {
    case <-ctx.Done():
        return ctx.Err()
    case result := <-s.workChan:
        return processResult(result)
    }
}

// 使用 sync.WaitGroup 等待 goroutine
func (s *Service) ProcessBatch(items []Item) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(items))

    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            if err := s.process(i); err != nil {
                errChan <- err
            }
        }(item)
    }

    wg.Wait()
    close(errChan)

    // 收集错误
    var errs []error
    for err := range errChan {
        errs = append(errs, err)
    }

    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

### 测试规范

```go
// 测试文件命名: <filename>_test.go
// 测试函数命名: Test<FunctionName>

package auth_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreateAgent(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Arrange
        ctx := context.Background()
        req := &CreateAgentRequest{Name: "test-agent"}

        // Act
        agent, err := service.Create(ctx, req)

        // Assert
        require.NoError(t, err)
        assert.NotEmpty(t, agent.ID)
        assert.Equal(t, "test-agent", agent.Name)
    })

    t.Run("duplicate name", func(t *testing.T) {
        // Test duplicate name scenario
    })
}

// 表驱动测试
func TestValidateStatus(t *testing.T) {
    tests := []struct {
        name    string
        status  string
        wantErr bool
    }{
        {"valid pending", StatusPending, false},
        {"valid running", StatusRunning, false},
        {"invalid status", "invalid", true},
        {"empty status", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateStatus(tt.status)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

## C++ 编码规范 (Agent)

### 代码格式

```cpp
// 使用 clang-format 格式化代码
// .clang-format 配置文件放在项目根目录
```

创建 `.clang-format` 文件：

```yaml
BasedOnStyle: Google
Language: Cpp
Standard: c++17
IndentWidth: 4
ColumnLimit: 100
BreakBeforeBraces: Attach
AllowShortFunctionsOnASingleLine: Inline
AllowShortIfStatementsOnASingleLine: Never
AllowShortLoopsOnASingleLine: false
```

### 命名规范

| 类型 | 命名风格 | 示例 |
|------|----------|------|
| 命名空间 | 小写蛇形 | `agent::websocket_client` |
| 类/结构体 | PascalCase | `WebSocketClient`, `TaskExecutor` |
| 函数/方法 | PascalCase | `Connect()`, `SendMessage()` |
| 成员变量 | 下划线后缀 | `agent_id_`, `is_connected_` |
| 局部变量 | 小写蛇形 | `result`, `error_code` |
| 常量 | k前缀+PascalCase | `kDefaultTimeout`, `kMaxRetryCount` |
| 枚举 | PascalCase | `enum class TaskStatus` |
| 枚举值 | k前缀+PascalCase | `kPending`, `kRunning` |
| 宏 | 全大写下划线 | `AGENT_VERSION_MAJOR` |

### 类定义

```cpp
// 头文件: include/agent/websocket_client.hpp
#ifndef AGENT_WEBSOCKET_CLIENT_HPP
#define AGENT_WEBSOCKET_CLIENT_HPP

#include <memory>
#include <string>
#include <functional>
#include <boost/asio.hpp>

namespace agent {

// 前向声明
class Message;

// 连接状态枚举
enum class ConnectionState {
    kDisconnected,
    kConnecting,
    kConnected,
    kAuthenticating,
    kAuthenticated
};

// 回调类型定义
using MessageCallback = std::function<void(const Message&)>;
using ErrorCallback = std::function<void(const std::string&)>;

// WebSocket 客户端类
class WebSocketClient {
public:
    // 构造和析构
    WebSocketClient(boost::asio::io_context& io_context);
    ~WebSocketClient();

    // 禁止拷贝
    WebSocketClient(const WebSocketClient&) = delete;
    WebSocketClient& operator=(const WebSocketClient&) = delete;

    // 允许移动
    WebSocketClient(WebSocketClient&&) noexcept = default;
    WebSocketClient& operator=(WebSocketClient&&) noexcept = default;

    // 公共方法
    void Connect(const std::string& url);
    void Disconnect();
    void SendMessage(const Message& message);

    // 设置回调
    void SetMessageCallback(MessageCallback callback);
    void SetErrorCallback(ErrorCallback callback);

    // 获取状态
    ConnectionState GetState() const { return state_; }
    bool IsConnected() const { return state_ == ConnectionState::kAuthenticated; }

private:
    // 私有方法
    void OnConnect(const boost::system::error_code& ec);
    void OnHandshake(const boost::system::error_code& ec);
    void DoRead();
    void HandleMessage(const std::string& data);

    // 成员变量（下划线后缀）
    boost::asio::io_context& io_context_;
    std::unique_ptr<class WebSocketImpl> impl_;
    ConnectionState state_ = ConnectionState::kDisconnected;
    MessageCallback message_callback_;
    ErrorCallback error_callback_;
};

}  // namespace agent

#endif  // AGENT_WEBSOCKET_CLIENT_HPP
```

### 源文件实现

```cpp
// 源文件: src/main/websocket_client.cpp
#include "agent/websocket_client.hpp"
#include "agent/message.hpp"
#include <boost/beast/core.hpp>
#include <boost/beast/websocket.hpp>
#include <nlohmann/json.hpp>

namespace agent {

namespace beast = boost::beast;
namespace websocket = beast::websocket;

// 内部实现类（PIMPL 模式）
class WebSocketImpl {
public:
    websocket::stream<beast::tcp_stream> ws;

    explicit WebSocketImpl(boost::asio::io_context& io_context)
        : ws(boost::asio::make_strand(io_context)) {}
};

// 常量定义
constexpr auto kDefaultTimeout = std::chrono::seconds(30);
constexpr auto kReconnectBaseDelay = std::chrono::seconds(5);

WebSocketClient::WebSocketClient(boost::asio::io_context& io_context)
    : io_context_(io_context)
    , impl_(std::make_unique<WebSocketImpl>(io_context)) {
}

WebSocketClient::~WebSocketClient() {
    Disconnect();
}

void WebSocketClient::Connect(const std::string& url) {
    state_ = ConnectionState::kConnecting;

    // 解析 URL 并连接
    // ... 实现代码
}

void WebSocketClient::Disconnect() {
    if (state_ == ConnectionState::kDisconnected) {
        return;
    }

    beast::error_code ec;
    impl_->ws.close(websocket::close_code::normal, ec);
    state_ = ConnectionState::kDisconnected;

    // 忽略关闭错误
    (void)ec;
}

void WebSocketClient::SendMessage(const Message& message) {
    if (!IsConnected()) {
        if (error_callback_) {
            error_callback_("Not connected");
        }
        return;
    }

    nlohmann::json j = message;
    impl_->ws.write(boost::asio::buffer(j.dump()));
}

void WebSocketClient::SetMessageCallback(MessageCallback callback) {
    message_callback_ = std::move(callback);
}

void WebSocketClient::SetErrorCallback(ErrorCallback callback) {
    error_callback_ = std::move(callback);
}

}  // namespace agent
```

### 错误处理

```cpp
// 使用 std::expected (C++23) 或自定义 Result 类型

// 自定义 Result 类型 (C++17)
template<typename T>
class Result {
public:
    // 成功构造
    static Result Success(T value) {
        return Result(std::move(value), {});
    }

    // 失败构造
    static Result Failure(std::string error) {
        return Result({}, std::move(error));
    }

    bool IsOk() const { return error_.empty(); }
    bool IsErr() const { return !error_.empty(); }

    const T& Value() const& { return value_; }
    T&& Value() && { return std::move(value_); }

    const std::string& Error() const& { return error_; }

private:
    Result(T value, std::string error)
        : value_(std::move(value))
        , error_(std::move(error)) {}

    T value_;
    std::string error_;
};

// 使用示例
Result<Task> TaskExecutor::Execute(const Command& cmd) {
    if (!ValidateCommand(cmd)) {
        return Result<Task>::Failure("Invalid command");
    }

    auto task = CreateTask(cmd);
    auto result = RunProcess(task);

    if (!result.IsOk()) {
        return Result<Task>::Failure(result.Error());
    }

    return Result<Task>::Success(std::move(task));
}
```

### 日志规范

```cpp
// 使用结构化日志
// 日志级别: TRACE, DEBUG, INFO, WARN, ERROR, FATAL

// 日志宏定义
#define LOG_TRACE(logger) SPDLOG_LOGGER_TRACE(logger)
#define LOG_DEBUG(logger) SPDLOG_LOGGER_DEBUG(logger)
#define LOG_INFO(logger)  SPDLOG_LOGGER_INFO(logger)
#define LOG_WARN(logger)  SPDLOG_LOGGER_WARN(logger)
#define LOG_ERROR(logger) SPDLOG_LOGGER_ERROR(logger)
#define LOG_FATAL(logger) SPDLOG_LOGGER_CRITICAL(logger)

// 使用示例
void WebSocketClient::HandleMessage(const std::string& data) {
    LOG_DEBUG(logger_) << "Received message: " << data;

    try {
        auto j = nlohmann::json::parse(data);
        std::string type = j["type"];

        LOG_INFO(logger_) << "Processing message type: " << type;

        if (message_callback_) {
            message_callback_(Message::FromJson(j));
        }
    } catch (const std::exception& e) {
        LOG_ERROR(logger_) << "Failed to parse message: " << e.what();
    }
}

// 日志格式
// [2024-01-15 10:30:45.123] [INFO] [websocket_client.cpp:123] Processing message type: heartbeat
```

### 配置文件

```yaml
# agent/config/agent.yaml
agent:
  id: ""
  token: ""
  server_url: "wss://server.example.com:443/api/v1/agent/ws"

connection:
  retry_interval: 5s
  max_retry_interval: 60s
  ping_interval: 10s
  pong_timeout: 5s

heartbeat:
  interval: 1s

metrics:
  interval: 1m

task:
  max_concurrent: 4
  queue_size: 100
  default_timeout: 5m

update:
  check_interval: 1h

logging:
  level: info
  file: "C:/ProgramData/AgentTeams/agent.log"
  max_size: 100MB
  max_files: 5
```

---

## 数据库规范

### 表命名

- 使用小写蛇形命名：`agents`, `tasks`, `agent_metrics`
- 多对多关联表：`agent_tags`, `user_roles`
- 使用复数形式

### 字段命名

- 使用小写蛇形命名：`agent_id`, `created_at`, `last_seen_at`
- 布尔字段使用 `is_` 前缀：`is_deleted`, `is_active`
- 时间字段使用 `_at` 后缀：`created_at`, `updated_at`, `deleted_at`

### 索引命名

```sql
-- 主键: pk_<table>
CONSTRAINT pk_agents PRIMARY KEY (id)

-- 唯一索引: uq_<table>_<columns>
CREATE UNIQUE INDEX uq_agents_name ON agents(name);

-- 普通索引: ix_<table>_<columns>
CREATE INDEX ix_agents_status ON agents(status);
CREATE INDEX ix_tasks_agent_created ON tasks(agent_id, created_at);

-- 外键: fk_<table>_<referenced_table>
CONSTRAINT fk_tasks_agent FOREIGN KEY (agent_id) REFERENCES agents(id)
```

### 迁移文件

```sql
-- 文件名: 001_create_agents.sql

-- +migrate Up
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    token VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'offline',
    version VARCHAR(20) NOT NULL DEFAULT '0.0.0',
    metadata JSONB DEFAULT '{}',
    last_seen_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_agents_name UNIQUE (name)
);

CREATE INDEX ix_agents_status ON agents(status);
CREATE INDEX ix_agents_deleted ON agents(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS agents;
```

---

## API 设计规范

### RESTful 端点

```
# 资源命名使用复数名词

# Agent 管理
POST   /api/v1/agents           # 创建 Agent
GET    /api/v1/agents           # 列出 Agents
GET    /api/v1/agents/:id       # 获取 Agent 详情
PUT    /api/v1/agents/:id       # 更新 Agent
DELETE /api/v1/agents/:id       # 删除 Agent
PUT    /api/v1/agents/:id/status # 更新状态

# Task 管理
POST   /api/v1/tasks            # 创建 Task
POST   /api/v1/tasks/batch      # 批量创建
GET    /api/v1/tasks            # 列出 Tasks
GET    /api/v1/tasks/:id        # 获取 Task 详情
DELETE /api/v1/tasks/:id        # 取消 Task

# WebSocket
WS     /api/v1/agent/ws         # Agent 连接端点
```

### 请求/响应格式

```json
// 成功响应
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "agent-001",
        "name": "web-server-01",
        "status": "online"
    }
}

// 分页响应
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [...],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 100,
            "total_pages": 5
        }
    }
}

// 错误响应
{
    "code": 10001,
    "message": "Agent not found",
    "data": null
}
```

### WebSocket 消息格式

```json
// 心跳
{
    "type": "heartbeat",
    "agent_id": "agent-001",
    "timestamp": "2024-01-15T10:30:45.123Z"
}

// 命令下发
{
    "type": "command",
    "id": "cmd-001",
    "command_type": "exec_shell",
    "params": {
        "shell": "cmd.exe",
        "command": "dir C:\\"
    },
    "timeout": 300
}

// 命令结果
{
    "type": "result",
    "id": "cmd-001",
    "status": "success",
    "exit_code": 0,
    "output": "...",
    "duration": 1.5
}

// 指标上报
{
    "type": "metrics",
    "agent_id": "agent-001",
    "timestamp": "2024-01-15T10:30:45.123Z",
    "data": {
        "cpu_usage": 45.2,
        "memory": {
            "total": 17179869184,
            "used": 8589934592,
            "percent": 50.0
        },
        "disk": {
            "total": 512110190592,
            "used": 256055095296,
            "percent": 50.0
        },
        "uptime": 86400
    }
}
```

### 错误码定义

| 范围 | 类别 |
|------|------|
| 0 | 成功 |
| 10000-10999 | 通用错误 |
| 10001 | 资源不存在 |
| 10002 | 参数无效 |
| 10003 | 权限不足 |
| 20000-20999 | 认证错误 |
| 20001 | 未认证 |
| 20002 | Token 过期 |
| 20003 | 凭证无效 |
| 30000-30999 | Agent 错误 |
| 30001 | Agent 离线 |
| 30002 | Agent 忙碌 |
| 40000-40999 | Task 错误 |
| 40001 | 任务超时 |
| 40002 | 任务失败 |
| 50000-50999 | 系统错误 |
| 50001 | 数据库错误 |
| 50002 | 内部错误 |

---

## Git 提交规范

### 分支命名

```
main                    # 主分支
develop                 # 开发分支
feature/<name>          # 功能分支: feature/agent-websocket
bugfix/<name>           # Bug 修复: bugfix/heartbeat-timeout
hotfix/<name>           # 紧急修复: hotfix/security-patch
release/<version>       # 发布分支: release/v1.0.0
```

### 提交消息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

#### Type 类型

| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响逻辑） |
| `refactor` | 重构 |
| `test` | 测试相关 |
| `chore` | 构建/工具相关 |
| `perf` | 性能优化 |

#### Scope 范围

| Scope | 说明 |
|-------|------|
| `agent` | Agent 端代码 |
| `server` | Server 端代码 |
| `api` | API 相关 |
| `auth` | 认证模块 |
| `task` | 任务模块 |
| `monitor` | 监控模块 |
| `db` | 数据库相关 |
| `deploy` | 部署相关 |

#### 示例

```bash
# 功能开发
feat(agent): implement WebSocket connection with auto-reconnect

- Add WebSocketClient class with TLS support
- Implement exponential backoff retry
- Add Challenge-Response authentication

Closes #123

# Bug 修复
fix(server): correct heartbeat timeout calculation

The heartbeat timeout was using milliseconds instead of seconds,
causing false offline detection.

Fixes #456

# 文档更新
docs(api): add OpenAPI specification for agent endpoints

# 重构
refactor(task): extract command execution to separate class
```

---

## 代码审查清单

### Go 代码审查

- [ ] 错误处理是否完整
- [ ] 是否有 goroutine 泄漏风险
- [ ] context 是否正确传递
- [ ] 并发访问是否使用锁保护
- [ ] 敏感信息是否正确处理
- [ ] 是否有单元测试

### C++ 代码审查

- [ ] 内存管理是否正确（无泄漏）
- [ ] 是否使用 RAII
- [ ] 异常安全是否考虑
- [ ] 线程安全是否保证
- [ ] 是否有适当的日志
- [ ] 是否有单元测试

---

## 工具配置

### Go 工具

```bash
# 安装工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest

# 运行检查
golangci-lint run ./...
goimports -w .
go test -race -cover ./...
```

### C++ 工具

```bash
# clang-format
clang-format -i src/**/*.cpp include/**/*.hpp

# clang-tidy
clang-tidy src/**/*.cpp -- -std=c++17

# CMake + Conan
mkdir build && cd build
conan install .. --build=missing
cmake ..
cmake --build .
ctest
```
