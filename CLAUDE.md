# CLAUDE.md

AgentTeams - 企业级 Windows 机器管理系统

- **Server (Go)**: REST API + WebSocket 网关
- **Agent (C++)**: Windows 客户端
- **Web (React)**: 管理后台前端

## 核心原则
1. 不要试图欺骗我
2. 不要用mock数据通过测试


## 规范文档

- [编码规范](docs/coding-standards.md) - **必须遵循**
- [设计文档](docs/design.md)
- [Git 规范](docs/git-conventions.md)

## 关键命令

```bash
# Server
cd server && golangci-lint run

# Agent
cd agent && mkdir build && cd build && conan install .. --build=missing && cmake .. && cmake --build .

# Web
cd web && npm install && npm run lint && npm run build
```

## Git Commit

```
<type>(<scope>): <subject>
Types: feat, fix, docs, refactor, test, chore
Scopes: agent, server, web, api, auth, task, monitor
```

