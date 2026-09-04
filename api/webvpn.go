package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.WebvpnSite{}, &webvpnSiteHandler{}, rest.RouteTypeALL, nil, "webvpn-site")
}

type webvpnSiteHandler struct {
	Id              int64     `json:"id"`
	Name            string    `json:"name"`
	HttpProxyId     int64     `json:"http_proxy_id"`
	TargetURL       string    `json:"target_url"`
	Prefix          string    `json:"prefix"`
	Hosts           string    `json:"hosts"`
	Replace         string    `json:"replace"`
	AllowedGroupIds string    `json:"allowed_group_ids"`
	IsProtected     int       `json:"is_protected"`
	Status          int       `json:"status"`
	Remark          string    `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (h *webvpnSiteHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "应用名称不能为空",
			})
			return false
		}
		if h.HttpProxyId <= 0 {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "必须选择关联的泛域名 HTTP 站点",
			})
			return false
		}
		h.TargetURL = strings.TrimSpace(h.TargetURL)
		if h.TargetURL == "" || (!strings.HasPrefix(h.TargetURL, "http://") && !strings.HasPrefix(h.TargetURL, "https://")) {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "目标地址必须是以 http:// 或 https:// 开头的合法 URL",
			})
			return false
		}

		// Standard WebVPN Prefix: s-<dashed-host>-<port> (or h- for http)
		u, err := url.Parse(h.TargetURL)
		if err == nil {
			targetHost := u.Hostname()
			targetPort := u.Port()
			schemePrefix := "s-"
			if u.Scheme == "http" {
				schemePrefix = "c-"
				if targetPort == "" {
					targetPort = "80"
				}
			} else {
				if targetPort == "" {
					targetPort = "443"
				}
			}
			dashed := strings.ReplaceAll(strings.ReplaceAll(targetHost, "-", "--"), ".", "-")
			h.Prefix = fmt.Sprintf("%s%s-%s", schemePrefix, dashed, targetPort)
		}

		if h.AllowedGroupIds == "" {
			h.AllowedGroupIds = "[]"
		}
		if h.Replace == "" {
			h.Replace = "{}"
		}
	}
	return true
}

func (h *webvpnSiteHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		service.SyncHTTPToCluster()
	}
}

func (h *webvpnSiteHandler) List(c *gin.Context) {
	var list []models.WebvpnSite
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if proxyIdStr := strings.TrimSpace(c.Query("http_proxy_id")); proxyIdStr != "" {
		if pid, err := strconv.ParseInt(proxyIdStr, 10, 64); err == nil && pid > 0 {
			session.Where("http_proxy_id = ?", pid)
		}
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	// Fetch all HttpProxy for mapping site hostname
	var proxies []models.HttpProxy
	_ = models.GetEngine().Find(&proxies)
	proxyMap := make(map[int64]models.HttpProxy)
	for _, p := range proxies {
		proxyMap[p.Id] = p
	}

	type WebvpnItemVO struct {
		models.WebvpnSite
		HttpProxyName     string `json:"http_proxy_name"`
		HttpProxyHostname string `json:"http_proxy_hostname"`
		FullAccessURL     string `json:"full_access_url"`
	}

	resList := make([]WebvpnItemVO, len(list))
	for i, item := range list {
		vo := WebvpnItemVO{WebvpnSite: item}
		if p, ok := proxyMap[item.HttpProxyId]; ok {
			vo.HttpProxyName = p.Name
			vo.HttpProxyHostname = p.Hostname

			// Compute full access url
			rootDomain := strings.TrimPrefix(p.Hostname, "*.")
			scheme := "http://"
			if p.TLS || p.H2 {
				scheme = "https://"
			}
			portSuffix := ""
			if p.Port != "80" && p.Port != "443" && p.Port != "" {
				portSuffix = ":" + p.Port
			}
			vo.FullAccessURL = fmt.Sprintf("%s%s.%s%s", scheme, item.Prefix, rootDomain, portSuffix)
		}
		resList[i] = vo
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resList,
	})
}
