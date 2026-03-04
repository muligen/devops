## Context

AgentTeams 项目当前有 14 个 spec 定义了系统行为，但缺乏自动化测试验证。每次代码变更后，需要手动验证功能完整性，效率低且容易遗漏回归问题。

**当前状态**:
- Server 有少量单元测试，无 E2E 测试
- Agent 无测试
- 测试环境需要手动搭建

**约束**:
- 测试需在隔离环境中运行，不污染开发环境
- 测试需快速执行，支持本地和 CI 运行
- 测试需覆盖 WebSocket 长连接场景

## Goals / Non-Goals

**Goals:**
- 建立可重复使用的 E2E 测试框架
- 测试环境一键启动（Docker Compose）
- 覆盖所有 14 个 spec 的核心场景
- CI 集成，每次 PR 自动运行测试
- 生成覆盖率报告

**Non-Goals:**
- Agent 端的 E2E 测试（Agent 是 C++，暂不在范围内）
- 性能/压力测试（后续单独建设）
- 前端 UI 测试（后续单独建设）

## Decisions

### 1. 测试框架选择

**选择**: Go 原生 testing + testify + testcontainers-go

**理由**:
- 与 Server 代码库一致，无需额外语言
- testcontainers-go 支持动态启动 PostgreSQL、Redis、RabbitMQ
- testify 提供丰富的断言和 mock 功能

**替代方案**:
- ❌ Jest + supertest: 需要额外 Node.js 环境
- ❌ Postman/Newman: 不适合 WebSocket 测试
- ❌ Cypress: 偏向前端，后端 API 测试不便

### 2. 测试环境策略

**选择**: testcontainers-go 动态容器

**理由**:
- 每次测试启动独立容器，隔离性好
- 支持本地和 CI 环境一致运行
- 无需预装 Docker Compose

**替代方案**:
- ❌ Docker Compose 静态环境: 需要手动管理生命周期
- ❌ 共享测试环境: 并发测试会冲突
- ❌ Mock 外部依赖: 无法验证真实集成

### 3. WebSocket 测试策略

**选择**: gorilla/websocket 客户端直连测试

**理由**:
- 可复用 Server 已有依赖
- 支持完整认证流程测试
- 可验证消息格式和时序

### 4. 测试数据管理

**选择**: Fixtures + Factory 模式

**理由**:
- Fixtures 提供基础数据（用户、Agent）
- Factory 按需创建测试数据
- 每个测试套件独立管理数据生命周期

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| testcontainers 启动慢 | 使用 parallel pool，复用容器 |
| WebSocket 测试不稳定 | 设置合理超时，添加重试机制 |
| CI 资源消耗大 | 使用 GitHub Actions 缓存，限制并发 |
| 测试覆盖不完整 | 优先覆盖核心路径，逐步完善 |

## Migration Plan

1. **Phase 1**: 搭建测试框架和环境
2. **Phase 2**: 为每个 spec 添加测试用例
3. **Phase 3**: CI 集成和覆盖率报告

**回滚策略**: 测试代码独立，可随时禁用或删除
