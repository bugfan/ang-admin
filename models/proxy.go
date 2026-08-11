package models

import (
	"time"

	"github.com/gin-gonic/gin"
)

type AppProxy struct {
	ID          int64     `xorm:"pk autoincr" json:"id"`
	Name        string    `xorm:"varchar(255)" json:"name"`
	Type        string    `xorm:"varchar(32)" json:"type"` // HTTP, WebVPN, TCP, UDP
	Hostname    string    `xorm:"varchar(255)" json:"hostname"`
	Port        int       `xorm:"int" json:"port"`
	TLS         bool      `xorm:"bool" json:"tls"`
	Certificate string    `xorm:"varchar(255)" json:"certificate"`
	ConfigJSON  string    `xorm:"text" json:"config_json"` // 附加的特性配置、规则等
	Status      int       `xorm:"int" json:"status"`
	Created     time.Time `xorm:"created" json:"created"`
	Updated     time.Time `xorm:"updated" json:"updated"`
}

type AppProxyContent struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
	TLS         bool   `json:"tls"`
	Certificate string `json:"certificate"`
	ConfigJSON  string `json:"config_json"`
	Status      int    `json:"status"`
}

func (c *AppProxyContent) Check(ctx *gin.Context, r interface{}, t int) bool {
	if c.Name == "" {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "name is required"})
		return false
	}
	return true
}
