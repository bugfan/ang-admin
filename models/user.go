package models

import "time"

type User struct {
	Id         int64     `xorm:"pk autoincr" json:"id"`
	Username   string    `xorm:"varchar(64) notnull unique" json:"username"`
	Password   string    `xorm:"varchar(255)" json:"password,omitempty"`
	FullName   string    `xorm:"varchar(64)" json:"full_name"`
	Email      string    `xorm:"varchar(128)" json:"email"`
	Mobile     string    `xorm:"varchar(32)" json:"mobile"`
	SourceType string    `xorm:"varchar(32) notnull default 'local'" json:"source_type"` // local, cas, radius
	SourceId   int64     `xorm:"bigint default 0" json:"source_id"`                      // 归属的认证方式 ID
	GroupIds   string    `xorm:"varchar(255) default '[]'" json:"group_ids"`             // 所属用户组 ID 数组，如 "[1,2]"
	Status     int       `xorm:"int default 1" json:"status"`                            // 1: 启用, 0: 禁用
	ExpireAt   string    `xorm:"varchar(32)" json:"expire_at"`                           // 格式：2026-12-31 23:59:59，留空永不过期
	Remark     string    `xorm:"varchar(255)" json:"remark"`
	CreatedAt  time.Time `xorm:"created" json:"created_at"`
	UpdatedAt  time.Time `xorm:"updated" json:"updated_at"`
}
