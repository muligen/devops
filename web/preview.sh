#!/bin/bash
# AgentTeams Web Preview (Production Build Preview)

echo "========================================"
echo "AgentTeams Web Preview"
echo "========================================"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 检查构建输出是否存在
if [ ! -d "dist" ]; then
    echo "[ERROR] Build output directory 'dist' not found."
    echo "Please run build first: ./build.sh"
    exit 1
fi

# 检查后端服务是否运行
echo "Checking backend server..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "[WARNING] Backend server (http://localhost:8080) is not available."
    echo "The preview will not be able to connect to the API."
    echo ""
fi

# 启动预览服务器
echo "Starting preview server..."
echo "Web will be available at: http://localhost:4173"
echo ""

npm run preview
