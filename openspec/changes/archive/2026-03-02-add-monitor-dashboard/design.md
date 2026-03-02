## Context

### 当前状态

- Server 提供 REST API 和 Agent WebSocket 连接
- 已有基础的 Dashboard 统计 API (`/api/v1/dashboard/stats`)
- 已有 Agent 管理、任务执行、告警规则的 API
- 缺乏前端界面，运维需要通过 API 或 curl 操作

### 约束

- 前端需要独立部署，不与 Server 耦合
- 需要支持实时数据更新
- 需要与现有认证体系 (JWT) 集成
- 需要遵循项目的编码规范

## Goals / Non-Goals

**Goals:**

- 提供 Web 监控面板，支持 Agent 状态查看和任务执行
- 实现实时数据推送 (WebSocket)
- 提供历史指标图表展示
- 支持告警规则管理和告警历史查看

**Non-Goals:**

- 不做移动端适配（当前仅支持桌面端）
- 不做复杂的权限管理（使用现有 RBAC）
- 不做数据导出功能（后续迭代）
- 不做多语言支持（当前仅中文）

## Decisions

### 1. 前端技术栈

**决策**: React 18 + TypeScript + Ant Design 5 + ECharts 5

**备选方案:**

| 方案 | 优点 | 缺点 |
|------|------|------|
| React + Ant Design | 生态丰富，企业级组件完善 | 包体积较大 |
| Vue + Element Plus | 学习曲线平缓，中文社区好 | 团队 React 熟悉度更高 |
| Next.js | SSR 支持 | 对纯后台系统收益不大 |

**理由**: Ant Design 的 Table、Form、Modal 等组件非常适合后台管理系统，ECharts 提供强大的监控图表能力。

### 2. 实时通信方案

**决策**: 前端使用独立 WebSocket 连接订阅实时数据

**架构:**

```
┌─────────────┐     WS /ws/dashboard     ┌─────────────┐
│   Frontend  │◄─────────────────────────│   Server    │
│  (React)    │                          │    (Go)     │
└─────────────┘                          └──────┬──────┘
                                                │
                    ┌───────────────────────────┼───────────────────────────┐
                    │                           │                           │
              ┌─────▼─────┐              ┌──────▼──────┐            ┌───────▼───────┐
              │ Agent WS  │              │   Metrics   │            │    Alerts     │
              │  Events   │              │   Events    │            │    Events     │
              └───────────┘              └─────────────┘            └───────────────┘
```

**推送内容:**

- Agent 状态变更 (online/offline)
- 实时指标 (每分钟聚合)
- 告警事件触发

### 3. 前端项目结构

**决策**: 独立 web 目录，与 server 平级

```
web/
├── src/
│   ├── api/           # API 调用封装
│   ├── components/    # 通用组件
│   ├── hooks/         # 自定义 Hooks
│   ├── pages/
│   │   ├── dashboard/ # Dashboard 总览
│   │   ├── agents/    # Agent 列表/详情
│   │   ├── tasks/     # 任务管理
│   │   └── alerts/    # 告警管理
│   ├── stores/        # Zustand 状态管理
│   ├── utils/         # 工具函数
│   └── App.tsx
├── public/
├── package.json
└── vite.config.ts
```

### 4. 告警历史存储

**决策**: 新增 `alert_events` 表存储告警事件

**表结构:**

```sql
CREATE TABLE alert_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID REFERENCES alert_rules(id),
    agent_id UUID REFERENCES agents(id),
    metric_value FLOAT NOT NULL,
    threshold FLOAT NOT NULL,
    status VARCHAR(20) NOT NULL, -- pending, acknowledged, resolved
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX ix_alert_events_status ON alert_events(status);
CREATE INDEX ix_alert_events_triggered ON alert_events(triggered_at);
```

### 5. 部署方式

**决策**: 前端构建为静态文件，可独立部署或由 Server 服务

**选项:**

- 开发环境: Vite 开发服务器代理到 Server
- 生产环境: Nginx 服务静态文件 + 反向代理 API

## Risks / Trade-offs

### 风险 1: WebSocket 连接稳定性

**风险**: 网络不稳定可能导致 WebSocket 频繁断开

**缓解措施**:
- 前端实现自动重连机制 (指数退避)
- 关键数据同时提供 REST API 作为 fallback
- 显示连接状态指示器

### 风险 2: 大量 Agent 时的性能

**风险**: 当 Agent 数量超过 1000 时，实时推送可能造成性能问题

**缓解措施**:
- 推送数据聚合 (每 5 秒批量推送)
- 前端虚拟滚动列表
- 支持分页加载，默认只显示前 100 个

### 风险 3: 前端状态管理复杂度

**风险**: 实时数据更新可能导致状态管理复杂

**缓解措施**:
- 使用 Zustand 简化状态管理
- WebSocket 数据直接更新 store
- 组件从 store 订阅数据

## Migration Plan

### 部署步骤

1. **后端**: 部署新的 API 和 WebSocket 端点
2. **前端**: 构建并部署静态文件
3. **数据库**: 执行 alert_events 表迁移

### 回滚策略

- 前端: 回退到上一版本静态文件
- 后端: API 变更为增量，不影响现有功能
- 数据库: 迁移为新增表，回滚无影响

## Open Questions

1. **是否需要暗色主题?** - 可以后续迭代支持
2. **是否需要任务执行日志流?** - 当前任务输出较长时需要，后续可加 SSE 支持
3. **是否需要 Dashboard 自定义布局?** - 当前固定布局，后续可支持拖拽
