# 应用层协议设计文档

## 1. 概述

本文档定义了基于 WebSocket/TCP 的多客户端通信系统的应用层协议。该协议支持：
- 文本消息传输
- 图片传输
- 任意文件传输
- 一对一通信
- 组通信（多人）
- 用户认证与在线状态管理
- 心跳保活

## 2. 消息格式

所有消息采用 JSON 格式，统一的消息结构如下：

```json
{
  "type": "string",
  "sub_type": "string",
  "from": "string",
  "to": ["string"],
  "group_id": "string",
  "timestamp": 1234567890000,
  "payload": {}
}
```

### 2.1 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 消息大类：auth, message, file, system, userlist, heartbeat |
| sub_type | string | 否 | 消息细分类型，根据 type 不同而不同 |
| from | string | 否 | 发送者用户 ID（服务器下发时填充） |
| to | array | 否 | 接收者用户 ID 列表 |
| group_id | string | 否 | 组通信标识（可选） |
| timestamp | number | 是 | 消息时间戳（Unix 毫秒） |
| payload | object | 是 | 实际业务数据 |

## 3. 消息类型详细定义

### 3.1 认证消息（type: auth）

#### 3.1.1 登录请求（sub_type: login）

**客户端 → 服务器**

```json
{
  "type": "auth",
  "sub_type": "login",
  "timestamp": 1732435200000,
  "payload": {
    "nickname": "用户昵称"
  }
}
```

#### 3.1.2 登录响应（sub_type: login_response）

**服务器 → 客户端**

成功：
```json
{
  "type": "auth",
  "sub_type": "login_response",
  "timestamp": 1732435200000,
  "payload": {
    "success": true,
    "user_id": "uuid-xxxx-xxxx",
    "nickname": "用户昵称",
    "message": "登录成功"
  }
}
```

失败：
```json
{
  "type": "auth",
  "sub_type": "login_response",
  "timestamp": 1732435200000,
  "payload": {
    "success": false,
    "message": "用户名已存在"
  }
}
```

### 3.2 文本消息（type: message, sub_type: text）

**客户端 → 服务器 → 客户端**

```json
{
  "type": "message",
  "sub_type": "text",
  "from": "user_id_1",
  "to": ["user_id_2", "user_id_3"],
  "timestamp": 1732435200000,
  "payload": {
    "text": "消息内容"
  }
}
```

### 3.3 图片消息（type: message, sub_type: image）

图片采用 Base64 编码传输（小图）或文件传输方式（大图）。

```json
{
  "type": "message",
  "sub_type": "image",
  "from": "user_id_1",
  "to": ["user_id_2"],
  "timestamp": 1732435200000,
  "payload": {
    "image_data": "data:image/png;base64,iVBORw0KGgoAAAA...",
    "file_name": "screenshot.png",
    "file_size": 12345
  }
}
```

### 3.4 文件传输（type: file）

#### 3.4.1 文件元数据（sub_type: file_meta）

发送文件前先发送元数据：

```json
{
  "type": "file",
  "sub_type": "file_meta",
  "from": "user_id_1",
  "to": ["user_id_2"],
  "timestamp": 1732435200000,
  "payload": {
    "file_id": "uuid-file-xxxx",
    "file_name": "document.pdf",
    "file_size": 1048576,
    "file_type": "application/pdf",
    "chunk_size": 65536,
    "total_chunks": 16
  }
}
```

#### 3.4.2 文件分片（sub_type: file_chunk）

```json
{
  "type": "file",
  "sub_type": "file_chunk",
  "from": "user_id_1",
  "to": ["user_id_2"],
  "timestamp": 1732435200000,
  "payload": {
    "file_id": "uuid-file-xxxx",
    "chunk_index": 0,
    "total_chunks": 16,
    "chunk_data": "base64编码的分片数据..."
  }
}
```

#### 3.4.3 文件传输完成（sub_type: file_complete）

```json
{
  "type": "file",
  "sub_type": "file_complete",
  "from": "user_id_1",
  "to": ["user_id_2"],
  "timestamp": 1732435200000,
  "payload": {
    "file_id": "uuid-file-xxxx",
    "file_name": "document.pdf",
    "success": true
  }
}
```

### 3.5 在线用户列表（type: userlist）

**服务器 → 客户端**

当有用户登录或登出时，服务器广播更新后的用户列表：

```json
{
  "type": "userlist",
  "timestamp": 1732435200000,
  "payload": {
    "users": [
      {
        "user_id": "uuid-1",
        "nickname": "用户1",
        "login_time": 1732435100000
      },
      {
        "user_id": "uuid-2",
        "nickname": "用户2",
        "login_time": 1732435150000
      }
    ]
  }
}
```

### 3.6 系统消息（type: system）

#### 3.6.1 用户上线通知（sub_type: user_online）

```json
{
  "type": "system",
  "sub_type": "user_online",
  "timestamp": 1732435200000,
  "payload": {
    "user_id": "uuid-1",
    "nickname": "用户1"
  }
}
```

#### 3.6.2 用户下线通知（sub_type: user_offline）

```json
{
  "type": "system",
  "sub_type": "user_offline",
  "timestamp": 1732435200000,
  "payload": {
    "user_id": "uuid-1",
    "nickname": "用户1"
  }
}
```

#### 3.6.3 错误消息（sub_type: error）

```json
{
  "type": "system",
  "sub_type": "error",
  "timestamp": 1732435200000,
  "payload": {
    "code": "USER_NOT_FOUND",
    "message": "目标用户不在线"
  }
}
```

### 3.7 心跳消息（type: heartbeat）

#### 3.7.1 心跳请求

**客户端 → 服务器**

```json
{
  "type": "heartbeat",
  "sub_type": "ping",
  "timestamp": 1732435200000,
  "payload": {}
}
```

#### 3.7.2 心跳响应

**服务器 → 客户端**

```json
{
  "type": "heartbeat",
  "sub_type": "pong",
  "timestamp": 1732435200001,
  "payload": {}
}
```

## 4. 消息流转规则

### 4.1 一对一通信

1. 客户端 A 发送消息，`to` 字段包含一个用户 ID
2. 服务器接收消息，验证目标用户在线
3. 服务器填充 `from` 字段为发送者 ID
4. 服务器转发消息给目标客户端 B

### 4.2 组通信

1. 客户端 A 发送消息，`to` 字段包含多个用户 ID（至少 2 个）
2. 服务器接收消息，验证目标用户
3. 服务器依次向 `to` 列表中的所有在线用户转发消息
4. 可选：服务器维护组 ID（`group_id`），简化后续消息发送

### 4.3 文件传输流程

1. 发送方发送文件元数据（`file_meta`）
2. 接收方收到元数据，准备接收
3. 发送方按序发送文件分片（`file_chunk`）
4. 接收方接收并组装分片
5. 传输完成后发送完成消息（`file_complete`）

### 4.4 用户状态管理

1. 客户端连接后发送登录请求
2. 服务器验证并分配 user_id
3. 服务器广播用户上线通知和更新用户列表
4. 客户端断开连接时，服务器广播下线通知和更新用户列表
5. 服务器定期检查心跳，超时则认为用户离线

## 5. 错误处理

### 5.1 错误码定义

| 错误码 | 说明 |
|--------|------|
| INVALID_MESSAGE | 消息格式错误 |
| NOT_AUTHENTICATED | 未认证 |
| USER_NOT_FOUND | 目标用户不存在或不在线 |
| NICKNAME_EXISTS | 昵称已被使用 |
| FILE_TOO_LARGE | 文件超过大小限制 |
| INVALID_RECIPIENT | 无效的接收者列表 |

### 5.2 错误响应示例

```json
{
  "type": "system",
  "sub_type": "error",
  "timestamp": 1732435200000,
  "payload": {
    "code": "USER_NOT_FOUND",
    "message": "用户 user_id_2 不在线",
    "original_message_type": "message"
  }
}
```

## 6. 安全与限制

### 6.1 大小限制

- 单条文本消息：不超过 10KB
- 图片消息（Base64）：不超过 5MB
- 文件大小：不超过 100MB
- 分片大小：64KB

### 6.2 频率限制

- 心跳间隔：30 秒
- 心跳超时：90 秒（3 次心跳未收到）
- 消息发送频率：建议不超过 10 条/秒

### 6.3 安全建议

- 生产环境应使用 WSS（WebSocket Secure）
- 可扩展添加消息签名验证
- 可扩展添加端到端加密

## 7. 协议版本

当前版本：**v1.0**

后续版本更新需保持向后兼容，或在消息中添加 `version` 字段标识协议版本。
