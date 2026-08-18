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
	RealIp       string `json:"real_ip"`
	TunnelType   string `json:"tunnel_type"`
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
