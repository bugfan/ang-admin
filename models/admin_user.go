package models

import (
	"time"
)

// AdminUser represents the admin user model in the database
type AdminUser struct {
	Id           int64     `xorm:"pk autoincr" json:"id"`
	Username     string    `xorm:"unique notnull" json:"username"`
	Password     string    `xorm:"notnull" json:"-"`
	Avatar       string    `xorm:"" json:"avatar"`
	Description  string    `xorm:"" json:"description"`
	IsSuperAdmin bool      `xorm:"default 0" json:"is_super_admin"`
	Roles        []string  `xorm:"-" json:"roles"`
	Permissions  []string  `xorm:"-" json:"permissions"`
	CreatedAt    time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
}

func (AdminUser) TableName() string {
	return "admin_user"
}
