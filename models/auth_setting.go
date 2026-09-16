package models

import "time"

// AuthSetting 认证模块的全局高级设置
type AuthSetting struct {
	Id          int64     `xorm:"pk autoincr" json:"id"`
	TokenName   string    `xorm:"varchar(255) notnull" json:"token_name"`
	TokenExpire int       `xorm:"notnull" json:"token_expire"` // 单位秒
	CreatedAt   time.Time `xorm:"created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated" json:"updated_at"`
}
