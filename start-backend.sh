#!/bin/bash

# 启动后端服务器脚本

cd "$(dirname "$0")/backend"

echo "========================================="
echo "正在启动聊天服务器..."
echo "========================================="

# 检查是否已编译
if [ ! -f "chat-server" ]; then
    echo "编译服务器..."
    go build -o chat-server main.go
    if [ $? -ne 0 ]; then
        echo "编译失败！"
        exit 1
    fi
fi

# 启动服务器
./chat-server
