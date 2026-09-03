package models

import "time"

type UserGroup struct {
	Id          int64     `xorm:"pk autoincr" json:"id"`
	Name        string    `xorm:"varchar(64) notnull unique" json:"name"`
	Description string    `xorm:"varchar(255)" json:"description"`
	IsDefault   bool      `xorm:"bool default false" json:"is_default"`
	CreatedAt   time.Time `xorm:"created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated" json:"updated_at"`
}
