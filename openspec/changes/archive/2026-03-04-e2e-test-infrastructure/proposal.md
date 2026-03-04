## Why

当前项目已有 14 个 spec 定义了各模块行为，但缺少端到端测试验证。每次变更后无法自动验证系统完整性，导致回归风险高、手动测试成本大。需要建立自动化 E2E 测试体系，确保所有 spec 定义的行为都能被自动验证。

## What Changes

- 新增 E2E 测试框架，支持 Server API 和 WebSocket 测试
- 为所有 14 个现有 spec 构建端到端测试用例
- 搭建隔离的测试环境（数据库、缓存、消息队列）
- 集成到 CI/CD 流程，每次提交自动运行测试
- 生成测试覆盖率报告

## Capabilities

### New Capabilities

- `e2e-test-infrastructure`: 测试框架和环境基础设施，包括测试数据库、Redis、RabbitMQ 的 Docker Compose 配置，测试数据 fixtures，测试辅助工具
- `e2e-test-agent-connection`: Agent WebSocket 连接认证流程的端到端测试
- `e2e-test-agent-heartbeat`: Agent 心跳机制和状态同步的端到端测试
- `e2e-test-agent-management`: Agent 注册、查询、删除管理的端到端测试
- `e2e-test-agent-metrics`: Agent 指标采集和存储的端到端测试
- `e2e-test-task-execution`: 任务下发执行的端到端测试
- `e2e-test-task-queue`: 批量任务队列的端到端测试
- `e2e-test-user-auth`: 用户认证授权的端到端测试
- `e2e-test-auto-update`: Agent 自动更新的端到端测试
- `e2e-test-monitoring`: 监控仪表盘和实时推送的端到端测试

### Modified Capabilities

无。此变更为纯新增测试能力，不修改现有 spec 的需求定义。

## Impact

- **新增代码**:
  - `server/tests/e2e/` - E2E 测试用例目录
  - `server/tests/fixtures/` - 测试数据 fixtures
  - `deployments/test/` - 测试环境 Docker Compose
- **依赖变更**:
  - 新增 Go 测试依赖: `testcontainers-go`, `testify`
- **CI/CD**:
  - 更新 GitHub Actions 添加 E2E 测试步骤
- **配置文件**:
  - 新增 `server/tests/e2e/test_config.yaml`
