package api

import (
	"log"
	"net/http"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.HttpProxy{}, &httpProxyHandler{}, rest.RouteTypeALL, nil, "http-proxy")
}

type httpProxyHandler struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Port         string `json:"port"`
	Hostname     string `json:"hostname"`
	HTTP         bool   `json:"http"`
	TLS          bool   `json:"tls"`
	H2           bool   `json:"h2"`
	HSTS         bool   `json:"hsts"`
	Certificate  string `json:"certificate"`
	ProxyHeaders string `json:"proxy_headers"`
	Compress     bool   `json:"compress"`
	Rules        string `json:"rules"`
	RealIp       string    `json:"real_ip"`
	RootCA       string    `json:"root_ca"`
	TunnelType   string    `json:"tunnel_type"`
	TunnelId     string `json:"tunnel_id"`
	TunnelToken  string `json:"tunnel_token"`
	DNSResolver  string `json:"dns_resolver"`
	LocationJSON string    `json:"location_json"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (h *httpProxyHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if h.Name == "" {
			h.Name = h.Hostname
		}
		if h.Port == "" {
			h.Port = "80"
		}
	}
	return true
}

func (h *httpProxyHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		log.Printf("[HTTP Proxy Change Event] Method: %s, ID: %d, Hostname: %s, Port: %s\n", method, h.Id, h.Hostname, h.Port)
		service.SyncHTTPToCluster()
	}
}

func (h *httpProxyHandler) List(c *gin.Context) {
	var list []models.HttpProxy
	session := models.GetEngine().NewSession()
	defer session.Close()

	if keyword := c.Query("keyword"); keyword != "" {
		pattern := "%" + keyword + "%"
		session.Where("name LIKE ? OR hostname LIKE ? OR port LIKE ? OR remark LIKE ?", pattern, pattern, pattern, pattern)
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询 HTTP 代理列表失败: " + err.Error()})
		return
	}

	resList := make([]httpProxyHandler, 0, len(list))
	for _, item := range list {
		resList = append(resList, httpProxyHandler{
			Id:           item.Id,
			Name:         item.Name,
			Port:         item.Port,
			Hostname:     item.Hostname,
			HTTP:         item.HTTP,
			TLS:          item.TLS,
			H2:           item.H2,
			HSTS:         item.HSTS,
			Certificate:  item.Certificate,
			ProxyHeaders: item.ProxyHeaders,
			Compress:     item.Compress,
			Rules:        item.Rules,
			RealIp:       item.RealIp,
			RootCA:       item.RootCA,
			TunnelType:   item.TunnelType,
			TunnelId:     item.TunnelId,
			TunnelToken:  item.TunnelToken,
			DNSResolver:  item.DNSResolver,
			LocationJSON: item.LocationJSON,
			Remark:       item.Remark,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}
