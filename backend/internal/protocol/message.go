package protocol

import (
	"encoding/json"
	"time"
)

// MessageType 消息类型
type MessageType string

const (
	TypeAuth      MessageType = "auth"
	TypeMessage   MessageType = "message"
	TypeFile      MessageType = "file"
	TypeSystem    MessageType = "system"
	TypeUserList  MessageType = "userlist"
	TypeHeartbeat MessageType = "heartbeat"
)

// SubType 消息子类型
type SubType string

const (
	SubTypeLogin         SubType = "login"
	SubTypeLoginResponse SubType = "login_response"
	SubTypeText          SubType = "text"
	SubTypeImage         SubType = "image"
	SubTypeFileMeta      SubType = "file_meta"
	SubTypeFileChunk     SubType = "file_chunk"
	SubTypeFileComplete  SubType = "file_complete"
	SubTypeUserOnline    SubType = "user_online"
	SubTypeUserOffline   SubType = "user_offline"
	SubTypeError         SubType = "error"
	SubTypePing          SubType = "ping"
	SubTypePong          SubType = "pong"
)

// Message 统一消息结构
type Message struct {
	Type      MessageType            `json:"type"`
	SubType   SubType                `json:"sub_type,omitempty"`
	From      string                 `json:"from,omitempty"`
	To        []string               `json:"to,omitempty"`
	GroupID   string                 `json:"group_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, subType SubType) *Message {
	return &Message{
		Type:      msgType,
		SubType:   subType,
		Timestamp: time.Now().UnixMilli(),
		Payload:   make(map[string]interface{}),
	}
}

// Encode 编码消息为 JSON 字节流
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode 解码 JSON 字节流为消息
func Decode(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// UserInfo 用户信息
type UserInfo struct {
	UserID    string `json:"user_id"`
	Nickname  string `json:"nickname"`
	LoginTime int64  `json:"login_time"`
}

// 错误码常量
const (
	ErrInvalidMessage   = "INVALID_MESSAGE"
	ErrNotAuthenticated = "NOT_AUTHENTICATED"
	ErrUserNotFound     = "USER_NOT_FOUND"
	ErrNicknameExists   = "NICKNAME_EXISTS"
	ErrFileTooLarge     = "FILE_TOO_LARGE"
	ErrInvalidRecipient = "INVALID_RECIPIENT"
)
