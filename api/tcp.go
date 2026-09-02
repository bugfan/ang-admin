package api

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	rest.Register(&models.TcpProxy{}, &tcpHandler{}, rest.RouteTypeALL, nil, "tcp-proxy")
}

type tcpHandler struct {
	Id              int64     `json:"id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Port            string    `json:"port"`
	Rules           string    `json:"rules"`
	BackendType     string    `json:"backend_type"`
	TunnelType      string    `json:"tunnel_type"`
	TunnelId        string    `json:"tunnel_id"`
	TunnelToken     string    `json:"tunnel_token"`
	UpstreamMethod  string    `json:"upstream_method"`
	UpstreamServers string    `json:"upstream_servers"`
	Remark          string    `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (t *tcpHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		portStr := strings.TrimSpace(t.Port)
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

		// Check for duplicate port in tcp_proxy
		targetId := t.Id
		if targetId == 0 {
			if idParam := g.Param("id"); idParam != "" {
				targetId, _ = strconv.ParseInt(idParam, 10, 64)
			}
		}

		session := models.GetEngine().Table("tcp_proxy").Where("port = ?", portStr)
		if targetId > 0 {
			session = session.Where("id != ?", targetId)
		}
		count, err := session.Count()
		if err == nil && count > 0 {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "portExists",
				"message":   "该监听端口已被其他 TCP 代理占用",
			})
			return false
		}

		// Check upstream servers duplicate targets
		if t.UpstreamServers != "" {
			var servers []struct {
				Target string `json:"target"`
				Weight int    `json:"weight"`
			}
			if err := json.Unmarshal([]byte(t.UpstreamServers), &servers); err == nil {
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

		if t.Name == "" {
			t.Name = "TCP-" + portStr
		}
	}
	return true
}

func (t *tcpHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	if g.Request.Method == http.MethodPost || g.Request.Method == http.MethodPut || g.Request.Method == http.MethodPatch || g.Request.Method == http.MethodDelete {
		service.SyncTCPToCluster()
	}
}

func (t *tcpHandler) List(c *gin.Context) {
	var list []models.TcpProxy
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := c.Query("name"); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if port := c.Query("port"); port != "" {
		session.Where("port LIKE ?", "%"+port+"%")
	}


	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询 TCP 代理列表失败: " + err.Error()})
		return
	}

	resList := make([]tcpHandler, 0, len(list))
	for _, item := range list {
		resList = append(resList, tcpHandler{
			Id:              item.Id,
			Name:            item.Name,
			Address:         item.Address,
			Port:            item.Port,
			Rules:           item.Rules,
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
