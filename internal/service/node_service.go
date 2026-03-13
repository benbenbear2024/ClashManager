package service

import (
	"clash-manager/internal/config"
	"clash-manager/internal/model"
	"fmt"
)

type NodeService struct{}

func NewNodeService() *NodeService {
	return &NodeService{}
}

func (s *NodeService) ListNodes() ([]model.Node, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	nodes := make([]model.Node, 0, len(cfg.Proxies))
	for i, proxy := range cfg.Proxies {
		node := model.Node{
			ID:       uint(i + 1),
			Name:     proxy.Name,
			Type:     proxy.Type,
			Server:   proxy.Server,
			Port:     proxy.Port,
			Cipher:   proxy.Cipher,
			Password: proxy.Password,
			UDP:      proxy.UDP,
			Username: proxy.Username,
			UUID:     proxy.UUID,
			Network:  proxy.Network,
			TLS:      proxy.TLS,
			SkipCert: proxy.SkipCert,
			Path:     proxy.Path,
			Host:     proxy.Host,
			ALPN:     proxy.ALPN,
			Rename:   proxy.Rename,
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (s *NodeService) CreateNode(node *model.Node) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	proxy := config.ProxyConfig{
		Name:     node.Name,
		Type:     node.Type,
		Server:   node.Server,
		Port:     node.Port,
		Cipher:   node.Cipher,
		Password: node.Password,
		UDP:      node.UDP,
		Username: node.Username,
		UUID:     node.UUID,
		Network:  node.Network,
		TLS:      node.TLS,
		SkipCert: node.SkipCert,
		Path:     node.Path,
		Host:     node.Host,
		ALPN:     node.ALPN,
		Rename:   node.Rename,
	}

	cfg.Proxies = append(cfg.Proxies, proxy)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

func (s *NodeService) UpdateNode(id uint, node *model.Node) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if id < 1 || int(id) > len(cfg.Proxies) {
		return fmt.Errorf("invalid node id: %d", id)
	}

	proxy := &cfg.Proxies[id-1]
	proxy.Name = node.Name
	proxy.Type = node.Type
	proxy.Server = node.Server
	proxy.Port = node.Port
	proxy.Cipher = node.Cipher
	proxy.Password = node.Password
	proxy.UDP = node.UDP
	proxy.Username = node.Username
	proxy.UUID = node.UUID
	proxy.Network = node.Network
	proxy.TLS = node.TLS
	proxy.SkipCert = node.SkipCert
	proxy.Path = node.Path
	proxy.Host = node.Host
	proxy.ALPN = node.ALPN
	proxy.Rename = node.Rename

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

func (s *NodeService) DeleteNode(id uint) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if id < 1 || int(id) > len(cfg.Proxies) {
		return fmt.Errorf("invalid node id: %d", id)
	}

	cfg.Proxies = append(cfg.Proxies[:id-1], cfg.Proxies[id:]...)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

func (s *NodeService) GetNodeByID(id uint) (*model.Node, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	if id < 1 || int(id) > len(cfg.Proxies) {
		return nil, fmt.Errorf("invalid node id: %d", id)
	}

	proxy := cfg.Proxies[id-1]
	node := &model.Node{
		ID:       id,
		Name:     proxy.Name,
		Type:     proxy.Type,
		Server:   proxy.Server,
		Port:     proxy.Port,
		Cipher:   proxy.Cipher,
		Password: proxy.Password,
		UDP:      proxy.UDP,
		Username: proxy.Username,
		UUID:     proxy.UUID,
		Network:  proxy.Network,
		TLS:      proxy.TLS,
		SkipCert: proxy.SkipCert,
		Path:     proxy.Path,
		Host:     proxy.Host,
		ALPN:     proxy.ALPN,
		Rename:   proxy.Rename,
	}

	return node, nil
}
