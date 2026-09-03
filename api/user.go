package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	rest.Register(&models.User{}, &userHandler{}, rest.RouteTypeALL, []string{"Password"}, "user")
}

type userHandler struct {
	Id         int64     `json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"password,omitempty"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Mobile     string    `json:"mobile"`
	SourceType string    `json:"source_type"`
	SourceId   int64     `json:"source_id"`
	GroupIds   string    `json:"group_ids"`
	Status     int       `json:"status"`
	ExpireAt   string    `json:"expire_at"`
	Remark     string    `json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (h *userHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost {
		h.Username = strings.TrimSpace(h.Username)
		if h.Username == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "usernameRequired",
				"message":   "用户名不能为空",
			})
			return false
		}

		if h.SourceType == "" {
			h.SourceType = "local"
		}

		// Check username unique
		var exist models.User
		has, err := x.Where("username = ?", h.Username).Get(&exist)
		if err == nil && has {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "usernameDuplicate",
				"message":   fmt.Sprintf("用户名 [%s] 已存在", h.Username),
			})
			return false
		}

		// For local user, password is required
		if h.SourceType == "local" {
			rawPwd := strings.TrimSpace(h.Password)
			if len(rawPwd) < 6 {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "passwordTooShort",
					"message":   "本地用户密码长度不能少于 6 位",
				})
				return false
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(rawPwd), bcrypt.DefaultCost)
			if err != nil {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "passwordHashError",
					"message":   "密码哈希加密失败",
				})
				return false
			}
			h.Password = string(hash)
		} else {
			h.Password = ""
		}

		// Ensure default group if group_ids is empty
		if strings.TrimSpace(h.GroupIds) == "" || h.GroupIds == "[]" {
			var defGroup models.UserGroup
			if hasDef, _ := x.Where("is_default = 1").Get(&defGroup); hasDef {
				h.GroupIds = fmt.Sprintf("[%d]", defGroup.Id)
			} else {
				h.GroupIds = "[]"
			}
		}

	} else if method == http.MethodPut || method == http.MethodPatch {
		h.Username = strings.TrimSpace(h.Username)
		if h.Username == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "usernameRequired",
				"message":   "用户名不能为空",
			})
			return false
		}

		// Check username unique
		var exist models.User
		has, err := x.Where("username = ? AND id != ?", h.Username, h.Id).Get(&exist)
		if err == nil && has {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":      1,
				"error_key": "usernameDuplicate",
				"message":   fmt.Sprintf("用户名 [%s] 已存在", h.Username),
			})
			return false
		}

		// If password is provided and non-empty, hash it
		rawPwd := strings.TrimSpace(h.Password)
		if rawPwd != "" {
			if len(rawPwd) < 6 {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "passwordTooShort",
					"message":   "密码长度不能少于 6 位",
				})
				return false
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(rawPwd), bcrypt.DefaultCost)
			if err != nil {
				g.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code":      1,
					"error_key": "passwordHashError",
					"message":   "密码哈希加密失败",
				})
				return false
			}
			h.Password = string(hash)
		} else {
			// Keep existing password
			var orig models.User
			if hasOrig, _ := x.ID(h.Id).Get(&orig); hasOrig {
				h.Password = orig.Password
			}
		}
	}
	return true
}

func (h *userHandler) List(c *gin.Context) {
	var list []models.User
	session := models.GetEngine().NewSession()
	defer session.Close()

	if query := strings.TrimSpace(c.Query("query")); query != "" {
		session.Where("username LIKE ? OR full_name LIKE ? OR mobile LIKE ? OR email LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	if st := strings.TrimSpace(c.Query("source_type")); st != "" {
		session.Where("source_type = ?", st)
	}

	if statusStr := strings.TrimSpace(c.Query("status")); statusStr != "" {
		if statusVal, err := strconv.Atoi(statusStr); err == nil {
			session.Where("status = ?", statusVal)
		}
	}

	if gidStr := strings.TrimSpace(c.Query("group_id")); gidStr != "" {
		session.Where("group_ids LIKE ?", "%"+gidStr+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	// Sanitize passwords
	for i := range list {
		list[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    list,
	})
}
