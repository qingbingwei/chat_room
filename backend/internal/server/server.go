package server

import (
	"chat-server/internal/client"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境需要限制）
	},
}

// Server WebSocket 服务器
type Server struct {
	Hub  *client.Hub
	Addr string
}

// NewServer 创建服务器
func NewServer(addr string) *Server {
	return &Server{
		Hub:  client.NewHub(),
		Addr: addr,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 启动 Hub
	go s.Hub.Run()

	// 注册 WebSocket 处理器
	http.HandleFunc("/ws", s.handleWebSocket)

	// 静态文件服务（可选）
	http.HandleFunc("/", s.handleRoot)

	log.Printf("WebSocket 服务器启动在 %s", s.Addr)
	return http.ListenAndServe(s.Addr, nil)
}

// handleWebSocket 处理 WebSocket 连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	client := client.NewClient(conn)

	// 启动读取协程（WritePump 已在 NewClient 中启动）
	go client.ReadPump(s.Hub)

	log.Printf("新客户端连接: %s", r.RemoteAddr)
}

// handleRoot 处理根路径
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>聊天服务器</title>
</head>
<body>
    <h1>WebSocket 聊天服务器</h1>
    <p>服务器正在运行...</p>
    <p>请使用客户端应用连接到 ws://` + r.Host + `/ws</p>
</body>
</html>
	`))
}
