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
	Rename      string `json:"rename,omitempty"`
	Source      string `json:"source,omitempty"`
	ExtraConfig string `json:"extraConfig,omitempty"`
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
