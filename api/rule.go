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
	rest.Register(&models.Rule{}, &ruleHandler{}, rest.RouteTypeALL, nil, "rule")
}

type ruleHandler struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Items     string    `json:"items"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *ruleHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if r.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameRequired",
				"message":   "规则名称不能为空",
			})
			return false
		}
	}
	return true
}

func (r *ruleHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		log.Printf("[Rule Change Event] Method: %s, ID: %d, Name: %s\n", method, r.Id, r.Name)
		service.SyncRuleToCluster()
	}
}

func (r *ruleHandler) List(c *gin.Context) {
	var list []models.Rule
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := c.Query("name"); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询规则列表失败: " + err.Error()})
		return
	}

	resList := make([]ruleHandler, 0, len(list))
	for _, item := range list {
		resList = append(resList, ruleHandler{
			Id:        item.Id,
			Name:      item.Name,
			Items:     item.Items,
			Remark:    item.Remark,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, resList)
}
