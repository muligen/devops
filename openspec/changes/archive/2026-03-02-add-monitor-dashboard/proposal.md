## Why

当前系统缺乏可视化监控界面，运维人员需要通过 API 或命令行来查看 Agent 状态和执行任务，效率较低。无法快速发现异常机器（高 CPU、低磁盘等），也无法直观查看历史指标趋势。

需要一个 Web 监控面板来：
- 实时监控所有 Agent 状态
- 快速执行远程任务
- 管理告警规则和查看告警历史
- 查看历史指标趋势图表

## What Changes

### 新增功能

- **Web 前端监控面板**: 基于 React + Ant Design 的管理界面
- **实时状态推送**: WebSocket 实时推送 Agent 状态变更
- **指标图表可视化**: ECharts 展示 CPU/内存/磁盘历史趋势
- **快速任务执行**: 从面板直接下发命令到单个或多个 Agent
- **告警历史记录**: 存储和展示告警事件历史

### API 变更

- 新增 `WS /api/v1/ws/dashboard` 前端订阅实时数据
- 新增 `GET /api/v1/alerts/history` 告警历史查询
- 增强 `GET /api/v1/dashboard/stats` 增加趋势数据
- 增强 `GET /api/v1/agents` 支持按资源使用排序

## Capabilities

### New Capabilities

- `dashboard-frontend`: Web 监控面板前端，包含 Dashboard、Agent 列表、Agent 详情、任务管理、告警管理等页面
- `realtime-push`: WebSocket 实时数据推送，支持 Agent 状态、指标、告警事件推送
- `metrics-chart`: 历史指标图表可视化，支持时间范围选择和多种图表类型
- `quick-task-exec`: 快速任务执行面板，支持单机和批量执行
- `alert-history`: 告警历史记录存储和查询

### Modified Capabilities

- `monitoring-dashboard`: 增加实时推送能力和更丰富的统计数据

## Impact

### 后端影响

- `server/internal/modules/monitor/`: 增强 Dashboard 统计、新增告警历史
- `server/internal/modules/agent/handler/`: 新增前端 WebSocket 端点
- `server/internal/modules/agent/domain/`: 新增告警事件模型

### 新增代码

- `web/`: React 前端项目目录
- `web/src/pages/dashboard/`: Dashboard 页面
- `web/src/pages/agents/`: Agent 列表和详情页
- `web/src/pages/tasks/`: 任务管理页
- `web/src/pages/alerts/`: 告警管理页

### 依赖变更

- 前端新增依赖: React 18, Ant Design 5, ECharts 5
- 后端无新增依赖（使用现有 gorilla/websocket）
