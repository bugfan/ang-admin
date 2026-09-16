package api

import (
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.AuthSetting{}, &authSettingHandler{}, rest.RouteTypeALL, nil, "auth-setting")
}

type authSettingHandler struct {
	Id          int64     `json:"id"`
	TokenName   string    `json:"token_name"`
	TokenExpire int       `json:"token_expire"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *authSettingHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	return true

}

func (h *authSettingHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		service.SyncHTTPToCluster()
	}

}
