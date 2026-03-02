# AgentTeams Server Deployment Guide

This guide covers deploying the AgentTeams Server in production environments.

## Prerequisites

### Infrastructure

- Kubernetes 1.24+ (recommended) or Docker Compose
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+
- MinIO (or S3-compatible storage)

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| Memory | 4 GB | 8+ GB |
| Storage | 50 GB | 100+ GB |

## Deployment Methods

### Method 1: Docker Compose (Development/Small Scale)

```bash
# Clone the repository
git clone https://github.com/your-org/agentteams.git
cd agentteams

# Copy environment file
cp .env.example .env

# Edit configuration
vim .env

# Start services
docker-compose up -d
```

### Method 2: Kubernetes (Production)

#### 1. Create Namespace

```bash
kubectl create namespace agentteams
```

#### 2. Create Secrets

```bash
kubectl create secret generic agentteams-secrets \
  --from-literal=jwt-secret=$(openssl rand -hex 32) \
  --from-literal=db-password=your-db-password \
  --from-literal=redis-password=your-redis-password \
  --from-literal=minio-root-user=admin \
  --from-literal=minio-root-password=your-minio-password \
  -n agentteams
```

#### 3. Deploy Infrastructure

```bash
# PostgreSQL
kubectl apply -f deployments/kubernetes/postgres.yaml -n agentteams

# Redis
kubectl apply -f deployments/kubernetes/redis.yaml -n agentteams

# RabbitMQ
kubectl apply -f deployments/kubernetes/rabbitmq.yaml -n agentteams

# MinIO
kubectl apply -f deployments/kubernetes/minio.yaml -n agentteams
```

#### 4. Deploy Server

```bash
kubectl apply -f deployments/kubernetes/server-deployment.yaml -n agentteams
kubectl apply -f deployments/kubernetes/server-service.yaml -n agentteams
```

#### 5. Configure Ingress

```bash
kubectl apply -f deployments/kubernetes/ingress.yaml -n agentteams
```

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `REDIS_URL` | Redis connection string | Yes |
| `RABBITMQ_URL` | RabbitMQ connection string | Yes |
| `MINIO_ENDPOINT` | MinIO endpoint | Yes |
| `MINIO_ACCESS_KEY` | MinIO access key | Yes |
| `MINIO_SECRET_KEY` | MinIO secret key | Yes |
| `JWT_SECRET` | JWT signing secret | Yes |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | No |

### ConfigMap Example

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentteams-config
  namespace: agentteams
data:
  LOG_LEVEL: "info"
  GIN_MODE: "release"
```

## Database Setup

### Run Migrations

```bash
# Using migration tool
migrate -path migrations -database "postgres://user:pass@localhost:5432/agentteams" up

# Or using the server's built-in migration
./server migrate
```

### Initial Admin User

```bash
# Create admin user
./server user create --username admin --role admin
```

## TLS Configuration

### Using cert-manager

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: agentteams-cert
  namespace: agentteams
spec:
  secretName: agentteams-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - agentteams.example.com
```

### Manual TLS

```bash
kubectl create secret tls agentteams-tls \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem \
  -n agentteams
```

## Scaling

### Horizontal Scaling

```bash
# Scale server deployment
kubectl scale deployment agentteams-server --replicas=3 -n agentteams

# Enable HPA
kubectl autoscale deployment agentteams-server \
  --min=2 --max=10 --cpu-percent=70 \
  -n agentteams
```

### WebSocket Considerations

For WebSocket connections with multiple replicas:
- Use Redis for session storage (already configured)
- Configure sticky sessions if needed

## Monitoring

### Health Check Endpoint

```bash
curl https://agentteams.example.com/health
```

### Prometheus Metrics

Metrics are exposed at `/metrics`:

```yaml
# Prometheus ServiceMonitor
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: agentteams
spec:
  selector:
    matchLabels:
      app: agentteams-server
  endpoints:
    - port: http
      path: /metrics
```

### Grafana Dashboard

Import the provided dashboard from `deployments/monitoring/grafana-dashboard.json`.

## Backup

### Database Backup

```bash
# PostgreSQL backup
kubectl exec -it postgres-pod -n agentteams -- \
  pg_dump -U postgres agentteams > backup.sql

# Restore
kubectl exec -i postgres-pod -n agentteams -- \
  psql -U postgres agentteams < backup.sql
```

### MinIO Backup

```bash
# Using mc (MinIO Client)
mc mirror local/agentteams /backup/agentteams
```

## Troubleshooting

### Check Logs

```bash
# Server logs
kubectl logs -f deployment/agentteams-server -n agentteams

# All pods
kubectl logs -l app=agentteams -n agentteams
```

### Common Issues

1. **Database connection failed**
   - Check credentials
   - Verify network connectivity
   - Check PostgreSQL logs

2. **Redis connection failed**
   - Verify Redis is running
   - Check credentials
   - Verify network policy

3. **WebSocket connections dropping**
   - Check load balancer timeout settings
   - Verify sticky sessions if needed
   - Check Redis session storage

## Upgrade

```bash
# Pull latest image
kubectl set image deployment/agentteams-server \
  server=agentteams/server:latest \
  -n agentteams

# Or with specific version
kubectl set image deployment/agentteams-server \
  server=agentteams/server:v1.1.0 \
  -n agentteams
```

## Rollback

```bash
# View rollout history
kubectl rollout history deployment/agentteams-server -n agentteams

# Rollback to previous version
kubectl rollout undo deployment/agentteams-server -n agentteams

# Rollback to specific revision
kubectl rollout undo deployment/agentteams-server --to-revision=2 -n agentteams
```
