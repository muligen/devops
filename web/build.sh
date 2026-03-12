#!/bin/bash
# AgentTeams Web Builder (Production Build)

echo "========================================"
echo "AgentTeams Web Builder (Production)"
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

# 检查是否需要安装依赖
if [ ! -d "node_modules" ]; then
    echo "[INFO] Installing dependencies..."
    npm install
    echo ""
fi

# 清理旧的构建
echo "Cleaning old build..."
rm -rf dist
echo ""

# 代码检查
echo "Running lint check..."
npm run lint
if [ $? -ne 0 ]; then
    echo "[WARNING] Lint check failed, but continuing with build..."
    echo "To fix lint issues, run: npm run lint:fix"
    echo ""
fi

# 构建
echo "Building for production..."
npm run build

if [ $? -eq 0 ]; then
    echo ""
    echo "[OK] Build completed successfully!"
    echo "Output directory: $SCRIPT_DIR/dist"
else
    echo ""
    echo "[ERROR] Build failed!"
    exit 1
fi
