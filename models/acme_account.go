package models

import (
	"time"
)

type AcmeAccount struct {
	Id           int64     `xorm:"pk autoincr" json:"id"`
	Name         string    `xorm:"'name' varchar(128) notnull" json:"name"`
	Provider      string    `xorm:"'provider' varchar(64) notnull" json:"provider"`
	DnsEnv        string    `xorm:"'dns_env' text" json:"dns_env"`
	Email         string    `xorm:"'email' varchar(128)" json:"email"`
	DirectoryUrl  string    `xorm:"'directory_url' varchar(128)" json:"directory_url"`
	KeyType       string    `xorm:"'key_type' varchar(32)" json:"key_type"`
	ChallengeType string    `xorm:"'challenge_type' varchar(32)" json:"challenge_type"`
	CreatedAt     time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
}

func (AcmeAccount) TableName() string {
	return "acme_account"
}
