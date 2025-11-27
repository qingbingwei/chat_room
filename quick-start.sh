#!/bin/bash

# 快速部署脚本 - 一键安装依赖并启动前后端

set -e

echo "========================================="
echo "    聊天系统一键部署脚本 (Linux/macOS)"
echo "========================================="
echo ""

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ -f /etc/debian_version ]]; then
        echo "debian"
    elif [[ -f /etc/redhat-release ]]; then
        echo "redhat"
    elif [[ -f /etc/arch-release ]]; then
        echo "arch"
    else
        echo "unknown"
    fi
}

OS=$(detect_os)
echo "检测到操作系统: $OS"
echo ""

# 安装 Go
install_go() {
    echo "   尝试自动安装 Go..."
    case $OS in
        macos)
            if command -v brew &> /dev/null; then
                echo "   使用 Homebrew 安装 Go..."
                brew install go
            else
                echo "   ❌ 请先安装 Homebrew: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                exit 1
            fi
            ;;
        debian)
            echo "   使用 apt 安装 Go..."
            sudo apt update
            sudo apt install -y golang-go
            ;;
        redhat)
            echo "   使用 yum/dnf 安装 Go..."
            if command -v dnf &> /dev/null; then
                sudo dnf install -y golang
            else
                sudo yum install -y golang
            fi
            ;;
        arch)
            echo "   使用 pacman 安装 Go..."
            sudo pacman -Sy --noconfirm go
            ;;
        *)
            echo "   ❌ 无法自动安装 Go，请手动安装"
            echo "   下载地址: https://go.dev/dl/"
            exit 1
            ;;
    esac
}

# 安装 Node.js
install_node() {
    echo "   尝试自动安装 Node.js..."
    case $OS in
        macos)
            if command -v brew &> /dev/null; then
                echo "   使用 Homebrew 安装 Node.js..."
                brew install node
            else
                echo "   ❌ 请先安装 Homebrew"
                exit 1
            fi
            ;;
        debian)
            echo "   使用 NodeSource 安装 Node.js LTS..."
            curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
            sudo apt install -y nodejs
            ;;
        redhat)
            echo "   使用 NodeSource 安装 Node.js LTS..."
            curl -fsSL https://rpm.nodesource.com/setup_lts.x | sudo bash -
            if command -v dnf &> /dev/null; then
                sudo dnf install -y nodejs
            else
                sudo yum install -y nodejs
            fi
            ;;
        arch)
            echo "   使用 pacman 安装 Node.js..."
            sudo pacman -Sy --noconfirm nodejs npm
            ;;
        *)
            echo "   ❌ 无法自动安装 Node.js，请手动安装"
            echo "   下载地址: https://nodejs.org/"
            exit 1
            ;;
    esac
}

# 检查并安装依赖
echo "1. 检查依赖..."

if ! command -v go &> /dev/null; then
    echo "   ⚠️  Go 未安装"
    read -p "   是否自动安装 Go? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        install_go
        # 重新加载环境变量
        export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
        if ! command -v go &> /dev/null; then
            echo "   ❌ Go 安装后未找到，请重新打开终端后再试"
            exit 1
        fi
    else
        echo "   ❌ 请先安装 Go 1.21+"
        exit 1
    fi
fi

if ! command -v node &> /dev/null; then
    echo "   ⚠️  Node.js 未安装"
    read -p "   是否自动安装 Node.js? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        install_node
        if ! command -v node &> /dev/null; then
            echo "   ❌ Node.js 安装后未找到，请重新打开终端后再试"
            exit 1
        fi
    else
        echo "   ❌ 请先安装 Node.js 18+"
        exit 1
    fi
fi

echo "✅ Go 版本: $(go version)"
echo "✅ Node.js 版本: $(node --version)"
echo ""

# 后端设置
echo "2. 设置后端..."
cd backend

if [ ! -d "vendor" ]; then
    echo "   安装 Go 依赖..."
    go mod tidy
fi

if [ ! -f "chat-server" ]; then
    echo "   编译后端..."
    go build -o chat-server main.go
fi

cd ..
echo "✅ 后端设置完成"
echo ""

# 前端设置
echo "3. 设置前端..."
cd frontend

if [ ! -d "node_modules" ]; then
    echo "   安装前端依赖..."
    npm install
fi

cd ..
echo "✅ 前端设置完成"
echo ""

# 启动服务
echo "========================================="
echo "准备启动服务..."
echo "========================================="
echo ""
echo "后端服务器将在后台运行..."
echo "前端开发服务器将在当前终端运行..."
echo ""
echo "按 Ctrl+C 停止前端服务器"
echo "要停止后端服务器，请运行: pkill -f chat-server"
echo ""
echo "前端地址: http://localhost:5173"
echo "后端地址: http://localhost:9090"
echo "WebSocket: ws://localhost:9090/ws"
echo ""
read -p "按 Enter 键继续启动..."

# 启动后端（后台运行）
cd backend
./chat-server &
BACKEND_PID=$!
echo "✅ 后端服务器已启动 (PID: $BACKEND_PID)"

# 等待后端启动
sleep 2

# 启动前端（前台运行）
cd ../frontend
echo "✅ 启动前端服务器..."
npm run dev

# 当前端退出时，清理后端
echo ""
echo "正在停止后端服务器..."
kill $BACKEND_PID 2>/dev/null
echo "✅ 服务已停止"
