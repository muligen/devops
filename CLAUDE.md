# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AgentTeams 是一个企业级 Windows 机器管理系统，包含：
- **Server (Go)**: 管理端，提供 REST API 和 WebSocket 网关
- **Agent (C++)**: Windows 客户端，支持远程命令执行、监控、自动更新

详细设计文档：
- [设计决策](docs/design.md)
- [编码规范](docs/coding-standards.md) - **必须遵循**
- [变更提案](openspec/changes/implement-agent-teams/proposal.md)

## Coding Standards

**重要**: 所有代码必须遵循 [编码规范](docs/coding-standards.md)。

### Go (Server) 规范摘要

```go
// 包名: 小写单词
package agent

// 导出函数: PascalCase
func CreateAgent(ctx context.Context, req *CreateAgentRequest) (*Agent, error)

// 私有函数: camelCase
func validateToken(token string) error

// 错误处理: 使用 %w 包装
return fmt.Errorf("failed to create agent: %w", err)

// 结构体: 导出字段在前，私有字段在后
type Agent struct {
    ID   string `json:"id"`
    name string // 私有字段
}
```

### C++ (Agent) 规范摘要

```cpp
// 类名: PascalCase
class WebSocketClient { ... };

// 函数名: PascalCase
void Connect(const std::string& url);

// 成员变量: 下划线后缀
std::string agent_id_;
bool is_connected_ = false;

// 常量: k前缀+PascalCase
constexpr auto kDefaultTimeout = std::chrono::seconds(30);

// 枚举: PascalCase, 枚举值 k前缀
enum class ConnectionState {
    kDisconnected,
    kConnected
};
```

## Project Structure

```
AgentTeams/
├── agent/                 # C++ Agent 代码
│   ├── src/
│   │   ├── main/          # 主进程
│   │   ├── heartbeat/     # 心跳工作进程
│   │   └── task/          # 任务工作进程
│   ├── include/           # 公共头文件
│   ├── tests/             # 单元测试
│   ├── CMakeLists.txt
│   └── conanfile.txt
├── server/                # Go Server 代码
│   ├── cmd/               # 入口程序
│   ├── internal/          # 私有代码
│   │   ├── modules/       # 业务模块
│   │   └── pkg/           # 内部公共包
│   ├── api/               # API 定义
│   └── go.mod
├── docs/                  # 文档
├── openspec/              # OpenSpec 变更管理
└── deployments/           # 部署配置
```

## Tech Stack

| 组件 | 技术 |
|------|------|
| Agent | C++17, Conan, CMake, MSVC, Boost.Beast |
| Server | Go 1.21+, Gin, GORM, gorilla/websocket |
| 数据库 | PostgreSQL 15+ |
| 缓存 | Redis 7+ |
| 消息队列 | RabbitMQ 3.12+ |
| 对象存储 | MinIO |

## Commands

### Go (Server)
```bash
cd server
go mod tidy
go test ./...
golangci-lint run
```

### C++ (Agent)
```bash
cd agent
mkdir build && cd build
conan install .. --build=missing
cmake ..
cmake --build .
ctest
```

## Git Commit Convention

详细规范见 [Git 提交规范](docs/git-conventions.md)

```
<type>(<scope>): <subject>

Types: feat, fix, docs, style, refactor, test, chore, perf, build, ci
Scopes: agent, server, api, auth, task, monitor, update, db, deploy
```

## Task Tracking

当前变更提案: [implement-agent-teams](openspec/changes/implement-agent-teams/)
- 17 个模块，184 个子任务
- 详见 [tasks.md](openspec/changes/implement-agent-teams/tasks.md)
