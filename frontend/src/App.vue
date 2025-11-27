<template>
  <div id="app">
    <!-- 登录对话框 -->
    <el-dialog
      v-model="showLogin"
      title="登录聊天室"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <el-form :model="loginForm" label-width="80px">
        <el-form-item label="昵称">
          <el-input
            v-model="loginForm.nickname"
            placeholder="请输入昵称"
            @keyup.enter="handleLogin"
            maxlength="20"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" @click="handleLogin" :loading="isConnecting">
          {{ isConnecting ? '连接中...' : '登录' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 主界面 -->
    <div v-if="isLoggedIn" class="chat-container">
      <!-- 头部 -->
      <div class="chat-header">
        <div class="header-left">
          <el-icon><ChatDotRound /></el-icon>
          <span class="title">聊天室</span>
        </div>
        <div class="header-right">
          <span class="user-info">欢迎，{{ currentUser.nickname }}</span>
          <el-button size="small" @click="handleLogout">退出</el-button>
        </div>
      </div>

      <!-- 主内容区 -->
      <div class="chat-content">
        <!-- 左侧：联系人列表 -->
        <div class="contact-list">
          <!-- 群组列表 -->
          <div class="section">
            <div class="section-header">
              <el-icon><ChatDotRound /></el-icon>
              <span>群组 ({{ groups.length }})</span>
            </div>
            <div class="section-body">
              <div
                v-for="group in groups"
                :key="group.group_id"
                :class="['contact-item', 'group-item', { active: currentGroup?.group_id === group.group_id }]"
                @click="selectGroup(group)"
              >
                <el-icon class="contact-icon"><ChatDotRound /></el-icon>
                <div class="contact-info">
                  <div class="contact-name">{{ group.group_name }}</div>
                  <div class="contact-detail">{{ group.members.length }}人</div>
                </div>
                <el-badge
                  v-if="unreadCounts[group.group_id] > 0"
                  :value="unreadCounts[group.group_id]"
                  :max="99"
                  class="unread-badge"
                />
              </div>
              <div v-if="groups.length === 0" class="empty-hint">
                选择多个用户创建群组
              </div>
            </div>
          </div>

          <!-- 在线用户列表 -->
          <div class="section">
            <div class="section-header">
              <el-icon><User /></el-icon>
              <span>在线用户 ({{ onlineUsers.length }})</span>
            </div>
            <div class="section-body">
              <div
                v-for="user in onlineUsers"
                :key="user.user_id"
                :class="['contact-item', { 
                  selected: isUserSelected(user.user_id),
                  active: !currentGroup && currentPrivateChat?.user_id === user.user_id
                }]"
                @click="handleUserClick(user, $event)"
                @contextmenu.prevent="handleUserClick(user, $event)"
              >
                <el-icon class="contact-icon"><User /></el-icon>
                <div class="contact-info">
                  <div class="contact-name">{{ user.nickname }}</div>
                </div>
                <el-badge
                  v-if="unreadCounts[user.user_id] > 0 && !isUserSelected(user.user_id)"
                  :value="unreadCounts[user.user_id]"
                  :max="99"
                  class="unread-badge"
                />
                <el-icon v-if="isUserSelected(user.user_id)" class="check-icon">
                  <Check />
                </el-icon>
              </div>
            </div>
          </div>

          <!-- 操作区 -->
          <div class="action-footer">
            <div v-if="isCreatingGroup && selectedUsers.length > 0" class="selection-info">
              <el-text size="small" type="info">
                已选择 {{ selectedUsers.length }} 人
              </el-text>
              <el-button size="small" text @click="cancelCreateGroup">
                取消
              </el-button>
            </div>
            <el-button
              v-if="isCreatingGroup && selectedUsers.length >= 2"
              type="primary"
              size="small"
              style="width: 100%"
              @click="confirmCreateGroup"
            >
              确认创建
            </el-button>
            <el-button
              v-else-if="!isCreatingGroup"
              type="success"
              size="small"
              style="width: 100%"
              @click="startCreateGroup"
            >
              <el-icon style="margin-right: 4px"><Plus /></el-icon>
              创建群组
            </el-button>
            <div v-else-if="isCreatingGroup && selectedUsers.length === 0" class="hint-text">
              请选择至少2个用户
            </div>
            <div v-else-if="isCreatingGroup && selectedUsers.length === 1" class="hint-text">
              已选1人，再选至少1人
            </div>
          </div>
        </div>

        <!-- 中间：消息显示区 -->
        <div class="message-area">
          <!-- 当前聊天对象显示 -->
          <div class="chat-target-header">
            <div v-if="currentGroup" class="target-info">
              <el-icon class="target-icon" color="#67c23a"><ChatDotRound /></el-icon>
              <div class="target-details">
                <div class="target-name">{{ currentGroup.group_name }}</div>
                <div class="target-members">群聊 · {{ currentGroup.members.length }}人</div>
              </div>
            </div>
            <div v-else-if="currentPrivateChat" class="target-info">
              <el-icon class="target-icon" color="#409eff"><User /></el-icon>
              <div class="target-details">
                <div class="target-name">{{ currentPrivateChat.nickname }}</div>
                <div class="target-members">私聊</div>
              </div>
            </div>
            <div v-else class="target-info empty">
              <el-icon class="target-icon" color="#c0c4cc"><ChatDotRound /></el-icon>
              <div class="target-details">
                <div class="target-name">请选择聊天对象</div>
                <div class="target-members">从左侧选择用户或群组</div>
              </div>
            </div>
          </div>
          <div class="message-list" ref="messageListRef">
            <div
              v-for="(msg, index) in filteredMessages"
              :key="index"
              :class="['message-item', msg.from === currentUser.user_id ? 'self' : 'other']"
            >
              <div class="message-header">
                <span class="message-nickname">{{ getUserNickname(msg.from) }}</span>
                <span class="message-time">{{ formatTime(msg.timestamp) }}</span>
              </div>
              <div class="message-content">
                <!-- 文本消息 -->
                <div v-if="msg.type === 'message' && msg.sub_type === 'text'" class="message-text">
                  {{ msg.payload.text }}
                </div>
                <!-- 图片消息 -->
                <div v-else-if="msg.type === 'message' && msg.sub_type === 'image'" class="message-image">
                  <el-image
                    :src="msg.payload.image_data"
                    :preview-src-list="[msg.payload.image_data]"
                    fit="contain"
                    style="max-width: 300px; max-height: 300px;"
                  />
                </div>
                <!-- 文件消息 -->
                <div v-else-if="msg.type === 'file'" class="message-file">
                  <el-icon><Document /></el-icon>
                  <span>{{ msg.payload.file_name }}</span>
                  <el-button
                    size="small"
                    text
                    @click="downloadFile(msg)"
                  >
                    下载
                  </el-button>
                </div>
                <!-- 系统消息 -->
                <div v-else-if="msg.type === 'system'" class="message-system">
                  {{ msg.payload.message || formatSystemMessage(msg) }}
                </div>
              </div>
            </div>
          </div>

          <!-- 消息输入区 -->
          <div class="message-input">
            <div class="input-toolbar">
              <el-upload
                :auto-upload="false"
                :show-file-list="false"
                :on-change="handleFileSelect"
                accept="image/*"
              >
                <el-button size="small" :icon="Picture">图片</el-button>
              </el-upload>
              <el-upload
                :auto-upload="false"
                :show-file-list="false"
                :on-change="handleFileSelect"
              >
                <el-button size="small" :icon="Document">文件</el-button>
              </el-upload>
            </div>
            <div class="input-area">
              <el-input
                v-model="messageInput"
                type="textarea"
                :rows="3"
                :placeholder="getInputPlaceholder()"
                @keydown.ctrl.enter="sendTextMessage"
                :disabled="!currentGroup && !currentPrivateChat"
              />
              <el-button
                type="primary"
                :disabled="!messageInput.trim() || (!currentGroup && !currentPrivateChat)"
                @click="sendTextMessage"
              >
                发送
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  ChatDotRound,
  User,
  Picture,
  Document,
  Check,
  Plus
} from '@element-plus/icons-vue'

// WebSocket 连接
let ws = null
const isConnecting = ref(false)
const isLoggedIn = ref(false)
const showLogin = ref(true)

// 用户信息
const loginForm = reactive({
  nickname: ''
})
const currentUser = reactive({
  user_id: '',
  nickname: ''
})

// 在线用户列表
const onlineUsers = ref([])
const selectedUsers = ref([])

// 群组列表
const groups = ref([])
const currentGroup = ref(null)
const currentPrivateChat = ref(null) // 当前私聊对象
const isCreatingGroup = ref(false) // 是否处于创建群组模式

// 消息列表
const messages = ref([])
const messageInput = ref('')
const messageListRef = ref(null)

// 未读消息计数 { group_id 或 user_id: count }
const unreadCounts = ref({})

// 文件接收缓存
const fileChunks = new Map()

// 过滤后的消息（只显示当前聊天对象的消息）
const filteredMessages = computed(() => {
  if (currentGroup.value) {
    // 群组消息：group_id 匹配或者 to 列表包含当前用户且是群组成员
    return messages.value.filter(msg => {
      if (msg.group_id === currentGroup.value.group_id) return true
      if (msg.type === 'system' && msg.sub_type === 'group_created' && 
          msg.payload?.group_id === currentGroup.value.group_id) return true
      return false
    })
  } else if (currentPrivateChat.value) {
    // 私聊消息：发送者或接收者是当前私聊对象
    return messages.value.filter(msg => {
      if (msg.type === 'system') return false // 私聊不显示系统消息
      const isFromTarget = msg.from === currentPrivateChat.value.user_id
      const isToTarget = msg.from === currentUser.user_id && 
                        msg.to?.includes(currentPrivateChat.value.user_id)
      return isFromTarget || isToTarget
    })
  }
  return [] // 未选择聊天对象时不显示任何消息
})

// 连接 WebSocket
const connectWebSocket = () => {
  return new Promise((resolve, reject) => {
    ws = new WebSocket('ws://localhost:9090/ws')

    ws.onopen = () => {
      console.log('WebSocket 连接已建立')
      resolve()
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        handleMessage(msg)
      } catch (error) {
        console.error('消息解析错误:', error)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket 错误:', error)
      isConnecting.value = false
      ElMessage.error('连接错误')
      reject(error)
    }

    ws.onclose = () => {
      console.log('WebSocket 连接已关闭')
      isLoggedIn.value = false
      isConnecting.value = false
      if (!showLogin.value) {
        ElMessage.warning('连接已断开')
        showLogin.value = true
      }
    }
  })
}

// 处理接收到的消息
const handleMessage = (msg) => {
  console.log('收到消息:', msg)

  switch (msg.type) {
    case 'auth':
      handleAuthMessage(msg)
      break
    case 'message':
      messages.value.push(msg)
      // 增加未读消息计数（如果不是当前聊天对象）
      if (msg.group_id) {
        // 群聊消息
        if (!currentGroup.value || currentGroup.value.group_id !== msg.group_id) {
          unreadCounts.value[msg.group_id] = (unreadCounts.value[msg.group_id] || 0) + 1
        }
      } else if (msg.from !== currentUser.user_id) {
        // 私聊消息（非自己发送的）
        if (!currentPrivateChat.value || currentPrivateChat.value.user_id !== msg.from) {
          unreadCounts.value[msg.from] = (unreadCounts.value[msg.from] || 0) + 1
        }
      }
      scrollToBottom()
      break
    case 'file':
      handleFileMessage(msg)
      break
    case 'system':
      handleSystemMessage(msg)
      break
    case 'userlist':
      handleUserListMessage(msg)
      break
    case 'heartbeat':
      // 心跳响应，无需处理
      break
  }
}

// 处理认证消息
const handleAuthMessage = (msg) => {
  console.log('收到认证消息:', msg)
  if (msg.sub_type === 'login_response') {
    console.log('登录响应 - 成功:', msg.payload.success)
    if (msg.payload.success) {
      currentUser.user_id = msg.payload.user_id
      currentUser.nickname = msg.payload.nickname
      isLoggedIn.value = true
      showLogin.value = false
      isConnecting.value = false
      console.log('登录成功，用户ID:', currentUser.user_id)
      ElMessage.success('登录成功')
    } else {
      console.log('登录失败:', msg.payload.message)
      ElMessage.error(msg.payload.message)
      isConnecting.value = false
    }
  }
}

// 处理文件消息
const handleFileMessage = (msg) => {
  if (msg.sub_type === 'file_meta') {
    // 接收文件元数据
    fileChunks.set(msg.payload.file_id, {
      meta: msg.payload,
      chunks: new Array(msg.payload.total_chunks),
      from: msg.from
    })
    messages.value.push({
      ...msg,
      type: 'file',
      payload: {
        ...msg.payload,
        status: 'receiving'
      }
    })
  } else if (msg.sub_type === 'file_chunk') {
    // 接收文件分片
    const fileData = fileChunks.get(msg.payload.file_id)
    if (fileData) {
      fileData.chunks[msg.payload.chunk_index] = msg.payload.chunk_data
      
      // 检查是否接收完成
      if (fileData.chunks.every(chunk => chunk !== undefined)) {
        // 所有分片接收完成
        const allData = fileData.chunks.join('')
        messages.value.push({
          type: 'file',
          sub_type: 'file_complete',
          from: fileData.from,
          timestamp: Date.now(),
          payload: {
            file_name: fileData.meta.file_name,
            file_data: allData
          }
        })
        fileChunks.delete(msg.payload.file_id)
      }
    }
  }
}

// 处理系统消息
const handleSystemMessage = (msg) => {
  console.log('处理系统消息:', msg.sub_type, msg.payload)
  
  if (msg.sub_type === 'error') {
    ElMessage.error(msg.payload.message)
  } else if (msg.sub_type === 'user_online') {
    ElNotification({
      title: '用户上线',
      message: `${msg.payload.nickname} 加入了聊天室`,
      type: 'success',
      duration: 2000
    })
    // 延迟请求用户列表，确保后端已完成注册
    setTimeout(() => {
      console.log('收到user_online后主动请求用户列表')
      // 服务器会自动广播，但我们可以通过心跳触发
    }, 100)
  } else if (msg.sub_type === 'user_offline') {
    ElNotification({
      title: '用户下线',
      message: `${msg.payload.nickname} 离开了聊天室`,
      type: 'info',
      duration: 2000
    })
  } else if (msg.sub_type === 'group_created') {
    // 收到群组创建通知
    console.log('收到群组创建通知:', msg.payload)
    
    // 从 payload 中获取成员信息
    const memberIds = msg.payload.member_ids || []
    const memberNicknames = msg.payload.members || []
    
    const group = {
      group_id: msg.payload.group_id,
      group_name: msg.payload.group_name,
      members: memberIds.map((id, index) => ({
        user_id: id,
        nickname: memberNicknames[index] || ''
      })),
      creator: msg.payload.creator,
      created_at: Date.now()
    }
    
    // 检查是否已存在该群组（避免重复添加）
    if (!groups.value.find(g => g.group_id === group.group_id)) {
      groups.value.push(group)
      ElNotification({
        title: '群组创建',
        message: `${msg.payload.creator} 创建了群组"${msg.payload.group_name}"`,
        type: 'success',
        duration: 3000
      })
    }
    // 注意：不将 group_created 系统消息添加到消息列表
    return
  }
  messages.value.push(msg)
}

// 处理用户列表消息
const handleUserListMessage = (msg) => {
  console.log('收到用户列表:', msg.payload.users)
  console.log('当前用户ID:', currentUser.user_id)
  
  if (!msg.payload.users || !Array.isArray(msg.payload.users)) {
    console.warn('无效的用户列表数据')
    return
  }
  
  // 过滤掉当前用户
  const filteredUsers = msg.payload.users.filter(u => {
    return u.user_id !== currentUser.user_id
  })
  
  console.log('过滤后的在线用户:', filteredUsers)
  onlineUsers.value = filteredUsers
}

// 登录
const handleLogin = async () => {
  if (!loginForm.nickname.trim()) {
    ElMessage.warning('请输入昵称')
    return
  }

  // 确保之前的连接已关闭
  if (ws && ws.readyState !== WebSocket.CLOSED) {
    ws.close()
  }

  isConnecting.value = true

  try {
    await connectWebSocket()

    // 发送登录请求
    const loginMsg = {
      type: 'auth',
      sub_type: 'login',
      timestamp: Date.now(),
      payload: {
        nickname: loginForm.nickname
      }
    }
    console.log('发送登录请求:', loginMsg)
    ws.send(JSON.stringify(loginMsg))
    console.log('登录请求已发送')

    // 启动心跳
    startHeartbeat()

    // 设置登录超时保护（10秒）
    setTimeout(() => {
      if (isConnecting.value) {
        console.warn('登录超时，重置连接状态')
        isConnecting.value = false
        ElMessage.error('登录超时，请重试')
        if (ws) {
          ws.close()
        }
      }
    }, 10000)
  } catch (error) {
    console.error('登录错误:', error)
    isConnecting.value = false
    ElMessage.error('连接失败，请重试')
  }
}

// 登出
const handleLogout = () => {
  if (ws) {
    ws.close()
  }
  isLoggedIn.value = false
  isConnecting.value = false
  showLogin.value = true
  selectedUsers.value = []
  messages.value = []
  onlineUsers.value = []
  groups.value = []
  currentGroup.value = null
  currentUser.user_id = ''
  currentUser.nickname = ''
  stopHeartbeat()
}

// 心跳定时器
let heartbeatTimer = null

const startHeartbeat = () => {
  heartbeatTimer = setInterval(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'heartbeat',
        sub_type: 'ping',
        timestamp: Date.now(),
        payload: {}
      }))
    }
  }, 30000)
}

const stopHeartbeat = () => {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
    heartbeatTimer = null
  }
}

// 用户选择
const isUserSelected = (userId) => {
  return selectedUsers.value.some(u => u.user_id === userId)
}

// 处理用户点击
const handleUserClick = (user, event) => {
  // 如果处于创建群组模式，只允许切换选中状态
  if (isCreatingGroup.value) {
    const index = selectedUsers.value.findIndex(u => u.user_id === user.user_id)
    if (index > -1) {
      selectedUsers.value.splice(index, 1)
    } else {
      selectedUsers.value.push(user)
    }
    return
  }
  
  // 非创建群组模式：点击进入私聊
  currentGroup.value = null
  currentPrivateChat.value = user
  selectedUsers.value = []
  // 清零该用户的未读消息
  if (unreadCounts.value[user.user_id]) {
    unreadCounts.value[user.user_id] = 0
  }
}

const clearSelection = () => {
  selectedUsers.value = []
  currentGroup.value = null
  currentPrivateChat.value = null
}

// 开始创建群组
const startCreateGroup = () => {
  isCreatingGroup.value = true
  selectedUsers.value = []
  currentGroup.value = null
  currentPrivateChat.value = null
  ElMessage.info('请选择群组成员（至少2人）')
}

// 取消创建群组
const cancelCreateGroup = () => {
  isCreatingGroup.value = false
  selectedUsers.value = []
}

// 确认创建群组
const confirmCreateGroup = () => {
  if (selectedUsers.value.length < 2) {
    ElMessage.warning('至少需要选择2个用户才能创建群组')
    return
  }
  
  ElMessageBox.prompt('请输入群组名称', '创建群组', {
    confirmButtonText: '创建',
    cancelButtonText: '取消',
    inputPattern: /.+/,
    inputErrorMessage: '群组名称不能为空'
  }).then(({ value }) => {
    createGroupWithName(value)
  }).catch(() => {
    // 取消创建
  })
}

// 选择群组
const selectGroup = (group) => {
  currentGroup.value = group
  currentPrivateChat.value = null
  selectedUsers.value = []
  // 清零该群组的未读消息
  if (unreadCounts.value[group.group_id]) {
    unreadCounts.value[group.group_id] = 0
  }
}

// 创建群组（实际执行创建）
const createGroupWithName = (groupName) => {
  const groupId = `group_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    
    // 群组成员包括选中的用户和创建者自己
    const allMembers = [
      { user_id: currentUser.user_id, nickname: currentUser.nickname },
      ...selectedUsers.value.filter(u => u.user_id !== currentUser.user_id) // 避免重复
    ]
    
    const group = {
      group_id: groupId,
      group_name: value,
      members: allMembers,
      creator: currentUser.user_id,
      created_at: Date.now()
    }
    
    groups.value.push(group)
    currentGroup.value = group
    selectedUsers.value = [] // 清除选中用户
    isCreatingGroup.value = false // 退出创建群组模式
    
    ElMessage.success(`群组"${groupName}"创建成功`)
    
    console.log('发送群组创建通知，所有成员:', allMembers)
    
    // 向群组所有成员（包括自己）发送群组创建通知
    const msg = {
      type: 'system',
      sub_type: 'group_created',
      to: allMembers.map(u => u.user_id),
      timestamp: Date.now(),
      payload: {
        group_id: groupId,
        group_name: groupName,
        creator: currentUser.nickname,
        members: allMembers.map(u => u.nickname),
        member_ids: allMembers.map(u => u.user_id)
      }
    }
    console.log('发送群组创建消息，接收者:', msg.to)
    ws.send(JSON.stringify(msg))
}

// 发送文本消息
const sendTextMessage = () => {
  if (!messageInput.value.trim()) {
    return
  }

  if (!currentGroup.value && !currentPrivateChat.value) {
    ElMessage.warning('请先选择聊天对象')
    return
  }

  const msg = {
    type: 'message',
    sub_type: 'text',
    to: currentGroup.value 
      ? currentGroup.value.members.map(u => u.user_id)
      : [currentPrivateChat.value.user_id],
    group_id: currentGroup.value ? currentGroup.value.group_id : undefined,
    timestamp: Date.now(),
    payload: {
      text: messageInput.value,
      group_name: currentGroup.value ? currentGroup.value.group_name : undefined
    }
  }

  ws.send(JSON.stringify(msg))
  
  // 添加到本地消息列表
  messages.value.push({
    ...msg,
    from: currentUser.user_id
  })

  messageInput.value = ''
  scrollToBottom()
}

// 处理文件选择
const handleFileSelect = async (file) => {
  if (selectedUsers.value.length === 0) {
    ElMessage.warning('请先选择接收者')
    return
  }

  const rawFile = file.raw
  const isImage = rawFile.type.startsWith('image/')

  // 读取文件
  const reader = new FileReader()
  reader.onload = (e) => {
    const base64Data = e.target.result

    if (isImage && rawFile.size < 5 * 1024 * 1024) {
      // 小于 5MB 的图片直接发送
      const msg = {
        type: 'message',
        sub_type: 'image',
        to: selectedUsers.value.map(u => u.user_id),
        timestamp: Date.now(),
        payload: {
          image_data: base64Data,
          file_name: rawFile.name,
          file_size: rawFile.size
        }
      }
      ws.send(JSON.stringify(msg))
      messages.value.push({
        ...msg,
        from: currentUser.user_id
      })
    } else {
      // 大文件分片发送
      sendFileInChunks(rawFile, base64Data)
    }

    scrollToBottom()
  }
  reader.readAsDataURL(rawFile)
}

// 分片发送文件
const sendFileInChunks = (file, base64Data) => {
  const fileId = generateFileId()
  const chunkSize = 64 * 1024 // 64KB
  const totalChunks = Math.ceil(base64Data.length / chunkSize)

  // 发送文件元数据
  const metaMsg = {
    type: 'file',
    sub_type: 'file_meta',
    to: selectedUsers.value.map(u => u.user_id),
    timestamp: Date.now(),
    payload: {
      file_id: fileId,
      file_name: file.name,
      file_size: file.size,
      file_type: file.type,
      chunk_size: chunkSize,
      total_chunks: totalChunks
    }
  }
  ws.send(JSON.stringify(metaMsg))

  // 发送分片
  for (let i = 0; i < totalChunks; i++) {
    const start = i * chunkSize
    const end = Math.min(start + chunkSize, base64Data.length)
    const chunkData = base64Data.substring(start, end)

    const chunkMsg = {
      type: 'file',
      sub_type: 'file_chunk',
      to: selectedUsers.value.map(u => u.user_id),
      timestamp: Date.now(),
      payload: {
        file_id: fileId,
        chunk_index: i,
        total_chunks: totalChunks,
        chunk_data: chunkData
      }
    }

    setTimeout(() => {
      ws.send(JSON.stringify(chunkMsg))
    }, i * 50) // 延迟发送避免阻塞
  }

  ElMessage.success('文件发送中...')
}

// 下载文件
const downloadFile = (msg) => {
  const link = document.createElement('a')
  link.href = msg.payload.file_data || msg.payload.image_data
  link.download = msg.payload.file_name
  link.click()
}

// 工具函数
const getUserNickname = (userId) => {
  if (userId === currentUser.user_id) {
    return currentUser.nickname
  }
  const user = onlineUsers.value.find(u => u.user_id === userId)
  return user ? user.nickname : '未知用户'
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const formatSystemMessage = (msg) => {
  if (msg.sub_type === 'user_online') {
    return `${msg.payload.nickname} 加入了聊天室`
  } else if (msg.sub_type === 'user_offline') {
    return `${msg.payload.nickname} 离开了聊天室`
  }
  return ''
}

const getInputPlaceholder = () => {
  if (currentGroup.value) {
    return `发送消息到 ${currentGroup.value.group_name}... (Ctrl+Enter 发送)`
  } else if (currentPrivateChat.value) {
    return `发送消息给 ${currentPrivateChat.value.nickname}... (Ctrl+Enter 发送)`
  } else {
    return '请先选择聊天对象...'
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

const generateFileId = () => {
  return Date.now() + '-' + Math.random().toString(36).substr(2, 9)
}

// 生命周期
onMounted(() => {
  // 可以在这里自动连接
})

onUnmounted(() => {
  if (ws) {
    ws.close()
  }
  stopHeartbeat()
})
</script>

<style scoped>
#app {
  width: 100vw;
  height: 100vh;
  background: #f5f5f5;
}

.chat-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: white;
}

.chat-header {
  height: 60px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  background: #409eff;
  color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 20px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.user-info {
  font-size: 14px;
}

.chat-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.contact-list {
  width: 280px;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  background: #fafafa;
  overflow-y: auto;
}

.section {
  background: white;
  margin-bottom: 10px;
}

.section-header {
  padding: 12px 15px;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  background: #f5f7fa;
}

.section-body {
  max-height: 300px;
  overflow-y: auto;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 15px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  background: white;
}

.contact-item:hover {
  background: #f0f2f5;
}

.contact-item.active {
  background: #e6f7ff;
  border-left: 3px solid #409eff;
}

.contact-item.selected {
  background: #e8f4ff;
}

.contact-item.group-item.active {
  background: #f0f9ff;
  border-left: 3px solid #67c23a;
}

.contact-icon {
  font-size: 20px;
  color: #909399;
}

.contact-item.active .contact-icon {
  color: #409eff;
}

.contact-item.group-item.active .contact-icon {
  color: #67c23a;
}

.contact-info {
  flex: 1;
  min-width: 0;
}

.contact-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contact-detail {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.check-icon {
  color: #409eff;
  font-size: 18px;
}

.unread-badge {
  margin-left: auto;
}

.empty-hint {
  padding: 20px 15px;
  text-align: center;
  color: #909399;
  font-size: 12px;
}

.action-footer {
  padding: 12px 15px;
  border-top: 1px solid #e4e7ed;
  background: white;
  margin-top: auto;
}

.selection-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.hint-text {
  padding: 10px;
  text-align: center;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}

.message-area {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chat-target-header {
  padding: 15px 20px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
}

.target-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.target-info.empty {
  opacity: 0.6;
}

.target-icon {
  font-size: 32px;
}

.target-details {
  flex: 1;
}

.target-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 2px;
}

.target-members {
  font-size: 12px;
  color: #909399;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f9f9f9;
}

.message-item {
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
}

.message-item.self {
  align-items: flex-end;
}

.message-item.other {
  align-items: flex-start;
}

.message-header {
  display: flex;
  gap: 10px;
  margin-bottom: 5px;
  font-size: 12px;
  color: #909399;
}

.message-content {
  max-width: 60%;
}

.message-text {
  padding: 10px 15px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  word-wrap: break-word;
}

.message-item.self .message-text {
  background: #409eff;
  color: white;
}

.message-image {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.message-file {
  padding: 10px 15px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 10px;
}

.message-system {
  padding: 5px 10px;
  background: #f4f4f5;
  border-radius: 4px;
  font-size: 12px;
  color: #909399;
  text-align: center;
}

.message-input {
  border-top: 1px solid #e4e7ed;
  padding: 15px;
  background: white;
}

.input-toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.input-area {
  display: flex;
  gap: 10px;
}

.input-area :deep(.el-textarea) {
  flex: 1;
}
</style>
