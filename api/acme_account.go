package api

import (
	"net/http"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.AcmeAccount{}, &acmeAccountHandler{}, rest.RouteTypeALL, nil, "acme-account")
}

type acmeAccountHandler struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	Provider      string    `json:"provider"`
	DnsEnv        string    `json:"dns_env"`
	Email         string    `json:"email"`
	DirectoryUrl  string    `json:"directory_url"`
	KeyType       string    `json:"key_type"`
	ChallengeType string    `json:"challenge_type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (h *acmeAccountHandler) List(c *gin.Context) {
	name := c.Query("name")
	provider := c.Query("provider")

	var providers []models.AcmeAccount
	session := models.GetEngine().Desc("id")

	if name != "" {
		session = session.Where("name LIKE ?", "%"+name+"%")
	}
	if provider != "" {
		session = session.Where("provider = ?", provider)
	}

	err := session.Find(&providers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, providers)
}

func (h *acmeAccountHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	return true
}

func (h *acmeAccountHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
}
