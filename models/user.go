package models

import (
	"encoding/json"
	"strings"
	"time"
)

// User 核心用户主体（账户体系）
// 本地账号密码、RADIUS 账号、CAS 账号、OAuth OpenID 均作为该主账户绑定的身份标识存在
type User struct {
	Id         int64     `xorm:"pk autoincr" json:"id"`
	Username   string    `xorm:"varchar(64) notnull unique" json:"username"` // 主账户唯一用户名/工号
	Password   string    `xorm:"varchar(255)" json:"password,omitempty"`     // 本地登录密码 (bcrypt)
	FullName   string    `xorm:"varchar(64)" json:"full_name"`
	Email      string    `xorm:"varchar(128)" json:"email"`
	Mobile     string    `xorm:"varchar(32)" json:"mobile"`
	SourceType string    `xorm:"varchar(32) notnull default 'local'" json:"source_type"` // 主账号注册源: local, cas, radius 等
	SourceId   int64     `xorm:"bigint default 0" json:"source_id"`                      // 归属的主认证方式 ID
	GroupIds   string    `xorm:"varchar(255) default '[]'" json:"group_ids"`             // 所属用户组 ID 数组，如 "[1,2]"
	Status     int       `xorm:"int default 1" json:"status"`                            // 1: 启用, 0: 禁用
	ExpireAt   string    `xorm:"varchar(32)" json:"expire_at"`                           // 格式：2026-12-31 23:59:59，留空永不过期

	// 第三方身份与协议绑定字段 (预留扩展账户体系)
	RadiusUsername string `xorm:"varchar(64) 'radius_username'" json:"radius_username"` // 绑定的 RADIUS 账号 ID
	CasUsername    string `xorm:"varchar(64) 'cas_username'" json:"cas_username"`       // 绑定的 CAS 学号/工号/账号
	OAuthOpenID    string `xorm:"varchar(128) 'oauth_openid'" json:"oauth_openid"`     // 绑定的 OAuth/OIDC OpenID
	Identities     string `xorm:"varchar(1024) default '{}' 'identities'" json:"identities"` // 通用绑定映射字典 JSON，如 {"radius":"rad01","cas":"c01"}

	Remark    string    `xorm:"varchar(255)" json:"remark"`
	CreatedAt time.Time `xorm:"created" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated" json:"updated_at"`
}

// GetRadiusAccount 获取该用户绑定的 RADIUS 账号
func (u *User) GetRadiusAccount() string {
	if strings.TrimSpace(u.RadiusUsername) != "" {
		return strings.TrimSpace(u.RadiusUsername)
	}
	if u.Identities != "" && u.Identities != "{}" {
		var m map[string]string
		if err := json.Unmarshal([]byte(u.Identities), &m); err == nil {
			if v, ok := m["radius"]; ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

// GetCasAccount 获取该用户绑定的 CAS 账号
func (u *User) GetCasAccount() string {
	if strings.TrimSpace(u.CasUsername) != "" {
		return strings.TrimSpace(u.CasUsername)
	}
	if u.Identities != "" && u.Identities != "{}" {
		var m map[string]string
		if err := json.Unmarshal([]byte(u.Identities), &m); err == nil {
			if v, ok := m["cas"]; ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

// GetOAuthAccount 获取该用户绑定的 OAuth OpenID
func (u *User) GetOAuthAccount() string {
	if strings.TrimSpace(u.OAuthOpenID) != "" {
		return strings.TrimSpace(u.OAuthOpenID)
	}
	if u.Identities != "" && u.Identities != "{}" {
		var m map[string]string
		if err := json.Unmarshal([]byte(u.Identities), &m); err == nil {
			if v, ok := m["oauth"]; ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}
