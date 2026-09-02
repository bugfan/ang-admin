package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/glob"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.DnsProxy{}, &dnsHandler{}, rest.RouteTypeALL, nil, "dns-proxy")
}

type dnsHandler struct {
	Id              int64  `json:"id"`
	Name            string `json:"name"`
	Address         string `json:"address"`
	Port            string `json:"port"`
	Rules           string `json:"rules"`
	HostsText       string `json:"hosts_text"`
	HostsJSON       string `json:"hosts_json"`
	BackendType     string `json:"backend_type"`
	TunnelType      string `json:"tunnel_type"`
	TunnelId        string `json:"tunnel_id"`
	TunnelToken     string `json:"tunnel_token"`
	UpstreamMethod  string    `json:"upstream_method"`
	UpstreamServers string    `json:"upstream_servers"`
	Remark          string    `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (d *dnsHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		portStr := strings.TrimSpace(d.Port)
		if portStr == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "portRequired",
				"message":   "监听端口不能为空",
			})
			return false
		}
		portNum, err := strconv.Atoi(portStr)
		if err != nil || portNum < 1 || portNum > 65535 {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "portInvalid",
				"message":   "监听端口必须为有效数字(1-65535)",
			})
			return false
		}

		// Currently DNS mode only supports QUIC tunnels
		if strings.TrimSpace(d.TunnelId) != "" {
			tunnelType := strings.ToLower(strings.TrimSpace(d.TunnelType))
			if tunnelType != "" && !strings.Contains(tunnelType, "quic") {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "quicTunnelOnly",
					"message":   "当前 DNS 模块仅支持关联 QUIC 类型的隧道",
				})
				return false
			}
		}

		// Validate and parse HostsText
		hostsText := strings.TrimSpace(d.HostsText)
		hostsA := make(map[string]string)
		hostsAAAA := make(map[string]string)

		if hostsText != "" {
			lines := strings.Split(hostsText, "\n")
			for lineNum, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
					continue
				}

				parts := strings.Fields(line)
				if len(parts) < 2 {
					g.AbortWithStatusJSON(http.StatusOK, gin.H{
						"code":    1,
						"message": fmt.Sprintf("第 %d 行 Hosts 格式错误: 每行需包含一个 IP 和至少一个域名 (如: 1.2.3.4 baidu.com *.baidu.com)", lineNum+1),
					})
					return false
				}

				ipStr := parts[0]
				parsedIP := net.ParseIP(ipStr)
				if parsedIP == nil {
					g.AbortWithStatusJSON(http.StatusOK, gin.H{
						"code":    1,
						"message": fmt.Sprintf("第 %d 行 IP 地址 [%s] 无效", lineNum+1, ipStr),
					})
					return false
				}

				isIPv4 := parsedIP.To4() != nil

				for _, domainPattern := range parts[1:] {
					domainPattern = strings.TrimSpace(domainPattern)
					if domainPattern == "" {
						continue
					}

					// Use gobwas/glob to compile & validate pattern
					_, globErr := glob.Compile(domainPattern)
					if globErr != nil {
						g.AbortWithStatusJSON(http.StatusOK, gin.H{
							"code":    1,
							"message": fmt.Sprintf("第 %d 行域名通配符 [%s] 校验错误: %v", lineNum+1, domainPattern, globErr),
						})
						return false
					}

					if isIPv4 {
						hostsA[domainPattern] = ipStr
					} else {
						hostsAAAA[domainPattern] = ipStr
					}
				}
			}
		}

		type structuredHosts struct {
			A    map[string]string `json:"A,omitempty"`
			AAAA map[string]string `json:"AAAA,omitempty"`
		}
		sh := structuredHosts{
			A:    hostsA,
			AAAA: hostsAAAA,
		}
		b, _ := json.Marshal(sh)
		d.HostsJSON = string(b)

		// Check upstream servers duplicate targets
		if d.UpstreamServers != "" {
			var servers []struct {
				Target string `json:"target"`
				Weight int    `json:"weight"`
			}
			if err := json.Unmarshal([]byte(d.UpstreamServers), &servers); err == nil {
				seen := make(map[string]bool)
				for _, s := range servers {
					tgt := strings.TrimSpace(s.Target)
					if tgt != "" {
						if seen[tgt] {
							g.AbortWithStatusJSON(http.StatusOK, gin.H{
								"code":      1,
								"error_key": "targetDuplicate",
								"message":   fmt.Sprintf("上游列表中存在重复的目标服务器地址 [%s]", tgt),
							})
							return false
						}
						seen[tgt] = true
					}
				}
			}
		}
	}
	return true
}

func (d *dnsHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	if g.Request.Method == http.MethodPost || g.Request.Method == http.MethodPut || g.Request.Method == http.MethodPatch || g.Request.Method == http.MethodDelete {
		service.SyncDNSToCluster()
	}
}

func (d *dnsHandler) List(c *gin.Context) {
	var list []models.DnsProxy
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := c.Query("name"); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if port := c.Query("port"); port != "" {
		session.Where("port LIKE ?", "%"+port+"%")
	}
	if address := c.Query("address"); address != "" {
		session.Where("address LIKE ?", "%"+address+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询 DNS 代理列表失败: " + err.Error()})
		return
	}

	resList := make([]dnsHandler, 0, len(list))
	for _, item := range list {
		resList = append(resList, dnsHandler{
			Id:              item.Id,
			Name:            item.Name,
			Address:         item.Address,
			Port:            item.Port,
			Rules:           item.Rules,
			HostsText:       item.HostsText,
			HostsJSON:       item.HostsJSON,
			BackendType:     item.BackendType,
			TunnelType:      item.TunnelType,
			TunnelId:        item.TunnelId,
			TunnelToken:     item.TunnelToken,
			UpstreamMethod:  item.UpstreamMethod,
			UpstreamServers: item.UpstreamServers,
			Remark:          item.Remark,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}
