@echo off
chcp 65001 >nul
title 聊天室前端服务

echo ========================================
echo        启动前端服务 (端口: 5173)
echo ========================================
echo.

cd /d "%~dp0frontend"

if not exist "node_modules" (
    echo 正在安装前端依赖...
    call npm install
    if %errorlevel% neq 0 (
        echo [错误] 依赖安装失败
        pause
        exit /b 1
    )
)

echo 前端服务启动中...
echo 访问地址: http://localhost:5173
echo.
call npm run dev
pause
