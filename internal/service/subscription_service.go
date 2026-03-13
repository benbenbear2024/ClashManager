package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"clash-manager/internal/config"
	"clash-manager/internal/model"
)

func min(a, b int) int {
	if a < b {
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(source.URL)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch subscription: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("subscription returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %v", err)
	}

	content := string(body)

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
		for i := len(existingNodes); i > 0; i-- {
			nodeService.DeleteNode(uint(i))
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

	decodedContent := content

	decoded, err := tryBase64Decode(content)
	if err == nil {
		decodedContent = decoded
	}

	lines := strings.Split(decodedContent, "\n")
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
