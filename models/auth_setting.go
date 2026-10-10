package models

import (
	"strings"
	"time"
)

// AuthSetting 认证模块的全局高级设置
type AuthSetting struct {
	Id              int64     `xorm:"pk autoincr" json:"id"`
	TokenName       string    `xorm:"varchar(255) notnull" json:"token_name"`
	TokenExpire     int       `xorm:"notnull" json:"token_expire"` // 单位秒
	CookieDomain    string    `xorm:"varchar(255) 'cookie_domain'" json:"cookie_domain"` // 认证Cookie作用域域名，如 .example.com
	EnableTicket    bool      `xorm:"bool 'enable_ticket'" json:"enable_ticket"` // 是否开启 SSO 重定向携带 URL 票据参数（跨顶级域名时启用）
	TicketParamName string    `xorm:"varchar(255) 'ticket_param_name'" json:"ticket_param_name"` // SSO 重定向携带的 URL 凭证参数名，留空则默认与 TokenName 相同
	CreatedAt       time.Time `xorm:"created" json:"created_at"`
	UpdatedAt       time.Time `xorm:"updated" json:"updated_at"`
}

// GetTicketParamName 获取跨域单点登录 URL 重定向时携带的参数名称
func (s *AuthSetting) GetTicketParamName() string {
	if !s.EnableTicket {
		return ""
	}
	if strings.TrimSpace(s.TicketParamName) != "" {
		return strings.TrimSpace(s.TicketParamName)
	}
	if strings.TrimSpace(s.TokenName) != "" {
		return strings.TrimSpace(s.TokenName)
	}
	return "_angt"
}
