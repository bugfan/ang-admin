package models

import (
	"time"
)

// User represents the user model in the database
type User struct {
	Id          int64     `xorm:"pk autoincr" json:"id"`
	Username    string    `xorm:"unique notnull" json:"username"`
	Password    string    `xorm:"notnull" json:"-"`
	Avatar      string    `xorm:"" json:"avatar"`
	Description string    `xorm:"" json:"description"`
	Roles       []string  `xorm:"-" json:"roles"`
	Permissions []string  `xorm:"-" json:"permissions"`
	CreatedAt   time.Time `xorm:"created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated" json:"updated_at"`
}
