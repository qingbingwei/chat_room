package client

import (
	"chat-server/internal/protocol"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client 客户端连接
type Client struct {
	ID            string
	Nickname      string
	Conn          *websocket.Conn
	Send          chan []byte
	LoginTime     int64
	LastHeartbeat int64
	mu            sync.RWMutex
}

// NewClient 创建新客户端
func NewClient(conn *websocket.Conn) *Client {
	client := &Client{
		Conn:          conn,
		Send:          make(chan []byte, 256),
		LastHeartbeat: time.Now().Unix(),
	}

	// 启动写入协程（异步）
	go client.WritePump()

	return client
}

// ReadPump 读取客户端消息
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("错误: %v", err)
			}
			break
		}

		log.Printf("收到原始消息: %s", string(message))

		// 解析消息
		msg, err := protocol.Decode(message)
		if err != nil {
			log.Printf("消息解析错误: %v", err)
			continue
		}

		log.Printf("解析后消息类型: %s, 子类型: %s", msg.Type, msg.SubType)

		// 更新心跳时间
		if msg.Type == protocol.TypeHeartbeat {
			c.mu.Lock()
			c.LastHeartbeat = time.Now().Unix()
			c.mu.Unlock()
		}

		// 处理消息
		hub.HandleMessage <- &ClientMessage{
			Client:  c,
			Message: msg,
		}
	}
}

// WritePump 向客户端发送消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		log.Printf("[WritePump] 用户 %s (%s) WritePump 已退出", c.Nickname, c.ID)
	}()

	log.Printf("[WritePump] 用户 %s (%s) WritePump 已启动", c.Nickname, c.ID)

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				log.Printf("[WritePump] Send 通道已关闭，退出")
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("[WritePump] 用户 %s (%s) 开始写入消息: %s", c.Nickname, c.ID, string(message))

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("[WritePump] NextWriter 错误: %v", err)
				return
			}
			w.Write(message)

			// 批量发送队列中的其他消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				log.Printf("[WritePump] Close 错误: %v", err)
				return
			}

			log.Printf("[WritePump] 用户 %s (%s) 消息写入完成", c.Nickname, c.ID)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage 发送消息到客户端
func (c *Client) SendMessage(msg *protocol.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	log.Printf("[SendMessage] 尝试发送消息到用户 %s (昵称: %s), 类型: %s, 子类型: %s", c.ID, c.Nickname, msg.Type, msg.SubType)

	// 对于认证响应等关键消息，使用阻塞发送确保消息不会丢失
	if msg.Type == protocol.TypeAuth {
		log.Printf("[SendMessage] 关键消息，使用阻塞发送")
		c.Send <- data
		log.Printf("[SendMessage] 消息已加入发送队列（阻塞）")
		return nil
	}

	// 对于其他消息，使用非阻塞发送避免死锁
	select {
	case c.Send <- data:
		log.Printf("[SendMessage] 消息已加入发送队列")
	default:
		log.Printf("[SendMessage] 发送队列已满，消息被丢弃")
	}

	return nil
}

// GetUserInfo 获取用户信息
func (c *Client) GetUserInfo() protocol.UserInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return protocol.UserInfo{
		UserID:    c.ID,
		Nickname:  c.Nickname,
		LoginTime: c.LoginTime,
	}
}

// ClientMessage 客户端消息包装
type ClientMessage struct {
	Client  *Client
	Message *protocol.Message
}

// Hub 客户端管理中心
type Hub struct {
	Clients       map[string]*Client
	NicknameMap   map[string]*Client
	Register      chan *Client
	Unregister    chan *Client
	HandleMessage chan *ClientMessage
	mu            sync.RWMutex
}

// NewHub 创建客户端管理中心
func NewHub() *Hub {
	return &Hub{
		Clients:       make(map[string]*Client),
		NicknameMap:   make(map[string]*Client),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		HandleMessage: make(chan *ClientMessage),
	}
}

// Run 运行客户端管理中心
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			if client.Nickname != "" {
				h.NicknameMap[client.Nickname] = client
			}
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				log.Printf("[Unregister] 用户 %s (%s) 断开连接", client.Nickname, client.ID)
				delete(h.Clients, client.ID)
				delete(h.NicknameMap, client.Nickname)
				close(client.Send)
				h.mu.Unlock()

				// 广播用户下线通知
				if client.ID != "" {
					h.BroadcastUserOffline(client)
					h.BroadcastUserList()
				}
			} else {
				log.Printf("[Unregister] 尝试注销不存在的客户端: %s (%s)", client.Nickname, client.ID)
				h.mu.Unlock()
			}

		case clientMsg := <-h.HandleMessage:
			h.processMessage(clientMsg)
		}
	}
}

// processMessage 处理客户端消息
func (h *Hub) processMessage(clientMsg *ClientMessage) {
	client := clientMsg.Client
	msg := clientMsg.Message

	switch msg.Type {
	case protocol.TypeAuth:
		h.handleAuth(client, msg)
	case protocol.TypeMessage, protocol.TypeFile, protocol.TypeSystem:
		// 消息、文件和系统通知都需要转发
		h.handleForward(client, msg)
	case protocol.TypeHeartbeat:
		h.handleHeartbeat(client, msg)
	default:
		log.Printf("[processMessage] 未知消息类型: %s", msg.Type)
	}
}

// handleAuth 处理认证消息
func (h *Hub) handleAuth(client *Client, msg *protocol.Message) {
	log.Printf("处理认证消息, SubType: %s", msg.SubType)
	if msg.SubType == protocol.SubTypeLogin {
		log.Printf("收到登录请求, Payload: %+v", msg.Payload)
		nickname, ok := msg.Payload["nickname"].(string)
		if !ok || nickname == "" {
			log.Printf("昵称为空或无效")
			h.sendError(client, protocol.ErrInvalidMessage, "昵称不能为空")
			return
		}
		log.Printf("用户 %s 尝试登录", nickname)

		// 检查昵称是否已存在
		h.mu.RLock()
		_, exists := h.NicknameMap[nickname]
		h.mu.RUnlock()

		if exists {
			response := protocol.NewMessage(protocol.TypeAuth, protocol.SubTypeLoginResponse)
			response.Payload = map[string]interface{}{
				"success": false,
				"message": "昵称已被使用",
			}
			client.SendMessage(response)
			return
		}

		// 生成用户 ID
		client.ID = uuid.New().String()
		client.Nickname = nickname
		client.LoginTime = time.Now().UnixMilli()

		// 直接注册客户端（同步）
		h.mu.Lock()
		h.Clients[client.ID] = client
		h.NicknameMap[client.Nickname] = client
		h.mu.Unlock()

		log.Printf("用户 %s (ID: %s) 注册成功", client.Nickname, client.ID)

		// 发送登录成功响应
		response := protocol.NewMessage(protocol.TypeAuth, protocol.SubTypeLoginResponse)
		response.Payload = map[string]interface{}{
			"success":  true,
			"user_id":  client.ID,
			"nickname": client.Nickname,
			"message":  "登录成功",
		}
		log.Printf("准备发送登录响应: %+v", response)
		err := client.SendMessage(response)
		if err != nil {
			log.Printf("发送登录响应失败: %v", err)
		} else {
			log.Printf("登录响应已发送到队列")
		}

		// 广播用户上线通知
		h.BroadcastUserOnline(client)

		// 延迟广播用户列表，确保客户端已处理登录响应
		go func() {
			time.Sleep(50 * time.Millisecond)
			h.BroadcastUserList()
		}()
	}
}

// handleForward 处理消息转发
func (h *Hub) handleForward(sender *Client, msg *protocol.Message) {
	if sender.ID == "" {
		h.sendError(sender, protocol.ErrNotAuthenticated, "未认证")
		return
	}

	// 设置发送者
	msg.From = sender.ID

	log.Printf("[handleForward] 转发消息: 类型=%s, 子类型=%s, 发送者=%s, 接收者数=%d",
		msg.Type, msg.SubType, sender.Nickname, len(msg.To))

	// 转发消息给目标用户
	if len(msg.To) == 0 {
		log.Printf("[handleForward] 接收者列表为空")
		h.sendError(sender, protocol.ErrInvalidRecipient, "接收者列表为空")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	successCount := 0
	for _, targetID := range msg.To {
		// 不转发给发送者自己（前端已经本地添加了）
		if targetID == sender.ID {
			log.Printf("[handleForward] 跳过发送者自己: %s (%s)", sender.Nickname, targetID)
			continue
		}

		if targetClient, ok := h.Clients[targetID]; ok {
			log.Printf("[handleForward] 转发到: %s (%s)", targetClient.Nickname, targetID)
			targetClient.SendMessage(msg)
			successCount++
		} else {
			log.Printf("[handleForward] 目标用户不存在: %s", targetID)
		}
	}
	log.Printf("[handleForward] 成功转发给 %d/%d 个用户（不含发送者）", successCount, len(msg.To))
}

// handleHeartbeat 处理心跳消息
func (h *Hub) handleHeartbeat(client *Client, msg *protocol.Message) {
	if msg.SubType == protocol.SubTypePing {
		response := protocol.NewMessage(protocol.TypeHeartbeat, protocol.SubTypePong)
		client.SendMessage(response)
	}
}

// BroadcastUserList 广播用户列表
func (h *Hub) BroadcastUserList() {
	h.mu.RLock()
	users := make([]protocol.UserInfo, 0, len(h.Clients))
	for _, client := range h.Clients {
		if client.ID != "" {
			users = append(users, client.GetUserInfo())
		}
	}
	h.mu.RUnlock()

	log.Printf("[BroadcastUserList] 广播用户列表，共 %d 个用户", len(users))

	msg := protocol.NewMessage(protocol.TypeUserList, "")
	msg.Payload = map[string]interface{}{
		"users": users,
	}

	data, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		if client.ID != "" {
			log.Printf("[BroadcastUserList] 发送用户列表到: %s (%s)", client.Nickname, client.ID)
			select {
			case client.Send <- data:
				log.Printf("[BroadcastUserList] 成功发送到: %s", client.Nickname)
			default:
				log.Printf("[BroadcastUserList] 发送失败（队列已满）: %s", client.Nickname)
			}
		}
	}
}

// BroadcastUserOnline 广播用户上线通知
func (h *Hub) BroadcastUserOnline(client *Client) {
	log.Printf("[BroadcastUserOnline] 广播用户上线: %s (%s)", client.Nickname, client.ID)

	msg := protocol.NewMessage(protocol.TypeSystem, protocol.SubTypeUserOnline)
	msg.Payload = map[string]interface{}{
		"user_id":  client.ID,
		"nickname": client.Nickname,
	}

	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()

	broadcastCount := 0
	successCount := 0
	for _, c := range h.Clients {
		if c.ID != "" && c.ID != client.ID {
			log.Printf("[BroadcastUserOnline] 发送上线通知到: %s (%s)", c.Nickname, c.ID)
			select {
			case c.Send <- data:
				broadcastCount++
				successCount++
				log.Printf("[BroadcastUserOnline] 成功发送到: %s", c.Nickname)
			default:
				broadcastCount++
				log.Printf("[BroadcastUserOnline] 发送失败（队列已满）: %s", c.Nickname)
			}
		}
	}
	log.Printf("[BroadcastUserOnline] 上线通知已发送给 %d/%d 个用户", successCount, broadcastCount)

	// 延迟再次广播用户列表，确保所有客户端同步
	if successCount > 0 {
		go func() {
			time.Sleep(100 * time.Millisecond)
			log.Printf("[BroadcastUserOnline] 延迟广播用户列表以确保同步")
			h.BroadcastUserList()
		}()
	}
}

// BroadcastUserOffline 广播用户下线通知
func (h *Hub) BroadcastUserOffline(client *Client) {
	if client.ID == "" {
		return
	}

	msg := protocol.NewMessage(protocol.TypeSystem, protocol.SubTypeUserOffline)
	msg.Payload = map[string]interface{}{
		"user_id":  client.ID,
		"nickname": client.Nickname,
	}

	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.Clients {
		if c.ID != "" && c.ID != client.ID {
			select {
			case c.Send <- data:
			default:
			}
		}
	}
}

// sendError 发送错误消息
func (h *Hub) sendError(client *Client, code, message string) {
	msg := protocol.NewMessage(protocol.TypeSystem, protocol.SubTypeError)
	msg.Payload = map[string]interface{}{
		"code":    code,
		"message": message,
	}
	client.SendMessage(msg)
}

// GetOnlineCount 获取在线用户数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, client := range h.Clients {
		if client.ID != "" {
			count++
		}
	}
	return count
}
