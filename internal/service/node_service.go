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
			AlterId:  proxy.AlterId,
			Network:  proxy.Network,
			TLS:      proxy.TLS,
			SkipCert: proxy.SkipCert,
			Path:     proxy.Path,
			Host:     proxy.Host,
			ALPN:     proxy.ALPN,
			Address:  proxy.Address,
			Rename:   proxy.Rename,
		}

		// 处理 Reality 相关字段
		if proxy.RealityOpts != nil {
			node.PublicKey = proxy.RealityOpts.PublicKey
			node.ShortID = proxy.RealityOpts.ShortID
			node.ServerName = proxy.RealityOpts.ServerName
			node.Fingerprint = proxy.RealityOpts.Fingerprint
			node.RealityShow = proxy.RealityOpts.Show
			node.RealityDebug = proxy.RealityOpts.Debug
		}

		// 处理其他字段
		node.Up = proxy.Up
		node.Down = proxy.Down
		node.HopInterval = proxy.HopInterval
		node.Flow = proxy.Flow
		node.ServerName = proxy.ServerName
		node.ClientFingerprint = proxy.ClientFingerprint

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
		Name:              node.Name,
		Type:              node.Type,
		Server:            node.Server,
		Port:              node.Port,
		Cipher:            node.Cipher,
		Password:          node.Password,
		UDP:               node.UDP,
		Username:          node.Username,
		Network:           node.Network,
		TLS:               node.TLS,
		SkipCert:          node.SkipCert,
		Path:              node.Path,
		Host:              node.Host,
		ALPN:              node.ALPN,
		Address:           node.Address,
		Rename:            node.Rename,
		Up:                node.Up,
		Down:              node.Down,
		HopInterval:       node.HopInterval,
		Flow:              node.Flow,
		ServerName:        node.ServerName,
		ClientFingerprint: node.ClientFingerprint,
	}

	// 根据节点类型设置特定字段
	if node.Type == "vmess" || node.Type == "vless" {
		proxy.UUID = node.UUID
		proxy.AlterId = node.AlterId
	}

	// For trojan and hysteria2 nodes, set SNI from Host
	if node.Type == "trojan" || node.Type == "hysteria2" {
		proxy.SNI = node.Host
	}

	// For vless nodes, set SNI from Host and handle Reality
	if node.Type == "vless" {
		proxy.SNI = node.Host
		if node.PublicKey != "" || node.ShortID != "" || node.ServerName != "" || node.Fingerprint != "" {
			proxy.RealityOpts = &config.RealityOpts{
				PublicKey:   node.PublicKey,
				ShortID:     node.ShortID,
				ServerName:  node.ServerName,
				Fingerprint: node.Fingerprint,
				Show:        node.RealityShow,
				Debug:       node.RealityDebug,
			}
		}
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
	proxy.Network = node.Network
	proxy.TLS = node.TLS
	proxy.SkipCert = node.SkipCert
	proxy.Path = node.Path
	proxy.Host = node.Host

	// 根据节点类型设置特定字段
	if node.Type == "vmess" || node.Type == "vless" {
		proxy.UUID = node.UUID
		proxy.AlterId = node.AlterId
	} else {
		proxy.UUID = ""
		proxy.AlterId = ""
	}

	// For trojan and hysteria2 nodes, set SNI from Host
	if node.Type == "trojan" || node.Type == "hysteria2" {
		proxy.SNI = node.Host
	}
	// For vless nodes, set SNI from Host and handle Reality
	if node.Type == "vless" {
		proxy.SNI = node.Host
		if node.PublicKey != "" || node.ShortID != "" || node.ServerName != "" || node.Fingerprint != "" {
			proxy.RealityOpts = &config.RealityOpts{
				PublicKey:   node.PublicKey,
				ShortID:     node.ShortID,
				ServerName:  node.ServerName,
				Fingerprint: node.Fingerprint,
				Show:        node.RealityShow,
				Debug:       node.RealityDebug,
			}
		} else {
			proxy.RealityOpts = nil
		}
	}
	proxy.ALPN = node.ALPN
	proxy.Address = node.Address
	proxy.Rename = node.Rename
	proxy.Up = node.Up
	proxy.Down = node.Down
	proxy.HopInterval = node.HopInterval
	proxy.Flow = node.Flow
	proxy.ServerName = node.ServerName
	proxy.ClientFingerprint = node.ClientFingerprint

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
