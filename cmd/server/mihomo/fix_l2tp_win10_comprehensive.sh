#!/bin/bash

# 综合修复 Windows 10 L2TP 连接问题脚本
echo "========================================"
echo "Windows 10 L2TP 连接问题综合修复脚本"
echo "========================================"

# 检查是否以root用户运行
if [ "$EUID" -ne 0 ]; then
    echo "错误：请以root用户运行此脚本"
    exit 1
fi

# 步骤1：安装必要的软件包
echo "[1/8] 安装必要的软件包..."
apt update
apt install -y xl2tpd ppp iptables libreswan

# 步骤2：备份原有配置
echo "[2/8] 备份原有配置..."
cp /etc/xl2tpd/xl2tpd.conf /etc/xl2tpd/xl2tpd.conf.bak 2>/dev/null || true
cp /etc/ipsec.conf /etc/ipsec.conf.bak 2>/dev/null || true
cp /etc/ipsec.secrets /etc/ipsec.secrets.bak 2>/dev/null || true
cp /etc/ppp/options.xl2tpd /etc/ppp/options.xl2tpd.bak 2>/dev/null || true

# 步骤3：配置 IPsec
echo "[3/8] 配置 IPsec..."
cat > /etc/ipsec.conf << 'EOF'
config setup
    virtual_private=%v4:10.0.0.0/8,%v4:192.168.0.0/16,%v4:172.16.0.0/12
    protostack=netkey
    nhelpers=0

conn L2TP-PSK-NAT
    rightsubnet=0.0.0.0/0
    also=L2TP-PSK-noNAT

conn L2TP-PSK-noNAT
    authby=secret
    pfs=no
    auto=add
    keyingtries=3
    rekey=no
    ikelifetime=8h
    keylife=1h
    type=transport
    left=%defaultroute
    leftprotoport=17/1701
    right=%any
    rightprotoport=17/%any
    dpddelay=40
    dpdtimeout=130
    dpdaction=clear
    ike=aes256-sha1-modp1024,aes128-sha1-modp1024,3des-sha1-modp1024
    esp=aes256-sha1,3des-sha1,aes128-sha1
EOF

# 步骤4：设置 IPsec 预共享密钥
echo "[4/8] 设置 IPsec 预共享密钥..."
read -p "请输入 IPsec 预共享密钥（默认为 'your_psk_key_here'）：" PSK
PSK=${PSK:-your_psk_key_here}

cat > /etc/ipsec.secrets << EOF
%any %any : PSK "$PSK"
EOF

# 步骤5：配置 xl2tpd
echo "[5/8] 配置 xl2tpd..."
cat > /etc/xl2tpd/xl2tpd.conf << 'EOF'
[global]
listen-addr = 0.0.0.0
ipsec saref = yes

[lns default]
ip range = 10.0.10.2-10.0.10.254
local ip = 10.0.10.1
refuse chap = yes
refuse pap = yes
require authentication = yes
ppp debug = yes
pppoptfile = /etc/ppp/options.xl2tpd
length bit = yes
EOF

# 步骤6：配置 PPP 选项
echo "[6/8] 配置 PPP 选项..."
cat > /etc/ppp/options.xl2tpd << 'EOF'
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
mtu 1400
mru 1400
noipx
EOF

# 步骤7：配置防火墙规则
echo "[7/8] 配置防火墙规则..."

# 获取主网络接口名称
MAIN_INTERFACE=$(ip route | grep default | awk '{print $5}')
if [ -z "$MAIN_INTERFACE" ]; then
    MAIN_INTERFACE=$(ip link | grep -v lo | grep UP | head -n 1 | awk '{print $2}' | sed 's/://')
fi

echo "检测到主网络接口: $MAIN_INTERFACE"

# 清除旧的规则
iptables -F
iptables -t nat -F
iptables -t mangle -F

# 允许 SSH
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# 允许 IPsec 和 L2TP 相关端口
iptables -A INPUT -p udp --dport 500 -j ACCEPT  # IKE
iptables -A INPUT -p udp --dport 4500 -j ACCEPT  # NAT-T
iptables -A INPUT -p udp --dport 1701 -j ACCEPT  # L2TP
iptables -A INPUT -p esp -j ACCEPT  # ESP 协议
iptables -A INPUT -p ah -j ACCEPT  # AH 协议

# 允许转发
iptables -A FORWARD -s 10.0.10.0/24 -j ACCEPT
iptables -A FORWARD -d 10.0.10.0/24 -j ACCEPT
if [ ! -z "$MAIN_INTERFACE" ]; then
    iptables -t nat -A POSTROUTING -s 10.0.10.0/24 -o $MAIN_INTERFACE -j MASQUERADE
else
    iptables -t nat -A POSTROUTING -s 10.0.10.0/24 -o ens18 -j MASQUERADE 2>/dev/null || true
    iptables -t nat -A POSTROUTING -s 10.0.10.0/24 -o eth0 -j MASQUERADE 2>/dev/null || true
fi

# 允许已建立的连接
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# 允许回环接口
iptables -A INPUT -i lo -j ACCEPT

# 设置默认策略
iptables -P INPUT DROP
iptables -P FORWARD ACCEPT
iptables -P OUTPUT ACCEPT

# 保存防火墙规则
mkdir -p /etc/iptables
iptables-save > /etc/iptables/rules.v4

# 确保防火墙规则在重启后仍然生效
echo "netfilter-persistent service enable" | bash 2>/dev/null || true


# 步骤8：启用 IP 转发并重启服务
echo "[8/8] 启用 IP 转发并重启服务..."

# 启用 IP 转发
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv4.conf.all.forwarding=1
sysctl -w net.ipv4.conf.default.forwarding=1
if [ ! -z "$MAIN_INTERFACE" ]; then
    sysctl -w net.ipv4.conf.$MAIN_INTERFACE.forwarding=1 2>/dev/null || true
fi

# 确保 IP 转发在重启后仍然启用
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv4.conf.all.forwarding=1" >> /etc/sysctl.conf
echo "net.ipv4.conf.default.forwarding=1" >> /etc/sysctl.conf
if [ ! -z "$MAIN_INTERFACE" ]; then
    echo "net.ipv4.conf.$MAIN_INTERFACE.forwarding=1" >> /etc/sysctl.conf 2>/dev/null || true
fi

# 重启服务
systemctl daemon-reload

# 停止服务
echo "停止服务..."
systemctl stop xl2tpd 2>/dev/null || true
systemctl stop ipsec 2>/dev/null || true

# 等待服务停止
sleep 3

# 启动服务
echo "启动 IPsec 服务..."
systemctl start ipsec
sleep 2

# 检查 IPsec 服务状态
echo "检查 IPsec 服务状态..."
systemctl status ipsec --no-pager

# 启动 xl2tpd 服务
echo "启动 xl2tpd 服务..."
systemctl start xl2tpd 2>/dev/null || true
sleep 2

# 检查 xl2tpd 服务状态
echo "检查 xl2tpd 服务状态..."
systemctl status xl2tpd --no-pager

# 启用服务自启
echo "启用服务自启..."
systemctl enable ipsec
systemctl enable xl2tpd 2>/dev/null || true

# 检查服务状态
echo ""
echo "========================================"
echo "服务状态检查"
echo "========================================"
echo "IPsec 服务状态："
systemctl status ipsec --no-pager
echo ""
echo "xl2tpd 服务状态："
systemctl status xl2tpd --no-pager
echo ""
echo "端口监听检查："
ss -tulnp | grep -E '(500|4500|1701)'
echo ""

# 显示 Windows 10 连接设置
echo "========================================"
echo "Windows 10 L2TP/IPsec 连接设置"
echo "========================================"
echo "1. VPN 类型: 使用 IPsec 的第 2 层隧道协议 (L2TP/IPsec)"
echo "2. 预共享密钥: $PSK"
echo "3. 用户名: user1-user253"
echo "4. 密码: pass1-pass253"
echo ""
echo "========================================"
echo "Windows 10 注册表修改（必须执行）"
echo "========================================"
echo "请在 Windows 10 客户端上执行以下注册表修改："
echo ""
echo "1. 修改注册表项 1（允许没有证书的 L2TP 连接）："
echo "   - 路径: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\PolicyAgent"
echo "   - 创建 DWORD 值: AssumeUDPEncapsulationContextOnSendRule"
echo "   - 设置值为: 2"
echo ""
echo "2. 修改注册表项 2（禁用 IPsec）："
echo "   - 路径: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\RasMan\Parameters"
echo "   - 创建 DWORD 值: ProhibitIpSec"
echo "   - 设置值为: 1"
echo ""
echo "3. 修改注册表项 3（增加 L2TP 超时）："
echo "   - 路径: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\RasMan\Parameters"
echo "   - 创建 DWORD 值: DisabledProtocols"
echo "   - 设置值为: 0"
echo ""
echo "4. 修改注册表项 4（启用老版 VPN 客户端）："
echo "   - 路径: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\RasMan\Parameters"
echo "   - 创建 DWORD 值: LegacyPptpIPsec"
echo "   - 设置值为: 1"
echo ""
echo "5. 重启 Windows 10 电脑"
echo ""
echo "========================================"
echo "其他可能的解决方案"
echo "========================================"
echo "1. 确保 Windows 10 已更新到最新版本"
echo "2. 临时禁用 Windows 防火墙和第三方安全软件"
echo "3. 检查网络连接，确保没有网络限制"
echo "4. 尝试使用有线连接而不是无线连接"
echo "5. 检查服务器的公网 IP 是否可访问"
echo ""
echo "========================================"
echo "修复完成！"
echo "========================================"
echo "请按照上述步骤在 Windows 10 客户端上进行设置，然后尝试重新连接 VPN。"
echo ""
echo "如果仍然无法连接，请检查服务器日志："
echo "  journalctl -u ipsec -n 50"
echo "  journalctl -u xl2tpd -n 50"
echo "  tail -f /var/log/syslog"
echo ""
