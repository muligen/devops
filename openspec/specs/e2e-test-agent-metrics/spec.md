## ADDED Requirements

### Requirement: Agent 指标上报
测试 SHALL 验证 Agent 定期上报系统指标。

#### Scenario: 指标上报
- **WHEN** Agent 已连接
- **THEN** Agent 每 30 秒上报 CPU、内存、磁盘指标

#### Scenario: 指标格式验证
- **WHEN** Agent 上报指标
- **THEN** 指标包含 cpu_usage, memory.total, memory.used, disk.total, disk.used 字段

### Requirement: 指标存储
测试 SHALL 验证 Server 存储指标数据。

#### Scenario: 指标入库
- **WHEN** Server 收到指标消息
- **THEN** 指标数据写入数据库 metrics 表

#### Scenario: 指标历史查询
- **WHEN** 请求 GET /api/v1/agents/{id}/metrics?range=1h
- **THEN** 返回最近 1 小时的指标历史数据

### Requirement: 指标聚合
测试 SHALL 验证指标数据聚合统计。

#### Scenario: 平均值统计
- **WHEN** 请求聚合指标统计
- **THEN** 返回指定时间范围内的平均值、最大值、最小值
