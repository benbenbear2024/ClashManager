# ClashManager Linux Build Script (PowerShell)
# Builds Linux executable for Debian 13.3 deployment

Write-Host "========================================" -ForegroundColor Green
Write-Host "ClashManager Linux Build Script" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

$PROJECT_ROOT = Get-Location
$BUILD_DIR = "$PROJECT_ROOT\build"
$OUTPUT_NAME = "clash-manager"

# Check required tools
Write-Host "[1/6] Checking environment..." -ForegroundColor Yellow

try {
    $goVersion = go version
    Write-Host "  Go: $goVersion" -ForegroundColor Gray
} catch {
    Write-Host "Error: Go is required" -ForegroundColor Red
    exit 1
}

try {
    $npmVersion = npm --version
    Write-Host "  npm: $npmVersion" -ForegroundColor Gray
} catch {
    Write-Host "Error: Node.js and npm are required" -ForegroundColor Red
    exit 1
}

# Clean old build files
Write-Host "[2/6] Cleaning old build files..." -ForegroundColor Yellow
if (Test-Path $BUILD_DIR) {
    Remove-Item -Path $BUILD_DIR -Recurse -Force
}
New-Item -ItemType Directory -Path $BUILD_DIR | Out-Null

# Build frontend
Write-Host "[3/6] Building frontend..." -ForegroundColor Yellow
Set-Location "$PROJECT_ROOT\web"

npm install
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: npm install failed" -ForegroundColor Red
    exit 1
}

npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Frontend build failed" -ForegroundColor Red
    exit 1
}

# Check frontend build result
if (-not (Test-Path "$PROJECT_ROOT\web\dist")) {
    Write-Host "Error: Frontend build failed, dist directory not found" -ForegroundColor Red
    exit 1
}

Write-Host "[4/6] Frontend build completed" -ForegroundColor Green

# Build backend (Linux AMD64)
Write-Host "[5/6] Building Linux AMD64 backend..." -ForegroundColor Yellow
Set-Location $PROJECT_ROOT

# Set cross-compilation environment variables
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

go build -ldflags "-s -w" -o "$BUILD_DIR\$OUTPUT_NAME" ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Backend build failed" -ForegroundColor Red
    exit 1
}

# Check build result
if (-not (Test-Path "$BUILD_DIR\$OUTPUT_NAME")) {
    Write-Host "Error: Backend build failed, output file not found" -ForegroundColor Red
    exit 1
}

# Copy necessary files
Write-Host "[6/6] Copying deployment files..." -ForegroundColor Yellow
New-Item -ItemType Directory -Path "$BUILD_DIR\data" -Force | Out-Null
if (Test-Path "$PROJECT_ROOT\README.md") {
    Copy-Item "$PROJECT_ROOT\README.md" "$BUILD_DIR\" -Force
}

# Create start script
$startScript = '#!/bin/bash

# ClashManager Start Script

APP_DIR=$(dirname "$(readlink -f "$0")")
cd "$APP_DIR"

# Check data directory
if [ ! -d "data" ]; then
    mkdir -p data
fi

# Start program
./clash-manager "$@"
'

$startScript | Out-File -FilePath "$BUILD_DIR\start.sh" -Encoding UTF8NoBOM

# Create systemd service file
$serviceFile = '[Unit]
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
'

$serviceFile | Out-File -FilePath "$BUILD_DIR\clash-manager.service" -Encoding UTF8NoBOM

# Create install script
$installScript = '#!/bin/bash

# ClashManager Install Script

set -e

echo "ClashManager Install Script"
echo "======================"

INSTALL_DIR="/opt/clash-manager"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Create install directory
echo "Creating install directory..."
mkdir -p "$INSTALL_DIR"

# Copy files
echo "Copying files..."
cp -r ./* "$INSTALL_DIR/"

# Set permissions
echo "Setting permissions..."
chmod +x "$INSTALL_DIR/clash-manager"
chmod +x "$INSTALL_DIR/start.sh"

# Create data directory
mkdir -p "$INSTALL_DIR/data"

# Install systemd service
echo "Installing systemd service..."
cp "$INSTALL_DIR/clash-manager.service" /etc/systemd/system/
systemctl daemon-reload
systemctl enable clash-manager

echo ""
echo "Installation complete!"
echo ""
echo "Usage:"
echo "  Start service: systemctl start clash-manager"
echo "  Stop service: systemctl stop clash-manager"
echo "  Check status: systemctl status clash-manager"
echo "  Manual run: $INSTALL_DIR/start.sh"
echo ""
echo "Access: http://localhost:8090"
echo ""
'

$installScript | Out-File -FilePath "$BUILD_DIR\install.sh" -Encoding UTF8NoBOM

# Create README
$readme = 'ClashManager Linux Build Package
========================================

Deployment Steps:
1. Copy all files from build directory to Debian 13.3 server
2. Run on server: sudo ./install.sh
3. Start service: sudo systemctl start clash-manager

Or run manually:
  ./start.sh

Access: http://localhost:8090

Files:
  - clash-manager: Main program (Linux AMD64)
  - start.sh: Manual start script
  - install.sh: Install script (requires root)
  - clash-manager.service: systemd service file
  - data/: Data directory

Requirements:
  - Debian 13.3 or compatible Linux distribution
  - systemd (for service management)
  - x86_64 architecture
'

$readme | Out-File -FilePath "$BUILD_DIR\README.txt" -Encoding UTF8NoBOM

# Display results
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Build Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Output directory: $BUILD_DIR" -ForegroundColor Cyan
Write-Host ""
Write-Host "Files:" -ForegroundColor Yellow
Get-ChildItem $BUILD_DIR | ForEach-Object {
    $size = if ($_.PSIsContainer) { "<DIR>" } else { "{0:N0} KB" -f ($_.Length / 1KB) }
    Write-Host "  $($_.Name) $size" -ForegroundColor Gray
}
Write-Host ""
Write-Host "Deployment Steps:" -ForegroundColor Yellow
Write-Host "1. Copy all files from $BUILD_DIR to Debian 13.3 server" -ForegroundColor White
Write-Host "2. Run on server: sudo ./install.sh" -ForegroundColor White
Write-Host "3. Start service: sudo systemctl start clash-manager" -ForegroundColor White
Write-Host ""
Write-Host "Or run manually:" -ForegroundColor Yellow
Write-Host '  ./start.sh' -ForegroundColor White
Write-Host ""

Set-Location $PROJECT_ROOT
