#!/bin/bash

# 离线依赖安装脚本
echo "========================================"
echo "离线依赖安装脚本"
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
echo "[1/2] 安装 Go 离线依赖..."

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
echo "[2/2] 安装 Node.js 和 npm..."

# 检查是否已安装 Node.js
if ! command -v npm >/dev/null 2>&1; then
    echo "正在安装 Node.js..."
    apt update
    apt install -y nodejs npm
else
    echo "Node.js 已安装，跳过"
fi

echo ""
echo "========================================"
echo "离线依赖安装完成！"
echo "========================================"
echo ""
echo "现在可以运行构建脚本："
echo "  ./build.sh"
echo ""
