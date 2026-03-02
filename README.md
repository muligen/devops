# AgentTeams

[![CI](https://github.com/muligen/devops/actions/workflows/ci.yml/badge.svg)](https://github.com/muligen/devops/actions/workflows/ci.yml)
[![Release](https://github.com/muligen/devops/actions/workflows/release.yml/badge.svg)](https://github.com/muligen/devops/actions/workflows/release.yml)

企业级 Windows 机器管理系统，支持远程命令执行、监控和自动更新。

## 功能特性

- 🔐 **认证授权**: JWT Token 认证，基于角色的访问控制
- 🖥️ **Agent 管理**: 自动注册、状态监控、在线/离线管理
- ⚡ **任务执行**: 远程命令执行，支持 Shell 命令、脚本
- 📊 **系统监控**: CPU、内存、磁盘实时监控，指标存储
- 🔔 **告警通知**: 自定义告警规则，Webhook/邮件通知
- 🔄 **自动更新**: Agent 版本管理，自动下载更新
- 📝 **审计日志**: 操作记录，追踪用户行为

## 系统架构

```
┌─────────────┐     WebSocket     ┌─────────────┐
│   Agent     │◄──────────────────►│   Server    │
│  (C++)      │     REST API       │   (Go)      │
└─────────────┘                    └──────┬──────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
              ┌─────▼─────┐        ┌──────▼──────┐      ┌───────▼───────┐
              │ PostgreSQL│        │    Redis    │      │   RabbitMQ    │
              │  (数据存储) │        │  (会话缓存)  │      │   (消息队列)   │
              └───────────┘        └─────────────┘      └───────────────┘
```

## 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| Agent | C++20, Boost.Beast, OpenSSL | MSVC 2022 |
| Server | Go, Gin, GORM, gorilla/websocket | Go 1.21+ |
| 数据库 | PostgreSQL | 15+ |
| 缓存 | Redis | 7+ |
| 消息队列 | RabbitMQ | 3.12+ |
| 对象存储 | MinIO | Latest |

## 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.21+ (开发)
- MSVC 2022 / Conan (Agent 开发)

### 启动服务

```bash
# 克隆仓库
git clone https://github.com/muligen/devops.git
cd devops

# 启动依赖服务
cd deployments/docker
docker-compose up -d postgres redis rabbitmq minio

# 启动 Server
cd ../../server
go run ./cmd/server -config configs/config.yaml

# 访问健康检查
curl http://localhost:8080/health
```

### 构建 Agent

```bash
cd agent
mkdir build && cd build
conan install .. --build=missing
cmake ..
cmake --build . --config Release
```

## 项目结构

```
AgentTeams/
├── agent/                 # C++ Agent
│   ├── src/
│   │   ├── main/          # 主进程
│   │   ├── heartbeat/     # 心跳 Worker
│   │   └── task/          # 任务 Worker
│   ├── include/           # 头文件
│   └── tests/             # 单元测试
├── server/                # Go Server
│   ├── cmd/               # 入口
│   ├── internal/
│   │   ├── modules/       # 业务模块
│   │   └── pkg/           # 公共包
│   └── configs/           # 配置
├── docs/                  # 文档
├── deployments/           # 部署配置
└── .github/               # GitHub Actions
```

## API 文档

详细的 API 文档见 [OpenAPI 规范](api/openapi.yaml)。

### 主要端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/auth/login` | POST | 用户登录 |
| `/api/v1/agents` | GET/POST | Agent 列表/注册 |
| `/api/v1/tasks` | GET/POST | 任务列表/创建 |
| `/api/v1/agents/:id/metrics` | GET | Agent 指标 |
| `/api/v1/agent/ws` | WS | Agent WebSocket |

## 开发指南

- [编码规范](docs/coding-standards.md)
- [Git 提交规范](docs/git-conventions.md)
- [设计文档](docs/design.md)

## 许可证

MIT License
