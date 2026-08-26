package models

import (
	"time"
)

// Tunnel represents the TLS-TUNNEL and QUIC-TUNNEL configuration model
type Tunnel struct {
	Id          int64     `xorm:"pk autoincr" json:"id"`
	Name        string    `xorm:"varchar(255)" json:"name"`
	Type        string    `xorm:"varchar(32) notnull" json:"type"`
	Port        string    `xorm:"varchar(32) notnull" json:"port"`
	SNI         string    `xorm:"'sni' varchar(255)" json:"sni,omitempty"`
	Certificate string    `xorm:"varchar(255)" json:"certificate,omitempty"`
	Remark      string    `xorm:"varchar(255)" json:"remark"`
	CreatedAt   time.Time `xorm:"created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated" json:"updated_at"`
}

func (Tunnel) TableName() string {
	return "tunnel"
}
