#!/bin/bash

# 综合修复脚本
echo "========================================"
echo "ClashManager 综合修复脚本"
echo "========================================"

# 检查是否以root用户运行
if [ "$EUID" -ne 0 ]; then
    echo "错误：请以root用户运行此脚本"
    exit 1
fi

# 显示菜单
echo "请选择要执行的修复操作："
echo "1. 修复 Windows 10 L2TP 连接问题"
echo "2. 修复 mihomo 服务启动参数"
echo "3. 修复 mihomo 开机自启问题"
echo "4. 执行所有修复操作"
echo ""
read -p "请输入选项编号：" choice

echo ""

case $choice in
    1|4)
        echo "========================================"
        echo "修复 Windows 10 L2TP 连接问题"
        echo "========================================"
        
        # 安装 Libreswan（IPsec 实现）
        echo "[1/6] 安装 Libreswan..."
        apt update
        apt install -y libreswan
        
        # 备份原有配置
        echo "[2/6] 备份原有配置..."
        cp /etc/xl2tpd/xl2tpd.conf /etc/xl2tpd/xl2tpd.conf.bak 2>/dev/null || true
        cp /etc/ipsec.conf /etc/ipsec.conf.bak 2>/dev/null || true
        
        # 配置 IPsec
        echo "[3/6] 配置 IPsec..."
        cat > /etc/ipsec.conf << 'EOF'
config setup
    virtual_private=%v4:10.0.0.0/8,%v4:192.168.0.0/16,%v4:172.16.0.0/12
    protostack=netkey
    listen=yes

conn L2TP-PSK-NAT
    rightsubnet=vhost:%priv
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
EOF
        
        # 设置 IPsec 预共享密钥
        echo "[4/6] 设置 IPsec 预共享密钥..."
        cat > /etc/ipsec.secrets << 'EOF'
%any %any : PSK "your_psk_key_here"
EOF
        
        # 修改 xl2tpd 配置
        echo "[5/6] 修改 xl2tpd 配置..."
        cat > /etc/xl2tpd/xl2tpd.conf << 'EOF'
[global]
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
        
        # 配置防火墙规则
        echo "[6/6] 配置防火墙规则..."
        
        # 清除旧的 L2TP 相关规则
        iptables -D FORWARD -s 10.0.10.0/24 -j ACCEPT 2>/dev/null || true
        iptables -D FORWARD -d 10.0.10.0/24 -j ACCEPT 2>/dev/null || true
        iptables -t nat -D POSTROUTING -s 10.0.10.0/24 -o eth0 -j MASQUERADE 2>/dev/null || true
        
        # 添加新的防火墙规则
        # 允许 UDP 500 (IKE)
        iptables -A INPUT -p udp --dport 500 -j ACCEPT
        # 允许 UDP 4500 (NAT-T)
        iptables -A INPUT -p udp --dport 4500 -j ACCEPT
        # 允许 UDP 1701 (L2TP)
        iptables -A INPUT -p udp --dport 1701 -j ACCEPT
        # 允许 ESP 协议
        iptables -A INPUT -p esp -j ACCEPT
        # 允许 AH 协议
        iptables -A INPUT -p ah -j ACCEPT
        
        # NAT 和转发规则
        iptables -t nat -A POSTROUTING -s 10.0.10.0/24 -o eth0 -j MASQUERADE
        iptables -A FORWARD -s 10.0.10.0/24 -j ACCEPT
        iptables -A FORWARD -d 10.0.10.0/24 -j ACCEPT
        
        # 保存防火墙规则
        mkdir -p /etc/iptables
        iptables-save > /etc/iptables/rules.v4
        
        # 启用 IP 转发
        echo "启用 IP 转发..."
        sysctl -w net.ipv4.ip_forward=1
        echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
        
        # 重启服务
        echo "重启服务..."
        systemctl restart xl2tpd
        systemctl enable ipsec
        systemctl restart ipsec
        
        # 检查服务状态
        echo ""
        echo "========================================"
        echo "配置完成！"
        echo "========================================"
        echo ""
        echo "IPsec 服务状态："
        systemctl status ipsec --no-pager
        echo ""
        echo "xl2tpd 服务状态："
        systemctl status xl2tpd --no-pager
        echo ""
        echo "========================================"
        echo "Windows 10 连接设置："
        echo "========================================"
        echo ""
        echo "1. VPN 类型: 使用 IPsec 的第 2 层隧道协议 (L2TP/IPsec)"
        echo "2. 预共享密钥: your_psk_key_here"
        echo "3. 用户名: user1-user253"
        echo "4. 密码: pass1-pass253"
        echo ""
        echo "如果仍然无法连接，请在 Windows 10 上执行以下操作："
        echo ""
        echo "方法1: 修改注册表（允许没有证书的 L2TP 连接）"
        echo "  1. 打开注册表编辑器 (regedit)"
        echo "  2. 导航到: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\PolicyAgent"
        echo "  3. 创建 DWORD 值: AssumeUDPEncapsulationContextOnSendRule"
        echo "  4. 设置值为: 1"
        echo "  5. 重启电脑"
        echo ""
        echo "方法2: 如果服务器在 NAT 后面，还需要修改注册表："
        echo "  1. 导航到: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\RasMan\Parameters"
        echo "  2. 创建 DWORD 值: ProhibitIpSec"
        echo "  3. 设置值为: 1"
        echo "  4. 重启电脑"
        echo ""
        ;;
    
    2|4)
        echo "========================================"
        echo "修复 mihomo 服务启动参数"
        echo "========================================"
        
        echo "[1/3] 停止 mihomo 服务..."
        systemctl stop mihomo
        
        echo "[2/3] 更新服务文件..."
        cat > /etc/systemd/system/mihomo.service << 'EOF'
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
        
        echo "[3/3] 重新加载并启动服务..."
        systemctl daemon-reload
        systemctl enable mihomo
        systemctl start mihomo
        
        echo ""
        echo "========================================"
        echo "服务状态："
        echo "========================================"
        systemctl status mihomo --no-pager
        
        echo ""
        echo "========================================"
        echo "检查端口监听："
        echo "========================================"
        ss -tulnp | grep -E '(7890|7891|9090)' || echo "端口未监听，请检查配置文件"
        
        echo ""
        ;;
    
    3|4)
        echo "========================================"
        echo "修复 mihomo 开机自启问题"
        echo "========================================"
        
        # 检查 mihomo 服务文件是否存在
        if [ ! -f /etc/systemd/system/mihomo.service ]; then
            echo "错误：mihomo 服务文件不存在"
            exit 1
        fi
        
        echo "[1/5] 检查 mihomo 服务文件..."
        cat /etc/systemd/system/mihomo.service
        
        echo ""
        echo "[2/5] 检查 mihomo 程序是否存在..."
        if [ ! -f /opt/mihomo/mihomo ]; then
            echo "错误：mihomo 程序不存在"
            exit 1
        fi
        ls -lh /opt/mihomo/mihomo
        
        echo ""
        echo "[3/5] 检查 mihomo 配置文件..."
        if [ ! -f /etc/mihomo/config.yaml ]; then
            echo "错误：mihomo 配置文件不存在"
            exit 1
        fi
        ls -lh /etc/mihomo/config.yaml
        
        echo ""
        echo "[4/5] 重新加载 systemd 并设置开机自启..."
        systemctl daemon-reload
        systemctl enable mihomo
        
        echo ""
        echo "[5/5] 检查服务状态..."
        echo ""
        echo "mihomo 服务是否已启用："
        systemctl is-enabled mihomo
        
        echo ""
        echo "mihomo 服务当前状态："
        systemctl status mihomo --no-pager
        
        echo ""
        echo "========================================"
        echo "检查完成！"
        echo "========================================"
        echo ""
        echo "如果服务未启动，请运行："
        echo "  systemctl start mihomo"
        echo ""
        echo "如果服务启动失败，请查看日志："
        echo "  journalctl -u mihomo -n 50"
        echo ""
        ;;
    
    *)
        echo "错误：无效的选项编号"
        exit 1
        ;;
esac

echo "========================================"
echo "修复操作完成！"
echo "========================================"
