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
	rest.Register(&models.SniProxy{}, &sniHandler{}, rest.RouteTypeALL, nil, "sni-proxy")
}

type sniHandler struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	SNI         string    `json:"sni"`
	ExtraSNI    string    `json:"extra_sni"` // JSON array of extra SNI patterns
	Port        string    `json:"port"`
	Rules       string    `json:"rules"`
	BackendType string    `json:"backend_type"`
	TunnelType  string    `json:"tunnel_type"`
	TunnelId    string    `json:"tunnel_id"`
	TunnelToken string    `json:"tunnel_token"`
	DNSResolver string    `json:"dns_resolver"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *sniHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		portStr := strings.TrimSpace(s.Port)
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

		if strings.TrimSpace(s.SNI) == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "sniRequired",
				"message":   "主 SNI 不能为空",
			})
			return false
		}

		// Check duplicate DNS resolvers
		if s.DNSResolver != "" {
			var resolvers []string
			if err := json.Unmarshal([]byte(s.DNSResolver), &resolvers); err == nil {
				seen := make(map[string]bool)
				for _, r := range resolvers {
					rTrim := strings.TrimSpace(r)
					if rTrim != "" {
						if seen[rTrim] {
							g.AbortWithStatusJSON(http.StatusOK, gin.H{
								"code":      1,
								"error_key": "targetDuplicate",
								"message":   fmt.Sprintf("上游列表中存在重复的目标服务器地址 [%s]", rTrim),
							})
							return false
						}
						seen[rTrim] = true
					}
				}
			}
		}

		if s.Name == "" {
			sniLabel := s.SNI
			if sniLabel == "" {
				sniLabel = "multi"
			}
			s.Name = "SNI-" + sniLabel
		}
	}
	return true
}

func (s *sniHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	if g.Request.Method == http.MethodPost || g.Request.Method == http.MethodPut || g.Request.Method == http.MethodPatch || g.Request.Method == http.MethodDelete {
		service.SyncSNIToCluster()
	}
}

func (s *sniHandler) List(c *gin.Context) {
	var list []models.SniProxy
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := c.Query("name"); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}

	if sni := c.Query("sni"); sni != "" {
		session.Where("sni LIKE ? OR extra_sni LIKE ?", "%"+sni+"%", "%"+sni+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询 SNI 代理列表失败: " + err.Error()})
		return
	}

	resList := make([]sniHandler, 0, len(list))
	for _, item := range list {
		resList = append(resList, sniHandler{
			Id:          item.Id,
			Name:        item.Name,
			SNI:         item.SNI,
			ExtraSNI:    item.ExtraSNI,
			Port:        item.Port,
			Rules:       item.Rules,
			BackendType: item.BackendType,
			TunnelType:  item.TunnelType,
			TunnelId:    item.TunnelId,
			TunnelToken: item.TunnelToken,
			DNSResolver: item.DNSResolver,
			Remark:      item.Remark,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}
