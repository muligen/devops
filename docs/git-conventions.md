# Git 提交规范

本文档定义了 AgentTeams 项目的 Git 提交规范，所有提交必须遵守此规范。

## 提交消息格式

```
<type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

### Type（类型）

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | feat(agent): 添加心跳超时检测 |
| `fix` | Bug 修复 | fix(server): 修复 WebSocket 连接泄漏 |
| `docs` | 文档更新 | docs: 更新部署指南 |
| `style` | 代码格式（不影响逻辑） | style(agent): 格式化代码缩进 |
| `refactor` | 重构（不增加功能，不修复 bug） | refactor(server): 重构任务调度逻辑 |
| `perf` | 性能优化 | perf(agent): 优化指标采集性能 |
| `test` | 测试相关 | test(server): 添加认证模块单元测试 |
| `build` | 构建系统或依赖更新 | build: 升级 Go 版本到 1.21 |
| `ci` | CI/CD 配置更新 | ci: 添加 GitHub Actions 工作流 |
| `chore` | 其他杂项 | chore: 更新 .gitignore |
| `revert` | 回滚提交 | revert: 回滚任务队列修改 |

### Scope（范围）

根据模块选择对应的 scope：

| Scope | 说明 |
|-------|------|
| `agent` | C++ Agent 相关 |
| `server` | Go Server 相关 |
| `api` | API 定义、OpenAPI 规范 |
| `auth` | 认证授权模块 |
| `task` | 任务管理模块 |
| `monitor` | 监控告警模块 |
| `update` | 自动更新模块 |
| `db` | 数据库相关 |
| `deploy` | 部署配置 |
| `docs` | 文档 |

### Subject（主题）

- 使用祈使句，首字母小写
- 不以句号结尾
- 简洁描述变更内容（50 字符以内）
- 使用中文或英文，保持一致性

### Body（正文）

- 详细说明变更内容
- 可以分多行
- 解释 "为什么" 而不是 "做了什么"

### Footer（脚注）

用于关联 Issue 或标记破坏性变更：

```
Closes #123
Fixes #456
BREAKING CHANGE: API 接口变更说明
```

## 示例

### 新功能

```
feat(task): 添加批量任务创建接口

支持一次性创建多个任务，减少 API 调用次数。
任务按优先级排序执行。

Closes #42
```

### Bug 修复

```
fix(websocket): 修复心跳超时判断逻辑

修复心跳超时计算错误导致 Agent 异常断开的问题。
原因是时间比较使用了错误的时间单位。

Fixes #88
```

### 破坏性变更

```
feat(api)!: 重构认证 API 响应格式

统一 API 响应格式，返回标准化的错误码。

BREAKING CHANGE:
- /api/v1/auth/login 响应格式变更
- refresh_token 字段移至 data.refresh_token
```

### 重构

```
refactor(monitor): 重构指标存储逻辑

将指标存储逻辑从 handler 中抽取到独立的 repository，
提高代码可测试性和可维护性。
```

## 分支命名规范

| 分支类型 | 命名格式 | 示例 |
|----------|----------|------|
| 主分支 | `main` | main |
| 开发分支 | `develop` | develop |
| 功能分支 | `feature/<name>` | feature/batch-task |
| 修复分支 | `fix/<name>` | fix/websocket-reconnect |
| 发布分支 | `release/<version>` | release/v1.0.0 |
| 热修复分支 | `hotfix/<version>` | hotfix/v1.0.1 |

## 版本标签规范

使用语义化版本号：`v<major>.<minor>.<patch>`

- **major**: 不兼容的 API 变更
- **minor**: 向后兼容的功能新增
- **patch**: 向后兼容的问题修复

示例：
- `v1.0.0` - 首个正式版本
- `v1.1.0` - 新增批量任务功能
- `v1.1.1` - 修复心跳超时问题

## 提交检查清单

提交前请确认：

- [ ] 提交消息符合规范格式
- [ ] 代码已通过 lint 检查
- [ ] 单元测试通过
- [ ] 不包含敏感信息（密码、密钥等）
- [ ] 不包含不必要的文件（构建产物、IDE 配置等）

## 工具配置

### Git Hooks (可选)

使用 commitlint 自动检查提交消息格式：

```bash
# 安装 commitlint
npm install -g @commitlint/cli @commitlint/config-conventional

# 配置
echo "module.exports = { extends: ['@commitlint/config-conventional'] };" > commitlint.config.js
```

### IDE 配置

VS Code 推荐安装：
- GitLens
- Conventional Commits
