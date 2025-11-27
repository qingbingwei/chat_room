package main

import (
	"chat-server/internal/server"
	"log"
)

func main() {
	// 创建并启动服务器
	srv := server.NewServer(":9090")

	log.Println("========================================")
	log.Println("聊天服务器启动中...")
	log.Println("监听地址: :9090")
	log.Println("WebSocket 端点: ws://localhost:9090/ws")
	log.Println("========================================")

	if err := srv.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
