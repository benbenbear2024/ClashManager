package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"clash-manager/internal/model"
)

// 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}

func ParseLink(link string) (*model.Node, error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return nil, fmt.Errorf("empty link")
	}

	if strings.HasPrefix(link, "ss://") {
		return parseShadowsocksLink(link)
	} else if strings.HasPrefix(link, "vmess://") {
		return parseVMessLink(link)
	} else if strings.HasPrefix(link, "trojan://") {
		return parseTrojanLink(link)
	} else if strings.HasPrefix(link, "vless://") {
		return parseVLESSLink(link)
	} else if strings.HasPrefix(link, "socks5://") {
		return parseSOCKS5Link(link)
	} else if strings.HasPrefix(link, "hysteria2://") || strings.HasPrefix(link, "hy2://") {
		return parseHysteria2Link(link)
	} else if strings.HasPrefix(link, "hysteria://") {
		return parseHysteriaLink(link)
	} else {
		// 尝试解析自定义格式
		return parseCustomFormat(link)
	}
}

// parseCustomFormat 解析自定义格式：server/port/username/password/date
// 支持两种分隔符：/ 或 |
// 格式示例：
//
//	192.168.1.100/1080/user1/pass123/2025-12-31
//	198.51.100.77|10808|bob|builder|2026-03-01
func parseCustomFormat(link string) (*model.Node, error) {
	// 确定分隔符 - 优先检查 | 分隔符
	var parts []string
	if strings.Contains(link, "|") {
		parts = strings.Split(link, "|")
	} else if strings.Count(link, "/") >= 3 {
		// 只有当至少有3个 / 时才认为是自定义格式（server/port/user/pass）
		parts = strings.Split(link, "/")
	} else {
		// 尝试作为标准 socks5 URL 解析
		return parseSOCKS5Link("socks5://" + link)
	}

	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid custom format, need at least 4 parts: server, port, username, password")
	}

	server := strings.TrimSpace(parts[0])
	port, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}
	username := strings.TrimSpace(parts[2])
	password := strings.TrimSpace(parts[3])

	// 确定节点名称：如果第5部分存在且不为空，则使用它；否则自动生成
	var name string
	if len(parts) >= 5 && strings.TrimSpace(parts[4]) != "" {
		name = strings.TrimSpace(parts[4])
	} else {
		// 生成节点名称：SK5_月份日期_5位随机数
		now := time.Now()
		name = fmt.Sprintf("SK5_%02d%02d_%s", now.Month(), now.Day(), generateRandomString(5))
	}

	node := &model.Node{
		Type:     "socks5",
		Server:   server,
		Port:     port,
		Username: username,
		Password: password,
		Name:     name,
		Rename:   server, // 将服务器地址保存到 Rename 字段
	}

	return node, nil
}

func parseShadowsocksLink(link string) (*model.Node, error) {
	parts := strings.SplitN(link, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss link")
	}

	var serverInfo, name string
	if idx := strings.Index(parts[1], "#"); idx != -1 {
		serverInfo = parts[1][:idx]
		name = parts[1][idx+1:]
	} else {
		serverInfo = parts[1]
	}

	// 分离 base64 编码部分和服务器地址
	// SS 链接格式: ss://base64(server:port:method:password)@server:port#name
	// 或者: ss://base64(method:password)@server:port#name
	var serverAddr string
	if atIdx := strings.Index(serverInfo, "@"); atIdx != -1 {
		serverAddr = serverInfo[atIdx+1:]
		serverInfo = serverInfo[:atIdx]
	}

	// 先进行 URL 解码（处理 %3D 等编码字符）
	serverInfo, _ = url.QueryUnescape(serverInfo)

	// 尝试多种 base64 解码方式
	var decoded []byte
	var err error

	// 1. 尝试标准 base64
	decoded, err = base64.StdEncoding.DecodeString(serverInfo)
	if err != nil {
		// 2. 尝试 URL 安全的 base64
		decoded, err = base64.URLEncoding.DecodeString(serverInfo)
		if err != nil {
			// 3. 尝试 RawStdEncoding（无填充）
			decoded, err = base64.RawStdEncoding.DecodeString(serverInfo)
			if err != nil {
				// 4. 尝试 RawURLEncoding（无填充的 URL 安全）
				decoded, err = base64.RawURLEncoding.DecodeString(serverInfo)
				if err != nil {
					return nil, fmt.Errorf("failed to decode ss link: %v", err)
				}
			}
		}
	}

	// 解析解码后的内容
	// 格式可能是: method:password 或 server:port:method:password
	var method, password string
	decodedStr := string(decoded)

	// 如果已经有 serverAddr，则 decoded 只包含 method:password
	if serverAddr != "" {
		// 从 serverAddr 解析服务器和端口
		serverParts := strings.SplitN(serverAddr, ":", 2)
		if len(serverParts) != 2 {
			return nil, fmt.Errorf("invalid server:port format in URL")
		}
		serverAddr = serverParts[0]
		port, _ := strconv.Atoi(serverParts[1])

		// decoded 包含 method:password
		methodPass := strings.SplitN(decodedStr, ":", 2)
		if len(methodPass) == 2 {
			method = methodPass[0]
			password = methodPass[1]
		} else {
			method = decodedStr
		}

		node := &model.Node{
			Type:     "ss",
			Server:   serverAddr,
			Port:     port,
			Cipher:   method,
			Password: password,
			UDP:      true, // SS 默认启用 UDP
		}

		if name != "" {
			decodedName, err := url.QueryUnescape(name)
			if err == nil {
				node.Name = decodedName
			} else {
				node.Name = name
			}
		}

		return node, nil
	}

	// 旧格式: server:port:method:password 全部在 base64 中
	infoParts := strings.SplitN(decodedStr, "@", 2)
	if len(infoParts) != 2 {
		return nil, fmt.Errorf("invalid ss info format")
	}

	method = infoParts[0]
	serverAndPort := infoParts[1]

	serverParts := strings.SplitN(serverAndPort, ":", 2)
	if len(serverParts) != 2 {
		return nil, fmt.Errorf("invalid server:port format")
	}

	port, err := strconv.Atoi(serverParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	node := &model.Node{
		Type:     "ss",
		Server:   serverParts[0],
		Port:     port,
		Cipher:   method,
		Password: "",
		UDP:      true, // SS 默认启用 UDP
	}

	if name != "" {
		decodedName, err := url.QueryUnescape(name)
		if err == nil {
			node.Name = decodedName
		} else {
			node.Name = name
		}
	}

	return node, nil
}

func parseVMessLink(link string) (*model.Node, error) {
	parts := strings.SplitN(link, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid vmess link")
	}

	// 尝试多种 base64 解码方式
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		// 尝试 URL 安全的 base64 解码
		decoded, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			// 尝试原始 base64 解码
			decoded, err = base64.RawStdEncoding.DecodeString(parts[1])
			if err != nil {
				// 尝试原始 URL 安全的 base64 解码
				decoded, err = base64.RawURLEncoding.DecodeString(parts[1])
				if err != nil {
					return nil, fmt.Errorf("failed to decode vmess link: %v", err)
				}
			}
		}
	}

	// 解析 JSON
	var vmessConfig struct {
		V    string `json:"v"`
		Ps   string `json:"ps"`
		Add  string `json:"add"`
		Port string `json:"port"`
		ID   string `json:"id"`
		Aid  string `json:"aid"`
		Scy  string `json:"scy"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		Sni  string `json:"sni"`
		Alpn string `json:"alpn"`
	}

	if err := json.Unmarshal(decoded, &vmessConfig); err != nil {
		return nil, fmt.Errorf("failed to parse vmess config: %v", err)
	}

	port, _ := strconv.Atoi(vmessConfig.Port)
	node := &model.Node{
		Type:    "vmess",
		Server:  vmessConfig.Add,
		Port:    port,
		UUID:    vmessConfig.ID,
		AlterId: vmessConfig.Aid,
		Cipher:  vmessConfig.Scy,
		Network: vmessConfig.Net,
		Path:    vmessConfig.Path,
		Host:    vmessConfig.Host,
		TLS:     vmessConfig.TLS == "tls" || vmessConfig.TLS == "True" || vmessConfig.TLS == "true",
	}

	if vmessConfig.Ps != "" {
		node.Name = vmessConfig.Ps
	}

	return node, nil
}

func parseTrojanLink(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid trojan url: %v", err)
	}

	port, _ := strconv.Atoi(u.Port())
	node := &model.Node{
		Type:     "trojan",
		Server:   u.Hostname(),
		Port:     port,
		Password: u.User.Username(),
		Network:  "tcp",
		TLS:      true,
		UDP:      true,
	}

	// 解析查询参数
	query := u.Query()

	// 设置 SNI
	if sni := query.Get("sni"); sni != "" {
		node.Host = sni
	} else {
		node.Host = u.Hostname()
	}

	// 设置跳过证书验证
	if query.Get("allowInsecure") == "1" {
		node.SkipCert = true
	}

	if name := u.Fragment; name != "" {
		decodedName, err := url.QueryUnescape(name)
		if err == nil {
			node.Name = decodedName
		} else {
			node.Name = name
		}
	}

	return node, nil
}

func parseVLESSLink(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid vless url: %v", err)
	}

	query := u.Query()

	port, _ := strconv.Atoi(u.Port())
	node := &model.Node{
		Type:    "vless",
		Server:  u.Hostname(),
		Port:    port,
		UUID:    u.User.Username(),
		Network: query.Get("type"),
		Path:    query.Get("path"),
		Host:    query.Get("host"),
		TLS:     query.Get("security") == "tls" || query.Get("security") == "reality",
		ALPN:    query.Get("alpn"),
		UDP:     true,   // VLESS 默认启用 UDP
		Cipher:  "auto", // VLESS 默认加密方式
		AlterId: "0",    // VLESS 默认 alterId
	}

	// 处理 Reality 配置
	if query.Get("security") == "reality" {
		node.PublicKey = query.Get("pbk")
		node.ShortID = query.Get("sid")
		node.ServerName = query.Get("servername")
		node.ClientFingerprint = query.Get("fp")
		if node.ClientFingerprint == "" {
			node.ClientFingerprint = "safari" // 默认指纹
		}
	}

	// 处理流控
	node.Flow = query.Get("flow")

	// 处理 SNI
	if sni := query.Get("sni"); sni != "" {
		node.Host = sni
	}

	if name := u.Fragment; name != "" {
		decodedName, err := url.QueryUnescape(name)
		if err == nil {
			node.Name = decodedName
		} else {
			node.Name = name
		}
	}

	return node, nil
}

func parseSOCKS5Link(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid socks5 url: %v", err)
	}

	port, _ := strconv.Atoi(u.Port())
	node := &model.Node{
		Type:   "socks5",
		Server: u.Hostname(),
		Port:   port,
	}

	if u.User != nil {
		node.Username = u.User.Username()
		if password, ok := u.User.Password(); ok {
			node.Password = password
		}
	}

	if name := u.Fragment; name != "" {
		decodedName, err := url.QueryUnescape(name)
		if err == nil {
			node.Name = decodedName
		} else {
			node.Name = name
		}
	}

	return node, nil
}

func parseHysteria2Link(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid hysteria2 url: %v", err)
	}

	query := u.Query()

	port, _ := strconv.Atoi(u.Port())

	// 解析 up、down、hop-interval 参数
	up, _ := strconv.Atoi(query.Get("up"))
	down, _ := strconv.Atoi(query.Get("down"))
	hopInterval, _ := strconv.Atoi(query.Get("hop-interval"))

	// 设置默认值
	if up == 0 {
		up = 30
	}
	if down == 0 {
		down = 30
	}
	if hopInterval == 0 {
		hopInterval = 60
	}

	node := &model.Node{
		Type:        "hysteria2",
		Server:      u.Hostname(),
		Port:        port,
		Password:    u.User.Username(), // 密码在 username 位置
		TLS:         true,
		Host:        query.Get("sni"),             // SNI 映射到 Host 字段
		SkipCert:    query.Get("insecure") == "1", // insecure 映射到 SkipCert 字段
		Up:          up,
		Down:        down,
		HopInterval: hopInterval,
	}

	if name := u.Fragment; name != "" {
		decodedName, err := url.QueryUnescape(name)
		if err == nil {
			node.Name = decodedName
		} else {
			node.Name = name
		}
	}

	return node, nil
}

func parseHysteriaLink(link string) (*model.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid hysteria url: %v", err)
	}

	query := u.Query()

	port, _ := strconv.Atoi(u.Port())
	node := &model.Node{
		Type:     "hysteria",
		Server:   u.Hostname(),
		Port:     port,
		Password: query.Get("auth"),
		Network:  query.Get("protocol"),
		TLS:      true,
	}

	if name := u.Fragment; name != "" {
		decodedName, err := url.QueryUnescape(name)
		if err == nil {
			node.Name = decodedName
		} else {
			node.Name = name
		}
	}

	return node, nil
}

func ExportLink(node *model.Node) (string, error) {
	switch node.Type {
	case "ss":
		return exportShadowsocksLink(node)
	case "vmess":
		return exportVMessLink(node)
	case "trojan":
		return exportTrojanLink(node)
	case "vless":
		return exportVLESSLink(node)
	case "socks", "socks5":
		return exportSOCKS5Link(node)
	case "hysteria2":
		return exportHysteria2Link(node)
	case "hysteria":
		return exportHysteriaLink(node)
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type)
	}
}

func ParseSubscription(content string) ([]model.Node, error) {
	// 尝试 base64 解码
	decodedContent := content
	decoded, err := TryBase64Decode(content)
	if err == nil {
		decodedContent = decoded
	}

	lines := strings.Split(decodedContent, "\n")
	nodes := make([]model.Node, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		node, err := ParseLink(line)
		if err != nil {
			continue
		}

		nodes = append(nodes, *node)
	}

	return nodes, nil
}

func TryBase64Decode(content string) (string, error) {
	// 先进行 URL 解码（处理 %3D 等编码字符）
	content, _ = url.QueryUnescape(content)

	// 尝试标准 base64
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil {
		return string(decoded), nil
	}

	// 尝试 URL 安全的 base64
	decoded, err = base64.URLEncoding.DecodeString(content)
	if err == nil {
		return string(decoded), nil
	}

	// 尝试 RawStdEncoding（无填充）
	decoded, err = base64.RawStdEncoding.DecodeString(content)
	if err == nil {
		return string(decoded), nil
	}

	// 尝试 RawURLEncoding（无填充的 URL 安全）
	decoded, err = base64.RawURLEncoding.DecodeString(content)
	if err == nil {
		return string(decoded), nil
	}

	// 尝试添加填充后解码
	padded := content
	pad := len(padded) % 4
	if pad > 0 {
		padded += strings.Repeat("=", 4-pad)

		// 再次尝试标准 base64
		decoded, err = base64.StdEncoding.DecodeString(padded)
		if err == nil {
			return string(decoded), nil
		}

		// 再次尝试 URL 安全的 base64
		decoded, err = base64.URLEncoding.DecodeString(padded)
		if err == nil {
			return string(decoded), nil
		}
	}

	return "", fmt.Errorf("all base64 decode attempts failed")
}

func exportShadowsocksLink(node *model.Node) (string, error) {
	userInfo := fmt.Sprintf("%s:%s", node.Cipher, node.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(userInfo))
	link := fmt.Sprintf("ss://%s@%s:%d", encoded, node.Server, node.Port)
	name := node.Rename
	if name == "" {
		name = node.Name
	}
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}
	return link, nil
}

func exportVMessLink(node *model.Node) (string, error) {
	u := url.URL{
		Scheme: "vmess",
		Host:   fmt.Sprintf("%s:%d", node.Server, node.Port),
	}

	q := u.Query()
	q.Set("id", node.UUID)
	q.Set("alterId", node.AlterId)
	q.Set("cipher", node.Cipher)
	q.Set("net", node.Network)
	q.Set("path", node.Path)
	q.Set("host", node.Host)
	if node.TLS {
		q.Set("tls", "true")
	}
	name := node.Rename
	if name == "" {
		name = node.Name
	}
	if name != "" {
		q.Set("ps", name)
	}

	u.RawQuery = q.Encode()
	encoded := base64.StdEncoding.EncodeToString([]byte(u.String()[8:]))
	return "vmess://" + encoded, nil
}

func exportTrojanLink(node *model.Node) (string, error) {
	name := node.Rename
	if name == "" {
		name = node.Name
	}
	u := url.URL{
		Scheme:   "trojan",
		Host:     fmt.Sprintf("%s:%d", node.Server, node.Port),
		User:     url.UserPassword(node.Password, ""),
		Fragment: name,
	}
	return u.String(), nil
}

func exportVLESSLink(node *model.Node) (string, error) {
	name := node.Rename
	if name == "" {
		name = node.Name
	}
	u := url.URL{
		Scheme:   "vless",
		Host:     fmt.Sprintf("%s:%d", node.Server, node.Port),
		User:     url.User(node.UUID),
		Fragment: name,
	}

	q := u.Query()
	q.Set("type", node.Network)
	q.Set("path", node.Path)
	q.Set("host", node.Host)
	if node.TLS {
		q.Set("security", "tls")
	}
	if node.ALPN != "" {
		q.Set("alpn", node.ALPN)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func exportSOCKS5Link(node *model.Node) (string, error) {
	name := node.Rename
	if name == "" {
		name = node.Name
	}
	u := url.URL{
		Scheme:   "socks5",
		Host:     fmt.Sprintf("%s:%d", node.Server, node.Port),
		Fragment: name,
	}

	if node.Username != "" {
		if node.Password != "" {
			u.User = url.UserPassword(node.Username, node.Password)
		} else {
			u.User = url.User(node.Username)
		}
	}

	return u.String(), nil
}

func exportHysteria2Link(node *model.Node) (string, error) {
	name := node.Rename
	if name == "" {
		name = node.Name
	}
	u := url.URL{
		Scheme:   "hysteria2",
		Host:     fmt.Sprintf("%s:%d", node.Server, node.Port),
		Fragment: name,
	}

	q := u.Query()
	if node.Password != "" {
		q.Set("password", node.Password)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func exportHysteriaLink(node *model.Node) (string, error) {
	name := node.Rename
	if name == "" {
		name = node.Name
	}
	u := url.URL{
		Scheme:   "hysteria",
		Host:     fmt.Sprintf("%s:%d", node.Server, node.Port),
		Fragment: name,
	}

	q := u.Query()
	if node.Password != "" {
		q.Set("auth", node.Password)
	}
	if node.Network != "" {
		q.Set("protocol", node.Network)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
