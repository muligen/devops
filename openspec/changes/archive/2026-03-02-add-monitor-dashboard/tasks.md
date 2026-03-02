## 1. 后端基础设施

- [x] 1.1 创建 alert_events 数据库迁移文件
- [x] 1.2 实现 AlertEvent 数据模型 (domain)
- [x] 1.3 实现 AlertEvent Repository (CRUD 操作)
- [x] 1.4 实现 AlertEvent Service (业务逻辑)
- [x] 1.5 添加告警事件触发时创建记录的逻辑

## 2. 后端 API 增强

- [x] 2.1 实现 GET /api/v1/alerts/history 告警历史查询接口
- [x] 2.2 实现 PUT /api/v1/alerts/history/:id/acknowledge 告警确认接口
- [x] 2.3 增强 GET /api/v1/dashboard/stats 添加趋势数据支持
- [x] 2.4 增强 GET /api/v1/agents 支持按资源排序 (sort, order 参数)
- [x] 2.5 编写 API 单元测试

## 3. 后端 WebSocket 推送

- [x] 3.1 创建前端 WebSocket Handler (独立于 Agent WebSocket)
- [x] 3.2 实现 JWT 认证中间件
- [x] 3.3 实现 Agent 状态变更推送 (online/offline)
- [x] 3.4 实现指标批量推送 (5秒聚合)
- [x] 3.5 实现告警事件推送 (triggered/resolved)
- [x] 3.6 实现 Dashboard 统计定期推送 (10秒)
- [x] 3.7 编写 WebSocket 连接测试

## 4. 前端项目初始化

- [x] 4.1 创建 web 目录，初始化 Vite + React + TypeScript 项目
- [x] 4.2 安装依赖 (antd, echarts, zustand, axios, react-router-dom)
- [x] 4.3 配置 Vite 代理和构建选项
- [x] 4.4 配置 ESLint 和 Prettier
- [x] 4.5 创建项目目录结构 (api, components, hooks, pages, stores, utils)

## 5. 前端认证模块

- [x] 5.1 实现登录页面 UI
- [x] 5.2 实现 authStore 状态管理 (token, user, login, logout)
- [x] 5.3 实现 API 请求拦截器 (添加 Authorization header)
- [x] 5.4 实现 401 响应处理 (自动跳转登录页)
- [x] 5.5 实现路由保护 (ProtectedRoute 组件)

## 6. 前端布局和导航

- [x] 6.1 创建主布局组件 (MainLayout)
- [x] 6.2 实现侧边导航菜单
- [x] 6.3 实现顶部用户信息栏
- [x] 6.4 实现页面路由配置
- [x] 6.5 实现连接状态指示器

## 7. Dashboard 页面

- [x] 7.1 创建 Dashboard 页面组件
- [x] 7.2 实现统计卡片组件 (在线/离线/任务/告警数量)
- [x] 7.3 实现 Agent 状态卡片网格组件
- [x] 7.4 实现任务执行趋势图表 (ECharts 折线图)
- [x] 7.5 实现最近告警事件列表组件
- [x] 7.6 对接 Dashboard API 和 WebSocket

## 8. Agent 列表页面

- [x] 8.1 创建 Agent 列表页面组件
- [x] 8.2 实现 Agent 表格组件 (Ant Design Table)
- [x] 8.3 实现状态筛选功能
- [x] 8.4 实现资源排序功能
- [x] 8.5 实现搜索功能
- [x] 8.6 实现实时状态更新 (WebSocket)

## 9. Agent 详情页面

- [x] 9.1 创建 Agent 详情页面组件
- [x] 9.2 实现基本信息卡片组件
- [x] 9.3 实现实时状态面板 (CPU/内存/磁盘仪表盘)
- [x] 9.4 实现历史指标图表组件 (ECharts)
- [x] 9.5 实现时间范围选择器 (1H/24H/7D)
- [x] 9.6 实现最近任务列表组件
- [x] 9.7 实现快速执行任务入口

## 10. 任务管理页面

- [x] 10.1 创建任务列表页面组件
- [x] 10.2 实现任务表格组件
- [x] 10.3 实现任务状态筛选功能
- [x] 10.4 实现任务详情抽屉/模态框
- [x] 10.5 实现任务输出展示组件
- [x] 10.6 实现任务取消功能

## 11. 快速执行任务面板

- [x] 11.1 创建任务执行对话框组件
- [x] 11.2 实现 Agent 选择组件 (支持多选)
- [x] 11.3 实现命令类型选择 (Shell/内置命令)
- [x] 11.4 实现 Shell 命令输入框
- [x] 11.5 实现内置命令下拉选择
- [x] 11.6 实现高级选项配置 (超时、优先级)
- [x] 11.7 实现执行结果实时展示
- [x] 11.8 实现批量执行汇总展示

## 12. 告警管理页面

- [x] 12.1 创建告警管理页面组件
- [x] 12.2 实现告警规则表格组件
- [x] 12.3 实现告警规则创建/编辑表单
- [x] 12.4 实现告警历史列表组件
- [x] 12.5 实现告警确认功能
- [x] 12.6 实现告警状态筛选
- [x] 12.7 实现告警实时推送提示

## 13. WebSocket 集成

- [x] 13.1 创建 useWebSocket Hook
- [x] 13.2 实现自动重连逻辑 (指数退避)
- [x] 13.3 实现消息分发机制 (按 type 分发到不同 store)
- [x] 13.4 实现 Dashboard 实时数据更新
- [x] 13.5 实现 Agent 列表实时更新
- [x] 13.6 实现告警实时推送通知

## 14. 前端优化和测试

- [x] 14.1 实现 Agent 列表虚拟滚动 (大数据量优化)
- [x] 14.2 实现图表数据缓存
- [x] 14.3 编写组件单元测试
- [x] 14.4 编写 E2E 测试 (可选)
- [x] 14.5 优化打包体积 (代码分割)

## 15. 部署配置

- [x] 15.1 创建前端 Dockerfile
- [x] 15.2 更新 docker-compose.yaml 添加 web 服务
- [x] 15.3 创建 Nginx 配置文件
- [x] 15.4 更新 Server CORS 配置
- [x] 15.5 编写前端部署文档
