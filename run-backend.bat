@echo off
chcp 65001 >nul
title 聊天室后端服务

echo ========================================
echo        启动后端服务 (端口: 9090)
echo ========================================
echo.

cd /d "%~dp0backend"

if not exist "chat-server.exe" (
    echo 正在编译后端...
    go build -o chat-server.exe main.go
    if %errorlevel% neq 0 (
        echo [错误] 编译失败
        pause
        exit /b 1
    )
)

echo 后端服务启动中...
echo WebSocket 地址: ws://localhost:9090/ws
echo.
chat-server.exe
pause
