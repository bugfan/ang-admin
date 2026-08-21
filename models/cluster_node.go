package models

import "time"

type ClusterNode struct {
	Id        int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name      string    `xorm:"varchar(255) notnull 'name'" json:"name"`
	Addr      string    `xorm:"varchar(255) notnull 'addr'" json:"addr"`
	Secret    string    `xorm:"varchar(255) 'secret'" json:"secret"`
	Status    int       `xorm:"int 'status'" json:"status"` // 1: online, 0: offline
	LastPing  time.Time `xorm:"datetime 'last_ping'" json:"last_ping"`
	Remark    string    `xorm:"varchar(255) 'remark'" json:"remark"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}
