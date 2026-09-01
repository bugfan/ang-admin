package models

import "time"

type SniProxy struct {
	Id           int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name         string    `xorm:"varchar(255) 'name'" json:"name"`
	SNI          string    `xorm:"varchar(255) 'sni'" json:"sni"`
	Port         string    `xorm:"varchar(50) 'port'" json:"port"`
	Rules        string    `xorm:"text 'rules'" json:"rules"`
	BackendType  string    `xorm:"varchar(50) 'backend_type'" json:"backend_type"` // "tunnel" or "dns"
	TunnelType   string    `xorm:"varchar(50) 'tunnel_type'" json:"tunnel_type"`
	TunnelId     string    `xorm:"varchar(100) 'tunnel_id'" json:"tunnel_id"`
	TunnelToken  string    `xorm:"varchar(255) 'tunnel_token'" json:"tunnel_token"`
	DNSResolver  string    `xorm:"varchar(255) 'dns_resolver'" json:"dns_resolver"`
	Remark       string    `xorm:"varchar(255) 'remark'" json:"remark"`
	CreatedAt    time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
}

func (SniProxy) TableName() string {
	return "sni_proxy"
}

func (s *SniProxy) BeforeInsert() {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = time.Now()
	}
}

func (s *SniProxy) BeforeUpdate() {
	s.UpdatedAt = time.Now()
}
