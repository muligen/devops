# 测试环境配置

## 服务端口

| 服务 | 地址 | 说明 |
|------|------|------|
| 后端 API | http://localhost:8080 | Go Server |
| 前端 Web | http://localhost:3000 | Nginx 容器 |
| PostgreSQL | localhost:5432 | 生产库 |
| PostgreSQL Test | localhost:5433 | 测试库 |
| Redis | localhost:6379 | 生产缓存 |
| Redis Test | localhost:6380 | 测试缓存 |
| RabbitMQ | localhost:5672 | 生产消息队列 |
| RabbitMQ Management | http://localhost:15672 | 管理界面 |
| RabbitMQ Test | localhost:5673 | 测试消息队列 |
| MinIO | http://localhost:9000 | 对象存储 |
| MinIO Console | http://localhost:9001 | 管理控制台 |

## Docker 容器

```bash
# 查看运行状态
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 生产环境容器
- agentteams-postgres    (端口 5432)
- agentteams-redis       (端口 6379)
- agentteams-rabbitmq    (端口 5672, 15672)
- agentteams-minio       (端口 9000, 9001)
- agentteams-web         (端口 3000)
- agentteams-server      (需要手动启动)

# 测试环境容器
- agentteams-test-postgres   (端口 5433)
- agentteams-test-redis      (端口 6380)
- agentteams-test-rabbitmq   (端口 5673, 15673)
- agentteams-test-minio      (端口 9002, 9003)
```

## 数据库连接

### PostgreSQL (生产)
```
Host: localhost
Port: 5432
Database: agentteams
User: postgres
Password: postgres

# 连接命令
docker exec -it agentteams-postgres psql -U postgres -d agentteams
```

### PostgreSQL (测试)
```
Host: localhost
Port: 5433
Database: agentteams_test
User: postgres
Password: postgres
```

## 测试账号

| 系统 | 用户名 | 密码 | 角色 |
|------|--------|------|------|
| Web 管理后台 | admin | admin123 | admin |
| RabbitMQ 管理 | guest | guest | - |
| MinIO | minioadmin | minioadmin | - |

## API 登录示例

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 使用 Token 访问 API
TOKEN="<your_token>"
curl http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer $TOKEN"
```

## 后端服务

### 编译
```bash
cd server
go build -o server.exe ./cmd/server
```

### 运行
```bash
./server.exe
# 或后台运行
nohup ./server.exe > server.log 2>&1 &
```

### 健康检查
```bash
curl http://localhost:8080/health
# 期望响应: {"status":"ok","version":"dev"}
```

## 前端服务

### 开发模式
```bash
cd web
npm install
npm run dev
# 访问 http://localhost:5173
```

### 生产构建
```bash
cd web
npm run build
# 输出到 dist/ 目录
```

### Docker 容器
```bash
# 重新构建并启动
docker-compose up -d --build web
```

## Agent 客户端

### 编译 (Windows)
```bash
cd agent
mkdir build && cd build
conan install .. --build=missing
cmake ..
cmake --build . --config Release
```

### 运行
```bash
./build/bin/agent.exe --config config.yaml
```

## 常用数据库查询

```sql
-- 查看 agents
SELECT id, name, status FROM agents;

-- 查看 metrics 数量
SELECT COUNT(*) as total, MAX(collected_at) as latest FROM agent_metrics;

-- 查看 tasks
SELECT id, agent_id, type, status, created_at FROM tasks ORDER BY created_at DESC LIMIT 10;

-- 查看 alert events
SELECT id, agent_id, status, message, triggered_at FROM alert_events ORDER BY triggered_at DESC LIMIT 10;
```

## 环境状态检查脚本

```bash
# 检查所有服务状态
echo "=== Docker 容器 ===" && docker ps --format "table {{.Names}}\t{{.Status}}" | grep agentteams
echo ""
echo "=== 后端健康 ===" && curl -s http://localhost:8080/health
echo ""
echo "=== 数据库连接 ===" && docker exec agentteams-postgres pg_isready -U postgres
echo ""
echo "=== Redis 连接 ===" && docker exec agentteams-redis redis-cli ping
echo ""
echo "=== RabbitMQ 状态 ===" && curl -s -u guest:guest http://localhost:15672/api/overview | grep -o '"running":true'
```

## 已知问题

见 [bugs.md](./bugs.md)
