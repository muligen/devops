# Bugs 记录

## 2026-03-10 发现的问题

### Bug #1: Server Dashboard WebSocket 未启动

**严重程度**: 高

**问题描述**:
`server/cmd/server/main.go` 创建了 `DashboardWSHandler` 但没有调用 `Start()` 方法，导致：
- 不会周期性推送 dashboard 统计数据
- 不会订阅 RabbitMQ 事件 (agent online/offline/heartbeat)
- 前端 WebSocket 连接后收不到任何实时推送消息

**影响范围**:
- 前端仪表盘数据不更新
- Agent 状态变化无法实时推送到前端
- 系统指标无法实时显示

**修复方案**:
1. 修改 `setupRouter()` 返回 `dashboardWSHandler`
2. 在 `main()` 中调用 `dashboardWSHandler.Start(ctx)`
3. 在 shutdown 时调用 `dashboardWSHandler.Stop()`
4. 添加 `StatsProviderFunc` 适配器解决接口类型不匹配问题

**修改文件**:
- `server/cmd/server/main.go`
- `server/internal/modules/monitor/handler/dashboard_websocket.go`

**状态**: ✅ 已修复

---

### Bug #2: 前端 Agent 状态不更新实时数据

**严重程度**: 中

**问题描述**:
`web/src/pages/dashboard/index.tsx` 中：
- 通过 HTTP API 一次性获取 `agents` 数据
- WebSocket 收到的 `metrics` 消息存到了 `dashboardStore.metrics`
- 但 `AgentStatusGrid` 组件显示的是 `agents` 数组，没有使用 `metrics` 来更新 CPU/内存/磁盘数据

**影响范围**:
- Agent 卡片显示的 CPU/内存/磁盘使用率是静态的
- 用户无法看到实时系统指标变化

**修复方案**:
使用 `useMemo` 将 WebSocket 推送的 metrics 数据合并到 agents 列表中

**修改文件**:
- `web/src/pages/dashboard/index.tsx`

**状态**: ✅ 已修复

---

### Bug #3: Agent 命令处理未实现

**严重程度**: 高

**问题描述**:
`agent/src/main/service.cpp` 第 236-240 行，WebSocket 消息回调中只有 `TODO` 注释，没有实际处理命令的逻辑：

```cpp
g_context->ws_client->SetMessageCallback(
    [](const std::string& message) {
        LOG_DEBUG("WebSocket message received: {}", message.substr(0, 100));
        // TODO: Handle incoming messages (commands from server)
    });
```

**影响范围**:
- 任务执行功能完全不可用
- Server dispatch 的命令无法被 Agent 执行
- 所有任务永远停留在 pending 状态

**修复方案**:
需要实现完整的命令处理逻辑，包括：
- 解析命令类型 (exec_shell, clean_disk, init_machine)
- 将命令放入任务队列
- 执行命令并返回结果

**状态**: ✅ 已修复

**修复说明**:
1. 在 `service.cpp` 中添加了命令队列和 worker 线程
2. 实现了 WebSocket 消息回调中的命令解析和处理
3. 添加了 `SendCommandResult()` 函数将执行结果发送回 Server
4. 集成现有的命令执行器 (exec_shell, clean_disk, init_machine)
5. 修改 CMakeLists.txt 将命令执行器代码编译到 main agent 中
6. 添加命令执行器的 factory 函数前向声明

**修改文件**:
- `agent/src/main/service.cpp`
- `agent/src/main/CMakeLists.txt`

---

### Bug #4: 前端历史指标图表时间范围切换无效

**严重程度**: 中

**问题描述**:
前端 Agent 详情页面的历史指标图表，用户点击 "1小时/24小时/7天" 按钮时图表数据不会变化。

**根本原因**:
前后端 API 参数不匹配：

1. **前端** (`web/src/api/agents.ts:29-34`) 传递 `range` 参数：
   ```typescript
   async getMetrics(id: string, range: string = '1h'): Promise<unknown> {
     const response = await apiClient.get(`/agents/${id}/metrics`, {
       params: { range },  // 传递 range=1h/24h/7d
     })
   }
   ```

2. **后端** (`server/internal/modules/monitor/handler/handler.go:520-537`) 只识别 `start` 和 `end` 参数：
   ```go
   func parseTimeRange(c *gin.Context) (start, end time.Time) {
       end = time.Now()
       start = end.Add(-1 * time.Hour) // 默认: 最近1小时
       // 只解析 start 和 end 参数，完全忽略 range 参数
   }
   ```

**影响范围**:
- 用户点击 "24 小时" 或 "7 天" 按钮时，图表数据不会变化
- 始终显示最近1小时的数据（默认值）
- 用户无法查看不同时间范围的历史指标

**修复方案**:
有两种修复方式：

方案A（后端适配）：
- 在后端 `parseTimeRange()` 函数中添加对 `range` 参数的解析
- 支持 `1h`, `24h`, `7d` 等简写格式

方案B（前端适配）：
- 前端根据 `range` 参数计算 `start` 和 `end` 时间戳
- 传递 RFC3339 格式的 `start` 和 `end` 参数给后端

**修改文件**:
- 方案A: `server/internal/modules/monitor/handler/handler.go`
- 方案B: `web/src/api/agents.ts`

**采用方案**: A（后端适配）

**状态**: ✅ 已修复

---

### Bug #5: 历史指标图表不显示（React Hooks 规则违反）

**严重程度**: 高

**问题描述**:
前端 Agent 详情页面的历史指标图表完全不显示，浏览器控制台报错 `Minified React error #310`。

**根本原因**:
`web/src/pages/agents/[id]/index.tsx` 中 `useMemo` hook 被放在条件返回（`if (loading)` 和 `if (!agent)`）之后，违反了 React Hooks 规则：
- React hooks 必须在每次渲染时以相同的顺序调用
- 条件返回会导致 hooks 调用次数不一致

**修复方案**:
将 `useMemo` 移到条件返回之前，确保 hooks 调用顺序一致。

**修改文件**:
- `web/src/pages/agents/[id]/index.tsx`

**状态**: ✅ 已修复

---

### Bug #6: 后端时区问题导致时间范围查询不准确

**严重程度**: 高

**问题描述**:
切换历史指标时间范围后，返回的数据时间范围不正确，无法查询到真实的历史数据。

**根本原因**:
- 后端 `parseTimeRange()` 使用 `time.Now()` 返回本地时间（如 UTC+8 的 18:00）
- 数据库中 `collected_at` 字段存储的是 UTC 时间（如 10:00）
- 时间比较时没有统一时区，导致查询条件错误

**影响范围**:
- 7天时间范围查询不到历史数据
- 时间范围计算偏差 8 小时

**修复方案**:
在 `parseTimeRange()` 中使用 `time.Now().UTC()` 统一使用 UTC 时间。

**修改文件**:
- `server/internal/modules/monitor/handler/handler.go`

**状态**: ✅ 已修复

---

### Bug #7: 前端未传递 limit 导致数据被截断

**严重程度**: 中

**问题描述**:
切换 24小时/7天 时间范围后，图表显示的数据量与 1小时 相同，没有显示更多历史数据。

**根本原因**:
- 后端默认 `limit=100`
- 前端 `getMetrics()` 没有传递 limit 参数
- 大时间范围的数据被截断到 100 条

**修复方案**:
前端根据时间范围传递合适的 limit：
- 1h: 120 条（约 2 个数据点/分钟）
- 24h: 500 条（约 1 个数据点/3分钟）
- 7d: 1000 条（约 1 个数据点/10分钟）

**修改文件**:
- `web/src/api/agents.ts`

**状态**: ✅ 已修复

---

## 修复日志

| 日期 | Bug ID | 修复内容 | 状态 |
|------|--------|----------|------|
| 2026-03-10 | #1 | Server Dashboard WebSocket 启动 | ✅ 已修复 |
| 2026-03-10 | #2 | 前端 Agent 实时状态更新 | ✅ 已修复 |
| 2026-03-11 | #3 | Agent 命令处理未实现 | ✅ 已修复 |
| 2026-03-10 | #4 | 前端历史指标图表时间范围切换无效 | ✅ 已修复 |
| 2026-03-10 | #5 | React Hooks 规则违反导致图表不显示 | ✅ 已修复 |
| 2026-03-10 | #6 | 后端时区问题导致时间范围查询不准确 | ✅ 已修复 |
| 2026-03-10 | #7 | 前端未传递 limit 导致数据被截断 | ✅ 已修复 |
| 2026-03-10 | #8 | WebSocket 频繁断开重连提示 | ✅ 已修复 |
| 2026-03-12 | #9 | Agent 执行 ping 命令时崩溃 | 🐛 暂不处理 |

---

### Bug #8: WebSocket 频繁断开重连提示

**严重程度**: 中

**问题描述**:
前端时不时弹出 "WebSocket 连接已断开，正在尝试重连..." 提示，用户体验差。

**根本原因**:
`useWebSocket` hook 在 `DashboardPage` 组件中调用：
- `DashboardPage` 是路由组件，导航到其他页面时会卸载
- 组件卸载时，`useWebSocket` 的 cleanup 函数调用 `disconnect()` 关闭 WebSocket
- 返回 Dashboard 时重新连接，触发短暂断开提示

**影响范围**:
- 用户离开仪表盘页面时 WebSocket 断开
- 无法在其他页面接收实时告警通知
- 频繁的断开/重连提示干扰用户

**修复方案**:
将 `useWebSocket` 移到 `AlertNotificationProvider` 组件，使 WebSocket 在整个应用生命周期内保持活跃。

**修改文件**:
- `web/src/components/common/AlertNotificationProvider.tsx`
- `web/src/pages/dashboard/index.tsx`

**状态**: ✅ 已修复 |

---

## 2026-03-12 发现的问题

### Bug #9: Agent 执行 ping 命令时崩溃

**严重程度**: 高

**问题描述**:
Agent 在执行 `ping` 命令时会立即崩溃，而简单的命令（如 `echo`, `dir`）正常工作。

**现象**:
- 日志显示 "Executing command" 后没有 "Sent command result"
- Agent 进程直接退出
- simple 命令正常，网络相关命令（如 `ping -n 4 8.8.8.8`）导致崩溃

**影响范围**:
- 用户无法使用 ping 命令检测网络连通性
- 需要网络诊断的任务无法执行

**已尝试的修复**:
1. 修复 stdout/stderr pipes 的可继承性设置 ❌
2. 尝试不同的 stdin 处理方式（NUL 设备）❌
3. 多次重建测试 ❌

**可能的根本原因**:
- `GetStdHandle(STD_INPUT_HANDLE)` 在特定环境下返回无效句柄
- CreateProcessA 调用时参数传递有问题
- 需要更详细的调试信息（crash dump）来确定

**变通方案**:
使用 PowerShell 的 `Test-Connection` 替代 `ping`：
```bash
Test-Connection -Count 1 -ComputerName 8.8.8.8
```

**修改文件**:
- `agent/src/task/commands/exec_shell.cpp`

**状态**: 🐛 暂不处理（需要 C++ 开发者调试）
