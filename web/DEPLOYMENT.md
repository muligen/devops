# AgentTeams Web Frontend 部署指南

## 目录

- [环境要求](#环境要求)
- [开发环境](#开发环境)
- [生产部署](#生产部署)
- [Docker 部署](#docker-部署)
- [Nginx 配置](#nginx-配置)
- [环境变量](#环境变量)
- [常见问题](#常见问题)

## 环境要求

- Node.js >= 18.0.0
- npm >= 9.0.0
- 现代浏览器 (Chrome, Firefox, Safari, Edge)

## 开发环境

### 安装依赖

```bash
cd web
npm install
```

### 启动开发服务器

```bash
npm run dev
```

开发服务器默认运行在 `http://localhost:3000`，会自动代理 API 请求到后端服务器。

### 代码检查

```bash
# ESLint 检查
npm run lint

# 自动修复
npm run lint:fix

# 格式化代码
npm run format
```

### 构建生产版本

```bash
npm run build
```

构建产物位于 `dist/` 目录。

## 生产部署

### 方式一：静态文件部署

1. 构建生产版本：

```bash
cd web
npm run build
```

2. 将 `dist/` 目录部署到任意静态文件服务器。

### 方式二：Docker 部署

1. 构建镜像：

```bash
cd web
docker build -t agentteams-web:latest .
```

2. 运行容器：

```bash
docker run -d \
  --name agentteams-web \
  -p 80:80 \
  -e API_URL=http://your-server:8080 \
  agentteams-web:latest
```

### 方式三：Docker Compose

在项目根目录使用 `docker-compose.yml`：

```bash
docker-compose up -d web
```

## Nginx 配置

如果使用独立的 Nginx 服务器，参考以下配置：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /var/www/agentteams-web/dist;
    index index.html;

    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/json;

    # API 代理
    location /api/ {
        proxy_pass http://your-server:8080/api/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket 代理
    location /ws/ {
        proxy_pass http://your-server:8080/ws/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }

    # SPA 回退
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `VITE_API_BASE_URL` | API 基础 URL | `/api/v1` |

创建 `.env.production` 文件设置生产环境变量：

```env
VITE_API_BASE_URL=/api/v1
```

## 架构说明

```
web/
├── src/
│   ├── api/           # API 调用封装
│   ├── components/    # React 组件
│   ├── hooks/         # 自定义 Hooks
│   ├── pages/         # 页面组件
│   ├── stores/        # Zustand 状态管理
│   ├── types/         # TypeScript 类型定义
│   ├── utils/         # 工具函数
│   └── App.tsx        # 应用入口
├── public/            # 静态资源
├── nginx.conf         # Docker Nginx 配置
└── Dockerfile         # Docker 构建文件
```

### 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI 组件**: Ant Design 5
- **图表**: ECharts 5
- **状态管理**: Zustand
- **HTTP 客户端**: Axios
- **路由**: React Router 6

## 常见问题

### 1. WebSocket 连接失败

检查 Nginx 是否正确配置了 WebSocket 代理：

```nginx
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "Upgrade";
```

### 2. API 请求跨域

确保后端服务器配置了正确的 CORS 头：

```go
router.Use(middleware.CORS([]string{"*"}))
```

### 3. 刷新页面 404

确保 Nginx 配置了 SPA 回退：

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### 4. 静态资源加载失败

检查 `vite.config.ts` 中的 `base` 配置，确保与部署路径一致。

## 性能优化

### 构建优化

项目已配置代码分割：

```typescript
manualChunks: {
  vendor: ['react', 'react-dom', 'react-router-dom'],
  antd: ['antd', '@ant-design/icons'],
  echarts: ['echarts', 'echarts-for-react'],
}
```

### 进一步优化建议

1. **启用 HTTP/2**: 减少连接开销
2. **CDN 加速**: 静态资源使用 CDN
3. **图片优化**: 使用 WebP 格式
4. **启用 Brotli 压缩**: 比 Gzip 更高的压缩率

## 监控与日志

建议配置前端错误监控：

- Sentry
- LogRocket
- 自定义错误上报

## 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建
cd web
npm install
npm run build

# Docker 部署
docker-compose up -d --build web
```
