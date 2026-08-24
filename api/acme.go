package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.AcmeConfig{}, &acmeConfigHandler{}, rest.RouteTypeALL, nil, "acme-config")
}

type acmeConfigHandler struct {
	Id            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	DirectoryUrl  string    `json:"directory_url"`
	KeyType       string    `json:"key_type"`
	ChallengeType string    `json:"challenge_type"`
	DnsProvider   string    `json:"dns_provider"`
	DnsEnv        string    `json:"dns_env"`
	Domains       string    `json:"domains"`
	CertId        string    `json:"cert_id"`
	DisableCname  bool      `json:"disable_cname"`
	AutoRenew     bool      `json:"auto_renew"`
	RenewDays     int       `json:"renew_days"`
	LastIssuedAt  time.Time `json:"last_issued_at"`
	LastStatus    string    `json:"last_status"`
	LastError     string    `json:"last_error"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (h *acmeConfigHandler) List(c *gin.Context) {
	name := c.Query("name")
	email := c.Query("email")
	domains := c.Query("domains")
	dnsProvider := c.Query("dns_provider")

	var configs []models.AcmeConfig
	session := models.GetEngine().Desc("id")

	if name != "" {
		session = session.Where("name LIKE ?", "%"+name+"%")
	}
	if email != "" {
		session = session.Where("email LIKE ?", "%"+email+"%")
	}
	if domains != "" {
		session = session.Where("domains LIKE ?", "%"+domains+"%")
	}
	if dnsProvider != "" {
		session = session.Where("dns_provider = ?", dnsProvider)
	}

	err := session.Find(&configs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

func (h *acmeConfigHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	return true
}

func (h *acmeConfigHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
}

// IssueAcmeCertHandler handles manual ad-hoc issuance (if needed)
func IssueAcmeCertHandler(c *gin.Context) {
	var req service.IssueCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "无效的请求参数: " + err.Error()})
		return
	}

	resp, err := service.IssueAcmeCertificate(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "证书签发成功！",
		"data":    resp,
	})
}

// IssueAcmeCertByConfigHandler handles issuance using saved config
func IssueAcmeCertByConfigHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "无效的配置 ID"})
		return
	}

	resp, err := service.IssueCertificateByConfigId(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "证书签发成功！",
		"data":    resp,
	})
}
