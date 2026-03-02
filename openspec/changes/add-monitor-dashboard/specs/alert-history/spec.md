## ADDED Requirements

### Requirement: 告警事件存储
系统 SHALL 存储告警触发事件。

#### Scenario: 告警触发记录
- **WHEN** 告警规则触发
- **THEN** 系统创建告警事件记录
- **AND** 记录规则 ID、Agent ID、指标值、阈值、触发时间

#### Scenario: 告警恢复记录
- **WHEN** 告警条件恢复正常
- **THEN** 系统更新告警事件状态为 resolved
- **AND** 记录恢复时间

### Requirement: 告警历史查询 API
系统 SHALL 提供告警历史查询 API。

#### Scenario: 查询告警历史
- **WHEN** 前端请求 GET /api/v1/alerts/history
- **THEN** 系统返回告警事件列表
- **AND** 支持按状态筛选（pending/acknowledged/resolved）
- **AND** 支持按时间范围筛选

#### Scenario: 分页查询
- **WHEN** 前端指定 page 和 page_size 参数
- **THEN** 系统返回分页后的结果

### Requirement: 告警确认
系统 SHALL 支持告警确认操作。

#### Scenario: 确认告警
- **WHEN** 用户点击告警的"确认"按钮
- **THEN** 系统更新告警状态为 acknowledged
- **AND** 记录确认人和确认时间

#### Scenario: 批量确认
- **WHEN** 用户选择多个告警并点击批量确认
- **THEN** 系统批量更新所有选中告警的状态

### Requirement: 告警历史展示
前端 SHALL 展示告警历史列表。

#### Scenario: 显示告警列表
- **WHEN** 用户访问告警历史页面
- **THEN** 系统显示告警事件列表
- **AND** 每条记录显示触发时间、Agent 名称、告警规则、当前状态

#### Scenario: 告警详情
- **WHEN** 用户点击告警记录
- **THEN** 系统显示详细信息，包含指标值、阈值、相关指标图表链接

#### Scenario: 状态过滤
- **WHEN** 用户选择状态过滤条件
- **THEN** 系统只显示该状态的告警记录

### Requirement: 告警统计
系统 SHALL 提供告警统计数据。

#### Scenario: 告警概览统计
- **WHEN** 前端请求 Dashboard 统计
- **THEN** 系统返回未处理告警数量
- **AND** 返回最近 24 小时告警触发数量
