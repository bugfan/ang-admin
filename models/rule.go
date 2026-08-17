package models

import "time"

type Rule struct {
	Id        int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name      string    `xorm:"varchar(255) 'name'" json:"name"`
	Items     string    `xorm:"text 'items'" json:"items"` // JSON string array e.g. [{"Matcher":{...},"Action":{...}}]
	Remark    string    `xorm:"varchar(255) 'remark'" json:"remark"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (Rule) TableName() string {
	return "rule"
}
