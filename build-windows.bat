@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo ClashManager Windows 构建脚本
echo ========================================
echo.

set "PROJECT_ROOT=%CD%"
set "BUILD_DIR=%PROJECT_ROOT%\build"
set "OUTPUT_NAME=clash-manager"

:: 检查必要的工具
echo [1/6] 检查环境...

where go >nul 2>nul
if errorlevel 1 (
    echo 错误: 需要安装 Go
    exit /b 1
)

where npm >nul 2>nul
if errorlevel 1 (
    echo 错误: 需要安装 Node.js 和 npm
    exit /b 1
)

echo 环境检查通过

:: 清理旧的构建文件
echo [2/6] 清理旧的构建文件...
if exist "%BUILD_DIR%" rmdir /s /q "%BUILD_DIR%"
mkdir "%BUILD_DIR%"

:: 构建前端
echo [3/6] 构建前端...
cd /d "%PROJECT_ROOT%\web"
call npm install
if errorlevel 1 (
    echo 错误: npm install 失败
    exit /b 1
)

call npm run build
if errorlevel 1 (
    echo 错误: 前端构建失败
    exit /b 1
)

:: 检查前端构建结果
if not exist "%PROJECT_ROOT%\web\dist" (
    echo 错误: 前端构建失败，dist目录不存在
    exit /b 1
)

echo [4/6] 前端构建完成

:: 构建后端（Linux AMD64）
echo [5/6] 构建 Linux AMD64 后端...
cd /d "%PROJECT_ROOT%"

:: 设置交叉编译环境变量
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -ldflags "-s -w" -o "%BUILD_DIR%\%OUTPUT_NAME%" ./cmd/server
if errorlevel 1 (
    echo 错误: 后端构建失败
    exit /b 1
)

:: 检查构建结果
if not exist "%BUILD_DIR%\%OUTPUT_NAME%" (
    echo 错误: 后端构建失败，输出文件不存在
    exit /b 1
)

:: 复制必要文件
echo [6/6] 复制部署文件...
if not exist "%BUILD_DIR%\data" mkdir "%BUILD_DIR%\data"
if exist "%PROJECT_ROOT%\README.md" copy "%PROJECT_ROOT%\README.md" "%BUILD_DIR%\" >nul

:: 创建启动脚本
echo #!/bin/bash > "%BUILD_DIR%\start.sh"
echo. >> "%BUILD_DIR%\start.sh"
echo # ClashManager 启动脚本 >> "%BUILD_DIR%\start.sh"
echo. >> "%BUILD_DIR%\start.sh"
echo APP_DIR=$(dirname "$(readlink -f "$0")") >> "%BUILD_DIR%\start.sh"
echo cd "$APP_DIR" >> "%BUILD_DIR%\start.sh"
echo. >> "%BUILD_DIR%\start.sh"
echo # 检查数据目录 >> "%BUILD_DIR%\start.sh"
echo if [ ! -d "data" ]; then >> "%BUILD_DIR%\start.sh"
echo     mkdir -p data >> "%BUILD_DIR%\start.sh"
echo fi >> "%BUILD_DIR%\start.sh"
echo. >> "%BUILD_DIR%\start.sh"
echo # 启动程序 >> "%BUILD_DIR%\start.sh"
echo ./clash-manager "$@" >> "%BUILD_DIR%\start.sh"

:: 创建 systemd 服务文件
echo [Unit] > "%BUILD_DIR%\clash-manager.service"
echo Description=ClashManager Service >> "%BUILD_DIR%\clash-manager.service"
echo After=network.target >> "%BUILD_DIR%\clash-manager.service"
echo. >> "%BUILD_DIR%\clash-manager.service"
echo [Service] >> "%BUILD_DIR%\clash-manager.service"
echo Type=simple >> "%BUILD_DIR%\clash-manager.service"
echo WorkingDirectory=/opt/clash-manager >> "%BUILD_DIR%\clash-manager.service"
echo ExecStart=/opt/clash-manager/clash-manager >> "%BUILD_DIR%\clash-manager.service"
echo Restart=on-failure >> "%BUILD_DIR%\clash-manager.service"
echo RestartSec=5 >> "%BUILD_DIR%\clash-manager.service"
echo. >> "%BUILD_DIR%\clash-manager.service"
echo [Install] >> "%BUILD_DIR%\clash-manager.service"
echo WantedBy=multi-user.target >> "%BUILD_DIR%\clash-manager.service"

:: 创建安装脚本
echo #!/bin/bash > "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo # ClashManager 安装脚本 >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo set -e >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo echo "ClashManager 安装脚本" >> "%BUILD_DIR%\install.sh"
echo echo "======================" >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo INSTALL_DIR="/opt/clash-manager" >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo # 检查是否为root用户 >> "%BUILD_DIR%\install.sh"
echo if [ "$EUID" -ne 0 ]; then >> "%BUILD_DIR%\install.sh"
echo     echo "请使用 root 权限运行此脚本" >> "%BUILD_DIR%\install.sh"
echo     exit 1 >> "%BUILD_DIR%\install.sh"
echo fi >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo # 创建安装目录 >> "%BUILD_DIR%\install.sh"
echo echo "创建安装目录..." >> "%BUILD_DIR%\install.sh"
echo mkdir -p "$INSTALL_DIR" >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo # 复制文件 >> "%BUILD_DIR%\install.sh"
echo echo "复制文件..." >> "%BUILD_DIR%\install.sh"
echo cp -r ./* "$INSTALL_DIR/" >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo # 设置权限 >> "%BUILD_DIR%\install.sh"
echo echo "设置权限..." >> "%BUILD_DIR%\install.sh"
echo chmod +x "$INSTALL_DIR/clash-manager" >> "%BUILD_DIR%\install.sh"
echo chmod +x "$INSTALL_DIR/start.sh" >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo # 创建数据目录 >> "%BUILD_DIR%\install.sh"
echo mkdir -p "$INSTALL_DIR/data" >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo # 安装 systemd 服务 >> "%BUILD_DIR%\install.sh"
echo echo "安装 systemd 服务..." >> "%BUILD_DIR%\install.sh"
echo cp "$INSTALL_DIR/clash-manager.service" /etc/systemd/system/ >> "%BUILD_DIR%\install.sh"
echo systemctl daemon-reload >> "%BUILD_DIR%\install.sh"
echo systemctl enable clash-manager >> "%BUILD_DIR%\install.sh"
echo. >> "%BUILD_DIR%\install.sh"
echo echo "" >> "%BUILD_DIR%\install.sh"
echo echo "安装完成！" >> "%BUILD_DIR%\install.sh"
echo echo "" >> "%BUILD_DIR%\install.sh"
echo echo "使用方法:" >> "%BUILD_DIR%\install.sh"
echo echo "  启动服务: systemctl start clash-manager" >> "%BUILD_DIR%\install.sh"
echo echo "  停止服务: systemctl stop clash-manager" >> "%BUILD_DIR%\install.sh"
echo echo "  查看状态: systemctl status clash-manager" >> "%BUILD_DIR%\install.sh"
echo echo "  手动运行: $INSTALL_DIR/start.sh" >> "%BUILD_DIR%\install.sh"
echo echo "" >> "%BUILD_DIR%\install.sh"
echo echo "访问地址: http://localhost:8080" >> "%BUILD_DIR%\install.sh"
echo echo "" >> "%BUILD_DIR%\install.sh"

:: 创建 README
echo ClashManager Linux 构建包 > "%BUILD_DIR%\README.txt"
echo ======================================== >> "%BUILD_DIR%\README.txt"
echo. >> "%BUILD_DIR%\README.txt"
echo 部署步骤: >> "%BUILD_DIR%\README.txt"
echo 1. 将 build 目录中的所有文件复制到 Debian 13.3 服务器 >> "%BUILD_DIR%\README.txt"
echo 2. 在服务器上运行: sudo ./install.sh >> "%BUILD_DIR%\README.txt"
echo 3. 启动服务: sudo systemctl start clash-manager >> "%BUILD_DIR%\README.txt"
echo. >> "%BUILD_DIR%\README.txt"
echo 或者手动运行: >> "%BUILD_DIR%\README.txt"
echo   ./start.sh >> "%BUILD_DIR%\README.txt"
echo. >> "%BUILD_DIR%\README.txt"
echo 访问地址: http://localhost:8080 >> "%BUILD_DIR%\README.txt"
echo. >> "%BUILD_DIR%\README.txt"
echo 文件说明: >> "%BUILD_DIR%\README.txt"
echo   - clash-manager: 主程序 >> "%BUILD_DIR%\README.txt"
echo   - start.sh: 手动启动脚本 >> "%BUILD_DIR%\README.txt"
echo   - install.sh: 安装脚本（需要root权限） >> "%BUILD_DIR%\README.txt"
echo   - clash-manager.service: systemd 服务文件 >> "%BUILD_DIR%\README.txt"
echo   - data/: 数据目录 >> "%BUILD_DIR%\README.txt"

echo.
echo ========================================
echo 构建完成！
echo ========================================
echo.
echo 输出目录: %BUILD_DIR%
echo.
echo 文件列表:
dir /b "%BUILD_DIR%"
echo.
echo 部署步骤:
echo 1. 将 %BUILD_DIR% 目录中的所有文件复制到 Debian 13.3 服务器
echo 2. 在服务器上运行: sudo ./install.sh
echo 3. 启动服务: sudo systemctl start clash-manager
echo.
echo 或者手动运行:
echo   ./start.sh
echo.

pause
