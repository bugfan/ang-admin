package api

import (
	"fmt"
	"log"
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
	rest.Register(&models.UserGroup{}, &userGroupHandler{}, rest.RouteTypeALL, nil, "user-group")
}

type userGroupHandler struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *userGroupHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameRequired",
				"message":   "用户组名称不能为空",
			})
			return false
		}

		// Check name unique
		var exist models.UserGroup
		has, err := x.Where("name = ? AND id != ?", h.Name, h.Id).Get(&exist)
		if err == nil && has {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "nameDuplicate",
				"message":   fmt.Sprintf("用户组名称 [%s] 已存在", h.Name),
			})
			return false
		}
	} else if method == http.MethodDelete {
		idStr := g.Param("id")
		idNum, _ := strconv.ParseInt(idStr, 10, 64)
		if idNum > 0 {
			var exist models.UserGroup
			has, _ := x.ID(idNum).Get(&exist)
			if has && exist.IsDefault {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "cannotDeleteDefaultGroup",
					"message":   "系统默认用户组不能删除",
				})
				return false
			}

			// Check if any users belong to this group
			matchPattern := fmt.Sprintf("%%%d%%", idNum)
			userCount, _ := x.Where("group_ids LIKE ?", matchPattern).Count(new(models.User))
			if userCount > 0 {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "groupHasUsers",
					"message":   fmt.Sprintf("该用户组下仍有 %d 名用户，请先调整用户归属后再删除", userCount),
				})
				return false
			}
		}
	}
	return true
}

func (h *userGroupHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		log.Printf("[UserGroup Change Event] Method: %s, ID: %d, Name: %s\n", method, h.Id, h.Name)
		service.SyncGroupToCluster()
	}
}

func (h *userGroupHandler) List(c *gin.Context) {
	var list []models.UserGroup
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	type GroupItemWithCount struct {
		models.UserGroup
		UserCount int64 `json:"user_count"`
	}

	resList := make([]GroupItemWithCount, len(list))
	for i, g := range list {
		matchPattern := fmt.Sprintf("%%%d%%", g.Id)
		uCount, _ := models.GetEngine().Where("group_ids LIKE ?", matchPattern).Count(new(models.User))
		resList[i] = GroupItemWithCount{
			UserGroup: g,
			UserCount: uCount,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resList,
	})
}
