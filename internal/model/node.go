package model

type Node struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Server      string `json:"server"`
	Port        int    `json:"port"`
	Cipher      string `json:"cipher,omitempty"`
	Password    string `json:"password,omitempty"`
	UDP         bool   `json:"udp,omitempty"`
	Username    string `json:"username,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	AlterId     string `json:"alterId,omitempty"`
	Network     string `json:"network,omitempty"`
	TLS         bool   `json:"tls,omitempty"`
	SkipCert    bool   `json:"skipCert,omitempty"`
	Path        string `json:"path,omitempty"`
	Host        string `json:"host,omitempty"`
	ALPN        string `json:"alpn,omitempty"`
	Address     string `json:"address,omitempty"`
	Rename      string `json:"rename,omitempty"`
	Source      string `json:"source,omitempty"`
	ExtraConfig string `json:"extraConfig,omitempty"`
	Up          int    `json:"up,omitempty"`          // Hysteria2 上行带宽
	Down        int    `json:"down,omitempty"`        // Hysteria2 下行带宽
	HopInterval int    `json:"hopInterval,omitempty"` // Hysteria2 跳跃间隔
	Flow        string `json:"flow,omitempty"`        // VLESS 流控
	ServerName  string `json:"serverName,omitempty"`   // VLESS Reality 服务器名称
	PublicKey   string `json:"publicKey,omitempty"`    // VLESS Reality 公钥
	ShortID     string `json:"shortId,omitempty"`     // VLESS Reality 短ID
	ClientFingerprint string `json:"clientFingerprint,omitempty"` // VLESS 客户端指纹
	Fingerprint string `json:"fingerprint,omitempty"` // Reality 指纹
	RealityShow bool   `json:"realityShow,omitempty"`   // Reality 是否显示
	RealityDebug bool  `json:"realityDebug,omitempty"`  // Reality 是否调试
}

type Rule struct {
	ID         uint   `json:"id"`
	Type       string `json:"type"`
	Payload    string `json:"payload"`
	Target     string `json:"target"`
	TargetType string `json:"targetType,omitempty"`
	Priority   int    `json:"priority"`
	NoResolve  bool   `json:"noResolve,omitempty"`
}
