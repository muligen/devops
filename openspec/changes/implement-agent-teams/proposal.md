## Why

企业需要一种集中化管理 Windows 机器的解决方案，用于远程执行命令、监控机器状态、批量管理任务。当前缺乏统一的 Agent 系统，导致运维效率低下、手动操作风险高、无法实时掌握机器状态。

本项目构建 AgentTeams 系统，包含 Windows Agent 客户端和 Server 管理端，实现远程命令执行、实时监控、自动更新等核心能力。

## What Changes

### Windows Agent (C++)

- **新增** 主进程架构：进程生命周期管理、自更新逻辑、Windows Service 入口
- **新增** 工作进程1（心跳进程）：心跳上报（1秒）、机器数据上报（1分钟）
- **新增** 工作进程2（任务进程）：WebSocket 连接、命令接收、并发执行、结果回传
- **新增** 内置命令：`init_machine`（初始化机器）、`clean_disk`（清理磁盘）、`exec_shell`（执行 Shell）
- **新增** 自更新机制：版本检测、文件下载、签名校验、热更新

### Server (Go)

- **新增** 模块化单体架构：Agent/Task/Monitor/Update/User 模块
- **新增** WebSocket 网关：Agent 连接管理、会话保持、认证鉴权
- **新增** REST API：Agent 管理、任务管理、版本管理、用户管理
- **新增** 事件驱动系统：RabbitMQ 消息队列、异步任务处理
- **新增** 数据存储：PostgreSQL（持久化）、Redis（缓存/会话）、MinIO（对象存储）

### 基础设施

- **新增** Docker Compose 开发环境
- **新增** 数据库迁移脚本
- **新增** API 文档（OpenAPI 规范）

## Capabilities

### New Capabilities

- `agent-connection`: Agent 与 Server 的 WebSocket 连接、认证、会话管理
- `agent-heartbeat`: 心跳上报、在线状态检测、断线重连
- `agent-metrics`: 机器数据采集与上报（CPU、内存、磁盘等）
- `task-execution`: 命令下发、并发执行、结果回传、超时处理
- `task-queue`: 任务队列、优先级调度、并发控制
- `auto-update`: Agent 自更新检测、版本比对、文件下载、签名校验、热更新
- `agent-management`: Agent 注册、查询、状态管理、删除
- `user-auth`: 用户登录、Token 管理、权限控制
- `monitoring-dashboard`: 实时监控、历史数据查询、告警规则

### Modified Capabilities

无（新项目）

## Impact

### 代码结构

```
AgentTeams/
├── agent/                 # Windows Agent (C++)
│   ├── src/
│   │   ├── main/          # 主进程
│   │   ├── heartbeat/     # 工作进程1
│   │   └── task/          # 工作进程2
│   ├── CMakeLists.txt
│   └── conanfile.txt
├── server/                # Server (Go)
│   ├── cmd/
│   ├── internal/
│   │   ├── modules/
│   │   └── pkg/
│   └── api/
├── docs/                  # 文档
│   └── design.md
└── openspec/              # OpenSpec 变更管理
```

### 技术依赖

**Agent (C++):**
- C++17
- Conan (依赖管理)
- CMake + MSVC
- WebSocket 客户端库
- JSON 解析库
- YAML 配置库

**Server (Go):**
- Go 1.21+
- Gin (Web 框架)
- GORM (ORM)
- gorilla/websocket
- go-redis
- RabbitMQ Go client
- MinIO Go client

### 外部系统

- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+
- MinIO

### 部署影响

- Agent: 安装为 Windows Service，端口出站 443 (WSS)
- Server: Kubernetes 部署，入站 443/80
