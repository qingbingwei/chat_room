#!/bin/bash

# 启动前端开发服务器脚本

cd "$(dirname "$0")/frontend"

echo "========================================="
echo "正在启动前端开发服务器..."
echo "========================================="

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
    echo "安装依赖..."
    npm install
    if [ $? -ne 0 ]; then
        echo "依赖安装失败！"
        exit 1
    fi
fi

# 启动开发服务器
npm run dev
