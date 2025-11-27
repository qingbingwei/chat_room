@echo off
chcp 65001 >nul
title 聊天室一键启动

echo ========================================
echo          一键启动聊天室
echo ========================================
echo.

:: 启动后端（新窗口）
echo 正在启动后端服务...
start "聊天室后端" cmd /c "%~dp0run-backend.bat"

:: 等待后端启动
timeout /t 2 /nobreak >nul

:: 启动前端（新窗口）
echo 正在启动前端服务...
start "聊天室前端" cmd /c "%~dp0run-frontend.bat"

echo.
echo ========================================
echo           服务已启动！
echo ========================================
echo.
echo 后端地址: ws://localhost:9090/ws
echo 前端地址: http://localhost:5173
echo.
echo 请等待前端编译完成后访问...
echo 按任意键关闭此窗口（服务将继续运行）
echo.
pause >nul
