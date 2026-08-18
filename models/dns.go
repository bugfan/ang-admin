package models

import "time"

type DnsProxy struct {
	Id              int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name            string    `xorm:"varchar(255) 'name'" json:"name"`
	Address         string    `xorm:"varchar(255) 'address'" json:"address"`
	Port            string    `xorm:"varchar(50) 'port'" json:"port"`
	Rules           string    `xorm:"text 'rules'" json:"rules"`                       // JSON string array e.g. ["rule1", "rule2"]
	HostsText       string    `xorm:"text 'hosts_text'" json:"hosts_text"`             // Raw multiline hosts text
	HostsJSON       string    `xorm:"text 'hosts_json'" json:"hosts_json"`             // Parsed JSON {"A":{...},"AAAA":{...}}
	BackendType     string    `xorm:"varchar(50) 'backend_type'" json:"backend_type"`  // "tunnel", "upstream", or "both"
	TunnelType      string    `xorm:"varchar(50) 'tunnel_type'" json:"tunnel_type"`    // "tls" or "quic"
	TunnelId        string    `xorm:"varchar(100) 'tunnel_id'" json:"tunnel_id"`      // Selected Tunnel ID
	TunnelToken     string    `xorm:"varchar(255) 'tunnel_token'" json:"tunnel_token"` // Selected Tunnel Token
	UpstreamMethod  string    `xorm:"varchar(50) 'upstream_method'" json:"upstream_method"` // "round_robin", "weight", "failover", "random"
	UpstreamServers string    `xorm:"text 'upstream_servers'" json:"upstream_servers"` // JSON string array of servers [{"target":"8.8.8.8:53","weight":1}]
	Remark          string    `xorm:"varchar(255) 'remark'" json:"remark"`
	CreatedAt       time.Time `xorm:"created" json:"created_at"`
	UpdatedAt       time.Time `xorm:"updated" json:"updated_at"`
}

func (DnsProxy) TableName() string {
	return "dns_proxy"
}

func (d *DnsProxy) BeforeInsert() {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now()
	}
}

func (d *DnsProxy) BeforeUpdate() {
	d.UpdatedAt = time.Now()
}
