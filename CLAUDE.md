# CLAUDE.md

AgentTeams - 企业级 Windows 机器管理系统

## 项目架构

| 模块 | 技术栈 | 说明 |
|------|--------|------|
| **Server** | Go + Gin + GORM | REST API + WebSocket 网关 |
| **Agent** | C++ + Boost | Windows 客户端，采集系统指标 |
| **Web** | React + Vite + Ant Design | 管理后台前端 |

## 核心原则

1. **不要试图欺骗我**
2. **不要用 mock 数据通过测试** - 所有数据必须是真实采集的

## 规范文档

- [编码规范](docs/coding-standards.md) - **必须遵循**
- [设计文档](docs/design.md)
- [Git 规范](docs/git-conventions.md)

## 辅助文档

- **[TEST_ENV.md](TEST_ENV.md)** - 测试环境配置参考，包含服务端口、数据库连接、测试账号等
- **[bugs.md](bugs.md)** - Bug 记录与修复日志，新问题记录在此并跟踪状态

## 测试环境

```bash
# 生产环境容器 (docker-compose up -d)
- agentteams-postgres   (端口 5432)
- agentteams-redis      (端口 6379)
- agentteams-rabbitmq   (端口 5672, 15672)
- agentteams-minio      (端口 9000, 9001)
- agentteams-web        (端口 3000)

# 后端服务 (本地运行)
cd server && go build -o server.exe ./cmd/server && ./server.exe

# 测试账号
用户名: admin  密码: admin123
```

## 关键命令

```bash
# Server
cd server && golangci-lint run && go build -o server.exe ./cmd/server

# Agent
cd agent && mkdir build && cd build && conan install .. --build=missing && cmake .. && cmake --build .

# Web
cd web && npm install && npm run lint && npm run build
```

## Git Commit 规范

```
<type>(<scope>): <subject>
Types: feat, fix, docs, refactor, test, chore
Scopes: agent, server, web, api, auth, task, monitor
```