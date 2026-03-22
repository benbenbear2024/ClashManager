#!/bin/bash

# 安装 Go 语言环境
echo "========================================"
echo "安装 Go 语言环境"
echo "========================================"

# 下载 Go 1.25.0（最新稳定版）
GO_VERSION="1.25.0"
GO_ARCH="amd64"

# 下载 Go
echo "下载 Go $GO_VERSION..."
wget -q https://golang.org/dl/go$GO_VERSION.linux-$GO_ARCH.tar.gz

# 解压到 /usr/local
echo "解压 Go 到 /usr/local..."
tar -xzf go$GO_VERSION.linux-$GO_ARCH.tar.gz -C /usr/local

# 设置环境变量
echo "设置环境变量..."
cat >> /etc/profile << EOF

# Go environment variables
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export GOBIN=$GOPATH/bin
EOF

# 立即生效
source /etc/profile

# 验证安装
echo "验证 Go 安装..."
go version

# 清理安装包
echo "清理安装包..."
rm -f go$GO_VERSION.linux-$GO_ARCH.tar.gz

echo "========================================"
echo "Go 安装完成！"
echo "========================================"
