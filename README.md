# 基于服务器转发的多客户端数据共享与交换系统

一个基于 WebSocket 的实时聊天系统，支持文本消息、图片、文件传输，以及一对一和组通信。

## 系统架构

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│  前端客户端  │◄───────►│   Go 服务器   │◄───────►│  前端客户端  │
│  (Vue3)    │ WebSocket│  (WebSocket) │ WebSocket│   (Vue3)    │
└─────────────┘         └──────────────┘         └─────────────┘
                               ▲
                               │
                               ▼
                        ┌──────────────┐
                        │  前端客户端   │
                        │   (Vue3)     │
                        └──────────────┘
```

## 技术栈

### 后端
- **语言**: Go 1.21+
- **WebSocket 库**: gorilla/websocket
- **UUID 生成**: google/uuid
- **协议**: 自定义 JSON 协议

### 前端
- **框架**: Vue 3 + Vite
- **UI 组件库**: Element Plus
- **通信**: WebSocket API

## 功能特性

- ✅ 用户登录认证（基于昵称）
- ✅ 实时在线用户列表
- ✅ 一对一私聊
- ✅ 多人组通信（选择多个用户）
- ✅ 文本消息发送
- ✅ 图片发送与预览
- ✅ 文件上传下载（支持大文件分片传输）
- ✅ 用户上下线通知
- ✅ 心跳保活机制
- ✅ 自动重连（前端）

## 项目结构

```
network_lab4/
├── backend/                 # 后端 Go 服务
│   ├── cmd/
│   │   └── server/
│   ├── internal/
│   │   ├── protocol/       # 协议定义
│   │   ├── client/         # 客户端管理
│   │   └── server/         # 服务器核心
│   ├── main.go            # 入口文件
│   └── go.mod
├── frontend/               # 前端 Vue 应用
│   ├── src/
│   │   ├── App.vue        # 主组件
│   │   └── main.js        # 入口文件
│   └── package.json
├── PROTOCOL.md            # 协议设计文档
├── TASKS.md               # 任务分解文档
├── README.md              # 本文件
├── start-backend.sh       # 后端启动脚本
└── start-frontend.sh      # 前端启动脚本
```

## 快速开始

### 前置条件

- Go 1.21 或更高版本
- Node.js 16 或更高版本
- npm 或 yarn

### 安装依赖

#### 后端依赖

```bash
cd backend
go mod tidy
```

#### 前端依赖

```bash
cd frontend
npm install
```

### 启动服务

#### 方式一：使用启动脚本（推荐）

**终端 1 - 启动后端：**
```bash
./start-backend.sh
```

**终端 2 - 启动前端：**
```bash
./start-frontend.sh
```

#### 方式二：手动启动

**启动后端服务器：**
```bash
cd backend
go run main.go
# 或者先编译
go build -o chat-server main.go
./chat-server
```

**启动前端开发服务器：**
```bash
cd frontend
npm run dev
```

### 访问应用

1. 后端服务器启动在: `http://localhost:9090`
2. WebSocket 端点: `ws://localhost:9090/ws`
3. 前端应用访问: `http://localhost:5173`（或终端显示的端口）

## 使用说明

### 登录

1. 打开浏览器访问前端地址
2. 在登录对话框中输入昵称
3. 点击"登录"按钮

### 发送消息

1. 在左侧用户列表中选择一个或多个用户
2. 在底部输入框输入消息
3. 点击"发送"按钮或按 Ctrl+Enter

### 发送图片/文件

1. 选择目标用户
2. 点击"图片"或"文件"按钮
3. 选择要发送的文件
4. 系统自动发送

### 查看消息

- 消息会实时显示在中间消息区域
- 自己的消息显示在右侧（蓝色）
- 其他人的消息显示在左侧（白色）
- 图片可以点击预览
- 文件可以点击下载

## 协议说明

详细的协议设计请参阅 [PROTOCOL.md](PROTOCOL.md)

### 消息类型

- `auth`: 认证消息（登录）
- `message`: 文本和图片消息
- `file`: 文件传输
- `system`: 系统通知
- `userlist`: 在线用户列表
- `heartbeat`: 心跳消息

## 开发说明

### 后端开发

主要模块：
- `internal/protocol`: 协议定义和编解码
- `internal/client`: 客户端连接管理和 Hub
- `internal/server`: WebSocket 服务器

### 前端开发

主要功能：
- WebSocket 连接管理
- 用户界面（Element Plus）
- 消息发送接收
- 文件处理

### 添加新功能

1. 在 `PROTOCOL.md` 中定义新的消息类型
2. 在后端 `protocol/message.go` 中添加相应的结构体
3. 在后端 `client/hub.go` 中添加处理逻辑
4. 在前端 `App.vue` 中添加 UI 和处理函数

## 测试

### 多客户端测试

1. 启动后端服务器
2. 打开多个浏览器窗口/标签页
3. 使用不同昵称登录
4. 测试各种消息类型的发送接收

### 测试场景

- [x] 用户登录/登出
- [x] 在线用户列表更新
- [x] 一对一文本消息
- [x] 多人组消息
- [x] 图片发送接收
- [x] 文件发送接收
- [x] 断线重连
- [x] 心跳保活

## 性能优化

- 文件分片传输（默认 64KB/片）
- WebSocket 消息批量发送
- 心跳间隔 30 秒
- 读写超时保护

## 安全考虑

- 昵称唯一性检查
- 消息大小限制
- 文件大小限制（建议 <100MB）
- 生产环境建议使用 WSS（WebSocket Secure）

## 已知限制

- 消息不持久化（仅内存存储）
- 无用户密码认证
- 无消息历史记录
- 无离线消息

## 未来改进

- [ ] 用户注册/密码登录
- [ ] 消息持久化（数据库）
- [ ] 消息历史记录
- [ ] 离线消息推送
- [ ] 群组管理（创建/加入/退出）
- [ ] 消息搜索功能
- [ ] 端到端加密
- [ ] 表情包支持
- [ ] 消息撤回
- [ ] 已读/未读状态

## 故障排查

### 后端无法启动

- 检查端口 9090 是否被占用
- 查看 Go 版本是否符合要求
- 确认依赖已正确安装

### 前端无法连接

- 确认后端服务已启动
- 检查 WebSocket 地址是否正确
- 查看浏览器控制台错误信息

### 消息发送失败

- 确认已选择接收者
- 检查 WebSocket 连接状态
- 查看后端日志

## 许可证

MIT License

## 作者

本项目为网络编程实验项目

## 参考文档

- [PROTOCOL.md](PROTOCOL.md) - 协议设计文档
- [TASKS.md](TASKS.md) - 任务分解文档
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket 库
- [Element Plus](https://element-plus.org/) - UI 组件库
- [Vue 3](https://vuejs.org/) - 前端框架
