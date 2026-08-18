package models

import "time"

type Rule struct {
	Id        int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name      string    `xorm:"varchar(255) 'name'" json:"name"`
	Items     string    `xorm:"text 'items'" json:"items"` // JSON string array e.g. [{"Matcher":{...},"Action":{...}}]
	Remark    string    `xorm:"varchar(255) 'remark'" json:"remark"`
	CreatedAt time.Time `xorm:"created" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated" json:"updated_at"`
}

func (Rule) TableName() string {
	return "rule"
}

func (r *Rule) BeforeInsert() {
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now()
	}
}

func (r *Rule) BeforeUpdate() {
	r.UpdatedAt = time.Now()
}
