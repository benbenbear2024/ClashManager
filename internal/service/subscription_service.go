package service

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"clash-manager/internal/config"
	"clash-manager/internal/model"

	"github.com/andybalholm/brotli"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type SubscriptionSource struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	URL            string     `json:"url"`
	NodeFilter     string     `json:"node_filter"`
	SyncMode       string     `json:"sync_mode"`
	Enabled        bool       `json:"enabled"`
	UpdateInterval int        `json:"updateInterval"`
	LastSync       *time.Time `json:"last_sync"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type SubscriptionService struct{}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{}
}

func (s *SubscriptionService) ListSources() ([]SubscriptionSource, error) {
	sourcesPath := config.GetSubscriptionSourcesPath()

	data, err := config.ReadFile(sourcesPath)
	if err != nil {
		return []SubscriptionSource{}, nil
	}

	var sources []SubscriptionSource
	if err := json.Unmarshal(data, &sources); err != nil {
		return []SubscriptionSource{}, nil
	}

	return sources, nil
}

func (s *SubscriptionService) CreateSource(source *SubscriptionSource) error {
	sources, err := s.ListSources()
	if err != nil {
		return err
	}

	source.ID = uint(len(sources) + 1)
	source.CreatedAt = time.Now()
	source.UpdatedAt = time.Now()

	sources = append(sources, *source)

	return s.saveSources(sources)
}

func (s *SubscriptionService) UpdateSource(id uint, source *SubscriptionSource) error {
	sources, err := s.ListSources()
	if err != nil {
		return err
	}

	if id < 1 || int(id) > len(sources) {
		return fmt.Errorf("invalid source id: %d", id)
	}

	source.ID = id
	source.UpdatedAt = time.Now()
	sources[id-1] = *source

	return s.saveSources(sources)
}

func (s *SubscriptionService) DeleteSource(id uint) error {
	sources, err := s.ListSources()
	if err != nil {
		return err
	}

	if id < 1 || int(id) > len(sources) {
		return fmt.Errorf("invalid source id: %d", id)
	}

	sources = append(sources[:id-1], sources[id:]...)

	return s.saveSources(sources)
}

func (s *SubscriptionService) SyncSource(id uint) (int, error) {
	sources, err := s.ListSources()
	if err != nil {
		return 0, err
	}

	if id < 1 || int(id) > len(sources) {
		return 0, fmt.Errorf("invalid source id: %d", id)
	}

	source := sources[id-1]

	// 创建带有重试机制的 HTTP 客户端
	var body []byte
	var contentType, contentEncoding string
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequest("GET", source.URL, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to create request: %v", err)
		}

		// 设置 User-Agent 和其他请求头
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		if err != nil {
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return 0, fmt.Errorf("failed to fetch subscription after %d retries: %v", maxRetries, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return 0, fmt.Errorf("subscription returned status: %d", resp.StatusCode)
		}

		// 保存响应头信息
		contentType = resp.Header.Get("Content-Type")
		contentEncoding = resp.Header.Get("Content-Encoding")

		// 打印响应头信息
		fmt.Printf("[SyncSource] ===== SUBSCRIPTION DEBUG (Attempt %d) =====\n", i+1)
		fmt.Printf("[SyncSource] Status: %d\n", resp.StatusCode)
		fmt.Printf("[SyncSource] Content-Encoding: %s\n", contentEncoding)
		fmt.Printf("[SyncSource] Content-Type: %s\n", contentType)

		// 检查内容编码并解压
		var reader io.ReadCloser = resp.Body
		if contentEncoding == "gzip" {
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				resp.Body.Close()
				fmt.Printf("[SyncSource] Failed to create gzip reader: %v\n", err)
				if i < maxRetries-1 {
					time.Sleep(time.Duration(i+1) * time.Second)
					continue
				}
				return 0, fmt.Errorf("failed to create gzip reader: %v", err)
			}
		} else if contentEncoding == "br" {
			// Wrap brotli.Reader in a struct that implements io.ReadCloser
			brReader := brotli.NewReader(resp.Body)
			reader = struct {
				io.Reader
				io.Closer
			}{
				Reader: brReader,
				Closer: resp.Body,
			}
		}

		// 读取响应体
		body, err = io.ReadAll(reader)
		reader.Close()
		// resp.Body.Close() 已经在 reader.Close() 中处理了

		if err != nil {
			fmt.Printf("[SyncSource] Failed to read body: %v\n", err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return 0, fmt.Errorf("failed to read response after %d retries: %v", maxRetries, err)
		}

		fmt.Printf("[SyncSource] Body length: %d\n", len(body))

		// 成功获取数据，跳出重试循环
		break
	}

	content := string(body)

	// 打印内容信息用于调试
	fmt.Printf("[SyncSource] ===== CONTENT ANALYSIS =====\n")
	fmt.Printf("[SyncSource] Content length: %d\n", len(content))

	// 检查内容是否是可读的文本
	printableCount := 0
	for i, b := range body {
		if i >= 100 {
			break
		}
		if b >= 32 && b < 127 || b == '\n' || b == '\r' || b == '\t' {
			printableCount++
		}
	}
	fmt.Printf("[SyncSource] First 100 bytes printable ratio: %d%%\n", printableCount)

	// 尝试直接作为 Base64 解码（订阅源通常是 base64 编码的）
	fmt.Printf("[SyncSource] Trying base64 decode...\n")
	decoded, err := TryBase64Decode(content)
	if err == nil {
		fmt.Printf("[SyncSource] Base64 decode successful, new length: %d\n", len(decoded))
		content = decoded
	} else {
		fmt.Printf("[SyncSource] Base64 decode failed: %v\n", err)
	}

	fmt.Printf("[SyncSource] Content preview (first 500 chars): %s\n", content[:min(500, len(content))])
	fmt.Printf("[SyncSource] Content preview (last 500 chars): %s\n", content[max(0, len(content)-500):])

	nodes, err := s.parseNodesFromContent(content)
	if err != nil {
		return 0, fmt.Errorf("failed to parse nodes: %v", err)
	}

	// 应用节点过滤
	filteredNodes := nodes
	if source.NodeFilter != "" {
		filters := strings.Split(source.NodeFilter, ",")
		filtered := make([]model.Node, 0)
		for _, node := range nodes {
			shouldExclude := false
			for _, filter := range filters {
				filter = strings.TrimSpace(filter)
				if filter != "" && strings.Contains(strings.ToLower(node.Name), strings.ToLower(filter)) {
					shouldExclude = true
					break
				}
			}
			if !shouldExclude {
				filtered = append(filtered, node)
			}
		}
		filteredNodes = filtered
	}

	nodeService := NewNodeService()
	existingNodes, err := nodeService.ListNodes()
	if err != nil {
		return 0, fmt.Errorf("failed to get existing nodes: %v", err)
	}

	count := 0
	switch source.SyncMode {
	case "replace":
			// 替换模式：清空所有节点
			for {
				currentNodes, err := nodeService.ListNodes()
				if err != nil {
					fmt.Printf("[SyncSource] Failed to list nodes: %v\n", err)
					break
				}
				if len(currentNodes) == 0 {
					break
				}
				// 从后往前删除节点
				if err := nodeService.DeleteNode(uint(len(currentNodes))); err != nil {
					fmt.Printf("[SyncSource] Failed to delete node: %v\n", err)
					break
				}
			}

			// 确保配置文件被清空
			cfg, err := config.LoadConfig()
			if err == nil && len(cfg.Proxies) > 0 {
				cfg.Proxies = []config.ProxyConfig{}
				config.SaveConfig(cfg)
			}

			// 重新获取现有节点（已清空）
			existingNames := make(map[string]bool)
			for _, node := range filteredNodes {
				originalName := node.Name
				newName := ""
				for j := 1; ; j++ {
					newName = fmt.Sprintf("Name%d", j)
					if !existingNames[newName] {
						break
					}
				}

				node.Rename = originalName
				node.Name = newName
				node.Source = source.Name

				if err := nodeService.CreateNode(&node); err != nil {
					fmt.Printf("[SyncSource] Failed to create node: %v\n", err)
					continue
				}

				existingNames[newName] = true
				count++
			}

	case "smart":
		// 智能合并模式：根据Rename字段判断节点是否存在
		renameToNode := make(map[string]model.Node)
		for _, n := range existingNodes {
			if n.Rename != "" {
				renameToNode[n.Rename] = n
			}
		}

		existingNames := make(map[string]bool)
		for _, n := range existingNodes {
			existingNames[n.Name] = true
		}

		for _, node := range filteredNodes {
			originalName := node.Name

			// 检查是否已存在（根据Rename字段）
			if existingNode, exists := renameToNode[originalName]; exists {
				// 节点存在，检查是否需要更新
				node.ID = existingNode.ID
				node.Name = existingNode.Name // 保持原有Name
				node.Rename = originalName
				node.Source = source.Name

				if err := nodeService.UpdateNode(existingNode.ID, &node); err != nil {
					continue
				}
				count++
			} else {
				// 节点不存在，创建新节点
				newName := ""
				for j := 1; ; j++ {
					newName = fmt.Sprintf("Name%d", j)
					if !existingNames[newName] {
						break
					}
				}

				node.Rename = originalName
				node.Name = newName
				node.Source = source.Name

				if err := nodeService.CreateNode(&node); err != nil {
					fmt.Printf("[SyncSource] Failed to create node: %v\n", err)
					continue
				}

				existingNames[newName] = true
				count++
			}
		}

	default: // append 模式
		// 追加模式：不判断节点是否存在，直接追加
		existingNames := make(map[string]bool)
		for _, n := range existingNodes {
			existingNames[n.Name] = true
		}

		for _, node := range filteredNodes {
			originalName := node.Name
			newName := ""
			for j := 1; ; j++ {
				newName = fmt.Sprintf("Name%d", j)
				if !existingNames[newName] {
					break
				}
			}

			node.Rename = originalName
			node.Name = newName
			node.Source = source.Name

			if err := nodeService.CreateNode(&node); err != nil {
				continue
			}

			existingNames[newName] = true
			count++
		}
	}

	// 更新最后同步时间
	now := time.Now()
	sources[id-1].LastSync = &now
	sources[id-1].UpdatedAt = now
	s.saveSources(sources)

	return count, nil
}

func (s *SubscriptionService) parseNodesFromContent(content string) ([]model.Node, error) {
	var nodes []model.Node

	fmt.Printf("[parseNodesFromContent] Raw content length: %d\n", len(content))
	fmt.Printf("[parseNodesFromContent] Raw content preview: %s\n", content[:min(200, len(content))])

	// 首先尝试对整个内容进行 base64 解码（针对 ID 3 这种情况）
	decodedContent, err := TryBase64Decode(content)
	if err == nil {
		fmt.Printf("[parseNodesFromContent] Base64 decode successful, new length: %d\n", len(decodedContent))
		fmt.Printf("[parseNodesFromContent] Decoded content preview: %s\n", decodedContent[:min(200, len(decodedContent))])
		content = decodedContent
	} else {
		fmt.Printf("[parseNodesFromContent] Base64 decode failed: %v\n", err)
	}

	// 尝试按换行符分割
	lines := strings.Split(content, "\n")
	fmt.Printf("[parseNodesFromContent] Total lines: %d\n", len(lines))

	// 如果只有一行，尝试按常见的分隔符分割
	if len(lines) == 1 {
		// 尝试按空格分割
		spaceLines := strings.Fields(content)
		if len(spaceLines) > 1 {
			fmt.Printf("[parseNodesFromContent] Split by space, found %d lines\n", len(spaceLines))
			lines = spaceLines
		} else {
			// 尝试按 # 分割（节点名称分隔符）
			hashLines := strings.Split(content, "#")
			if len(hashLines) > 1 {
				fmt.Printf("[parseNodesFromContent] Split by #, found %d lines\n", len(hashLines))
				lines = hashLines
			}
		}
	}

	// 处理每行内容
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 打印前10行用于调试
		if i < 10 {
			fmt.Printf("[parseNodesFromContent] Line %d: %s\n", i, line[:min(100, len(line))])
		}

		// 尝试直接解析
		node, err := ParseLink(line)
		if err != nil {
			// 如果解析失败，尝试 base64 解码后再解析
			decodedLine, decodeErr := TryBase64Decode(line)
			if decodeErr == nil {
				node, err = ParseLink(decodedLine)
			}
		}

		if err != nil {
			if i < 5 {
				fmt.Printf("[parseNodesFromContent] Failed to parse line %d: %v\n", i, err)
			}
			continue
		}

		nodes = append(nodes, *node)
	}

	// 如果没有解析到节点，尝试直接在内容中查找协议前缀
	if len(nodes) == 0 {
		fmt.Printf("[parseNodesFromContent] No nodes parsed, trying direct prefix search\n")
		
		// 支持的协议前缀
		prefixes := []string{
			"ss://",
			"trojan://",
			"vless://",
			"vmess://",
			"socks5://",
			"hysteria2://",
			"hy2://",
			"hysteria://",
			"ssr://",
			"http://",
			"https://",
		}

		// 查找所有链接
		links := []string{}
		currentPos := 0

		for currentPos < len(content) {
			// 查找下一个协议前缀
			nextLinkStart := -1
			nextPrefix := ""
			for _, prefix := range prefixes {
				pos := strings.Index(content[currentPos:], prefix)
				if pos != -1 {
					pos += currentPos
					if nextLinkStart == -1 || pos < nextLinkStart {
						nextLinkStart = pos
						nextPrefix = prefix
					}
				}
			}

			if nextLinkStart == -1 {
				break
			}

			// 查找链接的结束位置（下一个协议前缀或字符串结束）
			nextLinkEnd := len(content)
			for _, prefix := range prefixes {
				pos := strings.Index(content[nextLinkStart+len(nextPrefix):], prefix)
				if pos != -1 {
					pos += nextLinkStart + len(nextPrefix)
					if pos < nextLinkEnd {
						nextLinkEnd = pos
					}
				}
			}

			// 提取链接
			link := strings.TrimSpace(content[nextLinkStart:nextLinkEnd])
			if link != "" {
				links = append(links, link)
			}

			currentPos = nextLinkEnd
		}

		fmt.Printf("[parseNodesFromContent] Found %d links via prefix search\n", len(links))

		// 处理提取的链接
		for i, link := range links {
			// 打印前10个链接用于调试
			if i < 10 {
				fmt.Printf("[parseNodesFromContent] Link %d: %s\n", i, link[:min(100, len(link))])
			}

			// 尝试直接解析
			node, err := ParseLink(link)
			if err != nil {
				// 如果解析失败，尝试 base64 解码后再解析
				decodedLink, decodeErr := TryBase64Decode(link)
				if decodeErr == nil {
					node, err = ParseLink(decodedLink)
				}
			}

			if err != nil {
				if i < 5 {
					fmt.Printf("[parseNodesFromContent] Failed to parse link %d: %v\n", i, err)
				}
				continue
			}

			nodes = append(nodes, *node)
		}
	}

	fmt.Printf("[parseNodesFromContent] Parsed %d nodes\n", len(nodes))
	return nodes, nil
}

func (s *SubscriptionService) saveSources(sources []SubscriptionSource) error {
	sourcesPath := config.GetSubscriptionSourcesPath()

	data, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sources: %v", err)
	}

	if err := config.WriteFile(sourcesPath, data); err != nil {
		return fmt.Errorf("failed to write sources: %v", err)
	}

	return nil
}
