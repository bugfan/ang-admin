package entity

// TunnelConfig represents a single item in TLS-TUNNEL or QUIC-TUNNEL
type TunnelConfig struct {
	Port        string `json:"Port"`
	Certificate string `json:"Certificate,omitempty"`
	Auth        bool   `json:"Auth"`
}

// CertData represents the certificate data (Key & Cert)
type CertData struct {
	Key  string `json:"Key"`
	Cert string `json:"Cert"`
}

// CertConfig represents a single item in Certificate
type CertConfig struct {
	Type string   `json:"Type"`
	Data CertData `json:"Data"`
}

// MatcherConfig represents rule matcher configuration
type MatcherConfig struct {
	Name   string      `json:"Name"`
	Config interface{} `json:"Config"`
}

// ActionConfig represents rule action configuration
type ActionConfig struct {
	Name   string      `json:"Name"`
	Config interface{} `json:"Config"`
}

// RuleConfig represents a rule with Matcher and Action
type RuleConfig struct {
	Matcher MatcherConfig `json:"Matcher"`
	Action  ActionConfig  `json:"Action"`
}

// DNSHosts represents DNS host mapping (A and AAAA records)
type DNSHosts struct {
	A    map[string]string `json:"A,omitempty"`
	AAAA map[string]string `json:"AAAA,omitempty"`
}

// BackendTunnel represents tunnel info in backend config
type BackendTunnel struct {
	Type  string `json:"Type"`
	ID    string `json:"ID"`
	Token string `json:"Token,omitempty"`
}

// UpstreamServer represents target server and weight
type UpstreamServer struct {
	Target string `json:"Target"`
	Weight int    `json:"Weight"`
}

// UpstreamData represents data container for servers
type UpstreamData struct {
	Servers []UpstreamServer `json:"Servers"`
}

// UpstreamConfig represents upstream proxy method and server list
type UpstreamConfig struct {
	Method string       `json:"Method"`
	Data   UpstreamData `json:"Data"`
}

// DNSBackend represents DNS backend config
type DNSBackend struct {
	Tunnel   *BackendTunnel  `json:"Tunnel,omitempty"`
	Upstream *UpstreamConfig `json:"Upstream,omitempty"`
}

// DNSConfig represents a single item in DNS section
type DNSConfig struct {
	Address string      `json:"Address"`
	Port    string      `json:"Port"`
	Rule    []RuleConfig `json:"Rule,omitempty"`
	Hosts   *DNSHosts   `json:"Hosts,omitempty"`
	Backend *DNSBackend `json:"Backend,omitempty"`
}

// UDPBackend represents UDP backend config
type UDPBackend struct {
	Tunnel   *BackendTunnel  `json:"Tunnel,omitempty"`
	Upstream *UpstreamConfig `json:"Upstream,omitempty"`
}

// UDPConfig represents a single item in UDP section
type UDPConfig struct {
	Address string      `json:"Address"`
	Port    string      `json:"Port"`
	Timeout string      `json:"Timeout,omitempty"`
	Rule    []RuleConfig `json:"Rule,omitempty"`
	Backend *UDPBackend `json:"Backend,omitempty"`
}

// SNIBackend represents SNI backend config
type SNIBackend struct {
	Tunnel      *BackendTunnel `json:"Tunnel,omitempty"`
	DNSResolver []string       `json:"DNSResolver,omitempty"`
}

// SNIConfig represents a single item in SNI section
type SNIConfig struct {
	SNI      string       `json:"SNI"`
	ExtraSNI []string     `json:"ExtraSNI,omitempty"` // extra SNI patterns (exact or wildcard, e.g. *.example.com)
	Port     string       `json:"Port"`
	Rule     []RuleConfig `json:"Rule,omitempty"`
	Backend  *SNIBackend  `json:"Backend,omitempty"`
}

// TCPBackend represents TCP backend config
type TCPBackend struct {
	Tunnel   *BackendTunnel  `json:"Tunnel,omitempty"`
	Upstream *UpstreamConfig `json:"Upstream,omitempty"`
}

// TCPConfig represents a single item in TCP section
type TCPConfig struct {
	Address string      `json:"Address"`
	Port    string      `json:"Port"`
	Rule    []RuleConfig `json:"Rule,omitempty"`
	Backend *TCPBackend `json:"Backend,omitempty"`
}

// DNSResolverConfig represents a single item in DNSResolver section
type DNSResolverConfig struct {
	Servers []string `json:"Servers"`
	Timeout string   `json:"Timeout,omitempty"`
}

// HTTPFront represents front HTTP listener config
type HTTPFront struct {
	Port         string   `json:"Port"`
	Hostname     string   `json:"Hostname"`
	HTTP         bool     `json:"HTTP"`
	TLS          bool     `json:"TLS"`
	H2           bool     `json:"H2,omitempty"`
	HSTS         bool     `json:"HSTS"`
	Certificate  string   `json:"Certificate,omitempty"`
	ProxyHeaders []string `json:"ProxyHeaders"`
	Protected    bool     `json:"Protected,omitempty"`
}

// HTTPFeature represents HTTP features (like Compress)
type HTTPFeature struct {
	Compress bool `json:"Compress"`
}

// LocationUpstream represents upstream config for HTTP location
type LocationUpstream struct {
	Type string      `json:"Type"`
	Data interface{} `json:"Data"`
}

// HTTPLocation represents a location route in HTTP backend
type HTTPLocation struct {
	Path     string           `json:"Path"`
	Upstream LocationUpstream `json:"Upstream"`
}

// HTTPBackend represents HTTP backend config
type HTTPBackend struct {
	RealIp       string         `json:"RealIp"`
	Tunnel       *BackendTunnel `json:"Tunnel,omitempty"`
	DNSResolver  []string       `json:"DNSResolver,omitempty"`
	Location     []HTTPLocation `json:"Location,omitempty"`
}

// HTTPConfig represents a single item in HTTP section
type HTTPConfig struct {
	Front   HTTPFront    `json:"Front"`
	Feature HTTPFeature  `json:"Feature"`
	Rule    []RuleConfig `json:"Rule,omitempty"`
	Backend HTTPBackend  `json:"Backend"`
}

// TunnelFileConfig represents the tunnel.json configuration structure
type TunnelFileConfig struct {
	TLSTunnel    map[string]TunnelConfig `json:"TLS,omitempty"`
	QUICTunnel   map[string]TunnelConfig `json:"QUIC,omitempty"`
	TunnelClient map[string]string       `json:"CLIENT,omitempty"`
}

// CertificateFileConfig represents the certificate.json configuration structure
type CertificateFileConfig struct {
	Certificate map[string]CertConfig `json:"Certificate,omitempty"`
}

// ServerConfig represents the entire server.json configuration structure
type ServerConfig struct {
	DNS         map[string]DNSConfig         `json:"DNS,omitempty"`
	UDP         map[string]UDPConfig         `json:"UDP,omitempty"`
	SNI         map[string]SNIConfig         `json:"SNI,omitempty"`
	TCP         map[string]TCPConfig         `json:"TCP,omitempty"`
	DNSResolver map[string]DNSResolverConfig `json:"DNSResolver,omitempty"`
	HTTP        map[string]HTTPConfig        `json:"HTTP,omitempty"`
}
