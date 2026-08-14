package models

import (
	"time"
)

// TunnelClient represents a configured tunnel client item (Type, ID, Token)
// that can be referenced by proxies (HTTP, SNI, TCP, UDP, etc.) in server.json
type TunnelClient struct {
	Id        int64     `xorm:"pk autoincr" json:"id"`
	Name      string    `xorm:"'name' varchar(128) notnull" json:"name"`         // 配置名称/别名
	Type      string    `xorm:"'type' varchar(32) notnull" json:"type"`         // "tls" 或 "quic" (或 "TLS-TUNNEL" / "QUIC-TUNNEL")
	TunnelId  string    `xorm:"'tunnel_id' varchar(64) notnull" json:"tunnel_id"` // 所属 Tunnel 监听服务端 ID (如 "1", "2")
	Token     string    `xorm:"'token' varchar(255) notnull" json:"token"`       // 客户端鉴权 Token
	Remark    string    `xorm:"'remark' varchar(255)" json:"remark"`
	CreatedAt time.Time `xorm:"created" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated" json:"updated_at"`
}

func (TunnelClient) TableName() string {
	return "tunnel_client"
}
