## 1. 测试基础设施

- [x] 1.1 添加 Go 测试依赖 (testcontainers-go, testify)
- [x] 1.2 创建测试配置文件 server/tests/e2e/test_config.yaml
- [x] 1.3 实现 testcontainers 初始化函数 (PostgreSQL, Redis, RabbitMQ)
- [x] 1.4 创建测试数据 fixtures (用户、Agent 基础数据)
- [x] 1.5 实现 test factory 函数 (创建 Agent、Task 等)

## 2. 用户认证测试 (e2e-test-user-auth)

- [x] 2.1 实现登录成功/失败测试用例
- [x] 2.2 实现 Token 验证测试用例 (有效/过期/刷新)
- [x] 2.3 实现权限控制测试用例 (Admin/Operator/Viewer)

## 3. Agent 连接测试 (e2e-test-agent-connection)

- [x] 3.1 实现 WebSocket 连接建立测试
- [x] 3.2 实现 HMAC challenge-response 认证测试
- [x] 3.3 实现连接断开处理测试

## 4. Agent 心跳测试 (e2e-test-agent-heartbeat)

- [x] 4.1 实现心跳发送和响应测试
- [x] 4.2 实现心跳超时检测测试

## 5. Agent 管理测试 (e2e-test-agent-management)

- [x] 5.1 实现 Agent 注册测试
- [x] 5.2 实现 Agent 查询测试 (单个/列表/筛选)
- [x] 5.3 实现 Agent 删除测试

## 6. Agent 指标测试 (e2e-test-agent-metrics)

- [x] 6.1 实现指标上报测试
- [x] 6.2 实现指标存储测试
- [x] 6.3 实现指标历史查询测试

## 7. 任务执行测试 (e2e-test-task-execution)

- [x] 7.1 实现任务创建测试
- [x] 7.2 实现任务执行测试 (Shell 命令)
- [x] 7.3 实现任务超时和重试测试
- [x] 7.4 实现任务状态跟踪测试

## 8. 任务队列测试 (e2e-test-task-queue)

- [x] 8.1 实现批量任务创建测试
- [x] 8.2 实现任务队列管理测试
- [x] 8.3 实现任务取消测试

## 9. 自动更新测试 (e2e-test-auto-update)

- [x] 9.1 实现版本检查测试
- [x] 9.2 实现更新包下载测试
- [x] 9.3 实现更新执行和回滚测试

## 10. 监控仪表盘测试 (e2e-test-monitoring)

- [x] 10.1 实现仪表盘数据 API 测试
- [x] 10.2 实现实时数据推送测试 (WebSocket)
- [x] 10.3 实现历史数据查询测试

## 11. CI 集成

- [x] 11.1 创建 GitHub Actions E2E 测试 workflow
- [x] 11.2 配置测试覆盖率报告生成
- [x] 11.3 添加测试失败通知
