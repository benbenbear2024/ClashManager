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

type SubscriptionSource struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	SyncMode  string    `json:"sync_mode"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func (s *SubscriptionService) SyncSource(id uint) error {
	sources, err := s.ListSources()
	if err != nil {
		return err
	}

	if id < 1 || int(id) > len(sources) {
		return fmt.Errorf("invalid source id: %d", id)
	}

	source := sources[id-1]

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(source.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subscription returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	nodes, err := s.parseNodesFromContent(string(body))
	if err != nil {
		return fmt.Errorf("failed to parse nodes: %v", err)
	}

	nodeService := NewNodeService()
	existingNodes, err := nodeService.ListNodes()
	if err != nil {
		return fmt.Errorf("failed to get existing nodes: %v", err)
	}

	existingNames := make(map[string]bool)
	for _, n := range existingNodes {
		existingNames[n.Name] = true
	}

	for _, node := range nodes {
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
	}

	return nil
}

func (s *SubscriptionService) parseNodesFromContent(content string) ([]model.Node, error) {
	var nodes []model.Node

	lines := strings.Split(content, "\n")
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
