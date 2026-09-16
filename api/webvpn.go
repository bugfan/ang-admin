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
	rest.Register(&models.WebvpnDomain{}, &webvpnDomainHandler{}, rest.RouteTypeALL, nil, "webvpn-domain")
	rest.Register(&models.WebvpnSite{}, &webvpnSiteHandler{}, rest.RouteTypeALL, nil, "webvpn-site")
}

// -------------------------------------------------------------
// 1. WebVPN 服务 (WebvpnDomain) Handler
// -------------------------------------------------------------

type webvpnDomainHandler struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Hostname    string    `json:"hostname"`
	Port        string    `json:"port"`
	TLS         bool      `json:"tls"`
	H2          bool      `json:"h2"`
	Certificate string    `json:"certificate"`
	AuthId      int64     `json:"auth_id"`
	Fallback    string    `json:"fallback"`
	Status      int       `json:"status"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *webvpnDomainHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" && method == http.MethodPost {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "基础域名称不能为空",
			})
			return false
		}

		h.Hostname = strings.TrimSpace(h.Hostname)
		if h.Hostname == "" && method == http.MethodPost {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "泛域名不能为空",
			})
			return false
		}
		// 规范化通配符域名格式，如用户输入 example.com 或 webvpn.example.com，确保以 *. 开头
		if h.Hostname != "" && !strings.HasPrefix(h.Hostname, "*.") {
			if strings.HasPrefix(h.Hostname, "*") {
				h.Hostname = "*." + strings.TrimPrefix(h.Hostname, "*")
			} else {
				h.Hostname = "*." + h.Hostname
			}
		}

		if h.Port == "" && method == http.MethodPost {
			h.Port = "443"
		}

		if h.Fallback == "" && method == http.MethodPost {
			h.Fallback = "404"
		}
	}
	return true
}

func (h *webvpnDomainHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		service.SyncHTTPToCluster()
	}
}

func (h *webvpnDomainHandler) List(c *gin.Context) {
	var list []models.WebvpnDomain
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if hostname := strings.TrimSpace(c.Query("hostname")); hostname != "" {
		session.Where("hostname LIKE ?", "%"+hostname+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	// 统计关联的站点数量
	type WebvpnDomainVO struct {
		models.WebvpnDomain
		RootDomain string `json:"root_domain"`
		SiteCount  int64  `json:"site_count"`
	}

	resList := make([]WebvpnDomainVO, len(list))
	for i, dom := range list {
		count, _ := models.GetEngine().Where("domain_id = ?", dom.Id).Count(new(models.WebvpnSite))
		resList[i] = WebvpnDomainVO{
			WebvpnDomain: dom,
			RootDomain:   strings.TrimPrefix(dom.Hostname, "*."),
			SiteCount:    count,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resList,
	})
}

// -------------------------------------------------------------
// 2. WebVPN 站点 (WebvpnSite) Handler
// -------------------------------------------------------------

type webvpnSiteHandler struct {
	Id              int64     `json:"id"`
	Name            string    `json:"name"`
	DomainId        int64     `json:"domain_id"`
	TargetURL       string    `json:"target_url"`
	Prefix          string    `json:"prefix"`
	Hosts           string    `json:"hosts"`
	Replace         string    `json:"replace"`
	AllowedGroupIds string    `json:"allowed_group_ids"`
	IsProtected     int       `json:"is_protected"`
	TunnelId        int64     `json:"tunnel_id"`
	TunnelToken     string    `json:"tunnel_token"`
	TunnelType      string    `json:"tunnel_type"`
	Status          int       `json:"status"`
	Remark          string    `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (h *webvpnSiteHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" && method == http.MethodPost {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "应用名称不能为空",
			})
			return false
		}

		if h.DomainId <= 0 && method == http.MethodPost {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "必须选择所属的基础域",
			})
			return false
		}

		h.TargetURL = strings.TrimSpace(h.TargetURL)
		if h.TargetURL == "" && method == http.MethodPost {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "目标地址不能为空",
			})
			return false
		}
		if h.TargetURL != "" && (!strings.HasPrefix(h.TargetURL, "http://") && !strings.HasPrefix(h.TargetURL, "https://")) {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "目标地址必须是以 http:// 或 https:// 开头的合法 URL",
			})
			return false
		}

		// 标准 WebVPN 子域名前缀推导: s-<dashed-host>-<port> (http为 c-)
		if h.TargetURL != "" {
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
		}

		// 清洗关联域名列表 (Hosts)
		if h.Hosts != "" {
			var cleaned []string
			for _, line := range strings.Split(h.Hosts, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.Contains(line, "://") {
					if u, err := url.Parse(line); err == nil && u.Host != "" {
						// Keep scheme and host:port, but strip path
						line = u.Scheme + "://" + u.Host
					}
				} else {
					if idx := strings.Index(line, "/"); idx != -1 {
						line = line[:idx]
					}
					// (不再强行剥离端口，保留用户可能输入的非标准端口)
				}
				if line != "" {
					cleaned = append(cleaned, line)
				}
			}
			h.Hosts = strings.Join(cleaned, "\n")
		}

		if h.AllowedGroupIds == "" && method == http.MethodPost {
			h.AllowedGroupIds = "[]"
		}
		if h.Replace == "" && method == http.MethodPost {
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
	if domainIdStr := strings.TrimSpace(c.Query("domain_id")); domainIdStr != "" {
		if sid, err := strconv.ParseInt(domainIdStr, 10, 64); err == nil && sid > 0 {
			session.Where("domain_id = ?", sid)
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

	// 预加载 WebvpnDomain 映射
	var domains []models.WebvpnDomain
	_ = models.GetEngine().Find(&domains)
	domainMap := make(map[int64]models.WebvpnDomain)
	for _, s := range domains {
		domainMap[s.Id] = s
	}

	type WebvpnSiteItemVO struct {
		models.WebvpnSite
		DomainName     string `json:"domain_name"`
		DomainHostname string `json:"domain_hostname"`
		FullAccessURL  string `json:"full_access_url"`
	}

	resList := make([]WebvpnSiteItemVO, len(list))
	for i, item := range list {
		vo := WebvpnSiteItemVO{WebvpnSite: item}

		// 优先从 WebvpnDomain 关联
		if dom, ok := domainMap[item.DomainId]; ok {
			vo.DomainName = dom.Name
			vo.DomainHostname = dom.Hostname

			rootDomain := strings.TrimPrefix(dom.Hostname, "*.")
			scheme := "http://"
			if dom.TLS || dom.H2 {
				scheme = "https://"
			}
			portSuffix := ""
			if dom.Port != "80" && dom.Port != "443" && dom.Port != "" {
				portSuffix = ":" + dom.Port
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
