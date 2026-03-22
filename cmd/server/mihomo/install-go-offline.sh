#!/bin/bash

# 离线安装 Go 语言环境
echo "========================================"
echo "离线安装 Go 语言环境"
echo "========================================"

# 离线包路径
GO_PACKAGE="/tmp/go1.26.1.linux-amd64.tar.gz"
GO_VERSION="1.26.1"

# 检查离线包是否存在
if [ ! -f "$GO_PACKAGE" ]; then
    echo "错误：离线包 $GO_PACKAGE 不存在"
    exit 1
fi

echo "找到离线包：$GO_PACKAGE"

# 解压到 /usr/local
echo "解压 Go 到 /usr/local..."
tar -xzf $GO_PACKAGE -C /usr/local

# 设置环境变量
echo "设置环境变量..."
cat >> /etc/profile << 'EOF'

# Go environment variables
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export GOBIN=$GOPATH/bin
EOF

# 立即生效
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export GOBIN=$GOPATH/bin

# 验证安装
echo "验证 Go 安装..."
/usr/local/go/bin/go version

echo "========================================"
echo "Go $GO_VERSION 安装完成！"
echo "========================================"
echo ""
echo "环境变量已添加到 /etc/profile"
echo "请运行以下命令使环境变量生效："
echo "  source /etc/profile"
echo ""
