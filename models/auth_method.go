package models

import "time"

type AuthMethod struct {
	Id         int64     `xorm:"pk autoincr" json:"id"`
	Name       string    `xorm:"varchar(64) notnull unique" json:"name"`   // 如 "本地账号认证", "校园CAS", "企业Radius"
	Type       string    `xorm:"varchar(32) notnull" json:"type"`          // "local", "cas", "radius"
	Enabled    bool      `xorm:"bool notnull default true" json:"enabled"` // 是否启用
	Priority   int       `xorm:"int default 0" json:"priority"`            // 排序/优先级
	ConfigJSON string    `xorm:"text" json:"config_json"`                  // 各类型特定配置 JSON
	Remark     string    `xorm:"varchar(255)" json:"remark"`
	CreatedAt  time.Time `xorm:"created" json:"created_at"`
	UpdatedAt  time.Time `xorm:"updated" json:"updated_at"`
}
