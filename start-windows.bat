@echo off
chcp 65001 >nul
title 聊天室部署脚本

echo ========================================
echo      聊天室一键部署脚本 (Windows)
echo ========================================
echo.

:: 检查 Go 是否安装
echo [1/5] 检查 Go 环境...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go，请先安装 Go 1.21+
    echo 下载地址: https://go.dev/dl/
    pause
    exit /b 1
)
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo       Go 版本: %GO_VERSION%

:: 检查 Node.js 是否安装
echo [2/5] 检查 Node.js 环境...
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Node.js，请先安装 Node.js 18+
    echo 下载地址: https://nodejs.org/
    pause
    exit /b 1
)
for /f "tokens=1" %%i in ('node -v') do set NODE_VERSION=%%i
echo       Node.js 版本: %NODE_VERSION%

:: 安装后端依赖
echo [3/5] 安装后端依赖...
cd backend
go mod download
if %errorlevel% neq 0 (
    echo [错误] 后端依赖安装失败
    pause
    exit /b 1
)
echo       后端依赖安装完成

:: 编译后端
echo [4/5] 编译后端服务...
go build -o chat-server.exe main.go
if %errorlevel% neq 0 (
    echo [错误] 后端编译失败
    pause
    exit /b 1
)
echo       后端编译完成
cd ..

:: 安装前端依赖
echo [5/5] 安装前端依赖...
cd frontend
call npm install
if %errorlevel% neq 0 (
    echo [错误] 前端依赖安装失败
    pause
    exit /b 1
)
echo       前端依赖安装完成
cd ..

echo.
echo ========================================
echo          部署完成！
echo ========================================
echo.
echo 请使用以下方式启动：
echo   1. 双击 run-backend.bat 启动后端
echo   2. 双击 run-frontend.bat 启动前端
echo   3. 或者双击 run-all.bat 同时启动
echo.
echo 访问地址: http://localhost:5173
echo.
pause
