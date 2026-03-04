## ADDED Requirements

### Requirement: 版本检查
测试 SHALL 验证 Agent 自动更新版本检查机制。

#### Scenario: 检查新版本
- **WHEN** Agent 请求更新检查
- **THEN** Server 返回最新版本信息

#### Scenario: 无需更新
- **WHEN** Agent 版本已是最新
- **THEN** Server 返回 no_update 状态

### Requirement: 更新包下载
测试 SHALL 验证更新包下载流程。

#### Scenario: 下载更新包
- **WHEN** Agent 请求下载更新
- **THEN** Server 返回更新包下载链接和校验和

#### Scenario: 校验下载文件
- **WHEN** Agent 下载更新包后校验
- **THEN** SHA256 校验通过才能继续更新

### Requirement: 更新执行
测试 SHALL 验证更新执行和回滚机制。

#### Scenario: 成功更新
- **WHEN** 更新包校验通过
- **THEN** Agent 执行更新并重启服务

#### Scenario: 更新失败回滚
- **WHEN** 更新后服务启动失败
- **THEN** Agent 自动回滚到上一版本
