package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
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

	err := session.Desc("id").Find(&list)
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
// POST /api/auth-method/test
// Body: { "type": "radius", "config_json": "{...}" }
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
		cfg, err := service.ParseCasConfig(req.ConfigJSON)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("CAS 配置解析失败: %v", err)})
			return
		}
		if err := service.CasTestConnectivity(cfg); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": fmt.Sprintf("CAS 服务连通性测试失败: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": fmt.Sprintf("CAS 服务 [%s] 连通正常 (登录端点响应正常)", cfg.BaseURL),
		})
		return

	case "radius":
		// 使用真实的 RADIUS Access-Request 探测，而非 UDP 端口拨号
		cfg, err := service.ParseRadiusConfig(req.ConfigJSON)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
			return
		}
		if cfg.Host == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "RADIUS 服务器 Host 不能为空"})
			return
		}
		if cfg.Secret == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "RADIUS 共享密钥 (secret) 不能为空"})
			return
		}
		if err := service.RadiusTestConnectivity(cfg); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": fmt.Sprintf("RADIUS 连通性测试失败: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": fmt.Sprintf("RADIUS 服务 [%s:%d] 连通正常（服务器已正常响应 Access-Request）", cfg.Host, cfg.Port),
		})
		return

	default:
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "不支持的认证方式类型"})
	}
}

// AuthenticateByMethodHandler 使用指定认证方式对用户名/密码进行真实认证
// POST /api/auth-method/authenticate
// 支持两种调用方式：
//  1. 传 method_id: { "method_id": 1, "username": "xxx", "password": "xxx" }
//  2. 直接传配置: { "type": "radius", "config_json": "{...}", "username": "xxx", "password": "xxx" }
func AuthenticateByMethodHandler(c *gin.Context) {
	var req struct {
		MethodID   int64  `json:"method_id"`
		Type       string `json:"type"`
		ConfigJSON string `json:"config_json"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Ticket     string `json:"ticket"`
		ServiceURL string `json:"service_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "请求参数错误"})
		return
	}

	authType := strings.ToLower(strings.TrimSpace(req.Type))
	configJSON := req.ConfigJSON

	// 如果传了 method_id，则从数据库加载配置
	if req.MethodID > 0 {
		var m models.AuthMethod
		has, err := models.GetEngine().ID(req.MethodID).Get(&m)
		if err != nil || !has {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "认证方式不存在"})
			return
		}
		if !m.Enabled {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "该认证方式已被禁用"})
			return
		}
		authType = m.Type
		configJSON = m.ConfigJSON
	}

	switch authType {
	case "local":
		if req.Username == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "用户名不能为空"})
			return
		}
		result, err := service.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "本地认证成功",
			"data":    result,
		})

	case "radius":
		if req.Username == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": "用户名不能为空"})
			return
		}
		cfg, err := service.ParseRadiusConfig(configJSON)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("RADIUS 配置解析失败: %v", err)})
			return
		}
		result, err := service.RadiusAuthenticate(cfg, req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("RADIUS 认证请求失败: %v", err)})
			return
		}
		if !result.Success {
			msg := "RADIUS 认证被拒绝"
			if result.ReplyMessage != "" {
				msg = fmt.Sprintf("RADIUS 认证被拒绝: %s", result.ReplyMessage)
			}
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": msg, "radius_code": result.Code})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":          0,
			"message":       "RADIUS 认证成功",
			"radius_code":   result.Code,
			"reply_message": result.ReplyMessage,
		})

	case "cas":
		cfg, err := service.ParseCasConfig(configJSON)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("CAS 配置解析失败: %v", err)})
			return
		}
		if req.Ticket != "" {
			res, err := service.ValidateCasTicket(cfg, req.Ticket, req.ServiceURL)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("CAS Ticket 校验请求失败: %v", err)})
				return
			}
			if !res.Success {
				c.JSON(http.StatusOK, gin.H{
					"code":       1,
					"message":    fmt.Sprintf("CAS 认证失败: %s", res.ErrorMsg),
					"error_code": res.ErrorCode,
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"code":       0,
				"message":    "CAS Ticket 校验成功",
				"username":   res.Username,
				"attributes": res.Attributes,
			})
			return
		}

		loginURL := service.BuildCasLoginURL(cfg, req.ServiceURL)
		c.JSON(http.StatusOK, gin.H{
			"code":      0,
			"message":   "CAS 客户端就绪",
			"login_url": loginURL,
		})

	default:
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": fmt.Sprintf("认证方式 [%s] 暂不支持", authType)})
	}
}

func (h *authMethodHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		service.SyncHTTPToCluster()
	}
}
