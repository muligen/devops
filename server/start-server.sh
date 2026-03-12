#!/bin/bash
# AgentTeams Server Launcher

echo "========================================"
echo "AgentTeams Server Launcher"
echo "========================================"
echo ""

# 检查 Docker 是否运行
echo "Checking Docker..."
if ! docker ps &> /dev/null; then
    echo "[ERROR] Docker is not running or not installed."
    echo "Please start Docker Desktop first."
    exit 1
fi
echo "[OK] Docker is running"
echo ""

# 检查容器是否运行
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_COMPOSE_FILE="$PROJECT_DIR/docker-compose.prod.yml"

echo "Checking containers..."
if ! docker ps --filter "name=agentteams-postgres" --format "{{.Names}}" | grep -q "agentteams-postgres"; then
    echo "[INFO] Production containers are not running."
    echo "Starting production environment..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d
    echo ""
    echo "Waiting for services to be ready..."
    sleep 5

    # 等待 PostgreSQL 就绪
    max_attempts=30
    attempts=0
    while [ $attempts -lt $max_attempts ]; do
        if docker exec agentteams-postgres pg_isready -U postgres &> /dev/null; then
            echo "[OK] PostgreSQL is ready"
            break
        fi
        attempts=$((attempts + 1))
        echo "Waiting... ($attempts/$max_attempts)"
        sleep 2
    done

    if [ $attempts -ge $max_attempts ]; then
        echo "[ERROR] PostgreSQL failed to start."
        exit 1
    fi
else
    echo "[OK] Production containers are running"
fi
echo ""

# 等待 RabbitMQ 就绪
echo "Waiting for RabbitMQ..."
rbAttempts=0
rbMaxAttempts=15
while [ $rbAttempts -lt $rbMaxAttempts ]; do
    if docker exec agentteams-rabbitmq rabbitmqctl status &> /dev/null; then
        echo "[OK] RabbitMQ is ready"
        break
    fi
    rbAttempts=$((rbAttempts + 1))
    echo "Waiting for RabbitMQ... ($rbAttempts/$rbMaxAttempts)"
    sleep 2
done

if [ $rbAttempts -ge $rbMaxAttempts ]; then
    echo "[ERROR] RabbitMQ failed to start."
    exit 1
fi
echo ""

# 启动 Server
echo "Starting AgentTeams Server..."
echo ""

cd "$SCRIPT_DIR"
./server.exe
