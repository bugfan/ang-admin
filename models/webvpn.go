package models

import "time"

type WebvpnSite struct {
	Id              int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name            string    `xorm:"varchar(255) notnull 'name'" json:"name"`                                 // 资源应用名称，如 "中国知网"
	HttpProxyId     int64     `xorm:"index 'http_proxy_id'" json:"http_proxy_id"`                              // 关联的泛域名 HTTP 站点 ID
	TargetURL       string    `xorm:"varchar(512) notnull 'target_url'" json:"target_url"`                     // 目标真实地址，如 https://www.cnki.net
	Prefix          string    `xorm:"varchar(128) notnull 'prefix'" json:"prefix"`                             // 子域名前缀，如 s-cnki 或 s-www-cnki-net-443
	Hosts           string    `xorm:"text 'hosts'" json:"hosts"`                                               // 关联地址/代理域名列表 (每行一个域名)
	Replace         string    `xorm:"text 'replace'" json:"replace"`                                           // 扩展内容替换 JSON map, 如 {"知网": "XX"}
	AllowedGroupIds string    `xorm:"varchar(255) default '[]' 'allowed_group_ids'" json:"allowed_group_ids"` // 允许访问的用户组 ID 列表 JSON 数组
	IsProtected     int       `xorm:"tinyint default 1 'is_protected'" json:"is_protected"`                    // 1: 保护模式(需登录), 0: 公开模式(免登录)
	Status          int       `xorm:"tinyint default 1 'status'" json:"status"`                                // 1: 启用, 0: 禁用
	Remark          string    `xorm:"varchar(255) 'remark'" json:"remark"`                                     // 备注
	CreatedAt       time.Time `xorm:"created" json:"created_at"`
	UpdatedAt       time.Time `xorm:"updated" json:"updated_at"`
}

func (WebvpnSite) TableName() string {
	return "webvpn_site"
}

func (w *WebvpnSite) BeforeInsert() {
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now()
	}
	if w.UpdatedAt.IsZero() {
		w.UpdatedAt = time.Now()
	}
}

func (w *WebvpnSite) BeforeUpdate() {
	w.UpdatedAt = time.Now()
}
