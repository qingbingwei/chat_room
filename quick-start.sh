#!/bin/bash

# 快速部署脚本 - 一键启动前后端

echo "========================================="
echo "聊天系统快速部署"
echo "========================================="
echo ""

# 检查依赖
echo "1. 检查依赖..."

if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go 未安装，请先安装 Go 1.21+"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo "❌ 错误: Node.js 未安装，请先安装 Node.js 16+"
    exit 1
fi

echo "✅ Go 版本: $(go version)"
echo "✅ Node.js 版本: $(node --version)"
echo ""

# 后端设置
echo "2. 设置后端..."
cd backend

if [ ! -d "vendor" ]; then
    echo "   安装 Go 依赖..."
    go mod tidy
fi

if [ ! -f "chat-server" ]; then
    echo "   编译后端..."
    go build -o chat-server main.go
fi

cd ..
echo "✅ 后端设置完成"
echo ""

# 前端设置
echo "3. 设置前端..."
cd frontend

if [ ! -d "node_modules" ]; then
    echo "   安装前端依赖..."
    npm install
fi

cd ..
echo "✅ 前端设置完成"
echo ""

# 启动服务
echo "========================================="
echo "准备启动服务..."
echo "========================================="
echo ""
echo "后端服务器将在后台运行..."
echo "前端开发服务器将在当前终端运行..."
echo ""
echo "按 Ctrl+C 停止前端服务器"
echo "要停止后端服务器，请运行: pkill -f chat-server"
echo ""
echo "前端地址: http://localhost:5173"
echo "后端地址: http://localhost:9090"
echo "WebSocket: ws://localhost:9090/ws"
echo ""
read -p "按 Enter 键继续启动..."

# 启动后端（后台运行）
cd backend
./chat-server &
BACKEND_PID=$!
echo "✅ 后端服务器已启动 (PID: $BACKEND_PID)"

# 等待后端启动
sleep 2

# 启动前端（前台运行）
cd ../frontend
echo "✅ 启动前端服务器..."
npm run dev

# 当前端退出时，清理后端
echo ""
echo "正在停止后端服务器..."
kill $BACKEND_PID 2>/dev/null
echo "✅ 服务已停止"
