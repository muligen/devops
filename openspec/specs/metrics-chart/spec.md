## Requirements

### Requirement: 指标历史查询
系统 SHALL 支持查询 Agent 历史指标数据。

#### Scenario: 查询指定时间范围
- **WHEN** 用户选择时间范围（1小时/24小时/7天/自定义）
- **THEN** 系统返回该时间范围内的指标数据点

#### Scenario: 数据点聚合
- **WHEN** 查询时间范围超过 1 小时
- **THEN** 系统返回聚合后的数据点（每 5 分钟一个点）

### Requirement: CPU 使用率图表
系统 SHALL 提供 CPU 使用率趋势图表。

#### Scenario: 显示 CPU 图表
- **WHEN** 用户查看 Agent 详情页的 CPU 指标
- **THEN** 系统显示折线图，X 轴为时间，Y 轴为 CPU 百分比
- **AND** 图表支持悬停显示具体数值

### Requirement: 内存使用图表
系统 SHALL 提供内存使用趋势图表。

#### Scenario: 显示内存图表
- **WHEN** 用户查看 Agent 详情页的内存指标
- **THEN** 系统显示折线图，包含总内存和使用量两条线
- **AND** 显示内存使用百分比

### Requirement: 磁盘使用图表
系统 SHALL 提供磁盘使用趋势图表。

#### Scenario: 显示磁盘图表
- **WHEN** 用户查看 Agent 详情页的磁盘指标
- **THEN** 系统显示折线图，包含总磁盘和使用量两条线
- **AND** 显示磁盘使用百分比

### Requirement: 图表交互
指标图表 SHALL 支持交互操作。

#### Scenario: 图表缩放
- **WHEN** 用户拖动选择图表区域
- **THEN** 系统放大显示选中区域

#### Scenario: 时间范围切换
- **WHEN** 用户点击时间范围按钮（1H/24H/7D）
- **THEN** 系统重新加载并显示对应时间范围的数据

### Requirement: 多 Agent 对比
系统 SHALL 支持多个 Agent 指标对比。

#### Scenario: 选择对比 Agent
- **WHEN** 用户在图表中选择多个 Agent
- **THEN** 系统在同一图表中显示多个 Agent 的指标趋势
- **AND** 使用不同颜色区分
