package models

import "time"

type Rule struct {
	Id        int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name      string    `xorm:"varchar(255) 'name'" json:"name"`
	Matcher   string    `xorm:"text 'matcher'" json:"matcher"` // JSON string e.g. {"Name":"ip_matcher","Config":{"Address":["127.0.0.1"]}}
	Action    string    `xorm:"text 'action'" json:"action"`   // JSON string e.g. {"Name":"reset_conn_action","Config":{"Content":"reset"}}
	Remark    string    `xorm:"varchar(255) 'remark'" json:"remark"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (Rule) TableName() string {
	return "rule"
}
