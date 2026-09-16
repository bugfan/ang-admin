package api

import (
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
	rest.Register(&models.Auth{}, &authHandler{}, rest.RouteTypeALL, nil, "auth")
}

type authHandler struct {
	Id            int64     `json:"id"`
	Name          string    `json:"name"`
	AuthMethodIds string    `json:"auth_method_ids"`
	PortalUrl     string    `json:"portal_url"`
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (h *authHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameRequired",
				"message":   "认证配置名称不能为空",
			})
			return false
		}

		// Check name unique
		var exist models.Auth
		has, err := x.Where("name = ? AND id != ?", h.Name, h.Id).Get(&exist)
		if err == nil && has {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameDuplicate",
				"message":   fmt.Sprintf("认证配置名称 [%s] 已存在", h.Name),
			})
			return false
		}
	} else if method == http.MethodDelete {
		// 这里可以加一个校验：如果该 auth 正在被 HTTP/TCP 等规则使用，则不能删除
		// 暂时先不做协议关联的校验，先实现基础 CRUD
	}
	return true
}

func (h *authHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		service.SyncHTTPToCluster()
	}
}
