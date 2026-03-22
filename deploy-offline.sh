#!/bin/bash

# ClashManager 离线部署脚本
# 使用本地离线依赖包进行部署
echo "========================================"
echo "ClashManager 离线部署脚本"
echo "========================================"

# 检查是否以root用户运行
if [ "$EUID" -ne 0 ]; then
    echo "错误：请以root用户运行此脚本"
    exit 1
fi

# 检查 Go 是否已安装
if ! command -v go >/dev/null 2>&1; then
    echo "错误：Go 未安装，请先安装 Go"
    exit 1
fi

echo "Go 已安装，版本：$(go version)"

# 离线包路径
BROTLI_PACKAGE="/tmp/brotli-1.2.0.tar.gz"
GIN_PACKAGE="/tmp/gin-1.12.0.tar.gz"
YAML_PACKAGE="/tmp/yaml-3.0.1.zip"

# 检查离线包是否存在
if [ ! -f "$BROTLI_PACKAGE" ]; then
    echo "错误：brotli 离线包 $BROTLI_PACKAGE 不存在"
    exit 1
fi

if [ ! -f "$GIN_PACKAGE" ]; then
    echo "错误：gin 离线包 $GIN_PACKAGE 不存在"
    exit 1
fi

if [ ! -f "$YAML_PACKAGE" ]; then
    echo "错误：yaml 离线包 $YAML_PACKAGE 不存在"
    exit 1
fi

echo "所有离线包检查通过"

# 设置 GOPATH
if [ -z "$GOPATH" ]; then
    export GOPATH="$HOME/go"
    echo "GOPATH 未设置，使用默认值：$GOPATH"
fi

# 创建 GOPATH 目录
echo "创建 GOPATH 目录结构..."
mkdir -p $GOPATH/src/github.com/andybalholm
mkdir -p $GOPATH/src/github.com/gin-gonic
mkdir -p $GOPATH/src/gopkg.in
mkdir -p $GOPATH/pkg $GOPATH/bin

# 安装 unzip
echo "检查 unzip..."
if ! command -v unzip >/dev/null 2>&1; then
    echo "安装 unzip..."
    apt update
    apt install -y unzip
fi

# 安装离线依赖
echo "[1/5] 安装离线依赖..."

# 安装 brotli
echo "安装 brotli..."
if [ -d "$GOPATH/src/github.com/andybalholm/brotli" ]; then
    rm -rf $GOPATH/src/github.com/andybalholm/brotli
fi
tar -xzf $BROTLI_PACKAGE -C $GOPATH/src
cp -r $GOPATH/src/brotli-1.2.0 $GOPATH/src/github.com/andybalholm/brotli
rm -rf $GOPATH/src/brotli-1.2.0

# 安装 gin
echo "安装 gin..."
if [ -d "$GOPATH/src/github.com/gin-gonic/gin" ]; then
    rm -rf $GOPATH/src/github.com/gin-gonic/gin
fi
tar -xzf $GIN_PACKAGE -C $GOPATH/src
cp -r $GOPATH/src/gin-1.12.0 $GOPATH/src/github.com/gin-gonic/gin
rm -rf $GOPATH/src/gin-1.12.0

# 安装 yaml
echo "安装 yaml..."
if [ -d "$GOPATH/src/gopkg.in/yaml.v3" ]; then
    rm -rf $GOPATH/src/gopkg.in/yaml.v3
fi
unzip -o $YAML_PACKAGE -d $GOPATH/src
cp -r $GOPATH/src/yaml-3.0.1 $GOPATH/src/gopkg.in/yaml.v3
rm -rf $GOPATH/src/yaml-3.0.1

# 安装 Node.js 和 npm
echo "[2/5] 安装 Node.js 和 npm..."

# 检查是否已安装 Node.js
if ! command -v npm >/dev/null 2>&1; then
    echo "正在安装 Node.js..."
    apt update
    apt install -y nodejs npm
else
    echo "Node.js 已安装，跳过"
fi

# 构建 ClashManager
echo "[3/5] 构建 ClashManager..."

cd /media/ClashManager

# 清理旧的构建文件
rm -rf build
mkdir -p build

# 构建前端
echo "构建前端..."
cd web

# 清理旧的 node_modules
rm -rf node_modules package-lock.json

# 安装依赖
npm install

# 修复权限
find node_modules -type f -exec chmod +x {} \; 2>/dev/null || true

# 构建
npm run build

# 检查构建结果
if [ ! -d "../web/dist" ]; then
    echo "错误: 前端构建失败，dist目录不存在"
    exit 1
fi

# 构建后端
echo "构建后端..."
cd ..

# 设置交叉编译环境变量
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

# 构建
go build -ldflags '-s -w' -o build/clash-manager ./cmd/server

# 检查构建结果
if [ ! -f "build/clash-manager" ]; then
    echo "错误: 后端构建失败"
    exit 1
fi

# 复制必要文件
echo "[4/5] 复制部署文件..."
cp -r data build/ 2>/dev/null || mkdir -p build/data
cp README.md build/ 2>/dev/null || true

# 创建启动脚本
cat > build/start.sh << 'EOF'
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

chmod +x build/start.sh

# 创建 systemd 服务文件
cat > build/clash-manager.service << EOF
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
cat > build/install.sh << 'EOF'
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

chmod +x build/install.sh

# 部署到系统
echo "[5/5] 部署到系统..."

# 安装
echo "安装 ClashManager..."
cd build
sudo ./install.sh

# 启动服务
echo "启动 ClashManager 服务..."
systemctl start clash-manager

# 查看状态
echo "查看服务状态..."
systemctl status clash-manager

echo ""
echo "========================================"
echo "部署完成！"
echo "========================================"
echo ""
echo "访问地址: http://localhost:8090"
echo ""
echo "服务状态:"
echo "  启动: systemctl start clash-manager"
echo "  停止: systemctl stop clash-manager"
echo "  状态: systemctl status clash-manager"
echo ""
