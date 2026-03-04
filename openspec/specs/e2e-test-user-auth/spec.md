## ADDED Requirements

### Requirement: 用户登录
测试 SHALL 验证用户登录认证流程。

#### Scenario: 成功登录
- **WHEN** 用户提交正确的用户名和密码
- **THEN** 系统返回 access_token 和 refresh_token

#### Scenario: 登录失败
- **WHEN** 用户提交错误密码
- **THEN** 系统返回认证失败错误

### Requirement: Token 验证
测试 SHALL 验证 JWT Token 验证机制。

#### Scenario: 有效 Token 访问
- **WHEN** 请求携带有效 Authorization header
- **THEN** 系统允许访问受保护资源

#### Scenario: 过期 Token
- **WHEN** Token 已过期
- **THEN** 系统返回 token_expired 错误码

#### Scenario: Token 刷新
- **WHEN** 使用 refresh_token 刷新
- **THEN** 系统返回新的 access_token

### Requirement: 权限控制
测试 SHALL 验证基于角色的访问控制。

#### Scenario: Admin 权限
- **WHEN** Admin 用户请求管理接口
- **THEN** 系统允许访问

#### Scenario: Operator 权限限制
- **WHEN** Operator 用户尝试删除 Agent
- **THEN** 系统返回权限不足错误

#### Scenario: Viewer 只读访问
- **WHEN** Viewer 用户请求查询接口
- **THEN** 系统允许访问
