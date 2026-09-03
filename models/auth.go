package models

import "time"

// Auth 认证配置（策略）
type Auth struct {
	Id            int64     `xorm:"pk autoincr" json:"id"`
	Name          string    `xorm:"varchar(64) notnull unique" json:"name"` // 如 "内网通用认证策略"
	AuthMethodIds string    `xorm:"varchar(255) default '[]'" json:"auth_method_ids"` // 关联的认证方式 ID 数组 (JSON)
	TokenName     string    `xorm:"varchar(64) default 'ANG_TOKEN'" json:"token_name"` // Token 名称
	PortalUrl     string    `xorm:"varchar(255)" json:"portal_url"` // Portal 登录页地址
	TokenExpire   int       `xorm:"int default 86400" json:"token_expire"` // Token 过期时间（秒）
	Remark        string    `xorm:"varchar(255)" json:"remark"`
	CreatedAt     time.Time `xorm:"created" json:"created_at"`
	UpdatedAt     time.Time `xorm:"updated" json:"updated_at"`
}
