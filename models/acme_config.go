package models

import (
	"time"
)

type AcmeConfig struct {
	Id            int64     `xorm:"pk autoincr" json:"id"`
	Name          string    `xorm:"'name' varchar(128) notnull" json:"name"`
	Email         string    `xorm:"'email' varchar(128) notnull" json:"email"`
	DirectoryUrl  string    `xorm:"'directory_url' varchar(255) notnull" json:"directory_url"`
	KeyType       string    `xorm:"'key_type' varchar(32)" json:"key_type"`
	ChallengeType string    `xorm:"'challenge_type' varchar(32) notnull" json:"challenge_type"`
	DnsProvider   string    `xorm:"'dns_provider' varchar(64)" json:"dns_provider"`
	DnsEnv        string    `xorm:"'dns_env' text" json:"dns_env"`
	Domains       string    `xorm:"'domains' text notnull" json:"domains"`
	CertId        string    `xorm:"'cert_id' varchar(64)" json:"cert_id"`
	DisableCname  bool      `xorm:"'disable_cname' bool default 1" json:"disable_cname"`
	AutoRenew     bool      `xorm:"'auto_renew' bool" json:"auto_renew"`
	RenewDays     int       `xorm:"'renew_days' int default 30" json:"renew_days"`
	LastIssuedAt  time.Time `xorm:"'last_issued_at' datetime" json:"last_issued_at"`
	LastStatus    string    `xorm:"'last_status' varchar(32)" json:"last_status"`
	LastError     string    `xorm:"'last_error' text" json:"last_error"`
	CreatedAt     time.Time `xorm:"created" json:"created_at"`
	UpdatedAt     time.Time `xorm:"updated" json:"updated_at"`
}

func (AcmeConfig) TableName() string {
	return "acme_config"
}
