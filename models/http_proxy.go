package models

import "time"

type HttpProxy struct {
	Id           int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name         string    `xorm:"varchar(255) 'name'" json:"name"`
	Port         string    `xorm:"varchar(50) 'port'" json:"port"`
	Hostname     string    `xorm:"varchar(255) 'hostname'" json:"hostname"`
	HTTP         bool      `xorm:"bool 'http'" json:"http"`
	TLS          bool      `xorm:"bool 'tls'" json:"tls"`
	H2           bool      `xorm:"bool 'h2'" json:"h2"`
	HSTS         bool      `xorm:"bool 'hsts'" json:"hsts"`
	Certificate  string    `xorm:"varchar(255) 'certificate'" json:"certificate"`
	ProxyHeaders string    `xorm:"text 'proxy_headers'" json:"proxy_headers"` // JSON array e.g. ["X-Forwarded-For"]
	Compress     bool      `xorm:"bool 'compress'" json:"compress"`
	Rules        string    `xorm:"text 'rules'" json:"rules"`                 // JSON array e.g. ["规则1"]
	RealIp       string    `xorm:"varchar(255) 'real_ip'" json:"real_ip"`
	TunnelType   string    `xorm:"varchar(50) 'tunnel_type'" json:"tunnel_type"`
	TunnelId     string    `xorm:"varchar(100) 'tunnel_id'" json:"tunnel_id"`
	TunnelToken  string    `xorm:"varchar(255) 'tunnel_token'" json:"tunnel_token"`
	DNSResolver  string    `xorm:"varchar(255) 'dns_resolver'" json:"dns_resolver"`
	LocationJSON string    `xorm:"text 'location_json'" json:"location_json"` // JSON array of Locations
	Remark       string    `xorm:"varchar(255) 'remark'" json:"remark"`
	CreatedAt    time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
}

func (HttpProxy) TableName() string {
	return "http_proxy"
}

func (h *HttpProxy) BeforeInsert() {
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now()
	}
	if h.UpdatedAt.IsZero() {
		h.UpdatedAt = time.Now()
	}
}

func (h *HttpProxy) BeforeUpdate() {
	h.UpdatedAt = time.Now()
}
