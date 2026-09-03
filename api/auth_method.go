package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.AuthMethod{}, &authMethodHandler{}, rest.RouteTypeALL, nil, "auth-method")
}

type authMethodHandler struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Enabled    bool      `json:"enabled"`
	Priority   int       `json:"priority"`
	ConfigJSON string    `json:"config_json"`
	Remark     string    `json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (h *authMethodHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameRequired",
				"message":   "认证方式名称不能为空",
			})
			return false
		}

		h.Type = strings.ToLower(strings.TrimSpace(h.Type))
		if h.Type != "local" && h.Type != "cas" && h.Type != "radius" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "typeInvalid",
				"message":   "不支持的认证方式类型，目前仅支持 local, cas, radius",
			})
			return false
		}

		// Check name unique
		var exist models.AuthMethod
		has, err := x.Where("name = ? AND id != ?", h.Name, h.Id).Get(&exist)
		if err == nil && has {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameDuplicate",
				"message":   fmt.Sprintf("认证方式名称 [%s] 已存在", h.Name),
			})
			return false
		}

		// Validate config_json format
		if strings.TrimSpace(h.ConfigJSON) != "" {
			var js map[string]interface{}
			if err := json.Unmarshal([]byte(h.ConfigJSON), &js); err != nil {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "configJsonInvalid",
					"message":   "认证方式配置参数必须是合法的 JSON 格式",
				})
				return false
			}
		} else {
			h.ConfigJSON = "{}"
		}

	}
	return true
}

func (h *authMethodHandler) List(c *gin.Context) {
	var list []models.AuthMethod
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if t := strings.TrimSpace(c.Query("type")); t != "" {
		session.Where("type = ?", t)
	}

	err := session.Asc("priority").Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    list,
	})
}

// TestAuthMethodHandler 测试认证方式连通性
func TestAuthMethodHandler(c *gin.Context) {
	var req struct {
		Type       string `json:"type"`
		ConfigJSON string `json:"config_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": "请求参数错误",
		})
		return
	}

	sourceType := strings.ToLower(strings.TrimSpace(req.Type))
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigJSON), &config); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": "配置 JSON 格式不合法",
		})
		return
	}

	switch sourceType {
	case "local":
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "本地数据库认证方式配置正常",
		})
		return

	case "cas":
		baseURL, _ := config["base_url"].(string)
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "CAS Base URL 不能为空"})
			return
		}
		skipTls, _ := config["skip_tls_verify"].(bool)
		client := &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTls},
			},
		}
		resp, err := client.Get(baseURL)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": fmt.Sprintf("CAS 服务连接失败: %v", err),
			})
			return
		}
		defer resp.Body.Close()
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": fmt.Sprintf("CAS 服务连接成功 (HTTP 状态码: %d)", resp.StatusCode),
		})
		return

	case "radius":
		host, _ := config["host"].(string)
		host = strings.TrimSpace(host)
		if host == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "RADIUS 服务器 Host 不能为空"})
			return
		}
		portVal, ok := config["port"]
		port := 1812
		if ok {
			if pFloat, isFloat := portVal.(float64); isFloat {
				port = int(pFloat)
			}
		}
		addr := fmt.Sprintf("%s:%d", host, port)
		conn, err := net.DialTimeout("udp", addr, 3*time.Second)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": fmt.Sprintf("RADIUS 端口拨号失败: %v", err),
			})
			return
		}
		defer conn.Close()
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": fmt.Sprintf("RADIUS 服务 UDP 地址 [%s] 网络连通正常", addr),
		})
		return

	default:
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "不支持的认证方式类型"})
	}
}
