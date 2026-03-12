#!/bin/bash
# AgentTeams Server Stopper

echo "========================================"
echo "AgentTeams Server Stopper"
echo "========================================"
echo ""

# 停止 Server（如果运行中）
echo "Stopping AgentTeams Server..."
if pgrep -f "server.exe" > /dev/null; then
    pkill -f "server.exe"
    echo "[OK] Server stopped"
else
    echo "[INFO] Server is not running"
fi
echo ""
sleep 1

# 停止生产环境容器
echo "Stopping production containers..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_COMPOSE_FILE="$PROJECT_DIR/docker-compose.prod.yml"

docker-compose -f "$DOCKER_COMPOSE_FILE" down
echo "[OK] Containers stopped"
echo ""
