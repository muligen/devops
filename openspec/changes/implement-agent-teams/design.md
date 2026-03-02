## Context

本项目从零开始构建 AgentTeams 系统，包含：
- **Windows Agent (C++)**：运行在目标机器上的客户端
- **Server (Go)**：管理端，提供 WebSocket 网关和 REST API

详细设计文档已存在于 `docs/design.md`，本文档聚焦实现层面的技术决策。

**约束：**
- Agent 必须运行在 Windows 10/11 和 Windows Server 2016+
- Agent 需要以 Windows Service 方式运行
- 需要支持大规模 Agent 连接（设计目标：10000+ 并发）

## Goals / Non-Goals

**Goals:**
- 实现完整的 Agent-Server 通信链路
- 支持命令下发和并发执行
- 支持心跳和机器数据上报
- 支持 Agent 自更新
- 提供基础的管理 API

**Non-Goals:**
- 前端 UI（后续迭代）
- Linux/macOS 支持（后续迭代）
- 高级告警规则（后续迭代）
- 多租户支持（后续迭代）

## Decisions

### 1. 进程架构：主进程 + 工作进程

**决策：** Agent 采用主进程 + 两个工作进程的架构

```
agent.exe (主进程)
    ├── worker_heartbeat.exe (工作进程1: 心跳/数据上报)
    └── worker_task.exe (工作进程2: 任务执行)
```

**理由：**
- 主进程极简，负责进程管理和自更新，几乎不需要更新
- 工作进程独立，崩溃不影响其他进程
- 工作进程可独立更新，实现热更新

**备选方案：**
- 单进程多线程：更简单但无法独立更新模块
- 多进程 + 共享内存：复杂度高，收益不大

### 2. 通信协议：WebSocket over TLS

**决策：** Agent 与 Server 之间使用 WebSocket over TLS (WSS) 通信

**理由：**
- 双向实时通信，Server 可主动推送命令
- 穿透性好，走 443 端口
- 成熟的库支持（gorilla/websocket, Boost.Beast）

**备选方案：**
- gRPC：双向流支持好，但需要额外配置 HTTP/2
- 自定义 TCP：灵活但需要自己实现心跳、重连等逻辑

### 3. 消息格式：JSON

**决策：** 使用 JSON 作为消息格式

**理由：**
- 可读性好，调试方便
- 跨语言支持完善
- 性能够用（非高频交易场景）

**备选方案：**
- Protobuf：更高效但调试复杂
- MessagePack：折中方案但生态不如 JSON

### 4. 并发模型：线程池 + 进程池

**决策：** 任务执行使用线程池调度，每个命令独立子进程

**理由：**
- 线程池管理并发数量，防止资源耗尽
- 子进程隔离，命令崩溃不影响 Agent
- 可对子进程设置超时、资源限制

### 5. Server 架构：模块化单体

**决策：** Server 采用模块化单体架构，按领域划分模块

**理由：**
- 初期开发效率高
- 部署简单，单二进制
- 模块边界清晰，未来可拆微服务

**备选方案：**
- 微服务：初期复杂度高，运维成本大
- 单体无模块：后期难以维护

### 6. 数据库：PostgreSQL

**决策：** 使用 PostgreSQL 作为主数据库

**理由：**
- JSONB 支持，存储动态数据（命令参数、结果）
- 成熟稳定，社区活跃
- 支持 CTE、窗口函数等高级特性

### 7. 消息队列：RabbitMQ

**决策：** 使用 RabbitMQ 作为消息队列

**理由：**
- 任务队列场景成熟
- 支持延迟队列、死信队列
- 管理界面友好

**备选方案：**
- NATS：更轻量但功能较少
- Redis Streams：简单但不适合持久化任务

### 8. Agent 认证：Token + Challenge-Response

**决策：** 使用 Token + Challenge-Response 认证

**流程：**
1. Agent 连接时发送 AgentID
2. Server 返回随机 nonce
3. Agent 返回 HMAC(token, nonce)
4. Server 验证后建立会话

**理由：**
- 避免明文传输 Token
- 防止重放攻击
- 实现简单

## Risks / Trade-offs

### Risk 1: 大规模连接的性能瓶颈

**风险：** 单台 Server 可能无法支撑 10000+ WebSocket 连接

**缓解措施：**
- 使用连接池、异步 IO
- Redis 存储会话，支持水平扩展
- 设计上支持多实例部署

### Risk 2: Agent 端更新失败

**风险：** 自更新过程中断导致 Agent 不可用

**缓解措施：**
- 主进程几乎不更新
- 更新前备份旧版本
- 更新失败自动回滚
- 定期心跳检测，失败告警

### Risk 3: 网络不稳定导致命令丢失

**风险：** 网络断开时命令可能丢失

**缓解措施：**
- Server 端持久化待执行命令
- Agent 重连后同步未完成任务
- 命令执行结果持久化，支持重试

### Risk 4: 并发任务资源耗尽

**风险：** 大量并发任务耗尽 Agent 资源

**缓解措施：**
- 可配置最大并发数
- 资源监控，超限拒绝新任务
- 任务队列背压控制

## Migration Plan

本项目为新项目，无需迁移。

**部署顺序：**
1. 部署基础设施（PostgreSQL、Redis、RabbitMQ、MinIO）
2. 部署 Server
3. 打包 Agent 安装程序
4. 在目标机器安装 Agent
5. 验证连接和功能

**回滚策略：**
- Server：版本化部署，支持快速回滚
- Agent：保留旧版本，更新失败自动回滚

## Open Questions

1. **Agent 分发方式：** 如何大规模分发 Agent 安装包？GPO？脚本？
2. **日志收集：** Agent 日志是否需要上报到 Server？还是本地存储？
3. **证书管理：** TLS 证书如何管理和更新？
4. **监控集成：** 是否需要集成现有的监控系统（Prometheus、Zabbix 等）？

这些问题将在后续迭代中解决。
