#!/bin/bash

# 主部署脚本
echo "========================================"
echo "ClashManager 部署脚本"
echo "========================================"

# 检查是否以root用户运行
if [ "$EUID" -ne 0 ]; then
    echo "错误：请以root用户运行此脚本"
    exit 1
fi

# 安装mihomo
echo "[1/4] 安装mihomo程序..."
mkdir -p /opt/mihomo

# 解压mihomo程序
MIHOMO_FILE="/tmp/mihomo-linux-amd64-v3-go123-alpha-dd4eb63.gz"
if [ -f "$MIHOMO_FILE" ]; then
    gunzip -c "$MIHOMO_FILE" > /opt/mihomo/mihomo
    chmod +x /opt/mihomo/mihomo
    echo "mihomo程序安装成功"
else
    echo "错误：$MIHOMO_FILE 文件不存在"
    exit 1
fi

# 创建mihomo配置目录
mkdir -p /etc/mihomo

# 创建mihomo系统服务文件
cat > /etc/systemd/system/mihomo.service << EOF
[Unit]
Description=mihomo proxy
After=network.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/opt/mihomo/mihomo -f /etc/mihomo/config.yaml
Restart=on-failure
RestartSec=5s
User=root

[Install]
WantedBy=multi-user.target
EOF

# 重新加载系统服务
systemctl daemon-reload

echo "mihomo服务配置完成"

# 安装L2TP服务器
echo "[2/4] 安装L2TP服务器..."

# 安装必要的软件包
apt update
apt install -y xl2tpd ppp iptables

# 配置xl2tpd
cat > /etc/xl2tpd/xl2tpd.conf << EOF
[global]
ipsec saref = yes
listen-addr = 0.0.0.0

[lns default]
ip range = 10.0.10.2-10.0.10.254
local ip = 10.0.10.1
refuse chap = yes
refuse pap = yes
require authentication = yes
ppp debug = no
pppoptfile = /etc/ppp/options.xl2tpd
length bit = yes
EOF

# 配置ppp选项
cat > /etc/ppp/options.xl2tpd << EOF
require-mschap-v2
ms-dns 8.8.8.8
ms-dns 8.8.4.4
asyncmap 0
auth
crtscts
lock
hide-password
modem
debug
name l2tpd
proxyarp
lcp-echo-interval 30
lcp-echo-failure 4
EOF

# 创建账号文件
cat > /etc/ppp/chap-secrets << EOF
# Secrets for authentication using CHAP
# client    server    secret    IP addresses
EOF

# 生成253个账号
for i in $(seq 1 253); do
    echo "user$i l2tpd pass$i 10.0.10.$((i+1))" >> /etc/ppp/chap-secrets
done

# 配置防火墙
iptables -t nat -A POSTROUTING -s 10.0.10.0/24 -o eth0 -j MASQUERADE
iptables -A FORWARD -s 10.0.10.0/24 -j ACCEPT
iptables -A FORWARD -d 10.0.10.0/24 -j ACCEPT

# 保存防火墙规则
mkdir -p /etc/iptables
iptables-save > /etc/iptables/rules.v4

# 启动服务
systemctl enable xl2tpd
systemctl start xl2tpd

echo "L2TP服务器安装和配置完成"

# 配置mihomo
echo "[3/4] 配置mihomo..."
mkdir -p /etc/mihomo
if [ -f "config_mihomo.yaml" ]; then
    cp config_mihomo.yaml /etc/mihomo/config.yaml
    echo "mihomo配置文件复制成功"
else
    echo "警告：config_mihomo.yaml 文件不存在，使用默认配置"
    # 创建默认配置
    cat > /etc/mihomo/config.yaml << EOF
# mihomo 配置文件

# 监听端口
port: 7890
socks-port: 7891
redir-port: 7892
tproxy-port: 7893
mixed-port: 7894

# 外部控制器
external-controller: 127.0.0.1:9090

# 允许 LAN
allow-lan: true

# 模式：rule, global, direct
mode: rule

# 日志级别：info, warning, error, debug
log-level: info

# DNS 配置
dns:
  enable: true
  listen: 0.0.0.0:53
  enhanced-mode: fake-ip
  nameserver:
    - 8.8.8.8
    - 8.8.4.4
  fallback:
    - 1.1.1.1
    - 1.0.0.1

# 规则配置
rules:
  - DOMAIN-SUFFIX,google.com,PROXY
  - DOMAIN-SUFFIX,facebook.com,PROXY
  - DOMAIN-SUFFIX,twitter.com,PROXY
  - DOMAIN-SUFFIX,youtube.com,PROXY
  - GEOIP,CN,DIRECT
  - MATCH,PROXY

# 节点配置
proxies:
  # 这里会自动生成 L2TP 节点
  # 格式：Name1, Name2, ..., Name253
EOF
    echo "默认配置文件创建成功"
fi

# 启动mihomo服务
echo "启动mihomo服务..."
systemctl enable mihomo
systemctl start mihomo

# 配置监控脚本
echo "[4/4] 配置监控脚本..."
mkdir -p /opt/mihomo
cat > /opt/mihomo/monitor.sh << 'EOF'
#!/bin/bash

# 监控mihomo节点状态的脚本

# 检查mihomo服务是否运行
check_mihomo_service() {
    if systemctl is-active --quiet mihomo; then
        return 0
    else
        return 1
    fi
}

# 检查指定节点是否可用
check_mihomo_node() {
    local node_name=$1
    # 检查mihomo的API是否响应
    if curl -s http://127.0.0.1:9090/proxies | grep -q "$node_name"; then
        return 0
    else
        return 1
    fi
}

# 断开指定IP的L2TP连接
disconnect_l2tp() {
    local ip=$1
    # 查找并断开对应的PPP连接
    local ppp_interface=$(ip addr | grep "$ip" | awk '{print $NF}')
    if [ ! -z "$ppp_interface" ]; then
        pkill -f "$ppp_interface"
        echo "已断开IP $ip 的L2TP连接"
    fi
}

# 主监控循环
while true; do
    # 检查mihomo服务是否运行
    if ! check_mihomo_service; then
        echo "mihomo服务未运行，断开所有L2TP连接"
        # 断开所有L2TP连接
        pkill -f "ppp"
        sleep 5
        continue
    fi
    
    # 检查每个L2TP连接对应的mihomo节点
    for i in $(seq 1 253); do
        local user="user$i"
        local ip="10.0.10.$((i+1))"
        local node="Name$i"
        
        # 检查用户是否已连接
        if ip addr | grep -q "$ip"; then
            # 检查对应的节点是否可用
            if ! check_mihomo_node "$node"; then
                echo "节点 $node 不可用，断开用户 $user 的连接"
                disconnect_l2tp "$ip"
            fi
        fi
    done
    
    # 休眠一段时间后再次检查
    sleep 10
done
EOF

chmod +x /opt/mihomo/monitor.sh

# 创建监控服务
cat > /etc/systemd/system/mihomo-monitor.service << EOF
[Unit]
Description=mihomo monitor
After=network.target

[Service]
Type=simple
ExecStart=/opt/mihomo/monitor.sh
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# 启动监控服务
systemctl daemon-reload
systemctl enable mihomo-monitor
systemctl start mihomo-monitor

echo "========================================"
echo "部署完成！"
echo "========================================"
echo "以下是部署信息："
echo "1. mihomo程序已安装在 /opt/mihomo/"
echo "2. L2TP服务器已配置，IP范围：10.0.10.2-10.0.10.254"
echo "3. 已生成253个账号：user1-user253，密码：pass1-pass253"
echo "4. 每个账号对应一个mihomo节点：Name1-Name253"
echo "5. 监控服务已启动，当mihomo节点断开时会自动断开对应的L2TP连接"
echo "========================================"
