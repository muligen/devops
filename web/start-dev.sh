#!/bin/bash
# AgentTeams Web Launcher (Development Mode)

echo "========================================"
echo "AgentTeams Web Launcher (Dev)"
echo "========================================"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 检查 Node.js 是否安装
echo "Checking Node.js..."
if ! command -v node &> /dev/null; then
    echo "[ERROR] Node.js is not installed or not in PATH."
    exit 1
fi
echo "[OK] Node.js version: $(node --version)"
echo ""

# 检查后端服务是否运行
echo "Checking backend server..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "[WARNING] Backend server (http://localhost:8080) is not available."
    echo "Please start the server first: cd ../server && ./start-server.sh"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "[OK] Backend server is available"
fi
echo ""

# 检查 node_modules 是否存在
if [ ! -d "node_modules" ]; then
    echo "[INFO] node_modules not found. Installing dependencies..."
    npm install
    echo ""
fi

# 启动开发服务器
echo "Starting development server..."
echo "Web will be available at: http://localhost:3000"
echo ""

npm run dev
