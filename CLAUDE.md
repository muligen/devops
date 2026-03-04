# CLAUDE.md

AgentTeams - 企业级 Windows 机器管理系统

- **Server (Go)**: REST API + WebSocket 网关
- **Agent (C++)**: Windows 客户端

## 规范文档

- [编码规范](docs/coding-standards.md) - **必须遵循**
- [设计文档](docs/design.md)
- [Git 规范](docs/git-conventions.md)

## 关键命令

```bash
# Server
cd server && go test ./... && golangci-lint run

# Agent
cd agent && mkdir build && cd build && conan install .. --build=missing && cmake .. && cmake --build . && ctest
```

## Git Commit

```
<type>(<scope>): <subject>
Types: feat, fix, docs, refactor, test, chore
Scopes: agent, server, api, auth, task, monitor
```

## 当前任务

[implement-agent-teams](openspec/changes/implement-agent-teams/tasks.md)
