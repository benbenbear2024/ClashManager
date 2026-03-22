#!/bin/bash

# ClashManager 构建脚本
# 用于构建Linux可执行程序，可直接部署到Debian 13.3

set -e

echo "========================================"
echo "ClashManager Linux 构建脚本"
echo "========================================"

# 检查必要的工具
command -v go >/dev/null 2>&1 || { echo "错误: 需要安装 Go"; exit 1; }
command -v npm >/dev/null 2>&1 || { echo "错误: 需要安装 Node.js 和 npm"; exit 1; }

# 设置变量
PROJECT_ROOT=$(pwd)
BUILD_DIR="$PROJECT_ROOT/build"
OUTPUT_NAME="clash-manager"
VERSION=$(date +"%Y%m%d")

# 清理旧的构建文件
echo "[1/6] 清理旧的构建文件..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# 构建前端
echo "[2/6] 构建前端..."
cd "$PROJECT_ROOT/web"

# 清理旧的 node_modules
echo "清理旧的 node_modules..."
rm -rf node_modules package-lock.json

# 安装依赖
echo "安装前端依赖..."
npm install

# 修复 node_modules 权限
echo "修复 node_modules 权限..."
find node_modules -type f -exec chmod +x {} \; 2>/dev/null || true

# 构建
echo "构建前端..."
npm run build

# 检查前端构建结果
if [ ! -d "$PROJECT_ROOT/web/dist" ]; then
    echo "错误: 前端构建失败，dist目录不存在"
    exit 1
fi

echo "[3/6] 前端构建完成"

# 复制前端文件到web目录
echo "[4/6] 准备嵌入资源..."
cd "$PROJECT_ROOT"

# 确保web/dist目录存在
if [ ! -d "web/dist" ]; then
    echo "错误: web/dist 目录不存在"
    exit 1
fi

# 构建后端（Linux AMD64）
echo "[5/6] 构建 Linux AMD64 后端..."
cd "$PROJECT_ROOT"

# 设置交叉编译环境变量
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

# 检查是否安装了musl工具链用于静态链接
if command -v x86_64-linux-musl-gcc >/dev/null 2>&1; then
    echo "使用 musl 进行静态链接..."
    export CC=x86_64-linux-musl-gcc
    export CXX=x86_64-linux-musl-g++
    go build -ldflags '-s -w -linkmode external -extldflags "-static"' -o "$BUILD_DIR/$OUTPUT_NAME" ./cmd/server
else
    echo "使用默认 CGO 编译（需要目标系统有相应的库）..."
    go build -ldflags '-s -w' -o "$BUILD_DIR/$OUTPUT_NAME" ./cmd/server
fi

# 检查构建结果
if [ ! -f "$BUILD_DIR/$OUTPUT_NAME" ]; then
    echo "错误: 后端构建失败"
    exit 1
fi

# 复制必要文件
echo "[6/6] 复制部署文件..."
cp -r "$PROJECT_ROOT/data" "$BUILD_DIR/" 2>/dev/null || mkdir -p "$BUILD_DIR/data"
cp "$PROJECT_ROOT/README.md" "$BUILD_DIR/" 2>/dev/null || true

# 创建启动脚本
cat > "$BUILD_DIR/start.sh" << 'EOF'
#!/bin/bash

# ClashManager 启动脚本

APP_DIR=$(dirname "$(readlink -f "$0")")
cd "$APP_DIR"

# 检查数据目录
if [ ! -d "data" ]; then
    mkdir -p data
fi

# 启动程序
./clash-manager "$@"
EOF

chmod +x "$BUILD_DIR/start.sh"

# 创建 systemd 服务文件
cat > "$BUILD_DIR/clash-manager.service" << EOF
[Unit]
Description=ClashManager Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/clash-manager
ExecStart=/opt/clash-manager/clash-manager
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 创建安装脚本
cat > "$BUILD_DIR/install.sh" << 'EOF'
#!/bin/bash

# ClashManager 安装脚本

set -e

echo "ClashManager 安装脚本"
echo "======================"

INSTALL_DIR="/opt/clash-manager"

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用 root 权限运行此脚本"
    exit 1
fi

# 创建安装目录
echo "创建安装目录..."
mkdir -p "$INSTALL_DIR"

# 复制文件
echo "复制文件..."
cp -r ./* "$INSTALL_DIR/"

# 设置权限
echo "设置权限..."
chmod +x "$INSTALL_DIR/clash-manager"
chmod +x "$INSTALL_DIR/start.sh"

# 创建数据目录
mkdir -p "$INSTALL_DIR/data"

# 安装 systemd 服务
echo "安装 systemd 服务..."
cp "$INSTALL_DIR/clash-manager.service" /etc/systemd/system/
systemctl daemon-reload
systemctl enable clash-manager

echo ""
echo "安装完成！"
echo ""
echo "使用方法:"
echo "  启动服务: systemctl start clash-manager"
echo "  停止服务: systemctl stop clash-manager"
echo "  查看状态: systemctl status clash-manager"
echo "  手动运行: $INSTALL_DIR/start.sh"
echo ""
echo "访问地址: http://localhost:8090"
echo ""
EOF

chmod +x "$BUILD_DIR/install.sh"

# 打包
echo ""
echo "========================================"
echo "构建完成！"
echo "========================================"
echo ""
echo "输出文件: $BUILD_DIR/"
echo ""
echo "文件列表:"
ls -lh "$BUILD_DIR/"
echo ""
echo "部署步骤:"
echo "1. 将 $BUILD_DIR 目录复制到 Debian 13.3 服务器"
echo "2. 在服务器上运行: sudo ./install.sh"
echo "3. 启动服务: sudo systemctl start clash-manager"
echo ""
echo "或者手动运行:"
echo "  ./start.sh"
echo ""
